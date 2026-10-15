<script lang="ts">
	import { afterNavigate, goto } from "$app/navigation";
	import { page } from "$app/state";
	import { resolve } from "$app/paths";
	import Error from "@/lib/Error.svelte";
	import Icon from "@/lib/Icon.svelte";
	import Poster from "@/lib/poster/Poster.svelte";
	import PosterList from "@/lib/poster/PosterList.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import CreateTagModal from "@/lib/tag/CreateTagModal.svelte";
	import DeleteTagModal from "@/lib/tag/DeleteTagModal.svelte";
	import { req } from "@/lib/util/api";
	import infScroll from "@/lib/util/infScroll";
	import paginatedLoader, {
		PaginatedLoaderRunFnAction,
	} from "@/lib/util/paginatedLoader.svelte";
	import { clearActiveFilters, store } from "@/store.svelte.js";
	import type { Media, PaginationResponse } from "@/types";
	import { onDestroy, untrack } from "svelte";

	let meta = $derived.by(() => {
		return {
			tagId: Number(page.params.id),
		};
	});
	let tag = $derived(store.tags.find((t) => t.id === meta.tagId));
	let tagEditModalShown = $state(false);
	let tagDeleteModalShown = $state(false);
	let tagSearchText = $state(store.activeFilters?.query || "");

	const scroll = infScroll({ callback: onScrollToBottom });
	const dataLoader = paginatedLoader<Media, undefined>(load);

	let nextLoadParams: {
		page: number;
		[x: string]: unknown;
	} = $derived({
		page: dataLoader.state.page + 1,
		...store.sortAndFiltersForQueryParams,
	});

	async function load(signal: AbortSignal) {
		console.debug("tag load: loadParams:", nextLoadParams);
		if (nextLoadParams.page === dataLoader.state.page) {
			console.warn("tag load: Already on this page, not loading it again!");
			return;
		}
		if (!meta.tagId) {
			console.warn("tag load: Missing tag id!");
			return;
		}
		const r = await req.get<PaginationResponse<Media, undefined>>(
			`/tag/${meta.tagId}/watched`,
			{
				params: nextLoadParams,
				signal,
			},
		);
		scroll.dataLoaded();
		return r;
	}

	async function onScrollToBottom() {
		if (dataLoader.state.reqLoadError) {
			return;
		}
		dataLoader.runFn();
	}

	$effect(() => {
		if (store.sortAndFiltersForQueryParams) {
			untrack(() => {
				dataLoader.reset();
				dataLoader.runFn();
			});
		}
	});

	afterNavigate((e) => {
		if (!e.from?.route?.id?.toLowerCase()?.includes("/tag")) {
			return;
		}
		dataLoader.abortReq("navigated away");
		dataLoader.runFn(PaginatedLoaderRunFnAction.Reset);
	});

	onDestroy(() => {
		scroll.destroy();
		dataLoader.abortReq("page destroyed");
	});

	function handleSearchInput(e: Event) {
		const val = (e.target as HTMLInputElement).value;
		tagSearchText = val;
		store.activeFilters = {
			...store.activeFilters,
			query: val.trim() ? val.trim() : undefined,
		};
	}

	function clearSearchInput() {
		tagSearchText = "";
		store.activeFilters = {
			...store.activeFilters,
			query: undefined,
		};
	}

	function setTypeFilter(t: "all" | "movie" | "tv" | "game") {
		if (t === "all") {
			store.activeFilters = {
				...store.activeFilters,
				type: [],
			};
		} else {
			store.activeFilters = {
				...store.activeFilters,
				type: [t],
			};
		}
	}

	function setGenreFilter(g: string | undefined) {
		store.activeFilters = {
			...store.activeFilters,
			genre: g,
		};
	}

	function setYearFilter(y: number | undefined) {
		store.activeFilters = {
			...store.activeFilters,
			year: y,
		};
	}

	function setRatingFilter(r: number | undefined) {
		store.activeFilters = {
			...store.activeFilters,
			minRating: r,
		};
	}

	function removeFilter(
		key: "status" | "type" | "year" | "minRating" | "query" | "genre",
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
		} else if (key === "query") {
			tagSearchText = "";
			store.activeFilters = { ...store.activeFilters, query: undefined };
		}
	}

	function onTagDeleted() {
		tagDeleteModalShown = false;
		goto(resolve("/"));
	}
