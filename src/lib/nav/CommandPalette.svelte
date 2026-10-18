<script lang="ts">
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import Icon from "@/lib/Icon.svelte";
	import { req } from "@/lib/util/api";
	import { store } from "@/store.svelte";
	import { toggleTheme } from "@/lib/util/theme";
	import { userHasPermission } from "@/lib/util/helpers";
	import {
		MediaTypeE,
		UserPermission,
		type Media,
		type PaginationResponse,
		type SearchResponseMeta,
		type Icon as IconType,
	} from "@/types";
	import {
		getMediaLink,
		getPosterImage,
		getReleaseYear,
	} from "@/lib/dashboard/mediaHelpers";
	import { onDestroy, onMount } from "svelte";
	import { searchCache } from "@/lib/util/searchCache";

	interface Props {
		open: boolean;
		onClose: () => void;
		onOpenWebhookModal?: () => void;
	}

	let {
		open = $bindable(false),
		onClose,
		onOpenWebhookModal,
	}: Props = $props();

	let query = $state("");
	let selectedIndex = $state(0);
	let isSearching = $state(false);
	let searchResults = $state<Media[]>([]);
	let searchTimeout: number | undefined;
	let inputEl: HTMLInputElement | undefined = $state();
	let resultsContainerEl: HTMLDivElement | undefined = $state();
	let isMac = $state(false);

	interface PaletteAction {
		id: string;
		section: "Navigation" | "Quick Actions" | "Search Results";
		title: string;
		subtitle?: string;
		icon?: IconType;
		poster?: string;
		badge?: string;
		badgeColor?: string;
		action: () => void;
	}

	const baseNavigation: PaletteAction[] = [
		{
			id: "nav-library",
			section: "Navigation",
			title: "Library",
			subtitle: "Browse your movies, series and games",
			icon: "film",
			action: () => navigateTo("/"),
		},
		{
			id: "nav-discover",
			section: "Navigation",
			title: "Discover",
			subtitle: "Explore trending movies and upcoming series",
			icon: "compass",
			action: () => navigateTo("/discover"),
		},
		{
			id: "nav-calendar",
			section: "Navigation",
			title: "Release Calendar",
			subtitle: "Upcoming premiere dates and episode releases",
			icon: "calendar",
			action: () => navigateTo("/calendar"),
		},
		{
			id: "nav-import",
			section: "Navigation",
			title: "Import Media",
			subtitle: "Import from Letterboxd, Trakt or CSV",
			icon: "download",
			action: () => navigateTo("/import"),
		},
		{
			id: "nav-profile",
			section: "Navigation",
			title: "Profile & Statistics",
			subtitle: "View stats, watch history and account settings",
			icon: "person",
			action: () => navigateTo("/profile"),
		},
		{
			id: "nav-tags",
			section: "Navigation",
			title: "Tags & Collections",
			subtitle: "Manage your customized media tags",
			icon: "tag",
			action: () => navigateTo("/tag"),
		},
	];

	const baseQuickActions: PaletteAction[] = [
		{
			id: "action-theme",
			section: "Quick Actions",
			title: "Toggle Dark / Light Theme",
			subtitle: "Switch between dark and light color scheme",
			icon: "sparkles",
			action: () => {
				toggleTheme(store.appTheme === "dark" ? "light" : "dark");
				close();
			},
		},
		{
			id: "action-webhook",
			section: "Quick Actions",
			title: "Release Calendar Webhooks",
			subtitle: "Configure Discord, Ntfy or webhook alerts",
			icon: "notifications",
			action: () => {
				close();
				if (onOpenWebhookModal) {
					onOpenWebhookModal();
				} else {
					navigateTo("/calendar");
				}
			},
		},
		{
			id: "action-export",
			section: "Quick Actions",
			title: "Export Watchlist",
			subtitle: "Export your library data to CSV or JSON",
			icon: "download",
			action: () => navigateTo("/profile"),
		},
	];

	const allStaticActions = $derived.by(() => {
		const list = [...baseNavigation, ...baseQuickActions];
		if (
			store.userInfo?.permissions &&
			userHasPermission(store.userInfo.permissions, UserPermission.PERM_ADMIN)
		) {
			list.push({
				id: "nav-server",
				section: "Navigation",
				title: "Server Settings",
				subtitle: "Manage users, tasks, arr integrations and system",
				icon: "settings",
				action: () => navigateTo("/server"),
			});
		}
		return list;
	});

	const searchMediaActions = $derived.by<PaletteAction[]>(() => {
		return searchResults.map((m) => {
			const link = getMediaLink(m);
			const poster = getPosterImage(m);
			const year = getReleaseYear(m);
			const typeBadge =
				m.type === MediaTypeE.tmdbMovie
					? "Movie"
					: m.type === MediaTypeE.tmdbShow
						? "Series"
						: "Game";
			const badgeColor =
				m.type === MediaTypeE.tmdbMovie
					? "#38bdf8"
					: m.type === MediaTypeE.tmdbShow
						? "#a855f7"
						: "#10b981";

			return {
				id: `media-${m.type}-${m.ids?.tmdb ?? m.ids?.igdb ?? Math.random()}`,
				section: "Search Results",
				title: m.name || "Untitled",
				subtitle: year ? `${typeBadge} • ${year}` : typeBadge,
				poster: poster,
				icon:
					m.type === MediaTypeE.igdbGame
						? "gamepad"
						: m.type === MediaTypeE.tmdbShow
							? "tv"
							: "film",
				badge: typeBadge,
				badgeColor: badgeColor,
				action: () => navigateTo(link),
			};
		});
	});

	const filteredActions = $derived.by<PaletteAction[]>(() => {
		const q = query.trim().toLowerCase();
		if (!q) {
			return allStaticActions;
		}

		const staticMatches = allStaticActions.filter(
			(a) =>
				a.title.toLowerCase().includes(q) ||
				(a.subtitle && a.subtitle.toLowerCase().includes(q)),
		);

		return [...staticMatches, ...searchMediaActions];
	});

	// Keep selectedIndex within bounds
	$effect(() => {
		if (selectedIndex >= filteredActions.length) {
			selectedIndex = Math.max(0, filteredActions.length - 1);
		}
	});

	// Trigger debounced search when query changes
	$effect(() => {
		const q = query.trim();
		clearTimeout(searchTimeout);

		if (q.length < 2) {
			searchResults = [];
			isSearching = false;
			return;
		}

		// Instant cache check (0ms response)
		const cached = searchCache.get<
			PaginationResponse<Media, SearchResponseMeta>
		>(q, "multi", false, 1);
		if (cached?.results) {
			searchResults = cached.results.slice(0, 8);
			isSearching = false;
			return;
		}

		isSearching = true;
		searchTimeout = window.setTimeout(async () => {
			try {
				const res = await req.get<
					PaginationResponse<Media, SearchResponseMeta>
				>(`/search`, {
					params: { query: q, page: 1, type: "multi" },
				});
				if (res?.results) {
					searchCache.set(q, res, "multi", false, 1);
					searchResults = res.results.slice(0, 8);
				} else {
					searchResults = [];
				}
			} catch (err) {
				console.error("CommandPalette search failed:", err);
				searchResults = [];
			} finally {
				isSearching = false;
			}
		}, 220);
	});

	// Auto-scroll selected element into view
	$effect(() => {
		if (resultsContainerEl && filteredActions.length > 0) {
			const el = resultsContainerEl.querySelector<HTMLElement>(
				`[data-index="${selectedIndex}"]`,
			);
			if (el) {
				el.scrollIntoView({ block: "nearest" });
			}
		}
	});

	// Focus input when opened
	$effect(() => {
		if (open) {
			query = "";
			selectedIndex = 0;
			searchResults = [];
			setTimeout(() => {
				inputEl?.focus();
			}, 50);
		}
	});

	function navigateTo(path: string) {
		close();
		goto(resolve(path as any));
	}

	function close() {
		open = false;
		onClose();
	}

	function handleKeyDown(e: KeyboardEvent) {
		if (!open) return;

		if (e.key === "Escape") {
			e.preventDefault();
			close();
			return;
		}

		if (e.key === "ArrowDown") {
			e.preventDefault();
			if (filteredActions.length > 0) {
				selectedIndex = (selectedIndex + 1) % filteredActions.length;
			}
			return;
		}

		if (e.key === "ArrowUp") {
			e.preventDefault();
			if (filteredActions.length > 0) {
				selectedIndex =
					(selectedIndex - 1 + filteredActions.length) % filteredActions.length;
			}
			return;
		}

		if (e.key === "Enter") {
			e.preventDefault();
			const item = filteredActions[selectedIndex];
			if (item) {
				item.action();
			}
			return;
		}
	}

	onMount(() => {
		isMac =
			typeof navigator !== "undefined" &&
			/Mac|iPod|iPhone|iPad/.test(navigator.platform || navigator.userAgent);
		window.addEventListener("keydown", handleKeyDown);
		return () => {
			window.removeEventListener("keydown", handleKeyDown);
		};
	});

	onDestroy(() => {
		clearTimeout(searchTimeout);
	});
