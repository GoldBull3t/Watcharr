<script lang="ts">
	import { dateValid } from "../util/date";
	import Icon from "@/lib/Icon.svelte";

	interface Props {
		homepage?: string;
		title?: string;
		releaseDate?: Date;
		endDate?: Date;
		voteAverage?: number;
		voteCount?: number;
	}

	let { homepage, title, releaseDate, endDate, voteAverage, voteCount }: Props =
		$props();

	// if voteAvg bigger than 10, it is out of 100, so no need to * by 10
	const vote = $derived(
		voteAverage
			? Math.round(voteAverage > 10 ? voteAverage : voteAverage * 10) / 10
			: 0,
	);
	const titleSafe = $derived(title ? title : "Unknown Title");
	const releaseYear = $derived(
		dateValid(releaseDate) ? releaseDate.getFullYear() : undefined,
	);
	const endYear = $derived(
		dateValid(endDate) ? endDate.getFullYear() : undefined,
	);
</script>

<span class="title-container">
	<span class="title">
		{#if homepage}
			<a href={homepage} rel="external" target="_blank">{titleSafe}</a>
		{:else}
			<span class="t">{titleSafe}</span>
		{/if}
		{#if releaseYear}
			<span class="year">
				<span title={releaseDate?.toLocaleDateString()}>{releaseYear}</span
				>{#if endYear && endYear != releaseYear}
					<span title={endDate?.toLocaleDateString()}>-{endYear}</span>
				{/if}
			</span>
		{/if}
	</span>
	{#if vote > 0}
		<span
			class="rating"
			title={`Rating: ${vote} out of 10 (based on ${voteCount ?? 0} votes)`}
		>
			<Icon i="star" wh={16} />
			<span>{vote}</span>
		</span>
	{/if}
</span>

<style lang="scss">
	.title-container {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 12px;
		width: 100%;

		.title {
			display: flex;
			align-items: baseline;
			flex-wrap: wrap;
			gap: 8px;

			a,
			span.t {
				color: #ffffff;
				text-decoration: none;
				font-size: 28px;
				font-weight: 700;
				letter-spacing: -0.02em;
				line-height: 1.2;

				&:hover {
					text-decoration: underline;
				}
			}

			span.year {
				font-size: 18px;
				font-weight: 400;
				color: rgba(255, 255, 255, 0.6);
			}
		}

		.rating {
			display: inline-flex;
			align-items: center;
			gap: 5px;
			margin-left: auto;
			padding: 4px 10px;
			height: 28px;
			border-radius: 6px;
			background: rgba(245, 158, 11, 0.12);
			border: 1px solid rgba(245, 158, 11, 0.3);
			color: #fbbf24;
			font-size: 13.5px;
			font-weight: 700;
			letter-spacing: 0.02em;
			flex-shrink: 0;
		}
	}
</style>
