<script lang="ts">
	import Icon from "./Icon.svelte";

	interface Props {
		title: string | undefined;
		children?: import("svelte").Snippet;
	}

	let { title, children }: Props = $props();
	let listEl: HTMLUListElement | undefined = $state();

	function scrollList(dir: -1 | 1) {
		if (!listEl) return;
		const amount = Math.max(listEl.clientWidth * 0.75, 320);
		listEl.scrollBy({ left: dir * amount, behavior: "smooth" });
	}
</script>

<section class="horiz-section">
	{#if title}
		<div class="horiz-header">
			<h2>{title}</h2>
			<div class="scroll-controls">
				<button
					type="button"
					class="plain scroll-btn"
					onclick={() => scrollList(-1)}
					aria-label="Scroll left"
				>
					<Icon i="chevron" wh={14} facing="left" />
				</button>
				<button
					type="button"
					class="plain scroll-btn"
					onclick={() => scrollList(1)}
					aria-label="Scroll right"
				>
					<Icon i="chevron" wh={14} facing="right" />
				</button>
			</div>
		</div>
	{/if}
	<ul bind:this={listEl}>
		{@render children?.()}
	</ul>
</section>

<style lang="scss">
	.horiz-section {
		width: 100%;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.horiz-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;

		h2 {
			margin: 0;
			font-size: 18px;
			font-weight: 700;
			letter-spacing: -0.01em;
			color: #ffffff;
		}

		.scroll-controls {
			display: inline-flex;
			align-items: center;
			gap: 6px;

			.scroll-btn {
				display: inline-flex;
				align-items: center;
				justify-content: center;
				width: 30px;
				height: 30px;
				border-radius: 6px;
				background: rgba(255, 255, 255, 0.05);
				border: 1px solid rgba(255, 255, 255, 0.1);
				color: rgba(255, 255, 255, 0.75);
				cursor: pointer;
				transition:
					background 140ms ease,
					border-color 140ms ease,
					color 140ms ease;

				&:hover {
					background: rgba(255, 255, 255, 0.1);
					border-color: rgba(255, 255, 255, 0.22);
					color: #ffffff;
				}
			}
		}
	}

	ul {
		display: flex;
		flex-flow: row;
		gap: 12px;
		list-style: none;
		margin: 0;
		padding: 4px 0 8px 0;
		overflow-x: auto;
		overscroll-behavior-x: contain;
		scrollbar-width: none;

		&::-webkit-scrollbar {
			display: none;
		}
	}
</style>