</script>

{#if open}
	<div
		class="command-palette-backdrop"
		onclick={(e) => {
			if (e.target === e.currentTarget) close();
		}}
		onkeydown={(e) => {
			if (e.key === "Escape") close();
		}}
		tabindex="-1"
		role="dialog"
		aria-modal="true"
		aria-label="Command Palette"
	>
		<div class="command-palette-modal">
			<!-- Header with Search Input -->
			<div class="palette-input-wrap">
				<span class="palette-search-icon">
					{#if isSearching}
						<span class="palette-spinner"></span>
					{:else}
						<Icon i="search" wh={20} />
					{/if}
				</span>
				<input
					bind:this={inputEl}
					bind:value={query}
					type="text"
					placeholder="Type a command or search movies, shows, games..."
					autocomplete="off"
					spellcheck="false"
				/>
				{#if query}
					<button
						class="palette-clear-btn"
						onclick={() => (query = "")}
						aria-label="Clear input"
					>
						<Icon i="close" wh={14} />
					</button>
				{/if}
				<button
					type="button"
					class="palette-esc-badge"
					onclick={close}
					aria-label="Close command palette"
				>
					ESC
				</button>
			</div>

			<!-- Results Body -->
			<div class="palette-results" bind:this={resultsContainerEl}>
				{#if filteredActions.length === 0}
					<div class="palette-empty">
						<Icon i="search" wh={32} />
						<p>No results found for "<strong>{query}</strong>"</p>
						<span
							>Try searching by movie title, director or navigating to library</span
						>
					</div>
				{:else}
					{@const sections = Array.from(
						new Set(filteredActions.map((a) => a.section)),
					)}
					{#each sections as section}
						{@const sectionItems = filteredActions
							.map((item, originalIndex) => ({ item, originalIndex }))
							.filter(({ item }) => item.section === section)}
						<div class="palette-section">
							<div class="palette-section-title">{section}</div>
							{#each sectionItems as { item, originalIndex }}
								<button
									class="palette-item"
									class:selected={selectedIndex === originalIndex}
									data-index={originalIndex}
									onclick={() => item.action()}
									onmouseenter={() => (selectedIndex = originalIndex)}
								>
									<div class="item-visual">
										{#if item.poster}
											<img
												src={item.poster}
												alt={item.title || "Poster"}
												class="item-poster"
												loading="lazy"
												decoding="async"
											/>
										{:else if item.icon}
											<div class="item-icon-box">
												<Icon i={item.icon} wh={18} />
											</div>
										{/if}
									</div>

									<div class="item-details">
										<div class="item-title">{item.title}</div>
										{#if item.subtitle}
											<div class="item-subtitle">{item.subtitle}</div>
										{/if}
									</div>

									{#if item.badge}
										<span
											class="item-badge"
											style={item.badgeColor
												? `color: ${item.badgeColor}; border-color: ${item.badgeColor}40; background: ${item.badgeColor}15;`
												: ""}
										>
											{item.badge}
										</span>
									{/if}

									{#if selectedIndex === originalIndex}
										<span class="item-enter-hint" title="Press Enter to select">
											↵
										</span>
									{/if}
								</button>
							{/each}
						</div>
					{/each}
				{/if}
			</div>

			<!-- Footer Keyboard Hints -->
			<div class="palette-footer">
				<div class="footer-hint">
					<kbd>↑</kbd><kbd>↓</kbd> <span>Navigate</span>
				</div>
				<div class="footer-hint">
					<kbd>↵</kbd> <span>Select</span>
				</div>
				<div class="footer-hint">
					<kbd>ESC</kbd> <span>Close</span>
				</div>
				<div class="footer-right">
					<span>Watcharr Command Palette</span>
				</div>
			</div>
		</div>
	</div>
{/if}

<style lang="scss">
	.command-palette-backdrop {
		position: fixed;
		inset: 0;
		z-index: 9999;
		display: flex;
		align-items: flex-start;
		justify-content: center;
		padding: 80px 16px 20px;
		background: rgba(3, 7, 18, 0.72);
		backdrop-filter: blur(14px) saturate(160%);
		-webkit-backdrop-filter: blur(14px) saturate(160%);
		animation: paletteFadeIn 140ms cubic-bezier(0.16, 1, 0.3, 1);

		@media (max-width: 600px) {
			padding: 24px 12px;
		}
	}

	@keyframes paletteFadeIn {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	.command-palette-modal {
		display: flex;
		flex-direction: column;
		width: 100%;
		max-width: 620px;
		max-height: 540px;
		background: linear-gradient(
			180deg,
			rgba(17, 24, 39, 0.96) 0%,
			rgba(11, 15, 25, 0.98) 100%
		);
		border: 1px solid rgba(255, 255, 255, 0.12);
		border-radius: 14px;
		box-shadow:
			0 24px 70px -10px rgba(0, 0, 0, 0.85),
			0 0 0 1px rgba(255, 255, 255, 0.06);
		overflow: hidden;
		animation: paletteScaleIn 160ms cubic-bezier(0.16, 1, 0.3, 1);
	}

	@keyframes paletteScaleIn {
		from {
			transform: scale(0.97) translateY(-6px);
			opacity: 0;
		}
		to {
			transform: scale(1) translateY(0);
			opacity: 1;
		}
	}

	.palette-input-wrap {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 16px 18px;
		border-bottom: 1px solid rgba(255, 255, 255, 0.08);
		background: rgba(255, 255, 255, 0.02);

		.palette-search-icon {
			display: flex;
			align-items: center;
			justify-content: center;
			color: rgba(255, 255, 255, 0.5);
			flex-shrink: 0;

			.palette-spinner {
				width: 18px;
				height: 18px;
				border: 2px solid rgba(255, 255, 255, 0.15);
				border-top-color: #6366f1;
				border-radius: 50%;
				animation: spin 600ms linear infinite;
			}

			@keyframes spin {
				to {
					transform: rotate(360deg);
				}
			}
		}

		input {
			flex: 1;
			background: transparent;
			border: none;
			outline: none;
			color: #ffffff;
			font-size: 15.5px;
			font-family: inherit;
			font-weight: 500;
			letter-spacing: -0.01em;

			&::placeholder {
				color: rgba(255, 255, 255, 0.35);
				font-weight: 400;
			}
		}

		.palette-clear-btn {
			background: rgba(255, 255, 255, 0.08);
			border: none;
			color: rgba(255, 255, 255, 0.6);
			border-radius: 50%;
			width: 22px;
			height: 22px;
			display: flex;
			align-items: center;
			justify-content: center;
			cursor: pointer;
			transition: all 120ms ease;

			&:hover {
				background: rgba(255, 255, 255, 0.16);
				color: #ffffff;
			}
		}

		.palette-esc-badge {
			padding: 2.5px 6.5px;
			border-radius: 5px;
			background: rgba(255, 255, 255, 0.07);
			border: 1px solid rgba(255, 255, 255, 0.12);
			color: rgba(255, 255, 255, 0.5);
			font-size: 10px;
			font-weight: 700;
			font-family: inherit;
			letter-spacing: 0.04em;
			cursor: pointer;
			user-select: none;
			transition: all 120ms ease;

			&:hover {
				background: rgba(255, 255, 255, 0.14);
				color: #ffffff;
			}
		}
	}

	.palette-results {
		flex: 1;
		overflow-y: auto;
		padding: 8px 6px;
		max-height: 400px;

		&::-webkit-scrollbar {
			width: 6px;
		}
		&::-webkit-scrollbar-thumb {
			background: rgba(255, 255, 255, 0.12);
			border-radius: 3px;
		}

		.palette-empty {
			display: flex;
			flex-direction: column;
			align-items: center;
			justify-content: center;
			padding: 42px 20px;
			text-align: center;
			color: rgba(255, 255, 255, 0.4);
			gap: 8px;

			p {
				color: rgba(255, 255, 255, 0.75);
				font-size: 14.5px;
				margin: 4px 0 0;
			}

			span {
				font-size: 12.5px;
				color: rgba(255, 255, 255, 0.4);
			}
		}

		.palette-section {
			margin-bottom: 6px;

			.palette-section-title {
				padding: 8px 12px 4px;
				font-size: 11px;
				font-weight: 700;
				letter-spacing: 0.06em;
				text-transform: uppercase;
				color: rgba(255, 255, 255, 0.38);
				user-select: none;
			}
		}

		.palette-item {
			display: flex;
			align-items: center;
			gap: 12px;
			width: 100%;
			padding: 8px 12px;
			border-radius: 8px;
			border: 1px solid transparent;
			background: transparent;
			color: rgba(255, 255, 255, 0.88);
			text-align: left;
			cursor: pointer;
			transition:
				background 100ms ease,
				border-color 100ms ease,
				transform 100ms ease;

			&:hover,
			&.selected {
				background: rgba(99, 102, 241, 0.14);
				border-color: rgba(99, 102, 241, 0.35);
				color: #ffffff;
			}

			&.selected {
				box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
			}

			.item-visual {
				flex-shrink: 0;
				display: flex;
				align-items: center;
				justify-content: center;

				.item-poster {
					width: 28px;
					height: 40px;
					border-radius: 4px;
					object-fit: cover;
					background: #11141c;
					box-shadow: 0 2px 6px rgba(0, 0, 0, 0.4);
				}

				.item-icon-box {
					width: 32px;
					height: 32px;
					border-radius: 7px;
					background: rgba(255, 255, 255, 0.05);
					border: 1px solid rgba(255, 255, 255, 0.08);
					display: flex;
					align-items: center;
					justify-content: center;
					color: #c7d2fe;
				}
			}

			.item-details {
				flex: 1;
				min-width: 0;

				.item-title {
					font-size: 13.5px;
					font-weight: 600;
					line-height: 1.3;
					color: inherit;
					overflow: hidden;
					text-overflow: ellipsis;
					white-space: nowrap;
				}

				.item-subtitle {
					font-size: 12px;
					color: rgba(255, 255, 255, 0.48);
					overflow: hidden;
					text-overflow: ellipsis;
					white-space: nowrap;
					margin-top: 1px;
				}
			}

			.item-badge {
				flex-shrink: 0;
				font-size: 10px;
				font-weight: 700;
				text-transform: uppercase;
				letter-spacing: 0.04em;
				padding: 2px 6px;
				border-radius: 4px;
				background: rgba(255, 255, 255, 0.06);
				border: 1px solid rgba(255, 255, 255, 0.12);
				color: rgba(255, 255, 255, 0.7);
			}

			.item-enter-hint {
				flex-shrink: 0;
				display: flex;
				align-items: center;
				justify-content: center;
				width: 18px;
				height: 18px;
				font-size: 12px;
				font-weight: 700;
				color: #a5b4fc;
				background: rgba(99, 102, 241, 0.25);
				border-radius: 4px;
			}
		}
	}

	.palette-footer {
		display: flex;
		align-items: center;
		gap: 16px;
		padding: 10px 16px;
		background: rgba(8, 11, 18, 0.85);
		border-top: 1px solid rgba(255, 255, 255, 0.06);
		font-size: 11px;
		color: rgba(255, 255, 255, 0.45);

		.footer-hint {
			display: flex;
			align-items: center;
			gap: 5px;

			kbd {
				padding: 1px 4.5px;
				border-radius: 3.5px;
				background: rgba(255, 255, 255, 0.08);
				border: 1px solid rgba(255, 255, 255, 0.12);
				font-size: 10px;
				font-family: inherit;
				font-weight: 700;
				color: rgba(255, 255, 255, 0.75);
			}

			span {
				font-size: 11px;
			}
		}

		.footer-right {
			margin-left: auto;
			font-size: 10.5px;
			letter-spacing: 0.02em;
			color: rgba(255, 255, 255, 0.3);

			@media (max-width: 500px) {
				display: none;
			}
		}
	}
</style>
