<script lang="ts">
	import { afterNavigate, goto, onNavigate } from "$app/navigation";
	import { resolve } from "$app/paths";
	import { page } from "$app/state";
	import Error from "@/lib/Error.svelte";
	import Icon from "@/lib/Icon.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import tooltip from "@/lib/actions/tooltip";
	import FaceMenu from "@/lib/nav/FaceMenu.svelte";
	import FollowingMenu from "@/lib/nav/FollowingMenu.svelte";
	import TagMenu from "@/lib/tag/TagMenu.svelte";
	import { baseURL, req, updateUserSetting } from "@/lib/util/api";
	import { store } from "@/store.svelte";
	import {
		MediaTypeE,
		type Media,
		type PaginationResponse,
		type ServerFeatures,
		type Follow,
		type PrivateUser,
		type Tag,
		type UserSettings,
	} from "@/types";
	import { onMount } from "svelte";
	import HandbookModal from "@/lib/handbook/HandbookModal.svelte";
	import { t, i18n } from "@/lib/i18n";
	import { toggleTheme } from "@/lib/util/theme";

	interface Props {
		children?: import("svelte").Snippet;
	}

	let { children }: Props = $props();

	onNavigate((navigation) => {
		if (!document.startViewTransition) return;
		return new Promise((resolve) => {
			document.startViewTransition(async () => {
				resolve();
				await navigation.complete;
			});
		});
	});

	let navEl: HTMLElement | undefined = $state();
	let mainSearchEl: HTMLInputElement | undefined = $state();
	let subMenuShown = $state(false);
	let followingMenuShown = $state(false);
	let tagMenuShown = $state(false);
	let isMac = $state(false);
	let isOffline = $state(false);
	let scroll = window.scrollY;

	// Enhanced live search state
	let navSearchQuery = $state("");
	let navSearchFocused = $state(false);
	let navSearchResults: Media[] = $state([]);
	let navSearchLoading = $state(false);
	let navSelectedIndex = $state(-1);
	let navSearchDebounce: number;

	function getHighlightedParts(
		text: string,
		query: string,
	): { str: string; match: boolean }[] {
		if (!query || !query.trim() || !text) {
			return [{ str: text || "", match: false }];
		}
		const trimmed = query.trim();
		const escaped = trimmed.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
		const regex = new RegExp(`(${escaped})`, "gi");
		const parts = text.split(regex);
		return parts
			.filter((p) => p.length > 0)
			.map((p) => ({
				str: p,
				match: p.toLowerCase() === trimmed.toLowerCase(),
			}));
	}

	function performNavSearch(query: string) {
		clearTimeout(navSearchDebounce);
		if (!query || query.trim().length < 2) {
			navSearchResults = [];
			navSearchLoading = false;
			navSelectedIndex = -1;
			return;
		}
		navSearchLoading = true;
		navSearchDebounce = window.setTimeout(async () => {
			try {
				const res = await req.get<PaginationResponse<Media, unknown>>("/search", {
					params: {
						query: query.trim(),
						type: "multi",
						page: 1,
						preferMyList: "true",
					},
				});
				if (res?.results) {
					navSearchResults = res.results.slice(0, 6);
				} else {
					navSearchResults = [];
				}
			} catch (err) {
				console.error("Live search failed:", err);
				navSearchResults = [];
			} finally {
				navSearchLoading = false;
				navSelectedIndex = -1;
			}
		}, 200);
	}

	function getMediaHref(m: Media): string {
		switch (m.type) {
			case MediaTypeE.tmdbMovie:
				return `/movie/${m.ids.tmdb}`;
			case MediaTypeE.tmdbShow:
				return `/tv/${m.ids.tmdb}`;
			case MediaTypeE.igdbGame:
				return `/game/${m.ids.igdb}`;
			default:
				if (m.ids.tmdb) return `/movie/${m.ids.tmdb}`;
				return `/search?query=${encodeURIComponent(m.name || "")}`;
		}
	}

	function getMediaPosterUrl(m: Media): string | undefined {
		if (m.poster?.path) return `${baseURL}/${m.poster.path}`;
		if (m.extPosterPath) {
			if (m.type === MediaTypeE.igdbGame) {
				return `https://images.igdb.com/igdb/image/upload/t_cover_big/${m.extPosterPath}.jpg`;
			}
			return `https://image.tmdb.org/t/p/w92${m.extPosterPath}`;
		}
		return undefined;
	}

	function getMediaBadge(m: Media): string {
		switch (m.type) {
			case MediaTypeE.tmdbMovie:
				return "Movie";
			case MediaTypeE.tmdbShow:
				return "TV";
			case MediaTypeE.igdbGame:
				return "Game";
			default:
				return "Media";
		}
	}

	function getMediaYear(m: Media): string | undefined {
		if (!m.releaseDate) return undefined;
		const y = new Date(m.releaseDate).getFullYear();
		return isNaN(y) ? undefined : String(y);
	}

	function navigateToMedia(m: Media) {
		const href = getMediaHref(m);
		navSearchFocused = false;
		navSearchResults = [];
		navSearchQuery = "";
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		goto(resolve(href as any));
	}

	function submitFullSearch(q?: string) {
		const query = (q !== undefined ? q : navSearchQuery).trim();
		if (!query) return;
		navSearchFocused = false;
		navSearchResults = [];
		navSearchQuery = "";
		goto(resolve(`/search?query=${encodeURIComponent(query)}&preferMyList=true`));
	}

	function handleSearchKeyDown(ev: KeyboardEvent) {
		if (ev.key === "ArrowDown") {
			if (navSearchResults.length > 0) {
				ev.preventDefault();
				navSelectedIndex = (navSelectedIndex + 1) % navSearchResults.length;
			}
		} else if (ev.key === "ArrowUp") {
			if (navSearchResults.length > 0) {
				ev.preventDefault();
				navSelectedIndex =
					navSelectedIndex <= 0
						? navSearchResults.length - 1
						: navSelectedIndex - 1;
			}
		} else if (ev.key === "Enter") {
			ev.preventDefault();
			if (navSelectedIndex >= 0 && navSearchResults[navSelectedIndex]) {
				navigateToMedia(navSearchResults[navSelectedIndex]);
			} else {
				submitFullSearch();
			}
		} else if (ev.key === "Escape") {
			navSearchFocused = false;
			navSearchResults = [];
			mainSearchEl?.blur();
		}
	}

	function focusSearch() {
		mainSearchEl?.focus();
		mainSearchEl?.select();
	}

	function handleProfileClick() {
		if (!localStorage.getItem("token")) {
			goto(resolve("/login"));
		} else {
			closeAllSubMenus("sub");
			subMenuShown = !subMenuShown;
		}
	}

	function isPublicRoute(pathname: string): boolean {
		return pathname.startsWith("/tag/") || pathname.startsWith("/lists/");
	}

	async function getInitialData() {
		if (!localStorage.getItem("token")) {
			if (isPublicRoute(page.url.pathname)) {
				console.debug(
					"getInitialData: No token, but viewing public route:",
					page.url.pathname,
				);
				return;
			}
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
			const localLang =
				typeof window !== "undefined"
					? (localStorage.getItem("watcharr_lang") as "en" | "pt" | null)
					: null;
			if (localLang && (localLang === "en" || localLang === "pt")) {
				i18n.setLocale(localLang);
				if (s.uiLanguage !== localLang) {
					updateUserSetting("uiLanguage", localLang);
				}
			} else if (
				s.uiLanguage &&
				(s.uiLanguage === "en" || s.uiLanguage === "pt")
			) {
				i18n.setLocale(s.uiLanguage);
			}
			if (
				s.theme &&
				(s.theme === "system" || s.theme === "dark" || s.theme === "light")
			) {
				toggleTheme(s.theme);
			}
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
	 * Close open nav menus when clicking into the page outside of nav menu wrappers (#813).
	 */
	function handleDocumentPointerDown(ev: PointerEvent) {
		const target = ev.target as HTMLElement | null;
		if (
			!target?.closest(".nav-item-wrap") &&
			!target?.closest(".mobile-bottom-bar") &&
			!target?.closest(".search")
		) {
			closeAllSubMenus();
			navSearchFocused = false;
		}
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
			navSearchFocused = false;
		}
		scroll = Math.max(0, currentScrollY);
	}

	function handleGlobalKeybind(ev: KeyboardEvent) {
		if (
			ev.key === "Escape" &&
			(subMenuShown || followingMenuShown || tagMenuShown || navSearchFocused)
		) {
			closeAllSubMenus();
			navSearchFocused = false;
			navSearchResults = [];
			return;
		}
		const isCmdOrCtrl = ev.ctrlKey || ev.metaKey;
		const key = ev.key.toLowerCase();
		if (isCmdOrCtrl && (key === "k" || key === "p" || key === "s")) {
			ev.preventDefault();
			focusSearch();
			return;
		}
		if (key === "/" && !isCmdOrCtrl) {
			const target = ev.target as HTMLElement | null;
			const tag = target?.tagName?.toLowerCase();
			if (tag !== "input" && tag !== "textarea" && !target?.isContentEditable) {
				ev.preventDefault();
				focusSearch();
			}
		}
	}

	afterNavigate(() => {
		decideOnNavSplit();
		closeAllSubMenus();
		navSearchFocused = false;
		navSearchResults = [];
	});

	onMount(() => {
		isMac =
			typeof navigator !== "undefined" &&
			/Mac|iPod|iPhone|iPad/.test(navigator.platform || navigator.userAgent);
		isOffline = typeof navigator !== "undefined" ? !navigator.onLine : false;

		const handleOnline = () => {
			isOffline = false;
		};
		const handleOffline = () => {
			isOffline = true;
		};
		window.addEventListener("online", handleOnline);
		window.addEventListener("offline", handleOffline);

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
			window.document.addEventListener("pointerdown", handleDocumentPointerDown);

			return () => {
				window.removeEventListener("online", handleOnline);
				window.removeEventListener("offline", handleOffline);
				window.removeEventListener("resize", decideOnNavSplit);
				window.document.removeEventListener("scroll", docOnScroll);
				window.document.removeEventListener("keydown", handleGlobalKeybind);
				window.document.removeEventListener(
					"pointerdown",
					handleDocumentPointerDown,
				);
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
				<a
					href={resolve("/franchise")}
					class="nav-link {page.url?.pathname.startsWith('/franchise')
						? 'active'
						: ''}"
				>
					<Icon i="reel" wh={16} />
					<span>Sagas</span>
				</a>
			</div>
		</div>
		<div class="search" class:is-focused={navSearchFocused}>
			<span class="search-icon">
				<Icon i="search" wh={16} />
			</span>
			<input
				bind:this={mainSearchEl}
				type="text"
				placeholder={`${t("nav.search_placeholder")} (${isMac ? "⌘K" : "Ctrl+K"})`}
				bind:value={navSearchQuery}
				oninput={(e) => performNavSearch((e.target as HTMLInputElement).value)}
				onfocus={() => {
					navSearchFocused = true;
					if (navSearchQuery.trim().length >= 2 && navSearchResults.length === 0) {
						performNavSearch(navSearchQuery);
					}
				}}
				onkeydown={handleSearchKeyDown}
				aria-label={t("nav.search_placeholder")}
				autocomplete="off"
				spellcheck="false"
			/>
			{#if navSearchLoading}
				<span class="search-spinner" aria-hidden="true"></span>
			{:else if navSearchQuery}
				<button
					type="button"
					class="plain search-clear-btn"
					onclick={() => {
						navSearchQuery = "";
						navSearchResults = [];
						mainSearchEl?.focus();
					}}
					aria-label="Clear search"
				>
					<Icon i="close" wh={13} />
				</button>
			{:else}
				<kbd class="search-kbd">{isMac ? "⌘K" : "Ctrl+K"}</kbd>
			{/if}

			{#if navSearchFocused && (navSearchResults.length > 0 || (navSearchQuery.trim().length >= 2 && !navSearchLoading))}
				<div class="search-preview-popover" role="listbox">
					{#if navSearchResults.length > 0}
						<div class="preview-results-list">
							{#each navSearchResults as item, idx (item.ids.tmdb || item.ids.igdb || item.name || idx)}
								<a
									href={resolve(getMediaHref(item) as any)}
									class="preview-item {navSelectedIndex === idx ? 'selected' : ''}"
									onclick={(e) => {
										e.preventDefault();
										navigateToMedia(item);
									}}
									onmouseenter={() => (navSelectedIndex = idx)}
									role="option"
									aria-selected={navSelectedIndex === idx}
								>
									<div class="preview-thumb">
										{#if getMediaPosterUrl(item)}
											<img src={getMediaPosterUrl(item)} alt="" />
										{:else}
											<div class="preview-thumb-fallback">
												<Icon i="film" wh={14} />
											</div>
										{/if}
									</div>
									<div class="preview-info">
										<div class="preview-title-row">
											<span class="preview-title">
												{#each getHighlightedParts(item.name || "Untitled", navSearchQuery) as part}
													{#if part.match}
														<mark class="search-match">{part.str}</mark>
													{:else}
														{part.str}
													{/if}
												{/each}
											</span>
											{#if item.watched}
												<span class="preview-watched-tag">Na Lista</span>
											{/if}
										</div>
										<div class="preview-sub">
											<span class="preview-badge {getMediaBadge(item).toLowerCase()}">
												{getMediaBadge(item)}
											</span>
											{#if getMediaYear(item)}
												<span class="preview-year">{getMediaYear(item)}</span>
											{/if}
										</div>
									</div>
								</a>
							{/each}
						</div>
						<div class="preview-footer">
							<button
								type="button"
								class="plain preview-view-all"
								onclick={() => submitFullSearch()}
							>
								<span>Ver todos os resultados para <strong>"{navSearchQuery.trim()}"</strong></span>
								<kbd>Enter ↵</kbd>
							</button>
						</div>
					{:else if !navSearchLoading}
						<div class="preview-empty">
							<span>Nenhum resultado para "{navSearchQuery.trim()}"</span>
						</div>
					{/if}
				</div>
			{/if}
		</div>
		<div class="btns">
			{#if isOffline}
				<div
					class="offline-pill"
					role="status"
					aria-live="polite"
					title="Offline - Modo de leitura a partir da cache"
				>
					<span class="offline-dot"></span>
					<span>Offline</span>
				</div>
			{/if}
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
	<div class="search-small-bar">
		<Icon i="search" wh={15} />
		<input
			type="text"
			placeholder={`${t("nav.search_placeholder")}...`}
			bind:value={navSearchQuery}
			onkeydown={handleSearchKeyDown}
			onfocus={() => {
				navSearchFocused = true;
			}}
		/>
		<button
			type="button"
			class="plain search-small-go"
			onclick={() => submitFullSearch()}
			aria-label="Search"
		>
			<Icon i="arrow" wh={12} />
		</button>
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
	<a
		href={resolve("/search")}
		class="nav-tab {page.url?.pathname.startsWith('/search') ? 'active' : ''}"
		aria-label={t("nav.search")}
	>
		<Icon i="search" wh={20} />
		<span>{t("nav.search")}</span>
	</a>
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
	<a
		href={resolve("/profile")}
		class="nav-tab {page.url?.pathname.startsWith('/profile') ? 'active' : ''}"
		aria-label="Perfil & Definições"
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
		<span>{t("nav.account")}</span>
	</a>
</nav>

{#if page.state.handbook}
	<HandbookModal onClose={() => history.back()} />
{/if}

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
			max-width: 380px;
			position: relative;
			display: flex;
			align-items: center;

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

			.search-spinner {
				position: absolute;
				right: 12px;
				width: 14px;
				height: 14px;
				border: 2px solid rgba(255, 255, 255, 0.2);
				border-top-color: $brand-color;
				border-radius: 50%;
				animation: spin 0.6s linear infinite;
				z-index: 2;
			}

			@keyframes spin {
				to {
					transform: rotate(360deg);
				}
			}

			.search-clear-btn {
				position: absolute;
				right: 10px;
				display: flex;
				align-items: center;
				justify-content: center;
				width: 22px;
				height: 22px;
				border-radius: 50%;
				background: rgba(255, 255, 255, 0.08);
				color: $placeholder-color;
				cursor: pointer;
				z-index: 2;
				transition: all 120ms ease;

				&:hover {
					background: rgba(255, 255, 255, 0.16);
					color: #ffffff;
				}
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
					background-color: var(--input-bg, rgba(255, 255, 255, 0.06));
				}

				&:focus {
					background-color: var(--card-bg, $bg-color);
					border-color: $brand-color;
					box-shadow: 0 0 0 3px $brand-glow;
				}
			}

			.search-preview-popover {
				position: absolute;
				top: calc(100% + 8px);
				left: 0;
				right: 0;
				min-width: 320px;
				max-height: 420px;
				background: rgba(18, 21, 28, 0.98);
				backdrop-filter: blur(20px);
				-webkit-backdrop-filter: blur(20px);
				border: 1px solid var(--card-border, rgba(255, 255, 255, 0.12));
				border-radius: 12px;
				box-shadow: 0 16px 40px rgba(0, 0, 0, 0.45);
				z-index: 99995;
				overflow: hidden;
				display: flex;
				flex-direction: column;

				.preview-results-list {
					display: flex;
					flex-direction: column;
					padding: 6px;
					overflow-y: auto;
					max-height: 350px;
					gap: 2px;
				}

				.preview-item {
					display: flex;
					align-items: center;
					gap: 12px;
					padding: 6px 8px;
					border-radius: 8px;
					text-decoration: none;
					color: $text-color;
					transition: background-color 120ms ease;
					cursor: pointer;

					&:hover,
					&.selected {
						background-color: rgba(255, 255, 255, 0.08);
					}

					.preview-thumb {
						width: 28px;
						height: 40px;
						border-radius: 4px;
						overflow: hidden;
						flex-shrink: 0;
						background: rgba(255, 255, 255, 0.05);
						border: 1px solid rgba(255, 255, 255, 0.08);

						img {
							width: 100%;
							height: 100%;
							object-fit: cover;
						}

						.preview-thumb-fallback {
							width: 100%;
							height: 100%;
							display: flex;
							align-items: center;
							justify-content: center;
							color: $placeholder-color;
						}
					}

					.preview-info {
						display: flex;
						flex-direction: column;
						min-width: 0;
						flex: 1;
						gap: 2px;

						.preview-title-row {
							display: flex;
							align-items: center;
							gap: 6px;

							.preview-title {
								font-size: 13px;
								font-weight: 600;
								color: $text-color;
								white-space: nowrap;
								overflow: hidden;
								text-overflow: ellipsis;

								.search-match {
									background: rgba(245, 158, 11, 0.24);
									color: #fbbf24;
									font-weight: 700;
									border-radius: 2px;
									padding: 0 1px;
								}
							}

							.preview-watched-tag {
								font-size: 9px;
								font-weight: 700;
								padding: 1px 5px;
								border-radius: 9999px;
								background: rgba(245, 158, 11, 0.2);
								border: 1px solid rgba(245, 158, 11, 0.4);
								color: $brand-color;
								flex-shrink: 0;
								text-transform: uppercase;
							}
						}

						.preview-sub {
							display: flex;
							align-items: center;
							gap: 6px;
							font-size: 11px;
							color: $text-color-accent;

							.preview-badge {
								font-size: 10px;
								font-weight: 600;
								padding: 1px 4px;
								border-radius: 4px;
								background: rgba(255, 255, 255, 0.08);
								text-transform: uppercase;
								letter-spacing: 0.02em;

								&.movie {
									color: #60a5fa;
								}
								&.tv {
									color: #34d399;
								}
								&.game {
									color: #c084fc;
								}
							}

							.preview-year {
								font-weight: 500;
							}
						}
					}
				}

				.preview-footer {
					padding: 6px 8px;
					border-top: 1px solid rgba(255, 255, 255, 0.06);
					background: rgba(0, 0, 0, 0.2);

					.preview-view-all {
						width: 100%;
						display: flex;
						align-items: center;
						justify-content: space-between;
						padding: 6px 10px;
						border-radius: 6px;
						font-size: 12px;
						color: $text-color-accent;
						cursor: pointer;
						transition: all 120ms ease;

						strong {
							color: $text-color;
						}

						kbd {
							font-size: 10px;
							padding: 1px 5px;
							border-radius: 4px;
							background: rgba(255, 255, 255, 0.08);
							border: 1px solid rgba(255, 255, 255, 0.1);
							color: $brand-color;
						}

						&:hover {
							background: rgba(255, 255, 255, 0.08);
							color: #ffffff;
						}
					}
				}

				.preview-empty {
					padding: 18px 12px;
					text-align: center;
					font-size: 12.5px;
					color: $text-color-accent;
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
			padding: 4px 12px;
			border-radius: 9999px;
			border: 1px solid var(--input-border, rgba(255, 255, 255, 0.12));
			background-color: var(--input-bg, rgba(255, 255, 255, 0.04));
			color: $text-color;
			font-size: 13px;
			font-weight: 500;
			box-sizing: border-box;

			input {
				background: transparent;
				border: none;
				outline: none;
				color: $text-color;
				width: 100%;
				font-size: 13px;

				&::placeholder {
					color: $placeholder-color;
				}
			}

			.search-small-go {
				display: flex;
				align-items: center;
				justify-content: center;
				padding: 4px;
				color: $brand-color;
				cursor: pointer;
			}
		}

		.btns {
			display: flex;
			flex-flow: row;
			align-items: center;
			justify-content: flex-end;
			gap: 8px;
			flex-shrink: 0;

			.offline-pill {
				display: inline-flex;
				align-items: center;
				gap: 5px;
				padding: 2px 7px;
				font-size: 11px;
				font-weight: 600;
				letter-spacing: 0.03em;
				text-transform: uppercase;
				color: #d97706;
				background: rgba(217, 119, 6, 0.12);
				border: 1px solid rgba(217, 119, 6, 0.28);
				border-radius: 4px;
				user-select: none;
				white-space: nowrap;

				.offline-dot {
					width: 6px;
					height: 6px;
					border-radius: 50%;
					background: #f59e0b;
				}
			}

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

	:global {
		@media (prefers-reduced-motion: no-preference) {
			::view-transition-old(root) {
				animation: 140ms cubic-bezier(0.16, 1, 0.3, 1) both vt-fade-out;
			}
			::view-transition-new(root) {
				animation: 160ms cubic-bezier(0.16, 1, 0.3, 1) both vt-fade-in;
			}
		}

		@keyframes vt-fade-out {
			from { opacity: 1; }
			to { opacity: 0; }
		}

		@keyframes vt-fade-in {
			from { opacity: 0; }
			to { opacity: 1; }
		}
	}
</style>
