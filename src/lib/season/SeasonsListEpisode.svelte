<script lang="ts">
	import type {
		TMDBSeasonDetailsEpisode,
		WatchedStatus,
		Watched,
	} from "@/types";
	import Icon from "../Icon.svelte";
	import PosterRating from "../poster/PosterRating.svelte";
	import PosterStatus from "../poster/PosterStatus.svelte";
	import { notify } from "../util/notify";
	import { store } from "@/store.svelte";
	import { removeWatchedEpisode, updateWatchedEpisode } from "./api";
	import { onMount } from "svelte";
	import { toRelativeDate } from "../util/helpers";
	import RelativeDateBadge from "@/lib/content/RelativeDateBadge.svelte";

	interface Props {
		ep: TMDBSeasonDetailsEpisode;
		tvId?: number;
		watchedItem: Watched | undefined;
		onStatusChanged?: () => void;
	}

	let { ep, tvId, watchedItem, onStatusChanged }: Props = $props();

	let isExpanded = $state(false);

	const we = $derived(
		watchedItem?.watchedEpisodes?.find(
			(s) =>
				s.seasonNumber === ep.season_number &&
				s.episodeNumber === ep.episode_number,
		),
	);

	let isHidden: boolean = $state(!!store?.userSettings?.hideSpoilers);

	const airDate: Date | undefined = $derived.by(() => {
		if (!ep.air_date) {
			return;
		}
		const d = new Date(ep.air_date);
		if (isNaN(d.getTime())) {
			return;
		}
		return d;
	});
	const isUnaired: boolean = $derived.by(() => {
		if (!airDate) {
			return false;
		}
		return airDate.getTime() > new Date().getTime();
	});

	/**
	 * Re-sets `isHidden` state.
	 */
	function reSetIsHidden() {
		// If the episode status is "FINISHED", ensure `isHidden` is set to
		// `false` (so finished episodes aren't blurred when hideSpoilers is on).
		if (we?.status == "FINISHED") {
			isHidden = false;
		}
	}

	onMount(() => {
		reSetIsHidden();
	});

	async function handleStatusClick(type: WatchedStatus | "DELETE") {
		if (!watchedItem) {
			console.error("SeasonListEpisode: handleStatusClick: No watched item.");
			return;
		}
		if (type === "DELETE") {
			if (!we || !we.id) {
				notify({
					text: "Failed to find watched episode id. Please try refreshing.",
					type: "error",
				});
				console.error(
					"handleStatusClick(DELETE): `we` doesn't exist or have an id",
					we,
				);
				return;
			}
			removeWatchedEpisode(watchedItem, we.id);
			// NOTE: Similar to below where we `reSetIsHidden` to unhide spoilers
			// automatically if status is set to FINISHED, we WONT do the opposite
			// here and re-hide the spoilers (if unhidden) after removing an episode
			// because that would probably be annoying to users (eg: click to
			// show spoilers, then delete episode, spoilers re-hidden automatically).
			return;
		}
		await updateWatchedEpisode(
			watchedItem,
			ep.season_number,
			ep.episode_number,
			{
				status: type,
			},
		);
		reSetIsHidden();
		onStatusChanged?.();
	}

	function handleStarClick(rating: number) {
		if (!watchedItem) {
			console.error("handleStarClick: No watchedItem!");
			return;
		}
		updateWatchedEpisode(watchedItem, ep.season_number, ep.episode_number, {
			rating,
		});
	}
</script>

<li
	id={`ep-${ep.season_number}-${ep.episode_number}`}
	class={isHidden ? "dont-spoil" : ""}
