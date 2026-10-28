<script lang="ts">
	import Icon from "@/lib/Icon.svelte";
	import DetailedMenu from "@/lib/nav/DetailedMenu.svelte";
	import SortMenu from "@/lib/nav/SortMenu.svelte";
	import FilterMenu from "./FilterMenu.svelte";
import WatchTogetherModal from "./WatchTogetherModal.svelte";
	import WatchRouletteModal from "./WatchRouletteModal.svelte";
	import { resolve } from "$app/paths";
	import { store, defaultSort, clearActiveFilters } from "@/store.svelte";
	import { onDestroy, onMount } from "svelte";
	import { t } from "@/lib/i18n";

	export type GridDensity = "compact" | "comfortable" | "editorial";

	interface Props {
		viewMode: "grid" | "list";
		density?: GridDensity;
		bulkMode: boolean;
		onViewModeChange: (mode: "grid" | "list") => void;
		onDensityChange?: (density: GridDensity) => void;
		onToggleBulkMode: () => void;
	}

	let {
		viewMode,
		density = "comfortable",
		bulkMode,
		onViewModeChange,
		onDensityChange,
		onToggleBulkMode,
	}: Props = $props();

	let detailedMenuShown = $state(false);
	let sortMenuShown = $state(false);
	let filterMenuShown = $state(false);
let watchTogetherShown = $state(false);
	let watchRouletteShown = $state(false);

	// In-library search query
	let librarySearchText = $state(store.activeFilters?.query || "");
	let searchDebounceTimeout: number;

	$effect(() => {
		librarySearchText = store.activeFilters?.query || "";
	});

	function handleLibrarySearchInput(e: Event) {
		const val = (e.target as HTMLInputElement).value;
		librarySearchText = val;
		clearTimeout(searchDebounceTimeout);
		searchDebounceTimeout = window.setTimeout(() => {
			store.activeFilters = {
				...store.activeFilters,
				query: val.trim() ? val.trim() : undefined,
			};
		}, 300);
	}

	function clearLibrarySearch() {
		librarySearchText = "";
		store.activeFilters = {
			...store.activeFilters,
			query: undefined,
		};
	}

	// Active status tabs
	let statusTabs = $derived([
		{ id: "all", label: t("status.all") },
		{ id: "watching", label: t("status.watching") },
		{ id: "planned", label: t("status.planned") },
		{ id: "finished", label: t("status.finished") },
		{ id: "hold", label: t("status.hold") },
		{ id: "dropped", label: t("status.dropped") },
		{ id: "cancelled", label: t("status.cancelled") },
	]);

	let currentStatus = $derived.by(() => {
		const s = store.activeFilters?.status;
		if (!s || s.length === 0) return "all";
		const val = s[0]?.toLowerCase();
		if (val === "watching") return "watching";
		if (val === "planned") return "planned";
		if (val === "finished") return "finished";
		if (val === "hold") return "hold";
		if (val === "dropped") return "dropped";
		if (val === "cancelled" || val === "canceled") return "cancelled";
		return "all";
	});

	function setStatusTab(tabId: string) {
		store.activeFilters = {
			...store.activeFilters,
			status: tabId === "all" ? [] : [tabId],
		};
	}

	function removeFilter(
		key:
			| "status"
			| "type"
			| "year"
			| "minRating"
			| "query"
			| "genre"
			| "productionStatus",
	) {
		if (key === "status") {
			store.activeFilters = { ...store.activeFilters, status: [] };
		} else if (key === "type") {
			store.activeFilters = { ...store.activeFilters, type: [] };
		} else if (key === "year") {
			store.activeFilters = { ...store.activeFilters, year: undefined };
		} else if (key === "minRating") {
			store.activeFilters = { ...store.activeFilters, minRating: undefined };
		} else if (key === "genre") {
			store.activeFilters = { ...store.activeFilters, genre: undefined };
		} else if (key === "productionStatus") {
			store.activeFilters = {
				...store.activeFilters,
				productionStatus: undefined,
			};
		} else if (key === "query") {
			librarySearchText = "";
			store.activeFilters = { ...store.activeFilters, query: undefined };
		}
	}

	function handleClearAll() {
		librarySearchText = "";
		clearActiveFilters();
	}

	function handleDocumentClick(e: MouseEvent) {
		const target = e.target as HTMLElement | null;
		if (!target?.closest(".control-menu-wrap")) {
			detailedMenuShown = false;
			sortMenuShown = false;
			filterMenuShown = false;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === "Escape") {
			detailedMenuShown = false;
			sortMenuShown = false;
			filterMenuShown = false;
		}
	}

	onMount(() => {
		window.addEventListener("click", handleDocumentClick);
		window.addEventListener("keydown", handleKeydown);
	});

	onDestroy(() => {
		window.removeEventListener("click", handleDocumentClick);
		window.removeEventListener("keydown", handleKeydown);
		clearTimeout(searchDebounceTimeout);
	});
