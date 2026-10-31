<script lang="ts">
	import Activity from "@/lib/activity/Activity.svelte";
	import Error from "@/lib/Error.svelte";
	import HorizontalList from "@/lib/HorizontalList.svelte";
	import Icon from "@/lib/Icon.svelte";
	import PersonPoster from "@/lib/poster/PersonPoster.svelte";
	import SeasonsList from "@/lib/season/SeasonsList.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import ProvidersList from "@/lib/content/ProvidersList.svelte";
	import SimilarContent from "@/lib/content/SimilarContent.svelte";
	import Title from "@/lib/content/Title.svelte";
	import {
		contentExistsOnJellyfin,
		req,
		updateWatched,
		type UpdateWatchedOptions,
	} from "@/lib/util/api";
	import { notify } from "@/lib/util/notify";
	import { getTopCrew } from "@/lib/util/helpers.js";
	import { store } from "@/store.svelte.js";
	import type {
		Media,
		TMDBContentCredits,
		TMDBContentCreditsCrew,
		WatchedStatus,
	} from "@/types";
	import RequestShow from "@/lib/request/RequestShow.svelte";
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
	import CountAsPlayModal from "@/lib/watched/CountAsPlayModal.svelte";
	import { createSignal, type Signal } from "@/lib/util/signal.js";
	import Genres from "@/lib/content/Genres.svelte";
	import { MediaStatusShow } from "@/lib/types/mediaStatus.js";
	import SeriesProgress from "@/lib/season/SeriesProgress.svelte";
	import NextEpisodeCard from "@/lib/season/NextEpisodeCard.svelte";
	import MediaLightbox from "@/lib/content/MediaLightbox.svelte";
	import CancellationGuardBadge from "@/lib/content/CancellationGuardBadge.svelte";
	import CliffhangerWarningModal from "@/lib/content/CliffhangerWarningModal.svelte";
	import FranchiseBadgeLink from "@/lib/franchise/FranchiseBadgeLink.svelte";
	import ExternalRatingsBar from "@/lib/content/ExternalRatingsBar.svelte";
	import MediaDetailSkeleton from "@/lib/content/MediaDetailSkeleton.svelte";
	import { untrack } from "svelte";

	let { data } = $props();

	let requestModalShown = $state(false);
	let jellyfinUrl: string | undefined = $state();
	let arrRequestButtonComp: ArrRequestButton | undefined = $state();
	let show: Media | undefined = $state();
	let pageError: unknown | undefined = $state();
	let countAsPlayModalSignal: Signal<boolean> | undefined = $state();
	let galleryOpen = $state(false);
	let pendingStatusChange:
		| {
				status?: WatchedStatus;
				rating?: number;
				thoughts?: string;
				pinned?: boolean;
		  }
		| undefined = $state();
	let cliffhangerWarningShown = $state(false);
	let isRefreshing = $state(false);

	async function refreshMetadata() {
		if (!data.tvId || isRefreshing) return;
		isRefreshing = true;
		const nid = notify({
			type: "loading",
			text: "Refreshing show & episodes...",
		});
		try {
			const region = untrack(() => store.userSettings?.country);
			const resp = await req.post<Media>(
				`/content/tv/${data.tvId}/refresh`,
				{},
				{ params: { region } },
			);
			if (resp) {
				show = resp;
			}
			notify({
				id: nid,
				type: "success",
				text: "Show & episodes cache refreshed!",
			});
		} catch (err) {
			console.error("refreshMetadata failed:", err);
			notify({ id: nid, type: "error", text: "Failed to refresh metadata" });
		} finally {
			isRefreshing = false;
		}
	}

	async function loadShow(signal?: AbortSignal) {
		try {
			show = undefined;
			pageError = undefined;
			if (!data.tvId) {
				return;
			}
			const region = untrack(() => store.userSettings?.country);
			const resp = await req.get<Media>(`/content/tv/${data.tvId}`, {
				params: { region },
				signal,
			});
			if (signal?.aborted) return;
			if (resp) {
				if (resp.name && resp.ids.tmdb) {
					contentExistsOnJellyfin("tv", resp.name, resp.ids.tmdb).then((j) => {
						if (!signal?.aborted && j?.hasContent && j?.url !== "") {
							jellyfinUrl = j.url;
						}
					});
				}
				show = resp;
			} else {
				show = undefined;
			}
		} catch (err: unknown) {
			if (signal?.aborted) return;
			show = undefined;
			pageError = err;
		}
	}

	$effect(() => {
		const controller = new AbortController();
		if (data.tvId) {
			loadShow(controller.signal);
		}
		return () => controller.abort();
	});

	async function getTvCredits() {
		const credits = await req.get<
			TMDBContentCredits & { topCrew: TMDBContentCreditsCrew[] }
		>(`/content/tv/${data.tvId}/credits`);
		if (credits.crew?.length > 0) {
			credits.topCrew = getTopCrew(credits.crew);
		}
		return credits;
	}

	async function confirmCliffhangerAdd() {
		cliffhangerWarningShown = false;
		if (pendingStatusChange) {
			const { status, rating, thoughts, pinned } = pendingStatusChange;
			pendingStatusChange = undefined;
			await executeContentChange(status, rating, thoughts, pinned);
		}
	}

	async function executeContentChange(
		newStatus?: WatchedStatus,
		newRating?: number,
		newThoughts?: string,
		pinned?: boolean,
	): Promise<boolean> {
		try {
			if (!data.tvId) {
				console.error("contentChanged: no tvId");
				return false;
			}
			if (!show) {
				console.error("contentChanged: no show");
				return false;
			}
			const reqOpts: UpdateWatchedOptions = {
				contentId: data.tvId,
				contentType: "tv",
				status: newStatus,
				rating: newRating,
				thoughts: newThoughts,
				pinned: pinned,
			};
			// Series differ from other media in that people are likely to go
			// back out of the 'FINISHED' state when a new season releases
			// while they watch it, then go back to 'FINISHED' after only
			// watching the new season. Because of that use case, for series,
			// we will ask the user if any 'FINISHED' statuses set when plays>1
			// should count as a play..
			if (show.watched?.plays && newStatus == "FINISHED") {
				countAsPlayModalSignal = createSignal<boolean>();
				const allow = await countAsPlayModalSignal.promise;
				countAsPlayModalSignal = undefined;
				if (!allow) {
					reqOpts.letCountAsPlay = false;
				}
			}
			show.watched = await updateWatched(show.watched, reqOpts);
			if (show?.watched && newStatus) {
				if (newStatus === "FINISHED") {
					if (show.watched.watchedSeasons) {
						for (const s of show.watched.watchedSeasons) {
							s.status = "FINISHED";
						}
					}
					if (show.watched.watchedEpisodes) {
						for (const e of show.watched.watchedEpisodes) {
							e.status = "FINISHED";
						}
					}
				}
				show.watched = {
					...show.watched,
					watchedEpisodes: [...(show.watched.watchedEpisodes || [])],
					watchedSeasons: [...(show.watched.watchedSeasons || [])],
				};
			}
			return true;
		} catch {
			return false;
		}
	}

	async function contentChanged(
		newStatus?: WatchedStatus,
		newRating?: number,
		newThoughts?: string,
		pinned?: boolean,
	): Promise<boolean> {
		if (
			!show?.watched &&
			(show?.cancellation?.hasCliffhanger ||
				show?.cancellation?.status === "cliffhanger") &&
			newStatus
		) {
			pendingStatusChange = {
				status: newStatus,
				rating: newRating,
				thoughts: newThoughts,
				pinned: pinned,
			};
			cliffhangerWarningShown = true;
			return false;
		}
		return executeContentChange(newStatus, newRating, newThoughts, pinned);
	}
