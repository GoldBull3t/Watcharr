<script lang="ts">
	import PersonPoster from "@/lib/poster/PersonPoster.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import HorizontalList from "@/lib/HorizontalList.svelte";
	import { contentExistsOnJellyfin, req, updateWatched } from "@/lib/util/api";
	import { notify } from "@/lib/util/notify";
	import { store } from "@/store.svelte";
	import type {
		Media,
		TMDBContentCredits,
		TMDBContentCreditsCrew,
		WatchedStatus,
	} from "@/types";
	import { getTopCrew } from "@/lib/util/helpers.js";
	import Activity from "@/lib/activity/Activity.svelte";
	import Title from "@/lib/content/Title.svelte";
	import ProvidersList from "@/lib/content/ProvidersList.svelte";
	import Icon from "@/lib/Icon.svelte";
	import SimilarContent from "@/lib/content/SimilarContent.svelte";
	import RequestMovie from "@/lib/request/RequestMovie.svelte";
	import Error from "@/lib/Error.svelte";
	import FollowedThoughts from "@/lib/content/FollowedThoughts.svelte";
	import ArrRequestButton from "@/lib/request/ArrRequestButton.svelte";
	import tooltip from "@/lib/actions/tooltip.js";
	import AddToTagButton from "@/lib/tag/AddToTagButton.svelte";
	import Tag from "@/lib/tag/Tag.svelte";
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import PageBackdrop from "@/lib/generic/PageBackdrop.svelte";
	import MyReview from "@/lib/content/MyReview.svelte";
	import ViewTrailerButton from "@/lib/content/ViewTrailerButton.svelte";
	import PosterImage from "@/lib/content/PosterImage.svelte";
	import ExpandableText from "@/lib/content/ExpandableText.svelte";
	import WatchedDeleteBtn from "@/lib/content/WatchedDeleteBtn.svelte";
	import TopCrewList from "@/lib/content/TopCrewList.svelte";
	import { activityRemovedHook } from "@/lib/activity/activity.js";
	import Genres from "@/lib/content/Genres.svelte";
	import MovieCollection from "@/lib/content/MovieCollection.svelte";
	import MediaLightbox from "@/lib/content/MediaLightbox.svelte";
	import FranchiseBadgeLink from "@/lib/franchise/FranchiseBadgeLink.svelte";
	import ExternalRatingsBar from "@/lib/content/ExternalRatingsBar.svelte";
	import MediaDetailSkeleton from "@/lib/content/MediaDetailSkeleton.svelte";
	import { untrack } from "svelte";

	let { data } = $props();

	let requestModalShown = $state(false);
	let jellyfinUrl: string | undefined = $state();
	let arrRequestButtonComp: ArrRequestButton | undefined = $state();
	let movie: Media | undefined = $state();
	let pageError: unknown | undefined = $state();
	let galleryOpen = $state(false);
	let isRefreshing = $state(false);

	async function refreshMetadata() {
		if (!data.movieId || isRefreshing) return;
		isRefreshing = true;
		const nid = notify({ type: "loading", text: "Refreshing metadata..." });
		try {
			const region = untrack(() => store.userSettings?.country);
			const resp = await req.post<Media>(
				`/content/movie/${data.movieId}/refresh`,
				{},
				{ params: { region } },
			);
			if (resp) {
				movie = resp;
			}
			notify({ id: nid, type: "success", text: "Metadata updated!" });
		} catch (err) {
			console.error("refreshMetadata failed:", err);
			notify({ id: nid, type: "error", text: "Failed to refresh metadata" });
		} finally {
			isRefreshing = false;
		}
	}

	async function loadMovie(signal?: AbortSignal) {
		try {
			movie = undefined;
			pageError = undefined;
			if (!data.movieId) {
				return;
			}
			const region = untrack(() => store.userSettings?.country);
			const resp = await req.get<Media>(`/content/movie/${data.movieId}`, {
				params: { region },
				signal,
			});
			if (signal?.aborted) return;
			if (resp) {
				if (resp.name && resp.ids.tmdb) {
					contentExistsOnJellyfin("movie", resp.name, resp.ids.tmdb).then(
						(j) => {
							if (!signal?.aborted && j?.hasContent && j?.url !== "") {
								jellyfinUrl = j.url;
							}
						},
					);
				}
				movie = resp;
			} else {
				movie = undefined;
			}
		} catch (err: unknown) {
			if (signal?.aborted) return;
			movie = undefined;
			pageError = err;
		}
	}

	$effect(() => {
		const controller = new AbortController();
		if (data.movieId) {
			loadMovie(controller.signal);
		}
		return () => controller.abort();
	});

	async function getMovieCredits() {
		const credits = await req.get<
			TMDBContentCredits & { topCrew: TMDBContentCreditsCrew[] }
		>(`/content/movie/${data.movieId}/credits`);
		if (credits.crew?.length > 0) {
			credits.topCrew = getTopCrew(credits.crew);
		}
		return credits;
	}

	async function contentChanged(
		newStatus?: WatchedStatus,
		newRating?: number,
		newThoughts?: string,
		pinned?: boolean,
	): Promise<boolean> {
		try {
			if (!data.movieId) {
				console.error("contentChanged: no movieId");
				return false;
			}
			if (!movie) {
				console.error("contentChanged: no movie");
				return false;
			}
			movie.watched = await updateWatched(movie.watched, {
				contentId: data.movieId,
				contentType: "movie",
				status: newStatus,
				rating: newRating,
				thoughts: newThoughts,
				pinned: pinned,
			});
			return true;
		} catch {
			return false;
		}
	}
