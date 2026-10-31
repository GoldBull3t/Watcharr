<script lang="ts">
	import {
		type Media,
		type WatchedStatus,
		type GameCompletionStatus,
	} from "@/types";
	import Activity from "@/lib/activity/Activity.svelte";
	import Title from "@/lib/content/Title.svelte";
	import Error from "@/lib/Error.svelte";
	import FollowedThoughts from "@/lib/content/FollowedThoughts.svelte";
	import { req, updateWatched } from "@/lib/util/api.js";
	import tooltip from "@/lib/actions/tooltip.js";
	import Icon from "@/lib/Icon.svelte";
	import AddToTagButton from "@/lib/tag/AddToTagButton.svelte";
	import Tag from "@/lib/tag/Tag.svelte";
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import PageBackdrop from "@/lib/generic/PageBackdrop.svelte";
	import MyReview from "@/lib/content/MyReview.svelte";
	import GamingTrackerCard from "@/lib/content/GamingTrackerCard.svelte";
	import ViewTrailerButton from "@/lib/content/ViewTrailerButton.svelte";
	import ProvidersList from "@/lib/content/ProvidersList.svelte";
	import PosterImage from "@/lib/content/PosterImage.svelte";
	import ExpandableText from "@/lib/content/ExpandableText.svelte";
	import WatchedDeleteBtn from "@/lib/content/WatchedDeleteBtn.svelte";
	import { activityRemovedHook } from "@/lib/activity/activity.js";
	import Genres from "@/lib/content/Genres.svelte";
	import SimilarContent from "@/lib/content/SimilarContent.svelte";
	import ExternalRatingsBar from "@/lib/content/ExternalRatingsBar.svelte";
	import MediaDetailSkeleton from "@/lib/content/MediaDetailSkeleton.svelte";
	import HorizontalList from "@/lib/HorizontalList.svelte";

	let { data } = $props();

	let game: Media | undefined = $state();
	let pageError: unknown | undefined = $state();
	let backdropSrc = $derived.by(() => {
		const base = "https://images.igdb.com/igdb/image/upload/t_1080p/";

		if (game?.extBackdropPath) {
			return base + game.extBackdropPath + ".jpg";
		} else if (game?.extPosterPath) {
			return base + game.extPosterPath + ".jpg";
		}
	});

	async function loadGame(signal?: AbortSignal) {
		try {
			game = undefined;
			pageError = undefined;
			if (!data.gameId) {
				return;
			}
			game = await req.get<Media>(`/game/${data.gameId}`, {
				signal,
			});
		} catch (err: unknown) {
			if (signal?.aborted) return;
			game = undefined;
			pageError = err;
		}
	}

	$effect(() => {
		const controller = new AbortController();
		if (data.gameId) {
			loadGame(controller.signal);
		}
		return () => controller.abort();
	});

	async function contentChanged(
		newStatus?: WatchedStatus,
		newRating?: number,
		newThoughts?: string,
		pinned?: boolean,
		hoursPlayed?: number,
		platform?: string,
		gameCompletionStatus?: GameCompletionStatus,
	): Promise<boolean> {
		try {
			if (!data.gameId) {
				console.error("contentChanged: no gameId");
				return false;
			}
			if (!game) {
				console.error("contentChanged: no game");
				return false;
			}
			game.watched = await updateWatched(game.watched, {
				contentId: data.gameId,
				contentType: "game",
				status: newStatus,
				rating: newRating,
				thoughts: newThoughts,
				pinned: pinned,
				hoursPlayed: hoursPlayed,
				platform: platform,
				gameCompletionStatus: gameCompletionStatus,
			});
			return true;
		} catch {
			return false;
		}
	}
	async function handleHoursChanged(hours: number): Promise<boolean> {
		return contentChanged(undefined, undefined, undefined, undefined, hours);
	}

	async function handlePlatformChanged(platform: string): Promise<boolean> {
		return contentChanged(
			undefined,
			undefined,
			undefined,
			undefined,
			undefined,
			platform,
		);
	}

	async function handleCompletionStatusChanged(
		status: GameCompletionStatus,
	): Promise<boolean> {
		return contentChanged(
			undefined,
			undefined,
			undefined,
			undefined,
			undefined,
			undefined,
			status,
		);
	}

	async function handleStartTracking(): Promise<boolean> {
		return contentChanged("WATCHING");
	}
</script>

<svelte:head>
	<title>{game?.name ? `${game.name} - ` : ""}Game</title>
</svelte:head>

