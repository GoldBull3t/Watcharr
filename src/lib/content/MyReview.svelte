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
	<Rating rating={watched?.rating} onChange={onRatingChanged} />
	<Status status={watched?.status} onChange={onStatusChanged} />
	{#if watched}
		<MyThoughts
			{contentTitle}
			thoughts={watched?.thoughts}
			onChange={onThoughtsChanged}
		/>
		{#if typeof watched.plays == "number" && watched.plays > 0}
			<div class="plays-badge">
				<span>{watched.plays} {watched.plays > 1 ? "Plays" : "Play"}</span>
			</div>
		{/if}
	{/if}
</div>

<style lang="scss">
	.review {
		display: flex;
		flex-flow: column;
		gap: 12px;
		width: 100%;
		max-width: 400px;
		color: $text-color;
		margin-left: auto;
		margin-right: auto;
		margin-top: 24px;
		padding: 20px 24px;
		background: rgba(22, 27, 38, 0.75);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 18px;
		box-shadow: 0 12px 32px -6px rgba(0, 0, 0, 0.5);
		backdrop-filter: blur(16px);
		-webkit-backdrop-filter: blur(16px);

		.plays-badge {
			display: flex;
			justify-content: center;
			font-size: 12.5px;
			font-weight: 600;
			color: rgba(255, 255, 255, 0.5);
			letter-spacing: 0.04em;
			text-transform: uppercase;
		}

		@media screen and (max-width: 440px) {
			max-width: 100%;
			padding: 16px;
		}
	}
</style>
