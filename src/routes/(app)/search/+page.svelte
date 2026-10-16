<script lang="ts">
	import { page } from "$app/state";
	import { afterNavigate, goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import { req } from "@/lib/util/api.js";
	import Poster from "@/lib/poster/Poster.svelte";
	import PosterList from "@/lib/poster/PosterList.svelte";
	import { store } from "@/store.svelte.js";
	import Spinner from "@/lib/Spinner.svelte";
	import Icon from "@/lib/Icon.svelte";
	import PersonPoster from "@/lib/poster/PersonPoster.svelte";
	import {
		MediaTypeE,
		SearchType,
		type Media,
		type PaginationResponse,
		type PublicUser,
		type SearchRequest,
		type SearchResponseMeta,
	} from "@/types";
	import UsersList from "@/lib/UsersList.svelte";
	import { onDestroy, onMount } from "svelte";
	import Error from "@/lib/Error.svelte";
	import infScroll from "@/lib/util/infScroll.js";
	import paginatedLoader, {
		PaginatedLoaderRunFnAction,
	} from "@/lib/util/paginatedLoader.svelte.js";
	import MediaTypeFilter from "@/lib/search/MediaTypeFilter.svelte";
	import Filters from "./components/Filters.svelte";
	import PosterGridSkeleton from "@/lib/dashboard/PosterGridSkeleton.svelte";

	let { data } = $props();

	const scroll = infScroll({ callback: onScrollToBottom });
	const dataLoader = paginatedLoader<Media, SearchResponseMeta>(load);

	let searchType: SearchType | undefined = $derived.by(() => {
		const t = page.url.searchParams.get("type");
		if (t) {
			return t as SearchType;
		}
		return SearchType.multi;
	});

	let preferMyList: boolean = $derived.by(() => {
		const t = page.url.searchParams.get("preferMyList");
		return Boolean(t);
	});
	let showingResultsFromMyList: boolean = $derived(
		Boolean(
			dataLoader.state.data?.length > 0 && dataLoader.state.meta?.fromMyList,
		),
	);

	let nextLoadParams: SearchRequest = $derived({
		page: dataLoader.state.page + 1,
		query: store.searchQuery,
		type: searchType,
		preferMyList: preferMyList,
	});

	async function load(signal: AbortSignal) {
		console.debug("load: loadParams:", nextLoadParams);
		if (nextLoadParams.page === dataLoader.state.page) {
			console.warn("load: Already on this page, not loading it again!");
			return;
		}
		if (!nextLoadParams.query) {
			console.warn("load: There is no search query!");
			return;
		}
		const r = await req.get<PaginationResponse<Media, SearchResponseMeta>>(
			`/search`,
			{
				params: nextLoadParams,
				signal,
			},
		);
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

	function setActiveSearchFilter(to: SearchType | undefined) {
		console.debug("setActiveSearchFilter: to:", to);
		const curLocation = new URL(page.url);
		if (!to || searchType === to) {
			curLocation.searchParams.delete("type");
		} else {
			curLocation.searchParams.set("type", to);
		}
		// Running the goto will cause afterNavigate hook to be called,
		// which will run a fresh search, so nothing else to do here.
		goto(resolve(`/search?${curLocation.searchParams.toString()}`));
	}

	async function searchUsers(query: string) {
		return await req.get<PublicUser[]>(`/user/search`, {
			params: { q: query },
		});
	}

	onMount(() => {
		if (!store.searchQuery && data?.query) {
			store.searchQuery = data.query;
		}
		dataLoader.runFn(PaginatedLoaderRunFnAction.Reset);
	});

	afterNavigate((e) => {
		if (!e.from?.route?.id?.toLowerCase()?.includes("/search")) {
			return;
		}
		console.log(
			"afterNavigate: Query changed, performing search.",
			"searchParams:",
			page.url.searchParams,
		);
		// Sync state (so back button updates search correctly)
		store.searchQuery = data?.query ? data.query : "";
		dataLoader.abortReq("navigated away");
		dataLoader.runFn(PaginatedLoaderRunFnAction.Reset);
	});

	onDestroy(() => {
		console.debug("SEARCH PAGE DESTROYED");
		store.searchQuery = "";
		scroll.destroy();
		dataLoader.abortReq("page destroyed");
	});

	function dontPreferMyListClicked() {
		const curLocation = new URL(page.url);
		curLocation.searchParams.delete("preferMyList");
		// Running the goto will cause afterNavigate hook to be called,
		// which will run a fresh search, so nothing else to do here.
		goto(resolve(`/search?${curLocation.searchParams.toString()}`));
	}

	function applySearchQuery(q: string) {
		store.searchQuery = q;
		const curLocation = new URL(page.url);
		curLocation.searchParams.set("query", q);
		if (!curLocation.searchParams.has("preferMyList")) {
			curLocation.searchParams.set("preferMyList", "true");
		}
		goto(resolve(`/search?${curLocation.searchParams.toString()}`));
	}

	function appendSyntax(suffix: string) {
		const current = store.searchQuery || data?.query || "";
		const updated = current.trim() ? `${current.trim()} ${suffix}` : suffix;
		applySearchQuery(updated);
	}
</script>

<svelte:head>
	<title
		>Search Results{store.searchQuery ? ` for '${store.searchQuery}'` : ""} - Watcharr</title
	>
</svelte:head>

<div class="content">
	<div class="inner">
		{#if data?.query}
			<!-- Users Search Results Rail -->
			{#await searchUsers(data?.query) then results}
				{#if results?.length > 0}
					<UsersList users={results} />
				{/if}
			{:catch err}
				<Error pretty="Failed to load users!" error={err} />
			{/await}

			<!-- Search Results Header -->
			<div class="search-header">
				<div class="header-titles">
					<h1 class="search-title">
						Results for <span>"{data?.query}"</span>
					</h1>
					{#if dataLoader.state.data?.length > 0}
						<span class="results-badge"
							>{dataLoader.state.data.length} matches loaded</span
						>
					{/if}
				</div>

				<div class="search-toolbar">
					<MediaTypeFilter
						active={searchType}
						disabled={dataLoader.state.reqLoading || showingResultsFromMyList}
						onChange={(nowActive) => {
							setActiveSearchFilter(nowActive as SearchType | undefined);
						}}
					/>
					<Filters />
				</div>
			</div>

			<!-- Syntax Helper Chips for Active Search -->
			<div class="syntax-refine-bar">
				<span class="syntax-label">Refine search:</span>
				<button class="syntax-pill" onclick={() => appendSyntax("y:2025")}>
					+ Year (y:2025)
				</button>
				<button class="syntax-pill" onclick={() => appendSyntax("y:2024")}>
					+ Year (y:2024)
				</button>
				<button class="syntax-pill" onclick={() => appendSyntax("fy:2024")}>
					+ First Year (fy:2024)
				</button>
				<button class="syntax-pill" onclick={() => appendSyntax("imdb:")}>
					+ IMDb ID
				</button>
			</div>

			<!-- Personal List Priority Notice Banner -->
			{#if showingResultsFromMyList}
				<div class="from-my-list-banner">
					<div class="banner-content">
						<Icon i="reel" wh={20} />
						<div class="banner-texts">
							<strong>Showing matches from your library</strong>
							<span
								>Results were matched with your existing watchlist entries.</span
							>
						</div>
					</div>
					<button class="banner-action-btn" onclick={dontPreferMyListClicked}>
						Search Full Database
					</button>
				</div>
			{/if}

			<!-- Poster List -->
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
						<div class="empty-search-results">
							<div class="empty-icon-wrap">
								<Icon i="search" wh={44} />
							</div>
							<h3 class="norm">No Results Found</h3>
							<p class="empty-hint">
								We couldn't find anything matching "{data?.query}". Check for
								typos, try searching for a title with a year filter (e.g.
								<code>y:2024</code>), or enter an IMDb ID directly.
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
							dataLoader.runFn(
								PaginatedLoaderRunFnAction.ResetIfOnFirstOrNoPage,
							);
						}}
					/>
				</div>
			{/if}
		{:else}
			<!-- Empty Search Prompt -->
			<div class="search-empty-landing">
				<div class="landing-icon-wrap">
					<Icon i="search" wh={48} />
				</div>
				<h2>Search Watcharr</h2>
				<p>
					Find movies, television shows, video games, people, or community
					profiles across global databases.
				</p>
				<div class="quick-links">
					<span class="links-label">Quick exploration:</span>
					<button
						class="quick-pill"
						onclick={() => goto(resolve("/discover?type=movie"))}
					>
						<Icon i="film" wh={14} /> Movies
					</button>
					<button
						class="quick-pill"
						onclick={() => goto(resolve("/discover?type=show"))}
					>
						<Icon i="tv" wh={14} /> TV Shows
					</button>
					{#if store.serverFeatures?.games}
						<button
							class="quick-pill"
							onclick={() => goto(resolve("/discover?type=game"))}
						>
							<Icon i="gamepad" wh={14} /> Games
						</button>
					{/if}
				</div>

				<!-- Power Search Syntax Guide -->
				<div class="syntax-guide-card">
					<div class="guide-header">
						<Icon i="sparkles" wh={18} />
						<h3>Power Search Syntax</h3>
					</div>
					<p class="guide-intro">
						Type these prefixes directly into the search bar for advanced,
						laser-focused queries:
					</p>
					<div class="syntax-grid">
						<div class="syntax-item">
							<div class="syntax-badge"><code>y:YEAR</code></div>
							<div class="syntax-desc">
								Filter movies and shows by release year
							</div>
							<button
								class="syntax-example-btn"
								onclick={() => applySearchQuery("Inception y:2010")}
							>
								Try "Inception y:2010"
							</button>
						</div>
						<div class="syntax-item">
							<div class="syntax-badge"><code>fy:YEAR</code></div>
							<div class="syntax-desc">
								Filter TV shows by series debut / first air year
							</div>
							<button
								class="syntax-example-btn"
								onclick={() => applySearchQuery("Breaking Bad fy:2008")}
							>
								Try "Breaking Bad fy:2008"
							</button>
						</div>
						<div class="syntax-item">
							<div class="syntax-badge"><code>imdb:ID</code></div>
							<div class="syntax-desc">
								Direct lookup by IMDb title identifier
							</div>
							<button
								class="syntax-example-btn"
								onclick={() => applySearchQuery("imdb:tt0111161")}
							>
								Try "imdb:tt0111161"
							</button>
						</div>
						<div class="syntax-item">
							<div class="syntax-badge">
								<code>movie:ID</code> / <code>tv:ID</code>
							</div>
							<div class="syntax-desc">Direct TMDB content lookup by ID</div>
							<button
								class="syntax-example-btn"
								onclick={() => applySearchQuery("movie:550")}
							>
								Try "movie:550"
							</button>
						</div>
					</div>
				</div>
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

	.search-header {
		display: flex;
		flex-direction: column;
		gap: 16px;
		margin-bottom: 24px;
		padding-bottom: 20px;
		border-bottom: 1px solid rgba(255, 255, 255, 0.06);

		.header-titles {
			display: flex;
			align-items: baseline;
			gap: 12px;
			flex-wrap: wrap;

			.search-title {
				font-size: 24px;
				font-weight: 700;
				letter-spacing: -0.02em;
				color: $text-color;
				margin: 0;

				span {
					color: $accent-color;
				}
			}

			.results-badge {
				font-size: 13px;
				color: $text-muted;
			}
		}

		.search-toolbar {
			display: flex;
			flex-flow: row wrap;
			align-items: center;
			justify-content: space-between;
			gap: 12px;
		}

		@media screen and (max-width: 768px) {
			.search-toolbar {
				flex-direction: column;
				align-items: stretch;
			}
		}
	}

	.from-my-list-banner {
		display: flex;
		flex-flow: row wrap;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		padding: 14px 20px;
		margin-bottom: 24px;
		border-radius: 12px;
		background: rgba(229, 160, 0, 0.08);
		border: 1px solid rgba(229, 160, 0, 0.25);
		color: $text-color;

		.banner-content {
			display: flex;
			align-items: center;
			gap: 14px;
			color: #ffd060;

			.banner-texts {
				display: flex;
				flex-direction: column;
				gap: 2px;

				strong {
					font-size: 14px;
					font-weight: 600;
					color: #ffe082;
				}

				span {
					font-size: 12.5px;
					color: rgba(255, 255, 255, 0.7);
				}
			}
		}

		.banner-action-btn {
			padding: 7px 16px;
			border-radius: 20px;
			font-size: 13px;
			font-weight: 600;
			background: #ffd060;
			color: #1a1a1a;
			border: none;
			cursor: pointer;
			transition: all 150ms ease;

			&:hover {
				background: #fff;
				transform: translateY(-1px);
			}
		}
	}

	.empty-search-results {
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
			max-width: 400px;
			margin: 0;
		}
	}

	.search-empty-landing {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 80px 20px;
		text-align: center;

		.landing-icon-wrap {
			display: flex;
			align-items: center;
			justify-content: center;
			width: 90px;
			height: 90px;
			border-radius: 50%;
			background: rgba(255, 255, 255, 0.04);
			border: 1px solid rgba(255, 255, 255, 0.08);
			color: $accent-color;
			margin-bottom: 20px;
		}

		h2 {
			font-size: 26px;
			font-weight: 700;
			letter-spacing: -0.02em;
			margin: 0 0 8px 0;
			color: $text-color;
		}

		p {
			font-size: 15px;
			color: $text-muted;
			max-width: 440px;
			margin: 0 0 28px 0;
			line-height: 1.5;
		}

		.quick-links {
			display: flex;
			flex-flow: row wrap;
			align-items: center;
			justify-content: center;
			gap: 8px;

			.links-label {
				font-size: 13px;
				color: $text-muted;
				margin-right: 4px;
			}

			.quick-pill {
				display: inline-flex;
				align-items: center;
				gap: 6px;
				padding: 7px 14px;
				border-radius: 20px;
				font-size: 13px;
				font-weight: 500;
				color: $text-color;
				background: rgba(255, 255, 255, 0.05);
				border: 1px solid rgba(255, 255, 255, 0.1);
				cursor: pointer;
				transition: all 180ms ease;

				&:hover {
					background: rgba(255, 255, 255, 0.12);
					border-color: rgba(255, 255, 255, 0.2);
					color: #fff;
					transform: translateY(-1px);
				}
			}
		}

		.syntax-guide-card {
			margin-top: 48px;
			max-width: 800px;
			width: 100%;
			background: rgba(255, 255, 255, 0.02);
			border: 1px solid rgba(255, 255, 255, 0.06);
			border-radius: 16px;
			padding: 24px;
			text-align: left;

			.guide-header {
				display: flex;
				align-items: center;
				gap: 8px;
				color: $accent-color;
				margin-bottom: 6px;

				h3 {
					font-size: 16px;
					font-weight: 600;
					margin: 0;
					color: $text-color;
				}
			}

			.guide-intro {
				font-size: 13px;
				color: $text-muted;
				margin-bottom: 16px;
			}

			.syntax-grid {
				display: grid;
				grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
				gap: 12px;

				@media screen and (max-width: 480px) {
					grid-template-columns: 1fr;
				}
			}

			.syntax-item {
				background: rgba(255, 255, 255, 0.03);
				border: 1px solid rgba(255, 255, 255, 0.05);
				border-radius: 10px;
				padding: 12px 14px;
				display: flex;
				flex-direction: column;
				gap: 6px;

				.syntax-badge {
					code {
						background: rgba(54, 186, 152, 0.15);
						color: #36ba98;
						font-size: 12px;
						padding: 2px 6px;
						border-radius: 4px;
					}
				}

				.syntax-desc {
					font-size: 12px;
					color: $text-muted;
				}

				.syntax-example-btn {
					background: rgba(255, 255, 255, 0.04);
					border: 1px solid rgba(255, 255, 255, 0.08);
					border-radius: 6px;
					color: rgba(255, 255, 255, 0.8);
					font-size: 12px;
					padding: 4px 8px;
					text-align: left;
					cursor: pointer;
					width: max-content;
					margin-top: 4px;
					transition: all 140ms ease;

					&:hover {
						background: rgba(54, 186, 152, 0.15);
						border-color: rgba(54, 186, 152, 0.3);
						color: #36ba98;
					}
				}
			}
		}
	}

	.syntax-refine-bar {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 8px;
		margin-bottom: 20px;
		padding: 8px 12px;
		background: rgba(255, 255, 255, 0.02);
		border: 1px solid rgba(255, 255, 255, 0.05);
		border-radius: 8px;

		.syntax-label {
			font-size: 12px;
			font-weight: 600;
			color: rgba(255, 255, 255, 0.4);
			text-transform: uppercase;
			letter-spacing: 0.5px;
		}

		.syntax-pill {
			display: inline-flex;
			align-items: center;
			font-size: 12px;
			background: rgba(255, 255, 255, 0.05);
			border: 1px solid rgba(255, 255, 255, 0.08);
			color: rgba(255, 255, 255, 0.8);
			padding: 4px 10px;
			border-radius: 12px;
			cursor: pointer;
			transition: all 140ms ease;

			&:hover {
				background: rgba(54, 186, 152, 0.12);
				border-color: rgba(54, 186, 152, 0.25);
				color: #36ba98;
			}
		}
	}

	.loader-wrap,
	.error-wrap {
		display: flex;
		justify-content: center;
		margin: 40px 0 60px 0;
	}
</style>
