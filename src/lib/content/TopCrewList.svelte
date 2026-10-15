<script lang="ts">
	import { resolve } from "$app/paths";
	import type { TMDBContentCreditsCrew } from "@/types";

	interface Props {
		topCrew: TMDBContentCreditsCrew[];
	}

	let { topCrew }: Props = $props();
</script>

<div class="creators">
	{#each topCrew as crew (crew.credit_id)}
		<div class="crew-card">
			<a href={resolve(`/person/${crew.id}`)} class="crew-name">{crew.name}</a>
			<span class="crew-job">{crew.job}</span>
		</div>
	{/each}
</div>

<style lang="scss">
	.creators {
		display: flex;
		flex-wrap: wrap;
		justify-content: center;
		gap: 14px;
		margin: 10px auto;
		max-width: 1000px;

		.crew-card {
			display: flex;
			flex-flow: column;
			gap: 3px;
			min-width: 140px;
			padding: 10px 16px;
			border-radius: 12px;
			background: rgba(22, 27, 38, 0.6);
			border: 1px solid rgba(255, 255, 255, 0.08);
			box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
			backdrop-filter: blur(10px);
			-webkit-backdrop-filter: blur(10px);
			transition:
				background 150ms ease,
				border-color 150ms ease,
				transform 150ms ease;

			&:hover {
				background: rgba(30, 36, 52, 0.85);
				border-color: rgba(255, 255, 255, 0.16);
				transform: translateY(-1px);
			}

			.crew-name {
				font-size: 14px;
				font-weight: 700;
				color: #ffffff;
				text-decoration: none;

				&:hover {
					color: #a5b4fc;
					text-decoration: underline;
				}
			}

			.crew-job {
				font-size: 12px;
				font-weight: 500;
				color: rgba(255, 255, 255, 0.55);
			}
		}
	}
</style>