{#if pageError}
	<Error
		pretty="Failed to load game!"
		error={pageError}
		onRetry={() => loadGame()}
	/>
{:else if !game}
	<MediaDetailSkeleton type="game" />
{:else if Object.keys(game).length > 0}
	{#if backdropSrc}
		<PageBackdrop src={backdropSrc} />
	{/if}
	<div>
		<div class="content">
			<div class="details-wrap">
				<div class="details-container">
					{#if game.extPosterPath}
						<PosterImage
							src={"https://images.igdb.com/igdb/image/upload/t_cover_big/" +
								game.extPosterPath +
								".jpg"}
						/>
					{/if}

					<div class="details">
						<Title
							title={game.name}
							homepage={game.homepage}
							releaseDate={game.releaseDate
								? new Date(game.releaseDate)
								: undefined}
							voteAverage={game.rating}
							voteCount={game.ratingCount}
						/>

						<span class="quick-info">
							<Genres genres={game.genres} />
							{#if game.themes && game.themes.length > 0}
								<Genres genres={game.themes} />
							{/if}
							<Genres genres={game.gameModes} />
						</span>

						<ExternalRatingsBar media={game} />

						<ExpandableText text={game.summary} style="margin-bottom: 12px;" />

						{#if game.storyline && game.storyline !== game.summary}
							<ExpandableText
								text={game.storyline}
								style="margin-bottom: 18px;"
							/>
						{/if}

						<div class="btns">
							<ViewTrailerButton videos={game.videos} mediaTitle={game.name} />
							{#if game.watched}
								<div class="other-side">
									<AddToTagButton watchedItem={game.watched} />
									<button
										onclick={() => {
											if (game?.watched?.pinned) {
												contentChanged(undefined, undefined, undefined, false);
											} else {
												contentChanged(undefined, undefined, undefined, true);
											}
										}}
										use:tooltip={{
											text: `${game.watched?.pinned ? "Unpin from" : "Pin to"} top of list`,
											pos: "bot",
										}}
									>
										<Icon i={game.watched?.pinned ? "unpin" : "pin"} wh={19} />
									</button>
									<WatchedDeleteBtn
										watchedId={game.watched.id}
										watchedEntry={game.watched}
										mediaName={game.name}
										onDelete={() => {
											if (game) {
												game.watched = undefined;
											}
										}}
										onRestore={(restored) => {
											if (game) {
												game.watched = restored;
											}
										}}
									/>
								</div>
							{/if}
						</div>

						{#if game?.watched?.tags && game.watched.tags.length > 0}
							<div class="applied-tags-row">
								<span class="tags-label">
									<Icon i="tag" wh={14} />
									<span>Tags</span>
								</span>
								<div class="tags-list">
									{#each game.watched.tags as t (t.id)}
										<Tag
											tag={t}
											size="sm"
											onClick={() => goto(resolve(`/tag/${t.id}`))}
										/>
									{/each}
								</div>
							</div>
						{/if}

						{#if game.providers}
							<ProvidersList
								providers={game.providers}
								mediaType="game"
								mediaTitle={game.name}
							/>
						{/if}
					</div>
				</div>
			</div>

			<MyReview
				watched={game.watched}
				contentTitle={game.name}
				onRatingChanged={(n) => contentChanged(undefined, n)}
				onStatusChanged={(n) => contentChanged(n)}
				onThoughtsChanged={(newThoughts) => {
					return contentChanged(undefined, undefined, newThoughts);
				}}
			/>
		</div>

		<div class="page">
			<GamingTrackerCard
				{game}
				watched={game.watched}
				onHoursChanged={handleHoursChanged}
				onPlatformChanged={handlePlatformChanged}
				onCompletionStatusChanged={handleCompletionStatusChanged}
				onStartTracking={handleStartTracking}
			/>

			{#if data.gameId}
				<FollowedThoughts mediaType="game" mediaId={data.gameId} />
			{/if}

			{#if game.screenshots && game.screenshots.length > 0}
				<HorizontalList title="Screenshots">
					{#each game.screenshots.slice(0, 16) as shot (shot)}
						<a
							href={`https://images.igdb.com/igdb/image/upload/t_1080p/${shot}.jpg`}
							target="_blank"
							rel="noreferrer noopener external"
							class="game-screenshot-card"
						>
							<img
								src={`https://images.igdb.com/igdb/image/upload/t_screenshot_med/${shot}.jpg`}
								alt={game.name}
								loading="lazy"
							/>
						</a>
					{/each}
				</HorizontalList>
			{/if}

			{#if game.similar && game.similar?.length > 0}
				<SimilarContent similar={game.similar} />
			{/if}

			{#if game.watched}
				<Activity
					activity={game.watched.activity}
					onRemoved={(a) => activityRemovedHook(game?.watched, a)}
				/>
			{/if}
		</div>
	</div>
{:else}
	<Error error="Game not found" pretty="Game not found" />
{/if}

<style lang="scss">
	@use "../../../../lib/content/page.scss";

	.content {
		position: relative;
		color: white;

		.details-container .details {
			.quick-info {
				display: flex;
				gap: 10px;
				margin-bottom: 8px;
			}

			.btns {
				display: flex;
				flex-flow: row;
				flex-wrap: wrap;
				gap: 8px;
				margin-top: auto;

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
				border-radius: 12px;

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
		align-items: center;
		margin-left: auto;
		margin-right: auto;
		gap: 30px;
		padding: 20px 50px;
		max-width: 1200px;

		@media screen and (max-width: 500px) {
			padding: 20px;
		}

		.game-screenshot-card {
			display: block;
			flex-shrink: 0;
			width: 280px;
			height: 158px;
			border-radius: 12px;
			overflow: hidden;
			border: 1px solid rgba(255, 255, 255, 0.12);
			background: rgba(15, 23, 42, 0.6);
			transition:
				transform 160ms ease,
				border-color 160ms ease;

			&:hover {
				transform: translateY(-2px) scale(1.02);
				border-color: rgba(255, 255, 255, 0.28);
			}

			img {
				width: 100%;
				height: 100%;
				object-fit: cover;
			}
		}
	}
</style>
