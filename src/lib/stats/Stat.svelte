<script lang="ts">
	import type { ResolvedPathname } from "$app/types";
	import tooltip from "../actions/tooltip";

	interface Props {
		name: string;
		value: string | number;
		large?: boolean;
		/**
		 * **NOTE:** Make sure to use resolve() when passing in the href!
		 */
		href?: ResolvedPathname;
		disc?: string | undefined;
	}

	let {
		name,
		value,
		large = false,
		href = undefined,
		disc = undefined,
	}: Props = $props();
</script>

<a {href}>
	{#if disc}
		<div class="disclaimer" use:tooltip={{ text: disc, pos: "top" }}>*</div>
	{/if}
	<span class={large ? "large" : ""}>{value}</span>
	<span>{name}</span>
</a>

<style lang="scss">
	a {
		position: relative;
		display: flex;
		flex-flow: column;
		flex: 1 1 140px;
		padding: 18px 20px;
		background: rgba(255, 255, 255, 0.04);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 14px;
		text-decoration: none;
		color: $text-color;
		transition: all 180ms ease;

		&:hover {
			background: rgba(255, 255, 255, 0.07);
			border-color: rgba(255, 255, 255, 0.16);
			transform: translateY(-2px);
		}

		> span:first-of-type {
			font-weight: 700;
			font-size: 22px;
			letter-spacing: -0.02em;
			color: #fff;
			margin-top: auto;

			&.large {
				font-size: 32px;
				color: $accent-color;
			}
		}

		> span:last-of-type {
			font-size: 12px;
			font-weight: 500;
			color: rgba(255, 255, 255, 0.55);
			margin-top: 4px;
			text-transform: uppercase;
			letter-spacing: 0.04em;
		}

		.disclaimer {
			position: absolute;
			top: 8px;
			right: 10px;
			font-weight: 700;
			font-size: 14px;
			color: $text-muted;
			cursor: help;
		}
	}
</style>
