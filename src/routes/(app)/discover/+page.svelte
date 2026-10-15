<script lang="ts">
	import Spinner from "@/lib/Spinner.svelte";
	import { req } from "@/lib/util/api";
	import Poster from "@/lib/poster/Poster.svelte";
	import Icon from "@/lib/Icon.svelte";
	import MediaTypeFilter from "@/lib/search/MediaTypeFilter.svelte";
	import infScroll from "@/lib/util/infScroll";
	import paginatedLoader, {
		PaginatedLoaderRunFnAction,
	} from "@/lib/util/paginatedLoader.svelte";
	import {
		DiscoverFilter,
		MediaTypeE,
		SearchType,
		type DiscoverRequest,
		type Media,
		type PaginationResponse,
	} from "@/types";
	import { page } from "$app/state";
	import { afterNavigate, goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import { onDestroy, onMount } from "svelte";
	import PosterList from "@/lib/poster/PosterList.svelte";
	import Error from "@/lib/Error.svelte";
	import PersonPoster from "@/lib/poster/PersonPoster.svelte";
	import FilterDropDown from "./FilterDropDown.svelte";
	import PosterGridSkeleton from "@/lib/dashboard/PosterGridSkeleton.svelte";

	const scroll = infScroll({ callback: onScrollToBottom });
	const dataLoader = paginatedLoader<Media, undefined>(load);

	let discoverFilter: DiscoverFilter = $state(DiscoverFilter.trending);
	let discoverType: SearchType | undefined = $derived.by(() => {
		const t = page.url.searchParams.get("type");
		if (t) {
			return t as SearchType;
		}
		return SearchType.multi;
	});
	let nextLoadParams: DiscoverRequest = $derived({
		page: dataLoader.state.page + 1,
		type: discoverType,
		filter: discoverFilter,
	});

	async function load(signal: AbortSignal) {
		console.debug("load: loadParams:", nextLoadParams);
		if (nextLoadParams.page === dataLoader.state.page) {
			console.warn("load: Already on this page, not loading it again!");
			return;
		}
		const r = await req.get<PaginationResponse<Media, undefined>>(`/discover`, {
			params: nextLoadParams,
			signal,
		});
		scroll.dataLoaded();
		return r;
	}

	async function onScrollToBottom() {
		// If an error is being shown, no more infinite scroll.
		if (dataLoader.state.reqLoadError) {
			return;
		}
		dataLoader.runFn();
	}

	function setActiveDiscoverType(to: SearchType | undefined) {
		console.debug("setActiveDiscoverType: to:", to);
		const curLocation = new URL(page.url);
		if (!to || discoverType === to) {
			curLocation.searchParams.delete("type");
		} else {
			curLocation.searchParams.set("type", to);
		}
		// Running the goto will cause afterNavigate hook to be called,
		// which will run a fresh search, so nothing else to do here.
		goto(resolve(`/discover?${curLocation.searchParams.toString()}`));
	}

	onMount(() => {
		dataLoader.runFn(PaginatedLoaderRunFnAction.Reset);
	});

	afterNavigate(() => {
		console.log(
			"afterNavigate: Query changed, performing request.",
			"searchParams:",
			page.url.searchParams,
		);
		dataLoader.abortReq("navigated away");
		dataLoader.runFn(PaginatedLoaderRunFnAction.Reset);
	});

	onDestroy(() => {
		console.debug("DISCOVER PAGE DESTROYED");
		scroll.destroy();
		dataLoader.abortReq("page destroyed");
	});
</script>

<svelte:head>
	<title>Discover Content - Watcharr</title>
</svelte:head>

<div class="content">
	<div class="inner">
		<div class="discover-header">
			<div class="header-titles">
				<h1 class="discover-title">
					<Icon i="compass" wh={28} />
					Discover
				</h1>
				<p class="discover-subtitle">
					Explore trending movies, series, games, and upcoming releases.
				</p>
			</div>

			<div class="discover-toolbar">
				<div class="toolbar-typefilter">
					<MediaTypeFilter
						active={discoverType}
						disabled={false}
						onChange={(nowActive) => {
							discoverFilter = DiscoverFilter.trending;
							setActiveDiscoverType(nowActive as SearchType | undefined);
						}}
					/>
				</div>
				<div class="toolbar-filterdropdown">
					<FilterDropDown
						{discoverType}
						bind:active={discoverFilter}
						onChange={() => {
							console.log("Discover FilterDropDown Selected Change");
							dataLoader.runFn(PaginatedLoaderRunFnAction.Reset);
						}}
					/>
				</div>
			</div>
		</div>

		{#if dataLoader.state.reqLoading && (!dataLoader.state.data || dataLoader.state.data.length === 0)}
			<PosterGridSkeleton count={18} />
		{:else}
			<PosterList>
				{#if dataLoader.state.data?.length > 0}
					{#each dataLoader.state.data as w, i (`${i}-${w.type}`)}
						{#if w.type === MediaTypeE.tmdbPerson}
							<PersonPoster
								id={w.ids.tmdb}
								name={w.name}
								path={w.extPosterPath}
							/>
						{:else if w.type === MediaTypeE.tmdbMovie || w.type === MediaTypeE.tmdbShow || w.type === MediaTypeE.igdbGame}
							<Poster
								media={w}
								bind:watched={dataLoader.state.data[i].watched}
								fluidSize
							/>
						{/if}
					{/each}
				{:else if !dataLoader.state.reqLoading && !dataLoader.state.reqLoadError}
					<div class="empty-discover">
						<div class="empty-icon-wrap">
							<Icon i="compass" wh={48} />
						</div>
						<h3 class="norm">No Media Found</h3>
						<p class="empty-hint">
							Try selecting a different filter category or media type above.
						</p>
					</div>
				{/if}
			</PosterList>
		{/if}

		{#if dataLoader.state.reqLoading && dataLoader.state.data?.length > 0}
			<div class="loader-wrap">
				<Spinner />
			</div>
		{/if}

		{#if dataLoader.state.reqLoadError}
			<div class="error-wrap">
				<Error
					pretty="Failed to load results!"
					error={dataLoader.state.reqLoadError}
					onRetry={() => {
						dataLoader.state.reqLoadError = undefined;
						dataLoader.runFn(PaginatedLoaderRunFnAction.ResetIfOnFirstOrNoPage);
					}}
				/>
			</div>
		{/if}
	</div>
</div>

<style lang="scss">
	.content {
		display: flex;
		width: 100%;
		justify-content: center;
		padding: 24px 20px 80px 20px;

		.inner {
			width: 100%;
			max-width: 1400px;
		}
	}

	.discover-header {
		display: flex;
		flex-direction: column;
		gap: 16px;
		margin-bottom: 24px;
		padding-bottom: 20px;
		border-bottom: 1px solid rgba(255, 255, 255, 0.06);

		.header-titles {
			display: flex;
			flex-direction: column;
			gap: 4px;

			.discover-title {
				display: flex;
				align-items: center;
				gap: 10px;
				font-size: 26px;
				font-weight: 700;
				letter-spacing: -0.02em;
				color: $text-color;
				margin: 0;
			}

			.discover-subtitle {
				font-size: 14px;
				color: $text-muted;
				margin: 0;
			}
		}

		.discover-toolbar {
			display: flex;
			flex-flow: row wrap;
			align-items: center;
			justify-content: space-between;
			gap: 12px;

			.toolbar-typefilter {
				display: flex;
				align-items: center;
			}

			.toolbar-filterdropdown {
				margin-left: auto;
			}
		}

		@media screen and (max-width: 768px) {
			.discover-toolbar {
				flex-direction: column;
				align-items: stretch;

				.toolbar-filterdropdown {
					margin-left: 0;
					width: 100%;
				}
			}
		}
	}

	.empty-discover {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 60px 20px;
		text-align: center;
		grid-column: 1 / -1;

		.empty-icon-wrap {
			display: flex;
			align-items: center;
			justify-content: center;
			width: 80px;
			height: 80px;
			border-radius: 50%;
			background: rgba(255, 255, 255, 0.04);
			border: 1px solid rgba(255, 255, 255, 0.08);
			color: $text-muted;
			margin-bottom: 16px;
		}

		h3 {
			font-size: 20px;
			font-weight: 600;
			margin: 0 0 6px 0;
			color: $text-color;
		}

		.empty-hint {
			font-size: 14px;
			color: $text-muted;
			max-width: 360px;
			margin: 0;
		}
	}

	.loader-wrap,
	.error-wrap {
		display: flex;
		justify-content: center;
		margin: 40px 0 60px 0;
	}
</style>