</script>

<svelte:head>
	<title>{tag ? `${tag.name} - Tag` : "Tag"} - Watcharr</title>
</svelte:head>

{#if tag}
	<!-- Cinematic Tag Hero -->
	<header
		class="tag-hero"
		style={`--tag-accent: ${tag.bgColor || "#36BA98"}; --tag-color: ${tag.color || "#ffffff"};`}
	>
		<div class="hero-ambient-glow"></div>
		<div class="hero-content">
			<div class="tag-badge-row">
				<div
					class="tag-pill-badge"
					style={`background-color: ${tag.bgColor}; color: ${tag.color};`}
				>
					<Icon i="tag" wh={16} />
					<span>{tag.name}</span>
				</div>
				<span class="tag-count-chip">
					{dataLoader.state.data?.length || 0}
					{dataLoader.state.data?.length === 1 ? "item" : "items"}
				</span>
			</div>

			<h1 class="tag-title">{tag.name}</h1>

			{#if tag.description}
				<p class="tag-description">{tag.description}</p>
			{:else}
				<button
					class="tag-add-desc-btn plain"
					onclick={() => (tagEditModalShown = true)}
				>
					+ Add description or notes
				</button>
			{/if}

			<div class="tag-actions-row">
				<button
					class="btn secondary tag-btn"
					onclick={() => (tagEditModalShown = true)}
				>
					<Icon i="pencil" wh={14} />
					<span>Edit Tag</span>
				</button>
				<button
					class="btn danger-subtle tag-btn"
					onclick={() => (tagDeleteModalShown = true)}
				>
					<Icon i="trash" wh={14} />
					<span>Delete Tag</span>
				</button>
			</div>
		</div>
	</header>

	<!-- Filter & Search Controls Toolbar -->
	<section class="tag-toolbar">
		<div class="toolbar-search">
			<Icon i="search" wh={16} />
			<input
				type="text"
				placeholder={`Search inside ${tag.name}...`}
				value={tagSearchText}
				oninput={handleSearchInput}
				class="tag-search-input"
			/>
			{#if tagSearchText}
				<button class="plain clear-search" onclick={clearSearchInput}>
					<Icon i="close" wh={14} />
				</button>
			{/if}
		</div>

		<div class="toolbar-dropdowns">
			<select
				class="filter-select"
				value={store.activeFilters?.type?.[0] || "all"}
				onchange={(e) =>
					setTypeFilter(
						(e.target as HTMLSelectElement).value as
							"all" | "movie" | "tv" | "game",
					)}
				aria-label="Filter by type"
			>
				<option value="all">All Types</option>
				<option value="movie">Movies</option>
				<option value="tv">TV Shows</option>
				{#if store.serverFeatures?.games}
					<option value="game">Games</option>
				{/if}
			</select>

			<select
				class="filter-select"
				value={store.activeFilters?.genre || "all"}
				onchange={(e) => {
					const val = (e.target as HTMLSelectElement).value;
					setGenreFilter(val === "all" ? undefined : val);
				}}
				aria-label="Filter by genre"
			>
				<option value="all">All Genres</option>
				<option value="Action">Action</option>
				<option value="Adventure">Adventure</option>
				<option value="Animation">Animation</option>
				<option value="Comedy">Comedy</option>
				<option value="Crime">Crime</option>
				<option value="Drama">Drama</option>
				<option value="Fantasy">Fantasy</option>
				<option value="Horror">Horror</option>
				<option value="Mystery">Mystery</option>
				<option value="Romance">Romance</option>
				<option value="Science Fiction">Sci-Fi</option>
				<option value="Thriller">Thriller</option>
			</select>

			<select
				class="filter-select"
				value={store.activeFilters?.year
					? String(store.activeFilters.year)
					: "all"}
				onchange={(e) => {
					const val = (e.target as HTMLSelectElement).value;
					setYearFilter(val === "all" ? undefined : parseInt(val));
				}}
				aria-label="Filter by year"
			>
				<option value="all">All Years</option>
				<option value="2026">2026</option>
				<option value="2025">2025</option>
				<option value="2024">2024</option>
				<option value="2023">2023</option>
				<option value="2022">2022</option>
				<option value="2021">2021</option>
				<option value="2020">2020</option>
			</select>

			<select
				class="filter-select"
				value={store.activeFilters?.minRating
					? String(store.activeFilters.minRating)
					: "all"}
				onchange={(e) => {
					const val = (e.target as HTMLSelectElement).value;
					setRatingFilter(val === "all" ? undefined : parseFloat(val));
				}}
				aria-label="Filter by minimum rating"
			>
				<option value="all">All Ratings</option>
				<option value="9">★ 9+ Outstanding</option>
				<option value="8">★ 8+ Great</option>
				<option value="7">★ 7+ Good</option>
				<option value="6">★ 6+ Decent</option>
			</select>
		</div>
	</section>

	<!-- Active Filter Chips -->
	{#if store.hasActiveFilters}
		<div class="active-chips-bar">
			<span class="chips-label">Active filters:</span>
			{#if store.activeFilters?.type && store.activeFilters.type.length > 0}
				{#each store.activeFilters.type as t (t)}
					<span class="filter-chip">
						Type: {t === "tv" ? "Shows" : t === "movie" ? "Movies" : "Games"}
						<button class="plain chip-x" onclick={() => removeFilter("type")}>
							<Icon i="close" wh={12} />
						</button>
					</span>
				{/each}
			{/if}
			{#if store.activeFilters?.genre}
				<span class="filter-chip">
					Genre: {store.activeFilters.genre}
					<button class="plain chip-x" onclick={() => removeFilter("genre")}>
						<Icon i="close" wh={12} />
					</button>
				</span>
			{/if}
			{#if store.activeFilters?.year}
				<span class="filter-chip">
					Year: {store.activeFilters.year}
					<button class="plain chip-x" onclick={() => removeFilter("year")}>
						<Icon i="close" wh={12} />
					</button>
				</span>
			{/if}
			{#if store.activeFilters?.minRating}
				<span class="filter-chip">
					Rating: {store.activeFilters.minRating}+
					<button
						class="plain chip-x"
						onclick={() => removeFilter("minRating")}
					>
						<Icon i="close" wh={12} />
					</button>
				</span>
			{/if}
			{#if store.activeFilters?.query}
				<span class="filter-chip">
					Search: "{store.activeFilters.query}"
					<button class="plain chip-x" onclick={() => removeFilter("query")}>
						<Icon i="close" wh={12} />
					</button>
				</span>
			{/if}
			<button class="plain clear-all-btn" onclick={() => clearActiveFilters()}>
				Clear all
			</button>
		</div>
	{/if}

	<!-- Poster Grid -->
	<PosterList>
		{#if dataLoader.state.data?.length > 0}
			{#each dataLoader.state.data as w, i (w.watched?.id ? `w-${w.watched.id}` : `${i}-${w.type}`)}
				{#if w}
					<Poster
						bind:watched={dataLoader.state.data[i].watched}
						media={w}
						fluidSize={true}
					/>
				{/if}
			{/each}
		{:else if !dataLoader.state.reqLoading && !dataLoader.state.reqLoadError}
			<div class="empty-list">
				<div class="empty-icon-wrap">
					<Icon i="tag" wh={48} />
				</div>
				<h2 class="norm">
					{store.hasActiveFilters
						? "No matching media found"
						: "This tag is empty"}
				</h2>
				<p class="empty-desc">
					{#if store.hasActiveFilters}
						No items in "{tag.name}" match your current filters.
					{:else}
						Add movies, shows, or games to this tag from their media detail
						pages.
					{/if}
				</p>
				<div class="empty-actions">
					{#if store.hasActiveFilters}
						<button class="btn primary" onclick={() => clearActiveFilters()}>
							Clear Filters
						</button>
					{:else}
						<button class="btn primary" onclick={() => goto(resolve("/"))}>
							Browse Library
						</button>
					{/if}
				</div>
			</div>
		{/if}
	</PosterList>

	{#if dataLoader.state.reqLoading}
		<div class="loader-wrap">
			<Spinner />
		</div>
	{/if}

	{#if dataLoader.state.reqLoadError}
		<div class="error-wrap">
			<Error
				pretty="Failed to load tagged items!"
				error={dataLoader.state.reqLoadError}
				onRetry={() => {
					dataLoader.state.reqLoadError = undefined;
					dataLoader.runFn();
				}}
			/>
		</div>
	{/if}
{:else}
	<div class="tag-not-found">
		<Icon i="close-circle" wh={48} />
		<h2>Tag Not Found</h2>
		<p>The tag you are looking for does not exist or may have been deleted.</p>
		<button class="btn primary" onclick={() => goto(resolve("/"))}>
			Return to Library
		</button>
	</div>
{/if}

{#if tag && tagEditModalShown}
	<CreateTagModal
		existingTag={tag}
		onClose={() => (tagEditModalShown = false)}
	/>
{/if}

{#if tag && tagDeleteModalShown}
	<DeleteTagModal {tag} onClose={onTagDeleted} />
{/if}

<style lang="scss">
	.tag-hero {
		position: relative;
		overflow: hidden;
		padding: 40px 32px 32px 32px;
		margin: 0 auto 24px auto;
		max-width: 1400px;
		border-radius: 16px;
		background: linear-gradient(
			180deg,
			rgba(255, 255, 255, 0.04) 0%,
			rgba(255, 255, 255, 0.01) 100%
		);
		border: 1px solid rgba(255, 255, 255, 0.06);

		.hero-ambient-glow {
			position: absolute;
			top: -100px;
			left: 20%;
			width: 60%;
			height: 200px;
			background: radial-gradient(
				circle,
				var(--tag-accent) 0%,
				transparent 70%
			);
			opacity: 0.15;
			filter: blur(50px);
			pointer-events: none;
		}

		.hero-content {
			position: relative;
			display: flex;
			flex-direction: column;
			gap: 12px;
		}
	}

	.tag-badge-row {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.tag-pill-badge {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 4px 12px;
		border-radius: 20px;
		font-size: 13px;
		font-weight: 600;
		letter-spacing: 0.3px;
	}

	.tag-count-chip {
		font-size: 12px;
		color: rgba(255, 255, 255, 0.5);
		background: rgba(255, 255, 255, 0.06);
		padding: 3px 8px;
		border-radius: 12px;
	}

	.tag-title {
		font-size: 32px;
		font-weight: 800;
		letter-spacing: -0.5px;
		margin: 0;
		color: #ffffff;
	}

	.tag-description {
		font-size: 14px;
		line-height: 1.6;
		color: rgba(255, 255, 255, 0.7);
		max-width: 700px;
		margin: 0;
	}

	.tag-add-desc-btn {
		font-size: 13px;
		color: rgba(255, 255, 255, 0.4);
		text-align: left;
		cursor: pointer;
		width: max-content;
		padding: 0;

		&:hover {
			color: rgba(255, 255, 255, 0.8);
		}
	}

	.tag-actions-row {
		display: flex;
		align-items: center;
		gap: 10px;
		margin-top: 8px;

		.tag-btn {
			display: inline-flex;
			align-items: center;
			gap: 6px;
			padding: 6px 14px;
			font-size: 13px;
			font-weight: 500;
			border-radius: 8px;
			cursor: pointer;
		}
	}

	.btn.secondary {
		background: rgba(255, 255, 255, 0.08);
		color: #ffffff;
		border: 1px solid rgba(255, 255, 255, 0.1);

		&:hover {
			background: rgba(255, 255, 255, 0.14);
		}
	}

	.btn.danger-subtle {
		background: rgba(255, 68, 68, 0.1);
		color: #ff6b6b;
		border: 1px solid rgba(255, 68, 68, 0.2);

		&:hover {
			background: rgba(255, 68, 68, 0.2);
			color: #ff8888;
		}
	}

	.tag-toolbar {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		max-width: 1400px;
		margin: 0 auto 16px auto;
		padding: 0 4px;

		.toolbar-search {
			display: flex;
			align-items: center;
			gap: 8px;
			background: rgba(255, 255, 255, 0.05);
			border: 1px solid rgba(255, 255, 255, 0.08);
			border-radius: 8px;
			padding: 6px 12px;
			flex: 1 1 240px;
			max-width: 380px;

			.tag-search-input {
				background: transparent;
				border: none;
				outline: none;
				color: #ffffff;
				font-size: 13px;
				width: 100%;

				&::placeholder {
					color: rgba(255, 255, 255, 0.35);
				}
			}

			.clear-search {
				color: rgba(255, 255, 255, 0.4);
				display: flex;
				cursor: pointer;
				padding: 2px;

				&:hover {
					color: #ffffff;
				}
			}
		}

		.toolbar-dropdowns {
			display: flex;
			align-items: center;
			gap: 8px;
			flex-wrap: wrap;

			.filter-select {
				appearance: none;
				-webkit-appearance: none;
				-moz-appearance: none;
				background: rgba(255, 255, 255, 0.05);
				background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='rgba(255,255,255,0.5)' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
				background-repeat: no-repeat;
				background-position: right 10px center;
				border: 1px solid rgba(255, 255, 255, 0.08);
				border-radius: 8px;
				color: rgba(255, 255, 255, 0.85);
				color-scheme: dark;
				padding: 6px 28px 6px 12px;
				font-size: 13px;
				outline: none;
				cursor: pointer;
				transition: all 140ms ease;

				&:hover {
					background-color: rgba(255, 255, 255, 0.09);
					border-color: rgba(255, 255, 255, 0.2);
				}

				&:focus {
					border-color: var(--tag-accent, $brand-color);
					box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.2);
				}

				option {
					background: #18191f;
					color: #ffffff;
				}
			}
		}
	}

	.active-chips-bar {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 8px;
		max-width: 1400px;
		margin: 0 auto 16px auto;
		padding: 8px 12px;
		background: rgba(255, 255, 255, 0.02);
		border: 1px solid rgba(255, 255, 255, 0.05);
		border-radius: 8px;

		.chips-label {
			font-size: 12px;
			font-weight: 600;
			color: rgba(255, 255, 255, 0.4);
			text-transform: uppercase;
			letter-spacing: 0.5px;
		}

		.filter-chip {
			display: inline-flex;
			align-items: center;
			gap: 6px;
			font-size: 12px;
			background: rgba(255, 255, 255, 0.08);
			color: #ffffff;
			padding: 3px 8px;
			border-radius: 12px;

			.chip-x {
				display: flex;
				align-items: center;
				color: rgba(255, 255, 255, 0.6);
				cursor: pointer;

				&:hover {
					color: #ff6b6b;
				}
			}
		}

		.clear-all-btn {
			margin-left: auto;
			font-size: 12px;
			color: #36ba98;
			cursor: pointer;
			text-decoration: underline;

			&:hover {
				color: #5ce6c3;
			}
		}
	}

	.empty-list {
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
		padding: 60px 20px;
		max-width: 480px;
		margin: 0 auto;

		.empty-icon-wrap {
			display: flex;
			align-items: center;
			justify-content: center;
			width: 80px;
			height: 80px;
			border-radius: 50%;
			background: rgba(255, 255, 255, 0.03);
			border: 1px solid rgba(255, 255, 255, 0.06);
			margin-bottom: 20px;
			color: rgba(255, 255, 255, 0.3);
		}

		h2 {
			font-size: 20px;
			font-weight: 700;
			margin-bottom: 8px;
			color: #ffffff;
		}

		.empty-desc {
			font-size: 14px;
			color: rgba(255, 255, 255, 0.5);
			line-height: 1.5;
			margin-bottom: 20px;
		}

		.empty-actions {
			display: flex;
			gap: 10px;
		}
	}

	.tag-not-found {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		text-align: center;
		min-height: 50vh;
		gap: 16px;
		color: rgba(255, 255, 255, 0.6);

		h2 {
			color: #ffffff;
			font-size: 24px;
		}
	}

	.loader-wrap,
	.error-wrap {
		display: flex;
		justify-content: center;
		padding: 40px 0;
	}
</style>