</script>

<svelte:head>
	<title>{movie?.name ? `${movie.name} - ` : ""}Movie</title>
</svelte:head>

{#if pageError}
	<Error
		pretty="Failed to load movie!"
		error={pageError}
		onRetry={() => loadMovie()}
	/>
{:else if !movie}
	<MediaDetailSkeleton type="movie" />
{:else if Object.keys(movie).length > 0}
	{#if movie?.extBackdropPath}
		<PageBackdrop
			src={"https://www.themoviedb.org/t/p/w1920_and_h800_multi_faces" +
				movie.extBackdropPath}
		/>
	{/if}
	<div>
		<div class="content">
			<div class="details-wrap">
				<div class="details-container">
					{#if movie.extPosterPath}
						<PosterImage
							src={"https://image.tmdb.org/t/p/w500" + movie.extPosterPath}
							alt={movie.name}
							onclick={() => (galleryOpen = true)}
						/>
					{/if}

					<div class="details">
						<Title
							title={movie.name}
							homepage={movie.homepage}
							releaseDate={movie.releaseDate
								? new Date(movie.releaseDate)
								: undefined}
							voteAverage={movie.rating}
							voteCount={movie.ratingCount}
						/>

						<span class="quick-info">
							{#if movie.certification}
								<span class="runtime-pill cert-pill">
									{movie.certification}
								</span>
							{/if}
							{#if movie.runtime}
								<span class="runtime-pill">
									<Icon i="clock" wh={13} />
									{movie.runtime} min
								</span>
							{/if}

							<Genres genres={movie.genres} />
						</span>

						<ExternalRatingsBar media={movie} />

						<ExpandableText text={movie.summary} style="margin-bottom: 18px;" />

						<div class="btns">
							<ViewTrailerButton
								videos={movie.videos}
								backdropPath={movie.extBackdropPath}
								posterPath={movie.extPosterPath}
								mediaTitle={movie.name}
							/>
							{#if jellyfinUrl}
								<a
									class="btn"
									href={jellyfinUrl}
									rel="external"
									target="_blank"
								>
									{#if localStorage.getItem("useEmby")}
										<Icon i="emby" wh={14} />Play On Emby
									{:else}
										<Icon i="jellyfin" wh={14} />Play On Jellyfin
									{/if}
								</a>
							{/if}
							{#if (store.serverFeatures?.radarr || store.serverFeatures?.overseerr) && data.movieId}
								<ArrRequestButton
									type="movie"
									tmdbId={data.movieId}
									openRequestModal={() =>
										(requestModalShown = !requestModalShown)}
									onRequested={() => {
										if (!movie?.watched) {
											contentChanged("PLANNED");
										}
									}}
									bind:this={arrRequestButtonComp}
								/>
							{/if}
							<div class="other-side">
								<button
									onclick={refreshMetadata}
									disabled={isRefreshing}
									use:tooltip={{
										text: "Refresh TMDB metadata",
										pos: "bot",
									}}
								>
									<Icon i="refresh" wh={18} />
								</button>
								{#if movie.watched}
									<AddToTagButton watchedItem={movie.watched} />
									<button
										onclick={() => {
											if (movie?.watched?.pinned) {
												contentChanged(undefined, undefined, undefined, false);
											} else {
												contentChanged(undefined, undefined, undefined, true);
											}
										}}
										use:tooltip={{
											text: `${movie.watched?.pinned ? "Unpin from" : "Pin to"} top of list`,
											pos: "bot",
										}}
									>
										<Icon i={movie.watched?.pinned ? "unpin" : "pin"} wh={19} />
									</button>
									<WatchedDeleteBtn
										watchedId={movie.watched.id}
										watchedEntry={movie.watched}
										mediaName={movie.name}
										onDelete={() => {
											if (movie) {
												movie.watched = undefined;
											}
										}}
										onRestore={(restored) => {
											if (movie) {
												movie.watched = restored;
											}
										}}
									/>
								{/if}
							</div>
						</div>

						{#if movie?.watched?.tags && movie.watched.tags.length > 0}
							<div class="applied-tags-row">
								<span class="tags-label">
									<Icon i="tag" wh={14} />
									<span>Tags</span>
								</span>
								<div class="tags-list">
									{#each movie.watched.tags as t (t.id)}
										<Tag
											tag={t}
											size="sm"
											onClick={() => goto(resolve(`/tag/${t.id}`))}
										/>
									{/each}
								</div>
							</div>
						{/if}

						<ProvidersList
							providers={movie.providers}
							fullListLink={movie.providersFullListLink}
							fullListLinkText="JustWatch"
							mediaType="movie"
							mediaId={movie.ids.tmdb}
							mediaTitle={movie.name}
						/>
					</div>

					<MyReview
						watched={movie.watched}
						contentTitle={movie.name}
						onRatingChanged={(n) => contentChanged(undefined, n)}
						onStatusChanged={(n) => contentChanged(n)}
						onThoughtsChanged={(newThoughts) => {
							return contentChanged(undefined, undefined, newThoughts);
						}}
					/>
				</div>
			</div>
		</div>

		{#if requestModalShown}
			<RequestMovie
				content={movie}
				onClose={(reqResp) => {
					requestModalShown = false;
					if (reqResp) {
						arrRequestButtonComp?.setExistingRequest(reqResp);
					}
				}}
			/>
		{/if}

		<div class="page">
			{#if data.movieId && movie}
				<FranchiseBadgeLink
					mediaType="movie"
					tmdbId={data.movieId}
					collectionId={movie.collection?.id}
					collectionName={movie.collection?.name}
					hasCollectionBlock={Boolean(movie.collection?.id)}
				/>
			{/if}

			{#if data.movieId}
				<FollowedThoughts mediaType="movie" mediaId={data.movieId} />
			{/if}

			{#if movie.collection?.id}
				<MovieCollection
					collectionId={movie.collection.id}
					currentMovieId={data.movieId}
				/>
			{/if}

			{#await getMovieCredits()}
				<Spinner />
			{:then credits}
				{#if credits.topCrew?.length > 0}
					<TopCrewList topCrew={credits.topCrew} />
				{/if}

				{#if credits.cast?.length > 0}
					<HorizontalList title="Cast">
						{#each credits.cast?.slice(0, 50) as cast (cast.credit_id)}
							<PersonPoster
								id={cast.id}
								name={cast.name}
								path={cast.profile_path}
								role={cast.character}
								zoomOnHover={false}
							/>
						{/each}
					</HorizontalList>
				{/if}
			{:catch err}
				<Error error={err} pretty="Failed to load cast!" />
			{/await}

			{#if movie.similar}
				<SimilarContent similar={movie.similar} />
			{/if}

			{#if movie.watched}
				<Activity
					activity={movie.watched.activity}
					onRemoved={(a) => activityRemovedHook(movie?.watched, a)}
				/>
			{/if}
		</div>
	</div>

	{#if galleryOpen}
		<MediaLightbox
			videos={movie.videos}
			backdropPath={movie.extBackdropPath}
			posterPath={movie.extPosterPath}
			mediaTitle={movie.name}
			initialType="poster"
			onclose={() => (galleryOpen = false)}
		/>
	{/if}
{:else}
	Movie not found
{/if}

<style lang="scss">
	@use "../../../../lib/content/page.scss";

	.content {
		position: relative;
		color: white;

		.details-container .details {
			.quick-info {
				display: flex;
				align-items: center;
				flex-wrap: wrap;
				gap: 10px;
				margin-bottom: 8px;

				.runtime-pill {
					display: inline-flex;
					align-items: center;
					gap: 5px;
					font-size: 12.5px;
					font-weight: 600;
					color: rgba(255, 255, 255, 0.75);
					background: rgba(255, 255, 255, 0.06);
					border: 1px solid rgba(255, 255, 255, 0.12);
					border-radius: 6px;
					padding: 3px 8px;
				}
			}

			.btns {
				display: flex;
				flex-flow: row;
				flex-wrap: wrap;
				gap: 8px;
				margin-top: auto;

				a.btn,
				button {
					max-width: fit-content;
					overflow: hidden;
					animation: 50ms cubic-bezier(0.86, 0, 0.07, 1) forwards otherbtn;
					white-space: nowrap;
					gap: 6px;
					justify-content: flex-start;
					font-size: 14px;

					@keyframes otherbtn {
						from {
							width: 0px;
						}
						to {
							width: 100%;
						}
					}
				}

				.other-side {
					display: flex;
					flex-flow: row;
					gap: 8px;

					@media screen and (min-width: 900px) {
						margin-left: auto;
					}
				}
			}

			.applied-tags-row {
				display: flex;
				align-items: center;
				flex-wrap: wrap;
				gap: 8px;
				margin: 10px 0;
				padding: 8px 12px;
				background: rgba(255, 255, 255, 0.03);
				border: 1px solid rgba(255, 255, 255, 0.07);
				border-radius: 8px;

				.tags-label {
					display: inline-flex;
					align-items: center;
					gap: 5px;
					font-size: 11px;
					font-weight: 700;
					color: rgba(255, 255, 255, 0.5);
					text-transform: uppercase;
					letter-spacing: 0.6px;
				}

				.tags-list {
					display: inline-flex;
					flex-wrap: wrap;
					align-items: center;
					gap: 6px;
				}
			}
		}
	}

	.page {
		display: flex;
		flex-flow: column;
		align-items: stretch;
		width: 100%;
		max-width: 1360px;
		margin-inline: auto;
		gap: 28px;
		padding: 12px 32px 64px;

		@media screen and (max-width: 1120px) {
			padding: 12px 24px 48px;
		}

		@media screen and (max-width: 680px) {
			padding: 12px 16px 40px;
			gap: 22px;
		}
	}
</style>
