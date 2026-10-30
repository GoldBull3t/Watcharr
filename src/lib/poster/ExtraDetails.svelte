<!-- Extra Details View For Posters -->
<script lang="ts">
	import { store } from "@/store.svelte";
	import { RatingSystem } from "@/types";
	import {
		getOrdinalSuffix,
		monthsShort,
		watchedStatuses,
	} from "../util/helpers";
	import Icon from "../Icon.svelte";
	import { toShowableRating, toWhichThumb } from "../rating/helpers";
	import { t } from "@/lib/i18n";
	import type { PosterExtraDetails } from "./lib";

	let {
		rating,
		status,
		dateAdded,
		dateModified,
		lastWatched,
		seasonProgress,
		hoursPlayed,
		platform,
	}: PosterExtraDetails = $props();

	let isUsingThumbs = $derived(
		store.userSettings &&
			store.userSettings.ratingSystem === RatingSystem.Thumbs,
	);

	function formatDate(e: number) {
		if (!e) {
			return "Unknown";
		}
		const d = new Date(e);
		return `${d.getDate()}${getOrdinalSuffix(d.getDate())} ${monthsShort[d.getMonth()]} '${String(
			d.getFullYear(),
		).substring(2, 4)}`;
	}

	// Compute whether any item is actually eligible to be shown, avoiding empty "ghost badges"
	const hasItemsToShow = $derived(
		Boolean(
			store.wlDetailedView &&
				store.wlDetailedView.length > 0 &&
				((dateAdded && store.wlDetailedView.includes("dateAdded")) ||
					(dateModified && store.wlDetailedView.includes("dateModified")) ||
					(seasonProgress && store.wlDetailedView.includes("lastWatched")) ||
					(lastWatched && store.wlDetailedView.includes("lastWatched")) ||
					(typeof hoursPlayed === "number" && hoursPlayed > 0) ||
					(store.wlDetailedView.includes("statusRating") && (status || rating))),
		),
	);
</script>

