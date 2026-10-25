<script lang="ts">
	import { afterNavigate, goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import { page } from "$app/state";
	import Error from "@/lib/Error.svelte";
	import Icon from "@/lib/Icon.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import tooltip from "@/lib/actions/tooltip";
	import FaceMenu from "@/lib/nav/FaceMenu.svelte";
	import FollowingMenu from "@/lib/nav/FollowingMenu.svelte";
	import TagMenu from "@/lib/tag/TagMenu.svelte";
	import { baseURL, req } from "@/lib/util/api";
	import { isTouch, userHasPermission } from "@/lib/util/helpers";
	import { store } from "@/store.svelte";
	import {
		UserPermission,
		type ServerFeatures,
		type Follow,
		type PrivateUser,
		type Tag,
		type UserSettings,
	} from "@/types";
	import { onMount } from "svelte";
	import { SvelteURLSearchParams } from "svelte/reactivity";
	import HandbookModal from "@/lib/handbook/HandbookModal.svelte";
	import CommandPalette from "@/lib/nav/CommandPalette.svelte";
	import { t } from "@/lib/i18n";

	interface Props {
		children?: import("svelte").Snippet;
	}

	let { children }: Props = $props();

	let navEl: HTMLElement | undefined = $state();
	let mainSearchEl: HTMLInputElement | undefined = $state();
	let searchTimeout: number;
	let subMenuShown = $state(false);
	let followingMenuShown = $state(false);
	let tagMenuShown = $state(false);
	let commandPaletteOpen = $state(false);
	let isMac = $state(false);
	let scroll = window.scrollY;

	function handleProfileClick() {
		if (!localStorage.getItem("token")) {
			goto(resolve("/login"));
		} else {
			closeAllSubMenus("sub");
			subMenuShown = !subMenuShown;
		}
	}

	/**
	 * Handle KeyDown event on the search bar.
	 */
	function searchBarKeyDown(ev: KeyboardEvent) {
		if (
			ev.key === "ContextMenu" ||
			ev.key === "Home" ||
			ev.key === "End" ||
			ev.key === "PageDown" ||
			ev.key === "PageUp" ||
			ev.key === "NumLock" ||
			ev.key === "Escape" ||
			ev.key === "Tab" ||
			ev.key === "CapsLock" ||
			ev.key === "OS" ||
			ev.key === "ArrowLeft" ||
			ev.key === "ArrowRight" ||
			ev.key === "ArrowUp" ||
			ev.key === "ArrowDown" ||
			ev.key === "Control" ||
			ev.key === "Alt" ||
			ev.key === "AltGraph" ||
			ev.key === "Shift" ||
			ev.key === "Meta"
		) {
			return;
		}

		clearTimeout(searchTimeout);

		if (ev.key === "Enter") {
			handleSearch(ev);
		} else {
			searchTimeout = window.setTimeout(
				() => {
					handleSearch(ev);
				},
				isTouch() ? 800 : 400,
			);
		}
	}

	/**
	 * Processes the query in search bar after a new query is input and
	 * navigates us to the new search page.
	 */
	function handleSearch(ev: KeyboardEvent) {
		const target = ev.target as HTMLInputElement;
		if (!target) {
			console.error("handleSearch: Event target doesn't exist!");
			return;
		}
		if (!target.value.trim()) {
			console.debug("handleSearch: Search box has no valid value.");
			return;
		}
		// Only trim the start of the query incase the debounced search
		// fires before the user finishes typing the full query (which
		// would remove any last space if the query is multi-word).
		let query = target.value.trimStart();
		if (ev.key === "Enter") {
			// If enter is the key that triggered the search, we will
			// also trim the end of the query, since we can be sure
			// the user was done typing.
			query = query.trimEnd();
		}
		const currentSearchType = page.url.searchParams.get("type");
		const searchParams = new SvelteURLSearchParams({
			query: encodeURIComponent(query),
			preferMyList: "true",
		});
		if (page.route?.id === "/(app)/search" && currentSearchType) {
			// If we are already on the search page, we can attempt
			// to keep any existing type filter on the next query.
			searchParams.set("type", currentSearchType);
		}
		// Enable autofocus before running `goto` because on chromium
		// the .focus() call won't work, even after a timeout.
		// Using autofocus seems to work. Disables after goto runs.
		// https://github.com/sbondCo/Watcharr/issues/169
		target.autofocus = true;
		goto(resolve(`/search?${searchParams.toString()}`)).then(() => {
			// Use mainSearchEl if nav not split, otherwise use ev target.
			if (!document.body.classList.contains("split-nav") && mainSearchEl) {
				mainSearchEl.focus();
				mainSearchEl.autofocus = false;
			} else {
				target?.focus();
			}
			target.autofocus = false;
		});
	}

	async function getInitialData() {
		if (!localStorage.getItem("token")) {
			console.warn("getInitialData: No token found, redirecting to login!");
			goto(resolve("/login?again=1"));
			return;
		}
		const [u, s, f, fo, ts] = await Promise.all([
			req.get<PrivateUser>("/user"),
			req.get<UserSettings>("/user/settings").catch((e) => {
				console.warn("Failed to load user settings:", e);
				return undefined;
			}),
			req.get<ServerFeatures>("/features").catch((e) => {
				console.warn("Failed to load features:", e);
				return undefined;
			}),
			req.get<Follow[]>("/follow").catch((e) => {
				console.warn("Failed to load follows:", e);
				return [];
			}),
			req.get<Tag[]>("/tag").catch((e) => {
				console.warn("Failed to load tags:", e);
				return [];
			}),
		]);
		if (u) {
			store.userInfo = u;
		}
		if (s) {
			store.userSettings = s;
		}
		if (f) {
			store.serverFeatures = f;
		}
		if (fo) {
			store.follows = fo;
		}
		if (ts) {
			store.tags = ts;
		}
	}

	function closeAllSubMenus(except?: string) {
		if (except !== "sub") subMenuShown = false;
		if (except !== "following") followingMenuShown = false;
		if (except !== "tag") tagMenuShown = false;
	}

	/**
	 * Adds or removed `split-nav` tag to body depending
	 * on how big the main search bar is.
	 */
	function decideOnNavSplit() {
		if (window.innerWidth <= 305) {
			document.body.classList.add("split-nav");
			return;
		}
		const bigInput = navEl?.querySelector("input:not(.small)");
		if (bigInput) {
			const b = bigInput.getBoundingClientRect();
			console.debug("decideOnNavSplit: bigInput width:", b.width);
			if (b.width <= 45) {
				document.body.classList.add("split-nav");
				console.debug("decideOnNavSplit: Splitting nav.");
			} else {
				document.body.classList.remove("split-nav");
				console.debug("decideOnNavSplit: Unsplitting nav.");
			}
		} else {
			console.warn("decideOnNavSplit: bigInput not found!", bigInput);
		}
	}

	function docOnScroll() {
		const currentScrollY = window.scrollY;
		document.body.classList.add("nav-shown");
		if (currentScrollY > 80 && scroll < currentScrollY) {
			closeAllSubMenus();
		}
		scroll = Math.max(0, currentScrollY);
	}

	function focusSearch() {
		try {
			if (!mainSearchEl) {
				console.warn("focusSearch: mainSearchEl not defined!");
				return;
			}
			if (document.activeElement === mainSearchEl) {
				console.debug("focusSearch: mainSearchEl is already focused.");
				return;
			}
			mainSearchEl.focus();
		} catch (err) {
			console.error("focusSearch: Failed!", err);
		}
	}

	function handleGlobalKeybind(ev: KeyboardEvent) {
		const isCmdOrCtrl = ev.ctrlKey || ev.metaKey;
		const key = ev.key.toLowerCase();
		if (isCmdOrCtrl && (key === "k" || key === "p")) {
			ev.preventDefault();
			commandPaletteOpen = !commandPaletteOpen;
		}
	}

	afterNavigate(() => {
		decideOnNavSplit();
		closeAllSubMenus();
	});

	onMount(() => {
		isMac =
			typeof navigator !== "undefined" &&
			/Mac|iPod|iPhone|iPad/.test(navigator.platform || navigator.userAgent);

		if (typeof navigator !== "undefined" && "serviceWorker" in navigator) {
			navigator.serviceWorker
				.register("/sw.js")
				.then((reg) => {
					console.debug("Service Worker registered successfully:", reg.scope);
				})
				.catch((err) => {
					console.debug("Service Worker registration failed:", err);
				});
		}

		if (navEl) {
			decideOnNavSplit();
			window.addEventListener("resize", decideOnNavSplit);
			window.document.addEventListener("scroll", docOnScroll);
			window.document.addEventListener("keydown", handleGlobalKeybind);

			return () => {
				window.removeEventListener("resize", decideOnNavSplit);
				window.document.removeEventListener("scroll", docOnScroll);
				window.document.removeEventListener("keydown", handleGlobalKeybind);
			};
		} else {
			console.error(
				"navEl doesn't exist, failed to initialize up/down listener",
			);
		}
	});
</script>

<nav bind:this={navEl} class="top-nav">
	<div class="wrapper">
		<div class="left-side">
			<a href={resolve("/")} class="logo">
				<span class="large">Watcharr</span>
				<span class="small">W</span>
			</a>
			<div class="main-nav-links">
				<a
					href={resolve("/")}
					class="nav-link {page.url?.pathname === '/' ? 'active' : ''}"
				>
					<Icon i="film" wh={16} />
					<span>{t("nav.library")}</span>
				</a>
				<a
					href={resolve("/discover")}
					class="nav-link {page.url?.pathname.startsWith('/discover')
						? 'active'
						: ''}"
				>
					<Icon i="compass" wh={16} />
					<span>{t("nav.discover")}</span>
				</a>
				<a
					href={resolve("/calendar")}
					class="nav-link {page.url?.pathname.startsWith('/calendar')
						? 'active'
						: ''}"
				>
					<Icon i="calendar" wh={16} />
					<span>{t("nav.calendar")}</span>
				</a>
			</div>
		</div>
		<div
			class="search clickable"
			role="button"
			tabindex="0"
			onclick={() => (commandPaletteOpen = true)}
			onkeydown={(e) => {
				if (e.key === "Enter" || e.key === " ") {
					e.preventDefault();
					commandPaletteOpen = true;
				}
			}}
		>
			<span class="search-icon">
				<Icon i="search" wh={16} />
			</span>
			<input
				type="text"
				placeholder={`${t("nav.search_placeholder")} (${isMac ? "⌘K" : "Ctrl+K"})`}
				readonly
				tabindex="-1"
			/>
			<kbd class="search-kbd">{isMac ? "⌘K" : "Ctrl+K"}</kbd>
		</div>
		<div class="btns">
			<!-- Tags menu -->
			<div class="nav-item-wrap">
				<button
					class="plain other tag {tagMenuShown ||
					page.url?.pathname.startsWith('/tag')
						? 'active'
						: ''}"
					onclick={() => {
						closeAllSubMenus("tag");
						tagMenuShown = !tagMenuShown;
					}}
					use:tooltip={{ text: "Tags", pos: "bot", condition: !tagMenuShown }}
					aria-label="Tags"
				>
					<Icon i="tag" wh={20} />
				</button>
				{#if tagMenuShown}
					<TagMenu
						onTagClick={(tag) => {
							goto(resolve(`/tag/${tag.id}`));
							tagMenuShown = false;
						}}
						showManageBtn={true}
						menuConfig={{
							width: "220px",
							right: "0px",
							top: "45px",
							arrowRight: "10px",
						}}
					/>
				{/if}
			</div>

			<!-- Following menu -->
			<div class="nav-item-wrap">
				<button
					class="plain other following {followingMenuShown ? 'active' : ''}"
					onclick={() => {
						closeAllSubMenus("following");
						followingMenuShown = !followingMenuShown;
					}}
					use:tooltip={{
						text: "Following",
						pos: "bot",
						condition: !followingMenuShown,
					}}
					aria-label="Following"
				>
					<Icon i="people" wh={20} />
				</button>
				{#if followingMenuShown}
					<FollowingMenu
						close={() => (followingMenuShown = false)}
						menuConfig={{
							width: "190px",
							right: "0px",
							top: "45px",
							arrowRight: "10px",
						}}
					/>
				{/if}
			</div>

			<!-- Account menu -->
			<div class="nav-item-wrap">
				<button
					class="plain face {subMenuShown ? 'active' : ''}"
					onclick={handleProfileClick}
					use:tooltip={{
						text: store.userInfo?.username
							? store.userInfo.username
							: "Account",
						pos: "bot",
						condition: !subMenuShown,
					}}
					aria-label="Account menu"
				>
					{#if store.userInfo?.avatar?.path}
						<img
							class="nav-avatar-img"
							src={`${baseURL}/${store.userInfo.avatar.path}`}
							alt=""
						/>
					{:else if store.userInfo?.username}
						<span class="nav-avatar-initial">
							{store.userInfo.username.charAt(0).toUpperCase()}
						</span>
					{:else}
						<span class="nav-avatar-face">:)</span>
					{/if}
				</button>
				{#if subMenuShown}
					<FaceMenu />
				{/if}
			</div>
		</div>
	</div>
	<div
		class="search-small-bar"
		role="button"
		tabindex="0"
		onclick={() => (commandPaletteOpen = true)}
		onkeydown={(e) => {
			if (e.key === "Enter" || e.key === " ") {
				e.preventDefault();
				commandPaletteOpen = true;
			}
		}}
	>
		<Icon i="search" wh={15} />
		<span>Search or jump to...</span>
		<kbd>{isMac ? "⌘K" : "Ctrl+K"}</kbd>
	</div>
</nav>

{#await getInitialData()}
	<Spinner />
{:then}
	{@render children?.()}
{:catch err}
	<Error
		pretty="Couldn't fetch app data!"
		error={err}
		onRetry={() => {
			location.reload();
		}}
	/>
{/await}

<!-- Mobile Bottom Navigation Bar -->
<nav class="mobile-bottom-bar" aria-label="Mobile navigation">
	<a
		href={resolve("/")}
		class="nav-tab {page.url?.pathname === '/' ? 'active' : ''}"
	>
		<Icon i="film" wh={20} />
		<span>{t("nav.library")}</span>
	</a>
	<a
		href={resolve("/calendar")}
		class="nav-tab {page.url?.pathname.startsWith('/calendar') ? 'active' : ''}"
	>
		<Icon i="calendar" wh={20} />
		<span>{t("nav.calendar")}</span>
	</a>
	<a
		href={resolve("/discover")}
		class="nav-tab {page.url?.pathname.startsWith('/discover') ? 'active' : ''}"
	>
		<Icon i="compass" wh={20} />
		<span>{t("nav.discover")}</span>
	</a>
	<button
		class="nav-tab plain {commandPaletteOpen ? 'active' : ''}"
		onclick={() => (commandPaletteOpen = true)}
		aria-label="Search and Commands"
	>
		<Icon i="search" wh={20} />
		<span>{t("nav.search")}</span>
	</button>
	<button
		class="nav-tab plain {tagMenuShown ? 'active' : ''}"
		onclick={() => {
			closeAllSubMenus("tag");
			tagMenuShown = !tagMenuShown;
		}}
		aria-label="Tags menu"
	>
		<Icon i="tag" wh={20} />
		<span>Tags</span>
	</button>
	<button
		class="nav-tab plain {subMenuShown ? 'active' : ''}"
		onclick={handleProfileClick}
		aria-label="Account menu"
	>
		{#if store.userInfo?.avatar?.path}
			<img
				class="mobile-nav-avatar"
				src={`${baseURL}/${store.userInfo.avatar.path}`}
				alt=""
			/>
		{:else if store.userInfo?.username}
			<span class="mobile-nav-initial">
				{store.userInfo.username.charAt(0).toUpperCase()}
			</span>
		{:else}
			<Icon i="person" wh={20} />
		{/if}
		<span>Account</span>
	</button>
</nav>

{#if page.state.handbook}
	<HandbookModal onClose={() => history.back()} />
{/if}

<CommandPalette
	bind:open={commandPaletteOpen}
	onClose={() => (commandPaletteOpen = false)}
/>

<style lang="scss">
	nav.top-nav {
		view-transition-name: app-header;
		display: flex;
		flex-flow: column;
		margin-bottom: 24px;
		padding: 10px 24px;
		position: sticky;
		top: 0;
		gap: 6px;
		z-index: 99990;
		border-bottom: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
		background: rgba(13, 17, 23, 0.94);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.35);

		@media screen and (max-width: 768px) {
			padding: 8px 16px;
			margin-bottom: 16px;
		}

		.wrapper {
			display: flex;
			flex-flow: row;
			gap: 20px;
			justify-content: space-between;
			align-items: center;

			.left-side,
			.btns {
				flex: 1;
			}

			@media screen and (max-width: 435px) {
				gap: 12px;
			}

			:global(body.split-nav) & {
				@media screen and (max-width: 380px) {
					gap: 10px;
				}

				@media screen and (max-width: 375px) {
					gap: 8px;
				}

				@media screen and (max-width: 370px) {
					gap: 5px;
				}

				@media screen and (max-width: 350px) {
					gap: 0;
				}
			}
		}

		.left-side {
			display: flex;
			align-items: center;
			gap: 20px;

			a.logo {
				display: inline-flex;
				align-items: center;
				text-decoration: none;
				font-family:
					"Shrikhand",
					system-ui,
					-apple-system,
					BlinkMacSystemFont;
				font-size: 28px;
				letter-spacing: -0.01em;
				background: linear-gradient(135deg, $text-color 65%, $brand-color 100%);
				-webkit-background-clip: text;
				background-clip: text;
				-webkit-text-fill-color: transparent;
				transition:
					opacity 150ms ease,
					transform 150ms ease;

				&:hover,
				&:focus-visible {
					opacity: 0.9;
					transform: scale(1.02);
				}

				span.large {
					display: block;
				}

				span.small {
					display: none;
				}

				@media screen and (max-width: 620px) {
					span.large {
						display: none;
					}
					span.small {
						display: block;
					}
				}
			}

			.main-nav-links {
				display: flex;
				align-items: center;
				gap: 4px;

				@media screen and (max-width: 768px) {
					display: none;
				}

				.nav-link {
					display: inline-flex;
					align-items: center;
					gap: 6px;
					padding: 6px 12px;
					border-radius: 8px;
					text-decoration: none;
					color: $text-color-accent;
					font-size: 13.5px;
					font-weight: 500;
					transition: all 150ms ease;

					&:hover {
						color: $text-color;
						background-color: var(--accent-color, rgba(255, 255, 255, 0.06));
					}

					&.active {
						color: #ffffff;
						font-weight: 600;
						background-color: var(--accent-color, rgba(255, 255, 255, 0.12));
						border: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
					}
				}
			}
		}

		.search {
			width: 100%;
			max-width: 360px;
			position: relative;
			display: flex;
			align-items: center;
			cursor: pointer;

			&:hover input {
				border-color: rgba(99, 102, 241, 0.35);
				background-color: var(--input-bg, rgba(255, 255, 255, 0.07));
			}

			.search-icon {
				position: absolute;
				left: 14px;
				pointer-events: none;
				user-select: none;
				display: flex;
				align-items: center;
				color: $placeholder-color;
				fill: $placeholder-color;
				z-index: 2;
			}

			.search-kbd {
				position: absolute;
				right: 10px;
				padding: 2px 6px;
				font-size: 10px;
				font-family: inherit;
				font-weight: 600;
				color: $text-color-accent;
				background-color: var(--accent-color);
				border: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
				border-radius: 4px;
				pointer-events: none;
				user-select: none;

				@media screen and (max-width: 580px) {
					display: none;
				}
			}

			input {
				width: 100%;
				padding: 8px 34px 8px 38px;
				border-radius: 9999px;
				border: 1px solid var(--input-border, rgba(255, 255, 255, 0.12));
				background-color: var(--input-bg, rgba(255, 255, 255, 0.04));
				color: $text-color;
				font-size: 13.5px;
				font-weight: 500;
				text-align: left;
				box-shadow: none;
				outline: none;
				text-overflow: ellipsis;
				transition:
					border-color 150ms ease,
					box-shadow 150ms ease,
					background-color 150ms ease;

				&::placeholder {
					color: $placeholder-color;
				}

				&:hover {
					border-color: var(--card-border, rgba(255, 255, 255, 0.2));
				}

				&:focus {
					background-color: var(--card-bg, $bg-color);
					border-color: $brand-color;
					box-shadow: 0 0 0 3px $brand-glow;
				}
			}

			@media screen and (max-width: 666px) {
				&:focus-within + .btns button:not(.face) {
					display: none;
				}
			}
		}

		:global(body.split-nav) & {
			.search {
				opacity: 0;
				visibility: hidden;
			}

			.search-small-bar {
				display: flex;
			}
		}

		.search-small-bar {
			display: none;
			align-items: center;
			gap: 8px;
			width: 100%;
			max-width: 340px;
			margin: 4px auto 0 auto;
			padding: 7px 16px;
			border-radius: 9999px;
			border: 1px solid var(--input-border, rgba(255, 255, 255, 0.12));
			background-color: var(--input-bg, rgba(255, 255, 255, 0.04));
			color: $placeholder-color;
			font-size: 13px;
			font-weight: 500;
			cursor: pointer;
			user-select: none;
			transition: all 140ms ease;

			kbd {
				margin-left: auto;
				padding: 1px 5px;
				font-size: 10px;
				border-radius: 4px;
				background: rgba(255, 255, 255, 0.08);
				border: 1px solid rgba(255, 255, 255, 0.1);
				color: $text-color-accent;
			}

			&:hover {
				border-color: rgba(99, 102, 241, 0.4);
				background-color: rgba(255, 255, 255, 0.08);
				color: #ffffff;
			}
		}

		.btns {
			display: flex;
			flex-flow: row;
			align-items: center;
			justify-content: flex-end;
			gap: 8px;
			flex-shrink: 0;

			.nav-item-wrap {
				position: relative;
				display: inline-flex;
				align-items: center;
			}

			@media screen and (max-width: 640px) {
				.nav-item-wrap:has(button.following),
				.nav-item-wrap:has(button.tag) {
					display: none !important;
				}
			}

			button.other {
				display: inline-flex;
				align-items: center;
				justify-content: center;
				width: 36px;
				height: 36px;
				border-radius: 8px;
				border: 1px solid transparent;
				color: $text-color;
				fill: $text-color;
				position: relative;
				transition: all 150ms ease;

				&:hover,
				&:focus-visible {
					background-color: $accent-color;
					border-color: var(--card-border, rgba(255, 255, 255, 0.08));
					transform: translateY(-1px);
				}

				&:active {
					transform: translateY(0);
				}

				&.active {
					color: $brand-color;
					fill: $brand-color;
					background-color: var(--accent-color, rgba(255, 255, 255, 0.08));
					border-color: rgba(245, 158, 11, 0.4);
					box-shadow: 0 0 10px rgba(245, 158, 11, 0.2);
				}
			}

			button.face {
				display: inline-flex;
				align-items: center;
				justify-content: center;
				width: 34px;
				height: 34px;
				border-radius: 50%;
				overflow: hidden;
				border: 1.5px solid var(--card-border, rgba(255, 255, 255, 0.15));
				background: linear-gradient(
					135deg,
					rgba(245, 158, 11, 0.25),
					rgba(217, 119, 6, 0.35)
				);
				cursor: pointer;
				transition: all 150ms ease;
				padding: 0;
				margin-left: 4px;

				&:hover,
				&:focus-visible {
					border-color: $brand-color;
					transform: scale(1.06);
					box-shadow: 0 0 0 3px $brand-glow;
				}

				.nav-avatar-img {
					width: 100%;
					height: 100%;
					object-fit: cover;
				}

				.nav-avatar-initial {
					font-size: 13px;
					font-weight: 700;
					color: $text-color;
				}

				.nav-avatar-face {
					font-size: 13px;
					font-weight: bold;
					transform: rotate(90deg);
					color: $text-color;
				}
			}
		}
	}

	.mobile-bottom-bar {
		view-transition-name: app-bottom-nav;
		display: none;

		@media screen and (max-width: 768px) {
			display: flex;
			flex-flow: row nowrap;
			position: fixed;
			bottom: 0;
			left: 0;
			right: 0;
			top: auto;
			width: 100%;
			height: calc(56px + env(safe-area-inset-bottom, 0px));
			margin: 0;
			padding: 0 8px;
			padding-bottom: env(safe-area-inset-bottom, 0px);
			background: rgba(11, 14, 20, 0.88);
			backdrop-filter: blur(20px) saturate(180%);
			-webkit-backdrop-filter: blur(20px) saturate(180%);
			border-top: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
			border-bottom: none;
			z-index: 99990;
			align-items: center;
			justify-content: space-around;
			box-sizing: border-box;

			.nav-tab {
				display: flex;
				flex-flow: column;
				align-items: center;
				justify-content: center;
				gap: 3px;
				flex: 1;
				height: 100%;
				color: $text-color-accent;
				text-decoration: none;
				font-size: 11px;
				font-weight: 500;
				transition: all 140ms ease;
				border: none;
				background: transparent;
				cursor: pointer;

				span {
					font-size: 10.5px;
					line-height: 1;
					letter-spacing: 0.01em;
				}

				&:hover,
				&:focus-visible,
				&.active {
					color: $brand-color;

					:global(svg) {
						transform: translateY(-1px);
					}
				}

				.mobile-nav-avatar {
					width: 22px;
					height: 22px;
					border-radius: 50%;
					object-fit: cover;
					border: 1px solid var(--card-border, rgba(255, 255, 255, 0.15));
				}

				.mobile-nav-initial {
					width: 22px;
					height: 22px;
					border-radius: 50%;
					background: rgba(245, 158, 11, 0.15);
					border: 1px solid rgba(245, 158, 11, 0.25);
					display: flex;
					align-items: center;
					justify-content: center;
					font-size: 11px;
					font-weight: 700;
					color: $brand-color;
				}
			}
		}
	}

	@media screen and (max-width: 768px) {
		:global(body) {
			padding-bottom: calc(68px + env(safe-area-inset-bottom, 0px));
		}
	}
</style>
