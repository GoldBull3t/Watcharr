<script lang="ts">
	import { onMount } from "svelte";

	interface Props {
		type?: "wrapped" | "vertical";
		children?: import("svelte").Snippet;
	}

	let { type = "wrapped", children }: Props = $props();

	let ulEl: HTMLUListElement | undefined = $state();

	onMount(() => {
		if (ulEl) {
			ulEl.classList.add(type);
		}
	});
</script>

<div class="poster-list-wrapper">
	<ul bind:this={ulEl} class={type}>
		{@render children?.()}
	</ul>
</div>

<style lang="scss">
	.poster-list-wrapper {
		display: flex;
		justify-content: center;
		width: 100%;
	}

	ul {
		list-style: none;
		width: 100%;

		&:global(.vertical) {
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

		&:global(.wrapped) {
			display: grid;
			grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
			gap: 24px 18px;
			width: 100%;
			max-width: 1440px;
			padding: 0 4px;
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
		}

		&:global(.wrapped li) {
			display: flex;
			justify-content: center;
			width: 100%;
			min-width: 0;
			content-visibility: auto;
			contain-intrinsic-size: 170px 256px;
		}

		&:global(.wrapped .container) {
			width: 100% !important;
			max-width: 220px;
			min-width: unset !important;
			aspect-ratio: 170 / 256;
			height: auto !important;
			min-height: unset !important;
		}
	}
</style>
