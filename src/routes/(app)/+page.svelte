<script lang="ts">
	import Error from "@/lib/Error.svelte";
	import Icon from "@/lib/Icon.svelte";
	import Poster from "@/lib/poster/Poster.svelte";
	import PosterList from "@/lib/poster/PosterList.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import DashboardControls from "@/lib/dashboard/DashboardControls.svelte";
	import ContinueWatchingRail from "@/lib/dashboard/ContinueWatchingRail.svelte";
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
	} from "@/types";
	import { notify } from "@/lib/util/notify";
	import { onDestroy, onMount, untrack } from "svelte";

	const scroll = infScroll({ callback: onScrollToBottom });
	const dataLoader = paginatedLoader<Media, undefined>(load);

	// View mode: 'grid' (cinematic cards) vs 'list' (compact table/rows)
	let viewMode: "grid" | "list" = $state("grid");

	// Bulk actions mode
	let bulkMode: boolean = $state(false);
	let selectedIds: number[] = $state([]);

	// Dedicated continue watching state
	let continueWatchingList: Media[] = $state([]);

	// Active status tab pill
	let activeTab:
		"all" | "watching" | "planned" | "finished" | "hold" | "dropped" =
		$derived.by(() => {
			const statuses = store.activeFilters?.status;
			if (!statuses || statuses.length === 0) return "all";
			if (statuses.includes("watching")) return "watching";
			if (statuses.includes("planned")) return "planned";
			if (statuses.includes("finished")) return "finished";
			if (statuses.includes("hold")) return "hold";
			if (statuses.includes("dropped")) return "dropped";
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

	async function load(signal: AbortSignal) {
		console.debug("load: loadParams:", nextLoadParams);
		if (nextLoadParams.page === dataLoader.state.page) {
			console.warn("load: Already on this page, not loading it again!");
			return;
		}
		const r = await req.get<PaginationResponse<Media, undefined>>(`/watched`, {
			params: nextLoadParams,
			signal,
		});
		scroll.dataLoaded();
		return r;
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
	}

	async function onScrollToBottom() {
		if (dataLoader.state.reqLoadError) {
			return;
		}
		dataLoader.runFn();
	}

	function toggleSelectItem(id: number) {
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
	}

	function clearSelection() {
		selectedIds = [];
	}

	async function bulkUpdateStatus(newStatus: WatchedStatus) {
		if (selectedIds.length === 0) return;
		const nid = notify({
			text: `Updating ${selectedIds.length} items...`,
			type: "loading",
		});
		let successCount = 0;
		for (const id of selectedIds) {
			const item = dataLoader.state.data?.find((w) => w?.watched?.id === id);
			if (item?.watched) {
				try {
					const cType =
						item.type === MediaTypeE.tmdbMovie
							? "movie"
							: item.type === MediaTypeE.tmdbShow
								? "tv"
								: "game";
					const cId = item.ids.tmdb || item.ids.igdb || 0;
					await updateWatched(item.watched, {
						contentId: cId,
						contentType: cType,
						status: newStatus,
					});
					item.watched.status = newStatus;
					successCount++;
				} catch (err) {
					console.error(`Failed to bulk update item ${id}:`, err);
				}
			}
		}
		notify({
			id: nid,
			text: `Updated ${successCount} items to ${newStatus}`,
			type: "success",
		});
		selectedIds = [];
		dataLoader.reset();
		dataLoader.runFn();
		loadContinueWatching();
	}

	async function bulkUpdateRating(newRating: number) {
		if (selectedIds.length === 0) return;
		const nid = notify({
			text: `Rating ${selectedIds.length} items...`,
			type: "loading",
		});
		let successCount = 0;
		for (const id of selectedIds) {
			const item = dataLoader.state.data?.find((w) => w?.watched?.id === id);
			if (item?.watched) {
				try {
					const cType =
						item.type === MediaTypeE.tmdbMovie
							? "movie"
							: item.type === MediaTypeE.tmdbShow
								? "tv"
								: "game";
					const cId = item.ids.tmdb || item.ids.igdb || 0;
					await updateWatched(item.watched, {
						contentId: cId,
						contentType: cType,
						rating: newRating,
					});
					item.watched.rating = newRating;
					successCount++;
				} catch (err) {
					console.error(`Failed to bulk rate item ${id}:`, err);
				}
			}
		}
		notify({
			id: nid,
			text: `Rated ${successCount} items`,
			type: "success",
		});
		selectedIds = [];
		dataLoader.reset();
		dataLoader.runFn();
		loadContinueWatching();
	}

	async function bulkDeleteItems() {
		if (selectedIds.length === 0) return;
		if (
			!confirm(
				`Are you sure you want to remove ${selectedIds.length} selected items from your watched list?`,
			)
		) {
			return;
		}
		const nid = notify({
			text: `Removing ${selectedIds.length} items...`,
			type: "loading",
		});
		let successCount = 0;
		for (const id of selectedIds) {
			try {
				const ok = await removeWatched(id);
				if (ok) successCount++;
			} catch (err) {
				console.error(`Failed to bulk remove item ${id}:`, err);
			}
		}
		notify({
			id: nid,
			text: `Removed ${successCount} items from your list`,
			type: "success",
		});
		selectedIds = [];
		dataLoader.reset();
		dataLoader.runFn();
		loadContinueWatching();
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
		loadContinueWatching();
	});

	function toggleViewMode(mode: "grid" | "list") {
		viewMode = mode;
		localStorage.setItem("watcharr_view_mode", mode);
	}

	onDestroy(() => {
		scroll.destroy();
		dataLoader.abortReq("page destroyed");
	});
</script>

<svelte:head>
	<title>Watched List - Watcharr</title>
</svelte:head>

<div class="dashboard-container">
	<!-- Top Controls: View Options, Sort & Bulk Mode -->
	<DashboardControls
		{viewMode}
		{bulkMode}
		onViewModeChange={toggleViewMode}
		onToggleBulkMode={() => {
			bulkMode = !bulkMode;
			if (!bulkMode) selectedIds = [];
		}}
	/>

	<!-- Continue Watching Rail (Shown when items are currently in progress) -->
	{#if displayContinueWatching.length > 0 && activeTab === "all" && !store.hasActiveFilters}
		<ContinueWatchingRail items={displayContinueWatching} />
	{/if}

	<!-- Main Content Area: Grid View or List View -->
	{#if dataLoader.state.reqLoading && (!dataLoader.state.data || dataLoader.state.data.length === 0)}
		<PosterGridSkeleton count={12} />
	{:else if viewMode === "grid"}
		<PosterList>
			{#if dataLoader.state.data?.length > 0}
				{#each dataLoader.state.data as w, i (`${i}-${w.type}`)}
					{#if w}
						<Poster
							bind:watched={dataLoader.state.data[i].watched}
							media={w}
							fluidSize={true}
							onUpdated={() => loadContinueWatching()}
							onClick={bulkMode && w.watched?.id
								? () => toggleSelectItem(w.watched!.id)
								: undefined}
						>
							{#if bulkMode && w.watched?.id}
								{@const isSelected = selectedIds.includes(w.watched.id)}
								<button
									type="button"
									class="bulk-card-badge"
									class:selected={isSelected}
									onclick={(e) => {
										e.stopPropagation();
										toggleSelectItem(w.watched!.id);
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
			{:else if !dataLoader.state.reqLoading && !dataLoader.state.reqLoadError}
				<EmptyLibrary />
			{/if}
		</PosterList>
	{:else}
		<!-- List View Mode -->
		{#if dataLoader.state.data?.length > 0}
			<MediaTable
				data={dataLoader.state.data}
				{bulkMode}
				{selectedIds}
				onToggleSelect={toggleSelectItem}
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
	{#if bulkMode}
		<BulkActionBar
			selectedCount={selectedIds.length}
			onSelectAll={selectAllLoaded}
			onClear={clearSelection}
			onUpdateStatus={bulkUpdateStatus}
			onUpdateRating={bulkUpdateRating}
			onDelete={bulkDeleteItems}
			onClose={() => {
				bulkMode = false;
				selectedIds = [];
			}}
		/>
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
	}

	:global(.poster-list-wrapper) {
		:global(.poster) {
			transition: transform 160ms cubic-bezier(0.16, 1, 0.3, 1), box-shadow 160ms ease;

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
