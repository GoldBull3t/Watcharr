<script lang="ts">
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import { store } from "@/store.svelte";
	import type { MediaGenre } from "@/types";

	interface Props {
		genres?: MediaGenre[];
	}

	let { genres }: Props = $props();

	function handleGenreClick(genreName: string) {
		store.activeFilters = {
			...store.activeFilters,
			genre: genreName,
		};
		goto(resolve("/"));
	}
</script>

<div class="genres-list">
	{#if genres && genres.length > 0}
		{#each genres as g (g.id || g.name)}
			<button
				type="button"
				class="plain genre-chip"
				onclick={() => handleGenreClick(g.name)}
				title={`Filtrar biblioteca por ${g.name}`}
			>
				{g.name}
			</button>
		{/each}
	{:else}
		<!-- Generic "unknown" text, since this component
		 	 is also used for displaying Game Modes. -->
		<span>Unknown</span>
	{/if}
</div>

<style lang="scss">
	.genres-list {
		display: inline-flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 6px;

		.genre-chip {
			display: inline-flex;
			align-items: center;
			padding: 2px 8px;
			border-radius: 6px;
			font-size: 12px;
			font-weight: 500;
			color: rgba(255, 255, 255, 0.78);
			background: rgba(255, 255, 255, 0.05);
			border: 1px solid rgba(255, 255, 255, 0.1);
			cursor: pointer;
			transition:
				background-color 120ms ease,
				border-color 120ms ease,
				color 120ms ease;

			&:hover {
				background: rgba(255, 255, 255, 0.12);
				border-color: rgba(255, 255, 255, 0.25);
				color: #fff;
			}
		}
	}
</style>
