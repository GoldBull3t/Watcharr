<script lang="ts">
	import Icon from "../Icon.svelte";
	import { store } from "@/store.svelte";

	type FilterType = "movie" | "show" | "game" | "person";

	interface Props {
		active?: string;
		disabled?: boolean;
		onChange: (nowActive: FilterType) => void;
	}

	let { active, disabled, onChange }: Props = $props();
</script>

<div class:disabled>
	<button
		class="plain"
		data-active={active === "movie"}
		onclick={() => onChange("movie")}
	>
		<Icon i="film" wh={20} /> Movies
	</button>
	<button
		class="plain"
		data-active={active === "show"}
		onclick={() => onChange("show")}
	>
		<Icon i="tv" wh={20} /> TV Shows
	</button>
	{#if store.serverFeatures?.games}
		<button
			class="plain"
			data-active={active === "game"}
			onclick={() => onChange("game")}
		>
			<Icon i="gamepad" wh={20} /> Games
		</button>
	{/if}
	<button
		class="plain"
		data-active={active === "person"}
		onclick={() => onChange("person")}
	>
		<Icon i="people-nocircle" wh={20} /> People
	</button>
</div>

<style lang="scss">
	div {
		display: inline-flex;
		flex-flow: row;
		flex-wrap: wrap;
		gap: 4px;
		padding: 4px;
		background: rgba(255, 255, 255, 0.04);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 30px;
		backdrop-filter: blur(8px);
		-webkit-backdrop-filter: blur(8px);

		button {
			display: inline-flex;
			flex-flow: row;
			gap: 7px;
			align-items: center;
			height: 34px;
			padding: 0 14px;
			border-radius: 20px;
			font-size: 13.5px;
			font-weight: 500;
			color: $text-color;
			fill: $text-color;
			border: 1px solid transparent;
			background: transparent;
			cursor: pointer;
			transition: all 180ms ease;

			&:hover {
				background: rgba(255, 255, 255, 0.08);
				color: #fff;
				fill: #fff;
			}

			&[data-active="true"] {
				background: $accent-color;
				color: #fff;
				fill: #fff;
				font-weight: 600;
				box-shadow: 0 2px 8px rgba(0, 0, 0, 0.25);
			}

			&:focus-visible {
				outline: 2px solid $accent-color;
				outline-offset: 2px;
			}
		}

		&.disabled {
			opacity: 0.6;
			pointer-events: none;
		}

		@media screen and (max-width: 600px) {
			width: 100%;
			justify-content: center;
			border-radius: 16px;

			button {
				padding: 0 10px;
				font-size: 12.5px;
				height: 32px;
			}
		}
	}
</style>
