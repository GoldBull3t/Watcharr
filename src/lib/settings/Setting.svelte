<script lang="ts">
	import type { HandbookEntryId } from "../handbook/entries";
	import HandbookLink from "../handbook/HandbookLink.svelte";

	interface Props {
		title?: string;
		desc?: string;
		row?: boolean;
		tag?: string | undefined;
		children?: import("svelte").Snippet;
		handbook?: HandbookEntryId;
	}

	let {
		title = "",
		desc = "",
		row = false,
		tag = undefined,
		children,
		handbook,
	}: Props = $props();
</script>

<div class={[row ? "row" : "", "setting-ctr"].join(" ")}>
	{#if row}
		<div>
			<h4 class="norm">
				{title}
				{#if handbook}
					<HandbookLink {handbook} />
				{/if}
				{#if tag}
					<span class="tag">{tag}</span>
				{/if}
			</h4>
			<h5 class="norm">{desc}</h5>
		</div>
	{:else}
		<h4 class="norm">
			{title}
			{#if tag}
				<span class="tag">{tag}</span>
			{/if}
		</h4>
		<h5 class="norm">{desc}</h5>
	{/if}

	{@render children?.()}
</div>

<style lang="scss">
	.setting-ctr {
		padding: 16px 20px;
		border-radius: 14px;
		background: rgba(255, 255, 255, 0.025);
		border: 1px solid rgba(255, 255, 255, 0.06);
		transition: background-color 150ms ease;

		&:hover {
			background: rgba(255, 255, 255, 0.035);
		}

		&.row {
			display: flex;
			flex-flow: row;
			gap: 16px;
			align-items: center;
			justify-content: space-between;

			& > div:first-of-type {
				margin-right: auto;
			}
		}

		&:not(.row) {
			display: flex;
			flex-direction: column;
			gap: 10px;

			h5 {
				margin-bottom: 4px;
			}
		}
	}

	h4 {
		font-size: 15px;
		font-weight: 600;
		color: $text-color;
		display: flex;
		flex-flow: row;
		align-items: center;
		gap: 8px;
		margin: 0 0 3px 0;

		& > .tag {
			font-size: 10px;
			font-weight: 700;
			text-transform: uppercase;
			letter-spacing: 0.05em;
			background: rgba(245, 158, 11, 0.15);
			color: #fbbf24;
			border: 1px solid rgba(245, 158, 11, 0.3);
			padding: 2px 7px;
			border-radius: 10px;
		}
	}

	h5 {
		font-weight: 400;
		font-size: 13px;
		color: $text-muted;
		line-height: 1.45;
		margin: 0;
	}
</style>
