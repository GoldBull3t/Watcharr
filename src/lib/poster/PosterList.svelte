<script lang="ts">
	import { onMount } from "svelte";

	export type GridDensity = "compact" | "comfortable" | "editorial";

	interface Props {
		type?: "wrapped" | "vertical";
		density?: GridDensity;
		children?: import("svelte").Snippet;
	}

	let { type = "wrapped", density = "comfortable", children }: Props = $props();

	let ulEl: HTMLUListElement | undefined = $state();

	onMount(() => {
		if (ulEl) {
			ulEl.classList.add(type);
		}
	});
</script>

<div class="poster-list-wrapper">
	<ul bind:this={ulEl} class={`${type} density-${density}`}>
		{@render children?.()}
	</ul>
</div>

<style lang="scss">
	.poster-list-wrapper {
		position: relative;
		z-index: 1;
		display: flex;
		justify-content: center;
		width: 100%;
	}

	ul {
		list-style: none;
		width: 100%;

		&.vertical {
			display: flex;
			flex-flow: row;
			flex-wrap: nowrap;
			justify-content: flex-start;
			overflow-x: auto;
			padding: 12px 16px 20px 16px;
			margin: 0;
			gap: 16px;
			scroll-behavior: smooth;
		}

		/* Wrapped Base */
		&.wrapped {
			display: grid;
			width: 100%;
			max-width: 1440px;
			padding: 0 4px;
			transition:
				gap 200ms ease,
				margin 200ms ease;

			/* 1. Comfortable Density (Default) */
			&.density-comfortable,
			&:not(.density-compact):not(.density-editorial) {
				grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
				gap: 24px 18px;
				margin: 20px 0 36px 0;

				@media screen and (max-width: 640px) {
					grid-template-columns: repeat(auto-fill, minmax(135px, 1fr));
					gap: 16px 12px;
					padding: 0;
				}

				@media screen and (max-width: 340px) {
					grid-template-columns: repeat(2, 1fr);
					gap: 12px 10px;
					padding: 0;
				}

				:global(li) {
					display: flex;
					justify-content: center;
					width: 100%;
					min-width: 0;
					content-visibility: auto;
					contain-intrinsic-size: 170px 256px;
				}

				:global(.container) {
					width: 100% !important;
					max-width: 220px;
					min-width: unset !important;
					aspect-ratio: 170 / 256;
					height: auto !important;
					min-height: unset !important;
				}
			}

			/* 2. Compact Density (Dense Wall) */
			&.density-compact {
				grid-template-columns: repeat(auto-fill, minmax(118px, 1fr));
				gap: 14px 10px;
				margin: 14px 0 30px 0;

				@media screen and (max-width: 640px) {
					grid-template-columns: repeat(auto-fill, minmax(96px, 1fr));
					gap: 10px 8px;
					padding: 0;
				}

				@media screen and (max-width: 360px) {
					grid-template-columns: repeat(3, 1fr);
					gap: 8px 6px;
					padding: 0;
				}

				:global(li) {
					display: flex;
					justify-content: center;
					width: 100%;
					min-width: 0;
					content-visibility: auto;
					contain-intrinsic-size: 120px 180px;
				}

				:global(.container) {
					width: 100% !important;
					max-width: 155px;
					min-width: unset !important;
					aspect-ratio: 170 / 256;
					height: auto !important;
					min-height: unset !important;
					border-radius: 8px;
				}
			}

			/* 3. Editorial Density (Large Showcase Cards) */
			&.density-editorial {
				grid-template-columns: repeat(auto-fill, minmax(230px, 1fr));
				gap: 32px 24px;
				margin: 24px 0 44px 0;

				@media screen and (max-width: 768px) {
					grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
					gap: 22px 16px;
					padding: 0;
				}

				@media screen and (max-width: 440px) {
					grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
					gap: 16px 12px;
					padding: 0;
				}

				:global(li) {
					display: flex;
					justify-content: center;
					width: 100%;
					min-width: 0;
					content-visibility: auto;
					contain-intrinsic-size: 250px 375px;
				}

				:global(.container) {
					width: 100% !important;
					max-width: 320px;
					min-width: unset !important;
					aspect-ratio: 170 / 256;
					height: auto !important;
					min-height: unset !important;
					border-radius: 12px;
					box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
				}
			}
		}
	}
</style>
