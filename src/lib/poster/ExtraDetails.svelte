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
	import { page } from "$app/state";
	import type { PosterExtraDetails } from "./lib";

	let {
		rating,
		status,
		dateAdded,
		dateModified,
		lastWatched,
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
</script>

{#if (page.url?.pathname === "/" || page.url?.pathname.startsWith("/search")) && store.wlDetailedView && store.wlDetailedView.length > 0}
	<div class="extra-details">
		<!--
      The `if` statements can't be on their own line to look pretty
      because that will cause whitespace in the generated markup,
      which causes the :empty css tag to not work.
      Can be reverted when its possible to trim whitespace in svelte
      OR when :empty tag is updated in browsers to new spec and counts whitespace as empty.
    -->
		<div>
			{#if dateAdded && store.wlDetailedView.includes("dateAdded")}
				<span title="Date added to watch list">
					<i><Icon i="calendar" /></i>
					<span>
						{formatDate(Date.parse(dateAdded))}
					</span>
				</span>
			{/if}{#if dateModified && store.wlDetailedView.includes("dateModified")}
				<span title="Date last modified">
					<i><Icon i="pencil" wh={15} /></i>
					<span>
						{formatDate(Date.parse(dateModified))}
					</span>
				</span>
			{/if}{#if lastWatched && store.wlDetailedView.includes("lastWatched")}
				<span title="Latest season watched">
					<i><Icon i="play" wh={15} /></i>
					<span>{lastWatched}</span>
				</span>
			{/if}{#if typeof hoursPlayed === "number" && hoursPlayed > 0}
				<span title="Horas jogadas">
					<i>🎮</i>
					<span>{hoursPlayed}h{#if platform} • {platform}{/if}</span>
				</span>
			{/if}{#if store.wlDetailedView.includes("statusRating")}
				<span class="status-rating" title="Status and Rating">
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
		pointer-events: none !important;

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