</script>

<div
	class="sticky-subnav-wrapper"
	class:has-open-menu={sortMenuShown || filterMenuShown || detailedMenuShown}
>
	<header
		class="unified-command-bar"
		class:has-open-menu={sortMenuShown || filterMenuShown || detailedMenuShown}
	>
		<!-- Left Side: Status Segmented Tabs -->
		<nav class="status-tabs-container" aria-label="Media status tabs">
			<div class="status-tabs-track" role="tablist">
				{#each statusTabs as tab (tab.id)}
					<button
						type="button"
						role="tab"
						class="status-tab-btn"
						class:active={currentStatus === tab.id}
						onclick={() => setStatusTab(tab.id)}
						aria-selected={currentStatus === tab.id}
					>
						{tab.label}
					</button>
				{/each}
			</div>
		</nav>

		<!-- Right Side: Search, Filters, Sort, Details, Bulk, Grid/List -->
		<div class="actions-cluster">
			<!-- In-Library Search Input -->
			<div class="library-search-box">
				<Icon i="search" wh={14} />
				<input
					type="text"
					placeholder={t("dashboard.search_placeholder")}
					value={librarySearchText}
					oninput={handleLibrarySearchInput}
					aria-label="Filter library by title"
				/>
				{#if librarySearchText}
					<button
						type="button"
						class="plain clear-search-icon"
						onclick={clearLibrarySearch}
						aria-label="Clear search"
					>
						<Icon i="close" wh={12} />
					</button>
				{/if}
			</div>

			<!-- Filters Menu Popover -->
			<div class="control-menu-wrap" class:menu-open={filterMenuShown}>
				<button
					type="button"
					class="icon-toggle"
					class:active={filterMenuShown || store.hasSecondaryFilters}
					onclick={(e) => {
						e.stopPropagation();
						sortMenuShown = false;
						detailedMenuShown = false;
						filterMenuShown = !filterMenuShown;
					}}
					title="Secondary Filters (Type, Year, Rating, Genre)"
					aria-label="Open secondary filters"
				>
					<Icon i="filter" wh={14} />
					<span class="toggle-label">{t("dashboard.filter")}</span>
					{#if store.secondaryFilterCount > 0}
						<span class="filter-count-badge">{store.secondaryFilterCount}</span>
					{/if}
				</button>
				{#if filterMenuShown}
					<FilterMenu />
				{/if}
			</div>

			<!-- Sort Menu Popover -->
			<div class="control-menu-wrap" class:menu-open={sortMenuShown}>
				<button
					type="button"
					class="icon-toggle"
					class:active={sortMenuShown}
					onclick={(e) => {
						e.stopPropagation();
						detailedMenuShown = false;
						filterMenuShown = false;
						sortMenuShown = !sortMenuShown;
					}}
					title="Sort Order"
					aria-label="Sort order"
				>
					<Icon i="sort" wh={14} />
					<span class="toggle-label">{t("dashboard.sort")}</span>
					{#if store.activeSort?.length === 2 && store.activeSort[1] && JSON.stringify(store.activeSort) !== JSON.stringify(defaultSort)}
						<span class="sort-active-dot"></span>
					{/if}
				</button>
				{#if sortMenuShown}
					<SortMenu />
				{/if}
			</div>

			<!-- Detailed View Toggle -->
			<div class="control-menu-wrap" class:menu-open={detailedMenuShown}>
				<button
					type="button"
					class="icon-toggle"
					class:active={detailedMenuShown}
					onclick={(e) => {
						e.stopPropagation();
						sortMenuShown = false;
						filterMenuShown = false;
						detailedMenuShown = !detailedMenuShown;
					}}
					title="Detailed View Options"
					aria-label="Detailed view options"
				>
					<Icon i="eye" wh={14} />
					<span class="toggle-label">{t("dashboard.details")}</span>
				</button>
				{#if detailedMenuShown}
					<DetailedMenu />
				{/if}
			</div>

			<div class="toolbar-separator"></div>

			<!-- Bulk Selection Toggle -->
			<button
				type="button"
				class="icon-toggle bulk-btn"
				class:active={bulkMode}
				onclick={onToggleBulkMode}
				title={bulkMode ? "Exit Bulk Actions" : "Bulk Selection Mode"}
				aria-label="Bulk selection mode"
			>
				<Icon i="check" wh={14} />
				<span class="toggle-label">{t("dashboard.bulk")}</span>
			</button>

			<!-- Watch Together Button -->
			<button
				type="button"
				class="icon-toggle watch-together-btn"
				class:active={watchTogetherShown}
				onclick={() => {
					detailedMenuShown = false;
					filterMenuShown = false;
					sortMenuShown = false;
					watchTogetherShown = true;
				}}
				title="Ver em Conjunto (Cruzar Watchlists com Amigos/Parceiros)"
				aria-label="Ver em Conjunto"
			>
				<span style="font-size: 13px;">🍿</span>
				<span class="toggle-label">Ver em Conjunto</span>
			</button>

			<!-- Sagas & Franchises Button -->
			<a
				href={resolve("/franchise")}
				class="icon-toggle sagas-btn"
				title="Explorar Sagas, Coleções TMDB e Linhas Cronológicas"
				aria-label="Sagas"
			>
				<Icon i="reel" wh={14} />
				<span class="toggle-label">Sagas</span>
			</a>

			<!-- Smart Roulette Button -->
			<button
				type="button"
				class="icon-toggle roulette-btn"
				class:active={watchRouletteShown}
				onclick={() => {
					detailedMenuShown = false;
					filterMenuShown = false;
					sortMenuShown = false;
					watchRouletteShown = true;
				}}
				title="Roleta Inteligente: Sorteia da tua Watchlist por duração e mood"
				aria-label="Roleta O Que Ver"
			>
				<span style="font-size: 13px;">🎲</span>
				<span class="toggle-label">O Que Ver?</span>
			</button>

			<div class="toolbar-separator"></div>

			<!-- View Switcher & Grid Density (Compact / Comfortable / Editorial / List) -->
			<div
				class="view-mode-group"
				role="group"
				aria-label="Visual density and layout mode"
			>
				<button
					type="button"
					class="view-btn"
					class:active={viewMode === "grid" && density === "compact"}
					onclick={() => {
						onDensityChange?.("compact");
						if (viewMode !== "grid") onViewModeChange("grid");
					}}
					title="Compact Grid (Dense Wall)"
					aria-label="Compact grid (dense wall)"
					aria-pressed={viewMode === "grid" && density === "compact"}
				>
					<svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
						<rect x="1" y="1" width="3.6" height="3.6" rx="0.8" />
						<rect x="6.2" y="1" width="3.6" height="3.6" rx="0.8" />
						<rect x="11.4" y="1" width="3.6" height="3.6" rx="0.8" />
						<rect x="1" y="6.2" width="3.6" height="3.6" rx="0.8" />
						<rect x="6.2" y="6.2" width="3.6" height="3.6" rx="0.8" />
						<rect x="11.4" y="6.2" width="3.6" height="3.6" rx="0.8" />
						<rect x="1" y="11.4" width="3.6" height="3.6" rx="0.8" />
						<rect x="6.2" y="11.4" width="3.6" height="3.6" rx="0.8" />
						<rect x="11.4" y="11.4" width="3.6" height="3.6" rx="0.8" />
					</svg>
				</button>
				<button
					type="button"
					class="view-btn"
					class:active={viewMode === "grid" && density === "comfortable"}
					onclick={() => {
						onDensityChange?.("comfortable");
						if (viewMode !== "grid") onViewModeChange("grid");
					}}
					title="Comfortable Grid (Standard)"
					aria-label="Comfortable grid (standard)"
					aria-pressed={viewMode === "grid" && density === "comfortable"}
				>
					<svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
						<rect x="1" y="1" width="6" height="6" rx="1.2" />
						<rect x="9" y="1" width="6" height="6" rx="1.2" />
						<rect x="1" y="9" width="6" height="6" rx="1.2" />
						<rect x="9" y="9" width="6" height="6" rx="1.2" />
					</svg>
				</button>
				<button
					type="button"
					class="view-btn"
					class:active={viewMode === "grid" && density === "editorial"}
					onclick={() => {
						onDensityChange?.("editorial");
						if (viewMode !== "grid") onViewModeChange("grid");
					}}
					title="Editorial Grid (Large Showcase)"
					aria-label="Editorial grid (large showcase)"
					aria-pressed={viewMode === "grid" && density === "editorial"}
				>
					<svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
						<rect x="1" y="1" width="6.2" height="14" rx="1.5" />
						<rect x="8.8" y="1" width="6.2" height="14" rx="1.5" />
					</svg>
				</button>
				<div class="inner-divider"></div>
				<button
					type="button"
					class="view-btn"
					class:active={viewMode === "list"}
					onclick={() => onViewModeChange("list")}
					title="Table / List View"
					aria-label="Table / list view"
					aria-pressed={viewMode === "list"}
				>
					<svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
						<rect x="1" y="2" width="14" height="2.5" rx="1" />
						<rect x="1" y="6.75" width="14" height="2.5" rx="1" />
						<rect x="1" y="11.5" width="14" height="2.5" rx="1" />
					</svg>
				</button>
			</div>
		</div>
	</header>

	<!-- Active Secondary Filter Chips (Only rendered when secondary filters or search are active) -->
	{#if store.hasSecondaryFilters || store.activeFilters?.query}
		<div class="active-chips-dock">
			<span class="chips-title">Active:</span>

			{#if store.activeFilters?.type && store.activeFilters.type.length > 0}
				{#each store.activeFilters.type as t (t)}
					<span class="filter-chip">
						Type: {t === "tv" ? "Shows" : t === "movie" ? "Movies" : "Games"}
						<button
							type="button"
							class="plain chip-x-btn"
							onclick={() => removeFilter("type")}
							aria-label="Remove type filter"
						>
							<Icon i="close" wh={11} />
						</button>
					</span>
				{/each}
			{/if}

			{#if store.activeFilters?.genre}
				<span class="filter-chip">
					Genre: {store.activeFilters.genre}
					<button
						type="button"
						class="plain chip-x-btn"
						onclick={() => removeFilter("genre")}
						aria-label="Remove genre filter"
					>
						<Icon i="close" wh={11} />
					</button>
				</span>
			{/if}

			{#if store.activeFilters?.year}
				<span class="filter-chip">
					Year: {store.activeFilters.year}
					<button
						type="button"
						class="plain chip-x-btn"
						onclick={() => removeFilter("year")}
						aria-label="Remove year filter"
					>
						<Icon i="close" wh={11} />
					</button>
				</span>
			{/if}

			{#if store.activeFilters?.minRating}
				<span class="filter-chip">
					★ {store.activeFilters.minRating}+
					<button
						type="button"
						class="plain chip-x-btn"
						onclick={() => removeFilter("minRating")}
						aria-label="Remove rating filter"
					>
						<Icon i="close" wh={11} />
					</button>
				</span>
			{/if}

			{#if store.activeFilters?.productionStatus}
				<span class="filter-chip production-chip">
					Status: {store.activeFilters.productionStatus === "canceled"
						? "Cancelada pelos Produtores"
						: store.activeFilters.productionStatus === "returning"
							? "Em Exibição"
							: "Terminada"}
					<button
						type="button"
						class="plain chip-x-btn"
						onclick={() => removeFilter("productionStatus")}
						aria-label="Remove production status filter"
					>
						<Icon i="close" wh={11} />
					</button>
				</span>
			{/if}

			{#if store.activeFilters?.query}
				<span class="filter-chip">
					"{store.activeFilters.query}"
					<button
						type="button"
						class="plain chip-x-btn"
						onclick={() => removeFilter("query")}
						aria-label="Remove query filter"
					>
						<Icon i="close" wh={11} />
					</button>
				</span>
			{/if}

			<button
				type="button"
				class="plain clear-all-dock-btn"
				onclick={handleClearAll}
			>
				Clear all
			</button>
		</div>
	{/if}

	{#if watchTogetherShown}
		<WatchTogetherModal onClose={() => (watchTogetherShown = false)} />
	{/if}

	{#if watchRouletteShown}
		<WatchRouletteModal onClose={() => (watchRouletteShown = false)} />
	{/if}
</div>

<style lang="scss">
	.sticky-subnav-wrapper {
		position: sticky;
		top: 60px;
		z-index: 500;
		margin-bottom: 22px;

		&.has-open-menu {
			z-index: 900;
		}

		@media screen and (max-width: 768px) {
			top: 52px;
			margin-bottom: 16px;
		}
	}

	.unified-command-bar {
		position: relative;
		z-index: 50;
		display: flex;
		flex-flow: row nowrap;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		margin-bottom: 0;
		padding: 7px 12px;
		border-radius: 14px;
		background: rgba(13, 17, 26, 0.88);
		border: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
		box-shadow:
			0 12px 36px -4px rgba(0, 0, 0, 0.6),
			0 0 0 1px rgba(255, 255, 255, 0.05);
		backdrop-filter: blur(24px) saturate(180%);
		-webkit-backdrop-filter: blur(24px) saturate(180%);
		box-sizing: border-box;
		width: 100%;

		&.has-open-menu {
			z-index: 100;
		}

		@media screen and (max-width: 960px) {
			flex-direction: column;
			align-items: stretch;
			gap: 10px;
			padding: 8px;
		}

		/* Status Tabs */
		.status-tabs-container {
			display: flex;
			align-items: center;
			overflow-x: auto;
			scrollbar-width: none;
			-ms-overflow-style: none;
			flex-shrink: 1;

			&::-webkit-scrollbar {
				display: none;
			}

			.status-tabs-track {
				display: inline-flex;
				align-items: center;
				gap: 3px;
				padding: 3px;
				border-radius: 10px;
				background: rgba(0, 0, 0, 0.35);
				border: 1px solid var(--card-border, rgba(255, 255, 255, 0.06));

				.status-tab-btn {
					padding: 6px 12px;
					border-radius: 7px;
					border: none;
					background: transparent;
					color: $text-color-accent;
					font-size: 12.5px;
					font-weight: 500;
					cursor: pointer;
					white-space: nowrap;
					width: auto !important;
					flex-shrink: 0;
					transition: all 140ms ease;

					&:hover {
						color: #ffffff;
						background-color: rgba(255, 255, 255, 0.06);
					}

					&.active {
						color: #fbbf24;
						background-color: rgba(245, 158, 11, 0.16);
						border: 1px solid rgba(245, 158, 11, 0.35);
						font-weight: 600;
						box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
					}
				}
			}
		}

		/* Actions Cluster */
		.actions-cluster {
			display: flex;
			align-items: center;
			gap: 6px;
			flex-wrap: nowrap;
			flex-shrink: 0;

			@media screen and (max-width: 960px) {
				justify-content: flex-end;
				width: 100%;
				flex-wrap: wrap;
				overflow: visible;
			}

			/* In-Library Search Input */
			.library-search-box {
				display: flex;
				align-items: center;
				position: relative;
				width: 155px;
				height: 34px;
				flex-shrink: 0;
				transition: width 180ms ease;

				@media screen and (max-width: 600px) {
					width: 120px;
				}

				&:focus-within {
					width: 195px;

					@media screen and (max-width: 600px) {
						width: 150px;
					}
				}

				:global(svg) {
					position: absolute;
					left: 9px;
					color: $placeholder-color;
					pointer-events: none;
				}

				input {
					width: 100%;
					height: 34px;
					padding: 0 26px 0 28px;
					border-radius: 8px;
					border: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
					background-color: rgba(0, 0, 0, 0.3);
					color: $text-color;
					font-size: 12px;
					font-weight: 500;
					outline: none;
					transition: all 140ms ease;

					&::placeholder {
						color: $placeholder-color;
					}

					&:focus {
						border-color: #f59e0b;
						background-color: rgba(0, 0, 0, 0.5);
						box-shadow: 0 0 0 2px rgba(245, 158, 11, 0.15);
					}
				}

				.clear-search-icon {
					position: absolute;
					right: 7px;
					display: flex;
					align-items: center;
					justify-content: center;
					color: $placeholder-color;
					cursor: pointer;
					padding: 2px;
					background: transparent;
					border: none;
					width: auto !important;

					&:hover {
						color: $text-color;
					}
				}
			}

			.control-menu-wrap {
				position: relative;
				display: inline-flex;
				align-items: center;
				flex-shrink: 0;

				&.menu-open {
					z-index: 110;
				}
			}

			.icon-toggle {
				display: inline-flex;
				align-items: center;
				justify-content: center;
				gap: 5px;
				height: 34px;
				padding: 0 10px;
				border-radius: 8px;
				border: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
				background-color: rgba(255, 255, 255, 0.035);
				color: $text-color-accent;
				cursor: pointer;
				font-size: 12px;
				font-weight: 600;
				width: auto !important;
				max-width: none !important;
				flex-shrink: 0;
				white-space: nowrap;
				transition: all 140ms ease;

				&:hover {
					color: #ffffff;
					background-color: rgba(255, 255, 255, 0.07);
					border-color: rgba(255, 255, 255, 0.15);
				}

				&.active {
					color: #fbbf24;
					background-color: rgba(245, 158, 11, 0.14);
					border-color: rgba(245, 158, 11, 0.4);
					box-shadow: 0 0 10px rgba(245, 158, 11, 0.12);
				}

				.toggle-label {
					font-size: 12px;

					@media screen and (max-width: 600px) {
						display: none;
					}
				}

				.filter-count-badge {
					display: inline-flex;
					align-items: center;
					justify-content: center;
					min-width: 16px;
					height: 16px;
					padding: 0 4px;
					border-radius: 9999px;
					background-color: #f59e0b;
					color: #0b0e14;
					font-size: 10px;
					font-weight: 700;
					line-height: 1;
				}

				.sort-active-dot {
					width: 6px;
					height: 6px;
					background-color: #f59e0b;
					border-radius: 50%;
					margin-left: 2px;
					box-shadow: 0 0 6px #f59e0b;
				}
			}

			.toolbar-separator {
				width: 1px;
				height: 18px;
				background-color: var(--card-border, rgba(255, 255, 255, 0.1));
				margin: 0 2px;
				flex-shrink: 0;

				@media screen and (max-width: 600px) {
					display: none;
				}
			}

			/* View Mode & Density Switcher */
			.view-mode-group {
				display: inline-flex;
				align-items: center;
				padding: 2px;
				height: 34px;
				border-radius: 8px;
				background-color: rgba(0, 0, 0, 0.35);
				border: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
				flex-shrink: 0;
				gap: 1px;

				.inner-divider {
					width: 1px;
					height: 16px;
					background-color: var(--card-border, rgba(255, 255, 255, 0.12));
					margin: 0 2px;
					flex-shrink: 0;
				}

				.view-btn {
					display: flex;
					align-items: center;
					justify-content: center;
					height: 28px;
					padding: 0 7px;
					border-radius: 6px;
					border: none;
					background: transparent;
					color: $text-color-accent;
					cursor: pointer;
					width: auto !important;
					transition: all 130ms ease;

					&:hover {
						color: #ffffff;
						background-color: rgba(255, 255, 255, 0.06);
					}

					&.active {
						color: #fbbf24;
						background-color: rgba(245, 158, 11, 0.16);
						border: 1px solid rgba(245, 158, 11, 0.3);
						box-shadow: 0 1px 3px rgba(0, 0, 0, 0.35);
					}
				}
			}
		}
	}

	/* Active Filter Chips Dock */
	.active-chips-dock {
		position: relative;
		z-index: 2;
		display: flex;
		flex-flow: row wrap;
		align-items: center;
		gap: 6px;
		margin-top: 6px;
		margin-bottom: 0;
		padding: 6px 12px;
		border-radius: 10px;
		background: rgba(13, 17, 26, 0.9);
		border: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
		backdrop-filter: blur(16px);
		-webkit-backdrop-filter: blur(16px);
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
		font-size: 12px;

		.chips-title {
			font-weight: 600;
			color: $text-color-accent;
			font-size: 11px;
			text-transform: uppercase;
			letter-spacing: 0.04em;
			margin-right: 2px;
		}

		.filter-chip {
			display: inline-flex;
			align-items: center;
			gap: 5px;
			padding: 2px 8px;
			border-radius: 9999px;
			background-color: rgba(245, 158, 11, 0.1);
			border: 1px solid rgba(245, 158, 11, 0.25);
			color: #fef08a;
			font-weight: 500;
			font-size: 11.5px;

			.chip-x-btn {
				display: flex;
				align-items: center;
				justify-content: center;
				color: rgba(254, 240, 138, 0.7);
				cursor: pointer;
				padding: 0;
				background: transparent;
				border: none;
				width: auto !important;
				transition: color 130ms ease;

				&:hover {
					color: #ffffff;
				}
			}
		}

		.clear-all-dock-btn {
			font-size: 11.5px;
			color: #f59e0b;
			font-weight: 600;
			cursor: pointer;
			margin-left: 4px;
			background: transparent;
			border: none;
			width: auto !important;
			transition: opacity 130ms ease;

			&:hover {
				text-decoration: underline;
				opacity: 0.85;
			}
		}
	}
</style>
