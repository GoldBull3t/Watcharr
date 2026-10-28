<script lang="ts">
	import Error from "@/lib/Error.svelte";
	import Icon from "@/lib/Icon.svelte";
	import Poster from "@/lib/poster/Poster.svelte";
	import PosterList, { type GridDensity } from "@/lib/poster/PosterList.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import DashboardControls from "@/lib/dashboard/DashboardControls.svelte";
	import HeroSpotlightCarousel from "@/lib/dashboard/HeroSpotlightCarousel.svelte";
	import ContinueWatchingRail from "@/lib/dashboard/ContinueWatchingRail.svelte";
	import UpNextRail from "@/lib/dashboard/UpNextRail.svelte";
	import BulkActionBar from "@/lib/dashboard/BulkActionBar.svelte";
	import MediaTable from "@/lib/dashboard/MediaTable.svelte";
	import PosterGridSkeleton from "@/lib/dashboard/PosterGridSkeleton.svelte";
	import EmptyLibrary from "@/lib/dashboard/EmptyLibrary.svelte";
	import { req, updateWatched, removeWatched } from "@/lib/util/api";
	import infScroll from "@/lib/util/infScroll";
	import paginatedLoader from "@/lib/util/paginatedLoader.svelte";
	import { store } from "@/store.svelte";
	import {
		MediaTypeE,
		type Media,
		type PaginationResponse,
		type WatchedStatus,
		type UpNextItem,
	} from "@/types";
	import { notify, notifyUndo } from "@/lib/util/notify";
	import { onDestroy, onMount, untrack } from "svelte";
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import { getMediaLink } from "@/lib/dashboard/mediaHelpers";
	import { t } from "@/lib/i18n";

	const scroll = infScroll({ callback: onScrollToBottom });
	const dataLoader = paginatedLoader<Media, undefined>(load);

	// View mode: 'grid' (cinematic cards) vs 'list' (compact table/rows)
	let viewMode: "grid" | "list" = $state("grid");

	// Grid density: 'compact' (dense wall) vs 'comfortable' (standard) vs 'editorial' (large cards)
	let gridDensity: GridDensity = $state("comfortable");

	// Bulk actions mode & selection
	let bulkMode: boolean = $state(false);
	let selectedIds: number[] = $state([]);
	let selectedAllInLibrary: boolean = $state(false);
	let lastClickedIndex: number | undefined = $state();

	// Lasso (Marquee Drag) selection state
	let isLassoing = $state(false);
	let lassoStartX = $state(0);
	let lassoStartY = $state(0);
	let lassoCurrentX = $state(0);
	let lassoCurrentY = $state(0);
	let initialSelectedIds: number[] = [];

	let lassoBox = $derived({
		left: Math.min(lassoStartX, lassoCurrentX),
		top: Math.min(lassoStartY, lassoCurrentY),
		width: Math.abs(lassoCurrentX - lassoStartX),
		height: Math.abs(lassoCurrentY - lassoStartY),
	});

	// Dedicated continue watching state
	let continueWatchingList: Media[] = $state([]);

	// Active status tab pill
	let activeTab:
		| "all"
		| "watching"
		| "planned"
		| "finished"
		| "hold"
		| "dropped"
		| "cancelled" = $derived.by(() => {
		const statuses = store.activeFilters?.status;
		if (!statuses || statuses.length === 0) return "all";
		if (statuses.includes("watching")) return "watching";
		if (statuses.includes("planned")) return "planned";
		if (statuses.includes("finished")) return "finished";
		if (statuses.includes("hold")) return "hold";
		if (statuses.includes("dropped")) return "dropped";
		if (statuses.includes("cancelled") || statuses.includes("canceled"))
			return "cancelled";
		return "all";
	});

	// Display items for continue watching
	let displayContinueWatching = $derived.by(() => {
		if (continueWatchingList.length > 0) return continueWatchingList;
		if (!dataLoader.state.data) return [];
		return dataLoader.state.data.filter(
			(w) => w && w.watched && w.watched.status === "WATCHING",
		);
	});

	let nextLoadParams: {
		page: number;
		[x: string]: unknown;
	} = $derived({
		page: dataLoader.state.page + 1,
		...store.sortAndFiltersForQueryParams,
	});

	let isInitialLoad = $state(true);

	async function load(signal: AbortSignal) {
		console.debug("load: loadParams:", nextLoadParams);
		if (nextLoadParams.page === dataLoader.state.page) {
			console.warn("load: Already on this page, not loading it again!");
			return;
		}
		try {
			const r = await req.get<PaginationResponse<Media, undefined>>(
				`/watched`,
				{
					params: nextLoadParams,
					signal,
				},
			);
			scroll.dataLoaded();
			return r;
		} finally {
			isInitialLoad = false;
		}
	}

	let upNextList: UpNextItem[] = $state([]);

	async function loadUpNext() {
		try {
			const res = await req.get<UpNextItem[]>("/watched/up-next");
			if (res && Array.isArray(res)) {
				upNextList = res;
			}
		} catch (err) {
			console.warn("loadUpNext: Failed to fetch /watched/up-next:", err);
		}
	}

	async function loadContinueWatching() {
		try {
			const res = await req.get<Media[]>("/watched/continue");
			if (res && Array.isArray(res)) {
				continueWatchingList = res;
			}
		} catch (err) {
			console.warn(
				"loadContinueWatching: Failed to fetch /watched/continue:",
				err,
			);
		}
		loadUpNext();
	}

	async function onScrollToBottom() {
		if (dataLoader.state.reqLoadError) {
			return;
		}
		dataLoader.runFn();
	}

	function toggleSelectItem(id: number) {
		selectedAllInLibrary = false;
		if (selectedIds.includes(id)) {
			selectedIds = selectedIds.filter((i) => i !== id);
		} else {
			selectedIds = [...selectedIds, id];
		}
	}

	function selectAllLoaded() {
		if (!dataLoader.state.data) return;
		const allIds = dataLoader.state.data
			.filter((w) => w?.watched?.id)
			.map((w) => w.watched!.id);
		selectedIds = allIds;
		selectedAllInLibrary = false;
	}

	function selectAllEverywhere() {
		selectAllLoaded();
		selectedAllInLibrary = true;
	}

	function clearSelection() {
		selectedIds = [];
		selectedAllInLibrary = false;
		lastClickedIndex = undefined;
	}

	function handlePosterClick(
		watchedId: number | undefined,
		index: number,
		e?: MouseEvent,
	) {
		if (!watchedId) return;

		// 1. Shift + Click -> Range selection
		if (e?.shiftKey && lastClickedIndex !== undefined) {
			e.preventDefault();
			const start = Math.min(lastClickedIndex, index);
			const end = Math.max(lastClickedIndex, index);
			const rangeItems = (dataLoader.state.data ?? []).slice(start, end + 1);
			const rangeIds: number[] = [];
			for (const item of rangeItems) {
				if (item?.watched?.id) {
					rangeIds.push(item.watched.id);
				}
			}

			const merged = new Set(selectedIds);
			for (const id of rangeIds) {
				merged.add(id);
			}
			selectedIds = Array.from(merged);
			bulkMode = true;
			lastClickedIndex = index;
			return;
		}

		// 2. Ctrl / Cmd + Click or active bulkMode -> Toggle selection
		if (bulkMode || e?.ctrlKey || e?.metaKey) {
			e?.preventDefault();
			if (selectedIds.includes(watchedId)) {
				selectedIds = selectedIds.filter((id) => id !== watchedId);
			} else {
				selectedIds = [...selectedIds, watchedId];
			}
			lastClickedIndex = index;
			bulkMode = selectedIds.length > 0;
			return;
		}

		// 3. Normal Click -> Navigation
		const link = getMediaLink(dataLoader.state.data[index]);
		if (link) {
			goto(resolve(link));
		}
	}

	function handleLassoPointerDown(e: PointerEvent) {
		if (e.button !== 0) return;
		const target = e.target as HTMLElement;
		if (
			target.closest(
				"button, a, input, select, textarea, .buttons, .dock-shell, .rating-wrap, .ratings-dropdown, .dashboard-controls, .hero-carousel-container, .cw-rail-wrap",
			)
		) {
			return;
		}

		const onPoster = target.closest(".poster-body");
		if (onPoster && !e.shiftKey) {
			return;
		}

		lassoStartX = e.clientX;
		lassoStartY = e.clientY;
		lassoCurrentX = e.clientX;
		lassoCurrentY = e.clientY;
		isLassoing = true;
		try {
			(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
		} catch {}

		if (!e.shiftKey && !e.ctrlKey && !e.metaKey && !bulkMode) {
			initialSelectedIds = [];
		} else {
			initialSelectedIds = [...selectedIds];
		}
	}

	function handleLassoPointerMove(e: PointerEvent) {
		if (!isLassoing) return;
		lassoCurrentX = e.clientX;
		lassoCurrentY = e.clientY;

		const box = lassoBox;
		if (box.width < 6 && box.height < 6) return;

		const posterEls = document.querySelectorAll<HTMLElement>(
			"li[data-watched-id]",
		);
		const touchedIds = new Set<number>(initialSelectedIds);

		for (const el of posterEls) {
			const idStr = el.dataset.watchedId;
			if (!idStr) continue;
			const id = Number(idStr);
			if (!id) continue;

			const rect = el.getBoundingClientRect();
			const intersects = !(
				rect.right < box.left ||
				rect.left > box.left + box.width ||
				rect.bottom < box.top ||
				rect.top > box.top + box.height
			);

			if (intersects) {
				touchedIds.add(id);
			}
		}

		selectedIds = Array.from(touchedIds);
		if (selectedIds.length > 0) {
			bulkMode = true;
		}
	}

	function handleLassoPointerUp(e: PointerEvent) {
		if (isLassoing) {
			isLassoing = false;
			try {
				(e.currentTarget as HTMLElement).releasePointerCapture(e.pointerId);
			} catch {}
		}
	}

	function handleWindowKeyDown(e: KeyboardEvent) {
		if (e.key === "Escape") {
			if (selectedIds.length > 0 || bulkMode) {
				selectedIds = [];
				selectedAllInLibrary = false;
				bulkMode = false;
				lastClickedIndex = undefined;
			}
		} else if ((e.ctrlKey || e.metaKey) && (e.key === "a" || e.key === "A")) {
			const activeEl = document.activeElement;
			if (
				activeEl &&
				(activeEl.tagName === "INPUT" ||
					activeEl.tagName === "TEXTAREA" ||
					(activeEl as HTMLElement).isContentEditable)
			) {
				return;
			}
			if (dataLoader.state.data?.length > 0) {
				e.preventDefault();
				selectAllLoaded();
				bulkMode = true;
			}
		}
	}

	async function bulkUpdateStatus(newStatus: WatchedStatus) {
		const isAll = selectedAllInLibrary;
		if (selectedIds.length === 0 && !isAll) return;

		const idsToUpdate = [...selectedIds];
		const count = isAll
			? dataLoader.state.totalResults || idsToUpdate.length
			: idsToUpdate.length;

		selectedIds = [];
		selectedAllInLibrary = false;
		bulkMode = false;

		// Optimistic UI update
		if (dataLoader.state.data) {
			const idSet = new Set(idsToUpdate);
			for (const w of dataLoader.state.data) {
				if (w?.watched && (isAll || idSet.has(w.watched.id))) {
					w.watched.status = newStatus;
				}
			}
		}

		const nid = notify({
			text: t("dashboard.bulk_updating", { count }),
			type: "loading",
		});

		try {
			await req.put("/watched/bulk", {
				ids: isAll ? undefined : idsToUpdate,
				all: isAll,
				status: newStatus,
			});
			notify({
				id: nid,
				text: t("dashboard.bulk_updated_status", { count, status: newStatus }),
				type: "success",
			});
		} catch (err) {
			console.error("bulkUpdateStatus failed:", err);
			notify({
				id: nid,
				text: t("dashboard.bulk_error"),
				type: "error",
			});
		} finally {
			dataLoader.reset();
			dataLoader.runFn();
			loadContinueWatching();
		}
	}

	async function bulkUpdateRating(newRating: number) {
		const isAll = selectedAllInLibrary;
		if (selectedIds.length === 0 && !isAll) return;

		const idsToUpdate = [...selectedIds];
		const count = isAll
			? dataLoader.state.totalResults || idsToUpdate.length
			: idsToUpdate.length;

		selectedIds = [];
		selectedAllInLibrary = false;
		bulkMode = false;

		// Optimistic UI update
		if (dataLoader.state.data) {
			const idSet = new Set(idsToUpdate);
			for (const w of dataLoader.state.data) {
				if (w?.watched && (isAll || idSet.has(w.watched.id))) {
					w.watched.rating = newRating;
				}
			}
		}

		const nid = notify({
			text: t("dashboard.bulk_rating", { count }),
			type: "loading",
		});

		try {
			await req.put("/watched/bulk", {
				ids: isAll ? undefined : idsToUpdate,
				all: isAll,
				rating: newRating,
			});
			notify({
				id: nid,
				text: t("dashboard.bulk_rated", { count }),
				type: "success",
			});
		} catch (err) {
			console.error("bulkUpdateRating failed:", err);
			notify({
				id: nid,
				text: t("dashboard.bulk_error"),
				type: "error",
			});
		} finally {
			dataLoader.reset();
			dataLoader.runFn();
			loadContinueWatching();
		}
	}

	async function bulkDeleteItems() {
		const isAll = selectedAllInLibrary;
		if (selectedIds.length === 0 && !isAll) return;

		const idsToRemove = [...selectedIds];
		const count = isAll
			? dataLoader.state.totalResults || idsToRemove.length
			: idsToRemove.length;

		// 1. Clear selection and exit bulk mode immediately
		selectedIds = [];
		selectedAllInLibrary = false;
		bulkMode = false;

		// 2. Optimistic UI update: Remove items instantly from the screen
		if (isAll) {
			dataLoader.state.data = [];
		} else {
			const idSet = new Set(idsToRemove);
			if (dataLoader.state.data) {
				dataLoader.state.data = dataLoader.state.data.filter(
					(w) => !w?.watched?.id || !idSet.has(w.watched.id),
				);
			}
		}

		// 3. Immediate progress notification
		const nid = notify({
			text: t("dashboard.bulk_removing", { count }),
			type: "loading",
		});

		try {
			if (isAll) {
				await req.post("/watched/bulk-delete", { all: true, hard: true });
			} else {
				await req.post("/watched/bulk-delete", { ids: idsToRemove });
			}
			notify({
				id: nid,
				text: t("dashboard.bulk_removed", { count }),
				type: "success",
			});
		} catch (err) {
			console.error("bulkDeleteItems failed:", err);
			notify({
				id: nid,
				text: t("dashboard.bulk_error"),
				type: "error",
			});
		} finally {
			dataLoader.reset();
			dataLoader.runFn();
			loadContinueWatching();
		}
	}

	$effect(() => {
		if (store.sortAndFiltersForQueryParams) {
			untrack(() => {
				dataLoader.reset();
				dataLoader.runFn();
			});
		}
	});

	onMount(() => {
		const savedViewMode = localStorage.getItem("watcharr_view_mode") as
			"grid" | "list";
		if (savedViewMode === "grid" || savedViewMode === "list") {
			viewMode = savedViewMode;
		}
		const savedDensity = localStorage.getItem(
			"watcharr_grid_density",
		) as GridDensity;
		if (
			savedDensity === "compact" ||
			savedDensity === "comfortable" ||
			savedDensity === "editorial"
		) {
			gridDensity = savedDensity;
		}
		loadContinueWatching();
	});

	function toggleViewMode(mode: "grid" | "list") {
		viewMode = mode;
		localStorage.setItem("watcharr_view_mode", mode);
	}

	function changeDensity(newDensity: GridDensity) {
		gridDensity = newDensity;
		localStorage.setItem("watcharr_grid_density", newDensity);
	}

	onDestroy(() => {
		scroll.destroy();
		dataLoader.abortReq("page destroyed");
	});
</script>

<svelte:head>
	<title>Watched List - Watcharr</title>
</svelte:head>

<svelte:window onkeydown={handleWindowKeyDown} />

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
	class="dashboard-container"
	onpointerdown={handleLassoPointerDown}
	onpointermove={handleLassoPointerMove}
	onpointerup={handleLassoPointerUp}
	onpointercancel={handleLassoPointerUp}
>
	<!-- Hero Spotlight Carousel & Continue Watching Rail -->
	{#if displayContinueWatching.length > 0 && activeTab === "all" && !store.hasActiveFilters}
		<HeroSpotlightCarousel
			items={displayContinueWatching}
			onItemUpdated={loadContinueWatching}
		/>
		{#if upNextList.length > 0}
			<UpNextRail items={upNextList} onEpisodeWatched={loadContinueWatching} />
		{:else}
			<ContinueWatchingRail items={displayContinueWatching} />
		{/if}
	{/if}

	<!-- Top Controls: View Options, Density, Sort & Bulk Mode -->
	<DashboardControls
		{viewMode}
		density={gridDensity}
		{bulkMode}
		onViewModeChange={toggleViewMode}
		onDensityChange={changeDensity}
		onToggleBulkMode={() => {
			bulkMode = !bulkMode;
			if (!bulkMode) selectedIds = [];
		}}
	/>

	<!-- Main Content Area: Grid View or List View -->
	{#if isInitialLoad && dataLoader.state.reqLoading && (!dataLoader.state.data || dataLoader.state.data.length === 0)}
		<PosterGridSkeleton
			count={gridDensity === "compact"
				? 18
				: gridDensity === "editorial"
					? 8
					: 12}
			density={gridDensity}
		/>
	{:else if viewMode === "grid"}
		{#if dataLoader.state.data?.length > 0}
			<PosterList density={gridDensity}>
				{#each dataLoader.state.data as w, i (w ? `${w.type}-${w.watched?.id || w.ids?.tmdb || w.ids?.igdb || i}` : i)}
					{#if w}
						<Poster
							bind:watched={dataLoader.state.data[i].watched}
							media={w}
							fluidSize={true}
							selected={w.watched?.id
								? selectedIds.includes(w.watched.id)
								: false}
							onUpdated={() => loadContinueWatching()}
							onClick={(e) => handlePosterClick(w.watched?.id, i, e)}
						>
							{#if (bulkMode || selectedIds.length > 0) && w.watched?.id}
								{@const isSelected = selectedIds.includes(w.watched.id)}
								<button
									type="button"
									class="bulk-card-badge"
									class:selected={isSelected}
									onclick={(e) => {
										e.stopPropagation();
										handlePosterClick(w.watched!.id, i, e);
									}}
									aria-pressed={isSelected}
									aria-label="Toggle selection"
								>
									{#if isSelected}
										<Icon i="check" wh={13} />
									{/if}
								</button>
							{/if}
						</Poster>
					{/if}
				{/each}
			</PosterList>
		{:else if !dataLoader.state.reqLoading && !dataLoader.state.reqLoadError}
			<EmptyLibrary />
		{/if}
	{:else}
		<!-- List View Mode -->
		{#if dataLoader.state.data?.length > 0}
			<MediaTable
				data={dataLoader.state.data}
				{bulkMode}
				{selectedIds}
				onToggleSelect={(id, e, idx) => handlePosterClick(id, idx ?? 0, e)}
			/>
		{:else if !dataLoader.state.reqLoading && !dataLoader.state.reqLoadError}
			<EmptyLibrary />
		{/if}
	{/if}

	{#if dataLoader.state.reqLoading && dataLoader.state.data?.length > 0}
		<div class="loader-wrap">
			<Spinner />
		</div>
	{/if}

	{#if dataLoader.state.reqLoadError}
		<div class="error-wrap">
			<Error
				pretty="Failed to load library results!"
				error={dataLoader.state.reqLoadError}
				onRetry={() => {
					dataLoader.state.reqLoadError = undefined;
					dataLoader.runFn();
				}}
			/>
		</div>
	{/if}

	<!-- Bulk Selection Action Dock (Fixed at bottom center when items are selected) -->
	{#if bulkMode || selectedIds.length > 0}
		<BulkActionBar
			selectedCount={selectedAllInLibrary
				? dataLoader.state.totalResults || selectedIds.length
				: selectedIds.length}
			totalCount={dataLoader.state.totalResults}
			onSelectAll={selectAllLoaded}
			onSelectAllEverywhere={selectAllEverywhere}
			onClear={clearSelection}
			onUpdateStatus={bulkUpdateStatus}
			onUpdateRating={bulkUpdateRating}
			onDelete={bulkDeleteItems}
			onClose={() => {
				bulkMode = false;
				selectedIds = [];
				selectedAllInLibrary = false;
				lastClickedIndex = undefined;
			}}
		/>
	{/if}

	<!-- Floating Marquee Lasso Box -->
	{#if isLassoing && lassoBox.width > 5 && lassoBox.height > 5}
		<div
			class="lasso-box"
			style="left: {lassoBox.left}px; top: {lassoBox.top}px; width: {lassoBox.width}px; height: {lassoBox.height}px;"
		></div>
	{/if}
</div>

<style lang="scss">
	.dashboard-container {
		display: flex;
		flex-flow: column;
		width: 100%;
		max-width: 1440px;
		margin: 0 auto;
		padding: 0 16px 120px 16px;
		position: relative;
		user-select: none;
	}

	.lasso-box {
		position: fixed;
		pointer-events: none;
		z-index: 99998;
		background: rgba(99, 102, 241, 0.16);
		border: 1.5px solid rgba(129, 140, 248, 0.85);
		border-radius: 6px;
		box-shadow:
			0 0 16px rgba(99, 102, 241, 0.3),
			inset 0 0 12px rgba(99, 102, 241, 0.1);
		backdrop-filter: blur(1.5px);
		-webkit-backdrop-filter: blur(1.5px);
	}

	:global(.poster-list-wrapper) {
		:global(.poster) {
			transition:
				transform 160ms cubic-bezier(0.16, 1, 0.3, 1),
				box-shadow 160ms ease;

			&:has(.bulk-card-badge) {
				cursor: pointer;

				&:hover {
					transform: translateY(-2px);
				}

				&:active {
					transform: scale(0.985);
				}
			}

			&:has(.bulk-card-badge.selected) {
				outline: 2px solid $brand-color;
				outline-offset: -2px;
				border-radius: 8px;
			}
		}

		.bulk-card-badge {
			position: absolute;
			top: 10px;
			right: 10px;
			width: 24px;
			height: 24px;
			border-radius: 50%;
			border: 1.5px solid rgba(255, 255, 255, 0.45);
			background-color: rgba(10, 14, 22, 0.6);
			color: #ffffff;
			display: flex;
			align-items: center;
			justify-content: center;
			cursor: pointer;
			z-index: 40;
			backdrop-filter: blur(12px);
			-webkit-backdrop-filter: blur(12px);
			transition: all 160ms cubic-bezier(0.16, 1, 0.3, 1);
			padding: 0;

			&:hover {
				border-color: rgba(255, 255, 255, 0.85);
				transform: scale(1.1);
			}

			&.selected {
				background-color: $brand-color;
				border-color: #ffffff;
				box-shadow: 0 2px 8px rgba(245, 158, 11, 0.5);
				transform: scale(1.05);
			}
		}
	}

	.loader-wrap,
	.error-wrap {
		display: flex;
		justify-content: center;
		margin: 30px 0 60px 0;
	}
</style>
