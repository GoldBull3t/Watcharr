<script lang="ts">
	import type {
		MediaSeason,
		TMDBSeasonDetails,
		TMDBSeasonDetailsEpisode,
		Watched,
	} from "@/types";
	import { req } from "@/lib/util/api";
	import { updateWatchedEpisode } from "./api";
	import Icon from "@/lib/Icon.svelte";
	import { toRelativeDate } from "@/lib/util/helpers";
	import Spinner from "@/lib/Spinner.svelte";
	import RelativeDateBadge from "@/lib/content/RelativeDateBadge.svelte";

	interface Props {
		tvId: number;
		seasons: MediaSeason[];
		watchedItem: Watched | undefined;
		onEpisodeWatched?: () => void;
	}

	let { tvId, seasons, watchedItem, onEpisodeWatched }: Props = $props();

	let seasonCache: Record<number, TMDBSeasonDetails> = $state({});
	let loadingEpisode: boolean = $state(false);
	let markingWatched: boolean = $state(false);
	let currentEpisodeDetails: TMDBSeasonDetailsEpisode | undefined = $state();

	// Calculate target next episode
	const targetNext = $derived.by(() => {
		if (!seasons || seasons.length === 0) return undefined;
		const regularSeasons = seasons
			.filter((s) => s.number > 0)
			.sort((a, b) => a.number - b.number);
		if (regularSeasons.length === 0) return undefined;

		const finished = (watchedItem?.watchedEpisodes || [])
			.filter((we) => we.seasonNumber > 0 && we.status === "FINISHED")
			.sort((a, b) => {
				if (a.seasonNumber !== b.seasonNumber) {
					return b.seasonNumber - a.seasonNumber;
				}
				return b.episodeNumber - a.episodeNumber;
			});

		// If no episodes finished yet, start at season 1, episode 1
		if (finished.length === 0) {
			const firstSeason = regularSeasons[0];
			if (firstSeason && firstSeason.episodeCount > 0) {
				return {
					seasonNumber: firstSeason.number,
					episodeNumber: 1,
					seasonName: firstSeason.name || `Season ${firstSeason.number}`,
				};
			}
			return undefined;
		}

		const latest = finished[0];
		const currSeason = regularSeasons.find(
			(s) => s.number === latest.seasonNumber,
		);

		// If more episodes in the current season
		if (currSeason && latest.episodeNumber < currSeason.episodeCount) {
			return {
				seasonNumber: latest.seasonNumber,
				episodeNumber: latest.episodeNumber + 1,
				seasonName: currSeason.name || `Season ${latest.seasonNumber}`,
			};
		}

		// Move to the next season
		const nextSeason = regularSeasons.find(
			(s) => s.number > latest.seasonNumber,
		);
		if (nextSeason && nextSeason.episodeCount > 0) {
			return {
				seasonNumber: nextSeason.number,
				episodeNumber: 1,
				seasonName: nextSeason.name || `Season ${nextSeason.number}`,
			};
		}

		// All caught up!
		return null;
	});

	// Fetch episode details whenever targetNext changes
	$effect(() => {
		const controller = new AbortController();
		if (
			targetNext &&
			targetNext.seasonNumber !== undefined &&
			targetNext.episodeNumber !== undefined
		) {
			const sNum = targetNext.seasonNumber;
			const epNum = targetNext.episodeNumber;

			if (seasonCache[sNum]) {
				const ep = seasonCache[sNum].episodes?.find(
					(e) => e.episode_number === epNum,
				);
				currentEpisodeDetails = ep;
			} else {
				loadingEpisode = true;
				req
					.get<TMDBSeasonDetails>(`/content/tv/${tvId}/season/${sNum}`, {
						signal: controller.signal,
					})
					.then((res) => {
						if (!controller.signal.aborted && res) {
							seasonCache[sNum] = res;
							const ep = res.episodes?.find((e) => e.episode_number === epNum);
							currentEpisodeDetails = ep;
						}
					})
					.catch((err) => {
						if (controller.signal.aborted) return;
						console.warn("NextEpisodeCard: Failed to load season:", err);
						currentEpisodeDetails = undefined;
					})
					.finally(() => {
						if (!controller.signal.aborted) {
							loadingEpisode = false;
						}
					});
			}
		} else {
			currentEpisodeDetails = undefined;
		}
		return () => controller.abort();
	});

	const airDate = $derived.by(() => {
		if (!currentEpisodeDetails?.air_date) return undefined;
		const d = new Date(currentEpisodeDetails.air_date);
		return isNaN(d.getTime()) ? undefined : d;
	});

	const isUnaired = $derived.by(() => {
		if (!airDate) return false;
		return airDate.getTime() > Date.now();
	});

	async function markAsWatched() {
		if (!watchedItem || !targetNext || markingWatched) return;
		markingWatched = true;
		try {
			await updateWatchedEpisode(
				watchedItem,
				targetNext.seasonNumber,
				targetNext.episodeNumber,
				{ status: "FINISHED" },
			);
			onEpisodeWatched?.();
		} catch (err) {
			console.error("markAsWatched: Failed:", err);
		} finally {
			markingWatched = false;
		}
	}
