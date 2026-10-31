<script lang="ts">
	import type { Watched, WatchedStatus } from "@/types";
	import Rating from "../rating/Rating.svelte";
	import Status from "../Status.svelte";
	import MyThoughts from "./MyThoughts.svelte";

	interface Props {
		watched?: Watched;
		contentTitle?: string;
		onRatingChanged: (newRating: number) => Promise<boolean>;
		onStatusChanged: (newStatus: WatchedStatus) => Promise<boolean>;
		onThoughtsChanged: (newThoughts: string) => Promise<boolean>;
	}

	let {
		watched,
		contentTitle,
		onRatingChanged,
		onStatusChanged,
		onThoughtsChanged,
	}: Props = $props();
</script>

<div class="review">
	<div class="review-header">
		<span class="review-title">Your Log & Rating</span>
		{#if watched && typeof watched.plays == "number" && watched.plays > 0}
			<span class="plays-badge">
				{watched.plays}
				{watched.plays > 1 ? "Plays" : "Play"}
			</span>
		{/if}
	</div>
	<Rating rating={watched?.rating} onChange={onRatingChanged} />
	<Status status={watched?.status} onChange={onStatusChanged} />
	{#if watched}
		<MyThoughts
			{contentTitle}
			thoughts={watched?.thoughts}
			onChange={onThoughtsChanged}
		/>
	{/if}
</div>

<style lang="scss">
	.review {
		display: flex;
		flex-flow: column;
		gap: 14px;
		width: 100%;
		min-inline-size: 0;
		color: $text-color;
		padding: 20px;
		background: rgba(15, 19, 28, 0.86);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 10px;
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.45);
		backdrop-filter: blur(16px);
		-webkit-backdrop-filter: blur(16px);

		.review-header {
			display: flex;
			align-items: center;
			justify-content: space-between;
			gap: 8px;
			padding-bottom: 10px;
			border-bottom: 1px solid rgba(255, 255, 255, 0.06);

			.review-title {
				font-size: 11.5px;
				font-weight: 700;
				text-transform: uppercase;
				letter-spacing: 0.06em;
				color: rgba(255, 255, 255, 0.55);
			}
		}

		.plays-badge {
			display: inline-flex;
			align-items: center;
			padding: 2px 8px;
			border-radius: 6px;
			background: rgba(255, 255, 255, 0.06);
			border: 1px solid rgba(255, 255, 255, 0.1);
			font-size: 11px;
			font-weight: 600;
			color: rgba(255, 255, 255, 0.75);
			letter-spacing: 0.03em;
			text-transform: uppercase;
		}

		@media screen and (max-width: 440px) {
			padding: 16px;
		}
	}
</style>