{#if hasItemsToShow}
	<div class="extra-details">
		<div>
			{#if dateAdded && store.wlDetailedView.includes("dateAdded")}
				<span title={t("poster.date_added") || "Date added to watch list"}>
					<i><Icon i="calendar" /></i>
					<span>
						{formatDate(Date.parse(dateAdded))}
					</span>
				</span>
			{/if}{#if dateModified && store.wlDetailedView.includes("dateModified")}
				<span title={t("poster.date_modified") || "Date last modified"}>
					<i><Icon i="pencil" wh={15} /></i>
					<span>
						{formatDate(Date.parse(dateModified))}
					</span>
				</span>
			{/if}{#if seasonProgress && store.wlDetailedView.includes("lastWatched")}
				<div class="season-progress-pill" title={seasonProgress.displayText}>
					{#if seasonProgress.totalSeasons <= 10}
						<div class="segments-row">
							{#each seasonProgress.seasons as s}
								<span
									class="seg {s.status.toLowerCase()}"
									title={`T${s.number}: ${
										s.status === "FINISHED"
											? t("status.finished") || "Concluída"
											: s.status === "PLANNED"
												? t("status.planned") || "Planeada"
												: s.status === "WATCHING"
													? t("status.watching") || "A ver"
													: t("status.unwatched") || "Por ver"
									}`}
								></span>
							{/each}
						</div>
					{:else}
						<!-- Continuous sleek bar for shows with more than 10 seasons -->
						<div
							class="continuous-progress-track"
							title={`${seasonProgress.finishedSeasons.length}/${seasonProgress.totalSeasons} (${seasonProgress.percentage}%)`}
						>
							<div
								class="continuous-progress-fill"
								style="width: {seasonProgress.percentage}%;"
							></div>
						</div>
					{/if}
					<span class="season-label" class:caught-up={seasonProgress.isCaughtUp}>
						<i><Icon i={seasonProgress.isCaughtUp ? "check" : "play"} wh={11} /></i>
						<span>{seasonProgress.displayText}</span>
					</span>
				</div>
			{:else if lastWatched && store.wlDetailedView.includes("lastWatched")}
				<span title={t("poster.latest_season") || "Latest season watched"}>
					<i><Icon i="play" wh={15} /></i>
					<span>{lastWatched}</span>
				</span>
			{/if}{#if typeof hoursPlayed === "number" && hoursPlayed > 0}
				<span title={t("gaming.hours_played") || "Horas jogadas"}>
					<i>🎮</i>
					<span>{hoursPlayed}h{#if platform} • {platform}{/if}</span>
				</span>
			{/if}{#if store.wlDetailedView.includes("statusRating")}
				<span
					class="status-rating"
					title={t("poster.status_and_rating") || "Status and Rating"}
				>
					{#if !isUsingThumbs}
						<i><Icon i="star" /></i>
					{/if}
					<span class="rating-span">
						{#if isUsingThumbs}
							{@const r = toWhichThumb(rating)}
							{#if r === -1}
								<i><Icon i="thumb-down" /></i>
							{:else if r === 0}
								<span class="neutral-dash-wrap" aria-hidden="true">
									<svg width="12" height="3" viewBox="0 0 12 3">
										<rect width="12" height="3" rx="1.5" fill="currentColor" />
									</svg>
								</span>
							{:else if r === 1}
								<i><Icon i="thumb-up" /></i>
							{/if}
						{:else}
							{toShowableRating(rating)}
						{/if}
					</span>
					{#if status}
						<i><Icon i={watchedStatuses[status]} wh={15} /></i>
					{/if}
				</span>
			{/if}
		</div>
	</div>
{/if}

<style lang="scss">
	.extra-details {
		position: absolute;
		bottom: 8px;
		left: 50%;
		transform: translateX(-50%);
		display: flex;
		flex-flow: column;
		justify-content: center;
		align-items: center;
		font-size: 12px;
		width: calc(100% - 16px);
		max-width: 160px;
		color: white;
		background-color: rgba(15, 18, 25, 0.88);
		border: 1px solid rgba(255, 255, 255, 0.12);
		border-radius: 8px;
		box-shadow: 0 4px 14px rgba(0, 0, 0, 0.35);
		backdrop-filter: blur(8px);
		-webkit-backdrop-filter: blur(8px);
		transition: opacity 100ms ease-out;
		pointer-events: none;

		& > div {
			padding: 6px 10px;
			width: 100%;

			&:empty {
				padding: 0px;
			}

			& > span {
				display: flex;
				flex-flow: row;
				align-items: center;
				gap: 6px;
				height: 16px;
				font-weight: 600;
				pointer-events: auto;
				cursor: default;

				&.status-rating {
					gap: 8px;
				}

				&:not(:last-child) {
					margin-bottom: 4px;
				}

				i {
					display: flex;
					align-items: center;
					width: 14px;
					fill: white;
					color: white;
				}

				&.caught-up {
					color: #38bdf8;

					i {
						fill: #34d399;
						color: #34d399;
					}
				}
			}
		}

		.season-progress-pill {
			display: flex;
			flex-direction: column;
			gap: 4px;
			width: 100%;
			margin-bottom: 5px;
			pointer-events: auto;
			cursor: default;

			.segments-row {
				display: flex;
				flex-flow: row;
				gap: 2.5px;
				width: 100%;
				height: 4px;

				.seg {
					flex: 1;
					height: 100%;
					border-radius: 2px;
					background: rgba(255, 255, 255, 0.2);

					&.finished {
						background: #10b981;
						box-shadow: 0 0 4px rgba(16, 185, 129, 0.4);
					}

					&.watching {
						background: #3b82f6;
					}

					&.planned {
						background: #f59e0b;
					}

					&.unwatched {
						background: rgba(255, 255, 255, 0.15);
					}
				}
			}

			.continuous-progress-track {
				width: 100%;
				height: 4px;
				border-radius: 2px;
				background: rgba(255, 255, 255, 0.18);
				overflow: hidden;

				.continuous-progress-fill {
					height: 100%;
					border-radius: 2px;
					background: linear-gradient(90deg, #3b82f6 0%, #10b981 100%);
					transition: width 0.3s ease;
				}
			}

			.season-label {
				display: flex;
				flex-flow: row;
				align-items: center;
				gap: 4px;
				font-size: 10.5px;
				font-weight: 600;
				white-space: nowrap;
				overflow: hidden;
				text-overflow: ellipsis;
				color: rgba(255, 255, 255, 0.9);

				i {
					display: flex;
					align-items: center;
					width: 12px;
					fill: white;
					color: white;
				}

				&.caught-up {
					color: #38bdf8;

					i {
						fill: #34d399;
						color: #34d399;
					}
				}
			}
		}

		.status-rating i:last-of-type {
			margin-left: auto;
		}

		.status-rating .rating-span:empty {
			/* Can happen when content is not rated and using thumbs system */
			display: none;
		}
	}
</style>
