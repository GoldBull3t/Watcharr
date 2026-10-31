<script lang="ts">
	import { resolve } from "$app/paths";
	import type { TMDBContentCreditsCrew } from "@/types";
	import PersonSmartPopover from "@/lib/poster/PersonSmartPopover.svelte";
	import { isTouch } from "@/lib/util/helpers";
	import { onDestroy } from "svelte";

	interface Props {
		topCrew: TMDBContentCreditsCrew[];
	}

	let { topCrew }: Props = $props();

	let activeCrewId = $state<number | null>(null);
	let hoverTimeout: number | undefined;

	function handleMouseEnter(id: number) {
		if (isTouch() || !id) return;
		clearTimeout(hoverTimeout);
		hoverTimeout = window.setTimeout(() => {
			activeCrewId = id;
		}, 320);
	}

	function handleMouseLeave() {
		clearTimeout(hoverTimeout);
		activeCrewId = null;
	}

	onDestroy(() => {
		clearTimeout(hoverTimeout);
	});
</script>

<div class="creators">
	{#each topCrew as crew (crew.credit_id)}
		<div
			class="crew-card"
			role="group"
			onmouseenter={() => handleMouseEnter(crew.id)}
			onmouseleave={handleMouseLeave}
		>
			{#if activeCrewId === crew.id}
				<PersonSmartPopover
					personId={crew.id}
					personName={crew.name}
					role={crew.job}
				/>
			{/if}
			<a href={resolve(`/person/${crew.id}`)} class="crew-name">{crew.name}</a>
			<span class="crew-job">{crew.job}</span>
		</div>
	{/each}
</div>

<style lang="scss">
	.creators {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(170px, 1fr));
		gap: 10px;
		width: 100%;
		margin: 0;

		.crew-card {
			position: relative;
			display: flex;
			flex-flow: column;
			gap: 2px;
			padding: 10px 14px;
			border-radius: 8px;
			background: rgba(15, 19, 28, 0.72);
			border: 1px solid rgba(255, 255, 255, 0.08);
			transition:
				background 150ms ease,
				border-color 150ms ease;

			&:hover,
			&:focus-within {
				background: rgba(24, 30, 44, 0.9);
				border-color: rgba(255, 255, 255, 0.18);
				z-index: 100;
			}

			.crew-name {
				font-size: 13.5px;
				font-weight: 700;
				color: #ffffff;
				text-decoration: none;
				white-space: nowrap;
				overflow: hidden;
				text-overflow: ellipsis;

				&:hover {
					color: $brand-color;
					text-decoration: underline;
				}
			}

			.crew-job {
				font-size: 11px;
				font-weight: 600;
				text-transform: uppercase;
				letter-spacing: 0.04em;
				color: rgba(255, 255, 255, 0.48);
				white-space: nowrap;
				overflow: hidden;
				text-overflow: ellipsis;
			}
		}
	}
</style>