>
	<div class="thumb-container">
		{#if ep.still_path}
			<img
				src={`https://www.themoviedb.org/t/p/w500/${ep.still_path}`}
				alt={ep.name || `Episode ${ep.episode_number}`}
				loading="lazy"
				decoding="async"
			/>
		{:else}
			<div class="no-still">
				<Icon i="film" wh={28} />
			</div>
		{/if}
		<span class="ep-num-badge">E{ep.episode_number}</span>
	</div>

	<div class="info">
		<div class="ep-title-row">
			<div class="ep-title-wrap">
				<h4 class="episode-name">{ep.name}</h4>
				<div class="ep-meta">
					{#if ep.runtime}
						<span class="ep-meta-tag" title="Runtime: {ep.runtime} min">
							<Icon i="clock" wh={12} />
							{ep.runtime} min
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
			</div>

			{#if ep.vote_count > 0 && (!isUnaired || ep.vote_average > 0)}
				<span
					class="rating-badge"
					title={`TMDB Rating: ${ep.vote_average} out of 10 (${ep.vote_count} votes)`}
				>
					★ {Math.round(ep.vote_average * 10) / 10}
				</span>
			{/if}
		</div>

		{#if ep.overview}
			<p class="overview" class:expanded={isExpanded}>{ep.overview}</p>
		{/if}

		<div class="ep-actions-row">
			<button
				type="button"
				class="plain ep-expand-btn"
				onclick={() => (isExpanded = !isExpanded)}
				aria-expanded={isExpanded}
			>
				<Icon i="chevron" wh={11} />
				<span>{isExpanded ? "Menos detalhes" : "Mais detalhes"}</span>
			</button>

			{#if isExpanded}
				<div class="ep-expanded-details">
					<span class="detail-pill">
						<strong>Código:</strong>
						S{String(ep.season_number).padStart(2, "0")}E{String(
							ep.episode_number,
						).padStart(2, "0")}
					</span>
					{#if airDate}
						<span class="detail-pill">
							<strong>Exibição:</strong>
							{airDate.toLocaleDateString()}
						</span>
					{/if}
					{#if ep.vote_count > 0}
						<span class="detail-pill">
							<strong>Avaliações TMDB:</strong>
							{Math.round(ep.vote_average * 10) / 10}/10 ({ep.vote_count} votos)
						</span>
					{/if}
					{#if tvId}
						<a
							class="detail-link"
							href={`https://www.themoviedb.org/tv/${tvId}/season/${ep.season_number}/episode/${ep.episode_number}`}
							target="_blank"
							rel="noreferrer noopener external"
						>
							Ver ficha completa & elenco no TMDB ↗
						</a>
					{/if}
				</div>
			{/if}
		</div>
	</div>

	{#if watchedItem}
		<div class="status-rating-ctr">
			<div class="rating">
				<PosterRating
					rating={we?.rating}
					btnTooltip={`Episode ${ep.episode_number} Rating`}
					handleStarClick={(r) => handleStarClick(r)}
					minimal={true}
					direction="bot"
					hideStarWhenRated
				/>
			</div>
			<div class="status">
				<PosterStatus
					status={we?.status}
					btnTooltip={`Episode ${ep.episode_number} Status`}
					handleStatusClick={(t) => handleStatusClick(t)}
					minimal={true}
					direction="bot"
				/>
			</div>
		</div>
	{/if}
	{#if isHidden}
		<button class="plain spoiler-text" onclick={() => (isHidden = false)}>
			<Icon i="eye-closed" wh={34} />
			<span>Click To Reveal</span>
		</button>
	{/if}
</li>

<style lang="scss">
	li {
		display: flex;
		flex-flow: row;
		gap: 18px;
		position: relative;
		background: rgba(15, 19, 28, 0.82);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 10px;
		padding: 16px;
		transition:
			background 150ms ease,
			border-color 150ms ease;

		&:hover {
			background: rgba(26, 32, 46, 0.82);
			border-color: rgba(255, 255, 255, 0.15);
		}

		&:global(.ep-highlight-pulse) {
			animation: epPulse 1.4s ease-out;
		}

		@keyframes epPulse {
			0% {
				box-shadow: 0 0 0 0 rgba(245, 158, 11, 0.7);
				border-color: rgba(245, 158, 11, 0.9);
				transform: scale(1.015);
			}
			50% {
				box-shadow: 0 0 24px 6px rgba(245, 158, 11, 0.4);
				border-color: rgba(245, 158, 11, 0.6);
			}
			100% {
				box-shadow: 0 4px 16px rgba(0, 0, 0, 0.25);
				border-color: rgba(255, 255, 255, 0.08);
				transform: scale(1);
			}
		}

		&:hover,
		&:focus-within,
		&:has(:global(.shown)),
		&:has(:global(.open)) {
			z-index: 50 !important;
		}

		.thumb-container {
			position: relative;
			width: 200px;
			min-width: 200px;
			aspect-ratio: 16 / 9;
			border-radius: 10px;
			overflow: hidden;
			background: #000000;
			border: 1px solid rgba(255, 255, 255, 0.08);
			flex-shrink: 0;

			img {
				width: 100%;
				height: 100%;
				object-fit: cover;
			}

			.no-still {
				width: 100%;
				height: 100%;
				display: flex;
				align-items: center;
				justify-content: center;
				color: rgba(255, 255, 255, 0.25);
				background: rgba(255, 255, 255, 0.03);
			}

			.ep-num-badge {
				position: absolute;
				bottom: 6px;
				left: 6px;
				padding: 2px 7px;
				border-radius: 6px;
				background: rgba(0, 0, 0, 0.75);
				backdrop-filter: blur(8px);
				-webkit-backdrop-filter: blur(8px);
				border: 1px solid rgba(255, 255, 255, 0.15);
				font-size: 11.5px;
				font-weight: 700;
				color: #ffffff;
				letter-spacing: 0.04em;
			}
		}

		.info {
			display: flex;
			flex-flow: column;
			gap: 8px;
			flex: 1;
			min-width: 0;

			.ep-title-row {
				display: flex;
				justify-content: space-between;
				align-items: flex-start;
				gap: 12px;

				.ep-title-wrap {
					display: flex;
					flex-direction: column;
					gap: 4px;

					.episode-name {
						margin: 0;
						font-size: 16px;
						font-weight: 700;
						color: #ffffff;
						line-height: 1.3;
					}

					.ep-meta {
						display: flex;
						align-items: center;
						gap: 12px;
						font-size: 12.5px;
						color: rgba(255, 255, 255, 0.6);

						.ep-meta-tag {
							display: inline-flex;
							align-items: center;
							gap: 4px;
						}
					}
				}

				.rating-badge {
					font-size: 12.5px;
					font-weight: 700;
					color: #ffd060;
					padding: 3px 8px;
					border-radius: 8px;
					background: rgba(250, 204, 21, 0.1);
					border: 1px solid rgba(250, 204, 21, 0.25);
					flex-shrink: 0;
				}
			}

			.overview {
				margin: 0;
				font-size: 13.5px;
				line-height: 1.5;
				color: rgba(255, 255, 255, 0.72);
				display: -webkit-box;
				-webkit-line-clamp: 3;
				line-clamp: 3;
				-webkit-box-orient: vertical;
				overflow: hidden;

				&.expanded {
					display: block;
					-webkit-line-clamp: unset;
					line-clamp: unset;
					overflow: visible;
				}
			}

			.ep-actions-row {
				display: flex;
				flex-direction: column;
				gap: 8px;
				margin-top: 2px;

				.ep-expand-btn {
					display: inline-flex;
					align-items: center;
					gap: 5px;
					width: fit-content;
					padding: 2px 0;
					font-size: 11.5px;
					font-weight: 600;
					color: rgba(255, 255, 255, 0.5);
					cursor: pointer;

					&:hover {
						color: #fff;
					}
				}

				.ep-expanded-details {
					display: flex;
					flex-wrap: wrap;
					align-items: center;
					gap: 8px;
					padding-top: 6px;
					border-top: 1px solid rgba(255, 255, 255, 0.07);
					font-size: 12px;
					color: rgba(255, 255, 255, 0.75);

					.detail-pill {
						display: inline-flex;
						align-items: center;
						gap: 4px;
						padding: 3px 8px;
						border-radius: 6px;
						background: rgba(255, 255, 255, 0.04);
						border: 1px solid rgba(255, 255, 255, 0.08);

						strong {
							font-size: 10.5px;
							text-transform: uppercase;
							letter-spacing: 0.04em;
							color: rgba(255, 255, 255, 0.45);
						}
					}

					.detail-link {
						display: inline-flex;
						align-items: center;
						padding: 3px 8px;
						border-radius: 6px;
						background: rgba(99, 102, 241, 0.14);
						border: 1px solid rgba(99, 102, 241, 0.3);
						color: #c7d2fe;
						font-weight: 600;
						text-decoration: none;

						&:hover {
							background: rgba(99, 102, 241, 0.24);
							color: #fff;
						}
					}
				}
			}
		}

		.status-rating-ctr {
			display: flex;
			align-items: center;
			flex-flow: column-reverse;
			gap: 10px;
			margin-left: auto;
			flex-shrink: 0;
			position: relative;
			z-index: 5;

			&:hover,
			&:focus-within,
			&:has(:global(.shown)),
			&:has(:global(.open)) {
				z-index: 60 !important;
			}

			div {
				transition: width 100ms ease;

				&.rating {
					height: 40px;
					min-height: 40px;
					position: relative;
					overflow: visible;

					&:hover,
					&:focus-within,
					&:has(:global(.shown)),
					&:has(:global(.open)) {
						z-index: 70 !important;
					}
				}

				&.status {
					width: 48px;
					min-height: 42px;
					height: 42px;
					overflow: visible;
					position: relative;

					&:hover,
					&:focus-within,
					&:has(:global(.shown)),
					&:has(:global(.open)) {
						z-index: 70 !important;
					}
				}
			}
		}

		.spoiler-text {
			display: flex;
			flex-flow: column;
			align-items: center;
			justify-content: center;
			gap: 8px;
			position: absolute;
			inset: 0;
			border-radius: 16px;
			font-weight: 700;
			font-size: 18px;
			color: #ffffff;
			fill: #ffffff;
			opacity: 0;
			background: rgba(11, 14, 20, 0.85);
			backdrop-filter: blur(10px);
			-webkit-backdrop-filter: blur(10px);
			transition: opacity 150ms ease-in;
			cursor: pointer;

			&:hover,
			&:active,
			&:focus {
				opacity: 1;
			}
		}

		img,
		.episode-name,
		.overview {
			transition: filter 150ms ease-out;
		}

		&.dont-spoil {
			.episode-name,
			.overview {
				filter: blur(4px);
			}

			img {
				filter: blur(8px);
			}
		}
	}

	@media screen and (max-width: 680px) {
		li {
			flex-flow: column;
			align-items: stretch;

			.thumb-container {
				width: 100%;
			}

			.status-rating-ctr {
				flex-flow: row;
				justify-content: flex-end;
				margin-left: 0;
			}
		}
	}
</style>