</script>

<svelte:head>
	<title>{show?.name ? `${show.name} - ` : ""}Show</title>
</svelte:head>

{#if pageError}
	<Error
		pretty="Failed to load tv show!"
		error={pageError}
		onRetry={() => loadShow()}
	/>
{:else if !show}
	<MediaDetailSkeleton type="tv" />
{:else if Object.keys(show).length > 0}
	{#if show?.extBackdropPath}
		<PageBackdrop
			src={"https://www.themoviedb.org/t/p/w1920_and_h800_multi_faces" +
				show.extBackdropPath}
		/>
	{/if}
	<div>
		<div class="content">
			<div class="details-wrap">
				<div class="details-container">
					{#if show.extPosterPath}
						<PosterImage
							src={"https://image.tmdb.org/t/p/w500" + show.extPosterPath}
							alt={show.name}
							onclick={() => (galleryOpen = true)}
						/>
					{/if}

					<div class="details">
						<Title
							title={show.name}
							homepage={show.homepage}
							releaseDate={show.releaseDate
								? new Date(show.releaseDate)
								: undefined}
							endDate={(show.status === MediaStatusShow.Ended ||
								show.status === MediaStatusShow.Canceled) &&
							show.releaseDateLast
								? new Date(show.releaseDateLast)
								: undefined}
							voteAverage={show.rating}
							voteCount={show.ratingCount}
						/>

						<CancellationGuardBadge
							showId={data.tvId}
							status={show.status}
							cancellation={show.cancellation}
							onOverrideChanged={(updated) => {
								if (show) {
									show.cancellation = updated;
								}
							}}
						/>

						<span class="quick-info">
							{#if show.certification}
								<span class="runtime-pill cert-pill">
									{show.certification}
								</span>
							{/if}
							{#if show.runtime}
								<span class="runtime-pill">
									<Icon i="clock" wh={13} />
									~{show.runtime} min/ep
								</span>
							{/if}
							<Genres genres={show.genres} />
						</span>

						<ExternalRatingsBar media={show} />

						<ExpandableText text={show.summary} style="margin-bottom: 18px;" />

						<div class="btns">
							<ViewTrailerButton
								videos={show.videos}
								backdropPath={show.extBackdropPath}
								posterPath={show.extPosterPath}
								mediaTitle={show.name}
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
							{#if (store.serverFeatures?.sonarr || store.serverFeatures?.overseerr) && data.tvId}
								<ArrRequestButton
									type="tv"
									tmdbId={data.tvId}
									openRequestModal={() =>
										(requestModalShown = !requestModalShown)}
									onRequested={() => {
										if (!show?.watched) {
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
										text: "Refresh TMDB metadata & episodes",
										pos: "bot",
									}}
								>
									<Icon i="refresh" wh={18} />
								</button>
								{#if show.watched}
									<AddToTagButton watchedItem={show.watched} />
									<button
										onclick={() => {
											if (show?.watched?.pinned) {
												contentChanged(undefined, undefined, undefined, false);
											} else {
												contentChanged(undefined, undefined, undefined, true);
											}
										}}
										use:tooltip={{
											text: `${show.watched?.pinned ? "Unpin from" : "Pin to"} top of list`,
											pos: "bot",
										}}
									>
										<Icon i={show.watched?.pinned ? "unpin" : "pin"} wh={19} />
									</button>
									<WatchedDeleteBtn
										watchedId={show.watched.id}
										watchedEntry={show.watched}
										mediaName={show.name}
										onDelete={() => {
											if (show) {
												show.watched = undefined;
											}
										}}
										onRestore={(restored) => {
											if (show) {
												show.watched = restored;
											}
										}}
									/>
								{/if}
							</div>
						</div>

						{#if show?.watched?.tags && show.watched.tags.length > 0}
							<div class="applied-tags-row">
								<span class="tags-label">
									<Icon i="tag" wh={14} />
									<span>Tags</span>
								</span>
								<div class="tags-list">
									{#each show.watched.tags as t (t.id)}
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
							providers={show.providers}
							fullListLink={show.providersFullListLink}
							fullListLinkText="JustWatch"
							mediaType="tv"
							mediaId={show.ids.tmdb}
							mediaTitle={show.name}
						/>
					</div>

					<MyReview
						watched={show.watched}
						contentTitle={show.name}
						onRatingChanged={(n) => contentChanged(undefined, n)}
						onStatusChanged={(n) => contentChanged(n)}
						onThoughtsChanged={(newThoughts) => {
							return contentChanged(undefined, undefined, newThoughts);
						}}
					/>
				</div>
			</div>
		</div>

		{#if countAsPlayModalSignal}
			<CountAsPlayModal onDecision={countAsPlayModalSignal} />
		{/if}

		{#if requestModalShown}
			<RequestShow
				content={show}
				onClose={(reqResp) => {
					requestModalShown = false;
					if (reqResp) {
						arrRequestButtonComp?.setExistingRequest(reqResp);
					}
				}}
			/>
		{/if}

		{#if cliffhangerWarningShown}
			<CliffhangerWarningModal
				showName={show.name}
				warning={show.cancellation?.warning}
				description={show.cancellation?.description}
				onConfirm={confirmCliffhangerAdd}
				onCancel={() => {
					cliffhangerWarningShown = false;
					pendingStatusChange = undefined;
				}}
			/>
		{/if}

		<div class="page">
			{#if show.watched && show.seasons && show.seasons.length > 0 && data.tvId}
				<NextEpisodeCard
					tvId={data.tvId}
					seasons={show.seasons}
					watchedItem={show.watched}
					onEpisodeWatched={() => {
						if (show?.watched) {
							show.watched = {
								...show.watched,
								watchedEpisodes: [...(show.watched.watchedEpisodes || [])],
								watchedSeasons: [...(show.watched.watchedSeasons || [])],
							};
						}
					}}
				/>
			{/if}

			{#if data.tvId}
				<FranchiseBadgeLink mediaType="tv" tmdbId={data.tvId} />
				<FollowedThoughts mediaType="tv" mediaId={data.tvId} />
			{/if}

			{#if show.seasons && show.seasons.length > 0}
				{#if show.watched}
					<SeriesProgress seasons={show.seasons} watchedItem={show.watched} />
				{/if}

				{#if data?.tvId}
					<SeasonsList
						tvId={data.tvId}
						seasons={show.seasons}
						watchedItem={show.watched}
						lastViewedSeason={show.watched?.lastViewedSeason}
						lastViewedSeasonChanged={(wid, lvs) => {
							if (show?.watched && show.watched.id === wid) {
								show.watched.lastViewedSeason = lvs;
							}
						}}
						onStatusChanged={() => {
							if (show?.watched) {
								show.watched = {
									...show.watched,
									watchedEpisodes: [...(show.watched.watchedEpisodes || [])],
									watchedSeasons: [...(show.watched.watchedSeasons || [])],
								};
							}
						}}
					/>
				{/if}
			{/if}

			{#await getTvCredits()}
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

			{#if show.similar}
				<SimilarContent similar={show.similar} />
			{/if}

			{#if show.watched}
				<Activity
					activity={show.watched.activity}
					onRemoved={(a) => activityRemovedHook(show?.watched, a)}
				/>
			{/if}
		</div>
	</div>

	{#if galleryOpen}
		<MediaLightbox
			videos={show.videos}
			backdropPath={show.extBackdropPath}
			posterPath={show.extPosterPath}
			mediaTitle={show.name}
			initialType="poster"
			onclose={() => (galleryOpen = false)}
		/>
	{/if}
{:else}
	Show not found
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
				margin: 12px 0;
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