</script>

{#if targetNext === null}
	<div class="next-episode-card caught-up">
		<div class="caught-up-content">
			<span class="caught-up-icon">🎉</span>
			<div>
				<strong>All caught up!</strong>
				<p>You have finished all available episodes for this series.</p>
			</div>
		</div>
	</div>
{:else if targetNext}
	<div class="next-episode-card">
		<div class="card-header">
			<div class="header-tag">
				<span class="pulsing-dot"></span>
				<span class="tag-title">Next Up To Watch</span>
			</div>
			<span class="season-badge">
				S{targetNext.seasonNumber} · E{targetNext.episodeNumber}
			</span>
		</div>

		{#if loadingEpisode && !currentEpisodeDetails}
			<div class="card-loading">
				<Spinner />
			</div>
		{:else}
			<div class="card-body">
				<div class="thumb-wrap">
					{#if currentEpisodeDetails?.still_path}
						<img
							src={`https://www.themoviedb.org/t/p/w500/${currentEpisodeDetails.still_path}`}
							alt={currentEpisodeDetails.name || "Episode thumbnail"}
							loading="lazy"
							decoding="async"
						/>
					{:else}
						<div class="thumb-fallback">
							<Icon i="film" wh={28} />
						</div>
					{/if}
				</div>

				<div class="info-wrap">
					<div class="title-row">
						<h4 class="ep-title">
							{currentEpisodeDetails?.name ||
								`Episode ${targetNext.episodeNumber}`}
						</h4>
						{#if currentEpisodeDetails?.vote_average && currentEpisodeDetails.vote_count > 0}
							<span class="vote-badge">
								★ {Math.round(currentEpisodeDetails.vote_average * 10) / 10}
							</span>
						{/if}
					</div>

					<div class="meta-row">
						{#if currentEpisodeDetails?.runtime}
							<span class="meta-item">
								<Icon i="clock" wh={13} />
								{currentEpisodeDetails.runtime} min
							</span>
						{/if}
						{#if airDate}
							<RelativeDateBadge
								date={airDate}
								prefix={isUnaired ? "Estreia" : undefined}
								size="sm"
							/>
						{/if}
					</div>

					{#if currentEpisodeDetails?.overview}
						<p class="overview-text">
							{currentEpisodeDetails.overview}
						</p>
					{/if}

					<div class="action-row">
						{#if watchedItem}
							<button
								class="watch-btn"
								class:secondary={isUnaired}
								disabled={markingWatched}
								onclick={markAsWatched}
							>
								{#if markingWatched}
									<Spinner />
									Saving...
								{:else}
									<Icon i="check" wh={16} />
									{#if isUnaired}
										Mark as Watched Anyway
									{:else}
										Mark as Watched
									{/if}
								{/if}
							</button>
						{/if}
						{#if isUnaired}
							<span class="unaired-pill"> Unaired episode </span>
						{/if}
					</div>
				</div>
			</div>
		{/if}
	</div>
{/if}

<style lang="scss">
	.next-episode-card {
		width: 100%;
		background: linear-gradient(
			135deg,
			rgba(20, 25, 38, 0.88) 0%,
			rgba(14, 18, 28, 0.94) 100%
		);
		border: 1px solid rgba(245, 158, 11, 0.25);
		border-radius: 18px;
		padding: 20px 24px;
		margin: 18px 0;
		box-shadow:
			0 12px 36px -8px rgba(0, 0, 0, 0.6),
			0 0 0 1px rgba(255, 255, 255, 0.06);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		transition:
			border-color 200ms ease,
			box-shadow 200ms ease,
			transform 200ms ease;

		&:hover {
			border-color: rgba(245, 158, 11, 0.45);
			box-shadow:
				0 16px 42px -8px rgba(0, 0, 0, 0.7),
				0 0 20px rgba(245, 158, 11, 0.12);
		}

		&.caught-up {
			border-color: rgba(34, 197, 94, 0.25);
			background: rgba(18, 24, 32, 0.7);

			.caught-up-content {
				display: flex;
				align-items: center;
				gap: 16px;

				.caught-up-icon {
					font-size: 32px;
					line-height: 1;
				}

				strong {
					font-size: 16px;
					color: #f1f5f9;
				}

				p {
					margin: 4px 0 0;
					font-size: 13.5px;
					color: rgba(255, 255, 255, 0.6);
				}
			}
		}
	}

	.card-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 16px;

		.header-tag {
			display: flex;
			align-items: center;
			gap: 8px;

			.pulsing-dot {
				width: 8px;
				height: 8px;
				border-radius: 50%;
				background: #f59e0b;
				box-shadow: 0 0 8px #f59e0b;
				animation: pulse 2s infinite;
			}

			.tag-title {
				font-size: 12px;
				font-weight: 700;
				letter-spacing: 0.08em;
				text-transform: uppercase;
				color: #fbbf24;
			}
		}

		.season-badge {
			padding: 4px 10px;
			border-radius: 999px;
			background: rgba(245, 158, 11, 0.15);
			border: 1px solid rgba(245, 158, 11, 0.3);
			color: #fbbf24;
			font-size: 12.5px;
			font-weight: 600;
		}
	}

	.card-loading {
		display: flex;
		justify-content: center;
		padding: 30px;
	}

	.card-body {
		display: flex;
		gap: 20px;
		align-items: flex-start;

		@media screen and (max-width: 680px) {
			flex-direction: column;
		}
	}

	.thumb-wrap {
		flex-shrink: 0;
		width: 220px;
		aspect-ratio: 16 / 9;
		border-radius: 12px;
		overflow: hidden;
		background: #000000;
		border: 1px solid rgba(255, 255, 255, 0.08);

		img {
			width: 100%;
			height: 100%;
			object-fit: cover;
		}

		.thumb-fallback {
			width: 100%;
			height: 100%;
			display: flex;
			align-items: center;
			justify-content: center;
			color: rgba(255, 255, 255, 0.3);
			background: rgba(255, 255, 255, 0.03);
		}

		@media screen and (max-width: 680px) {
			width: 100%;
		}
	}

	.info-wrap {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 8px;
		min-width: 0;

		.title-row {
			display: flex;
			align-items: center;
			justify-content: space-between;
			gap: 12px;

			.ep-title {
				margin: 0;
				font-size: 18px;
				font-weight: 700;
				color: #ffffff;
				line-height: 1.3;
			}

			.vote-badge {
				font-size: 13px;
				font-weight: 700;
				color: #ffd060;
				padding: 2px 8px;
				border-radius: 8px;
				background: rgba(250, 204, 21, 0.1);
				border: 1px solid rgba(250, 204, 21, 0.25);
				flex-shrink: 0;
			}
		}

		.meta-row {
			display: flex;
			align-items: center;
			flex-wrap: wrap;
			gap: 14px;
			font-size: 13px;
			color: rgba(255, 255, 255, 0.6);

			.meta-item {
				display: inline-flex;
				align-items: center;
				gap: 5px;
			}
		}

		.overview-text {
			margin: 4px 0 8px;
			font-size: 13.5px;
			line-height: 1.5;
			color: rgba(255, 255, 255, 0.72);
			display: -webkit-box;
			-webkit-line-clamp: 3;
			line-clamp: 3;
			-webkit-box-orient: vertical;
			overflow: hidden;
		}

		.action-row {
			display: flex;
			align-items: center;
			flex-wrap: wrap;
			gap: 12px;
			margin-top: 6px;

			.watch-btn {
				display: inline-flex;
				align-items: center;
				gap: 8px;
				padding: 8px 16px;
				border-radius: 8px;
				background: $brand-color;
				color: #0b0e14;
				font-size: 13px;
				font-weight: 600;
				border: none;
				cursor: pointer;
				box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
				transition:
					background-color 150ms ease,
					opacity 150ms ease;

				&:hover:not(:disabled) {
					filter: brightness(1.05);
				}

				&:active:not(:disabled) {
					transform: translateY(0);
				}

				&:disabled {
					opacity: 0.6;
					cursor: not-allowed;
				}

				&.secondary {
					background: rgba(255, 255, 255, 0.08);
					color: #ffffff;
					border: 1px solid rgba(255, 255, 255, 0.15);
					box-shadow: none;

					&:hover:not(:disabled) {
						background: rgba(255, 255, 255, 0.14);
						border-color: rgba(255, 255, 255, 0.25);
					}
				}
			}

			.unaired-pill {
				font-size: 12px;
				padding: 4px 10px;
				border-radius: 999px;
				background: rgba(251, 191, 36, 0.12);
				border: 1px solid rgba(251, 191, 36, 0.28);
				color: #fbbf24;
				font-weight: 500;
			}
		}
	}

	@keyframes pulse {
		0%,
		100% {
			opacity: 1;
			transform: scale(1);
		}
		50% {
			opacity: 0.5;
			transform: scale(0.85);
		}
	}
</style>
