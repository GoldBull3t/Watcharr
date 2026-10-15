<script lang="ts">
	import { store } from "@/store.svelte";
	import Menu, { type MenuConfig } from "../Menu.svelte";
	import Icon from "../Icon.svelte";

	interface Props {
		menuConfig?: MenuConfig;
	}

	let { menuConfig }: Props = $props();

	const sortOptions = [
		{ id: "DATEADDED", label: "Date Added" },
		{ id: "LASTCHANGED", label: "Last Changed" },
		{ id: "LASTFIN", label: "Last Finished" },
		{ id: "RATING", label: "Rating" },
		{ id: "ALPHA", label: "Alphabetical" },
		{ id: "DATERELEASED", label: "Release Date" },
	] as const;

	function sortClicked(type: string) {
		window.scrollTo({ top: 0 });
		let mode = "UP";
		if (store.activeSort[0] === type) {
			if (store.activeSort[1] === "UP") {
				mode = "DOWN";
			} else if (store.activeSort[1] === "DOWN") {
				mode = "";
			}
		}
		if (!mode) {
			store.activeSort = [];
			return;
		}
		store.activeSort = [type, mode];
	}

	function getSortState(id: string): "UP" | "DOWN" | undefined {
		if (store.activeSort[0] !== id) return undefined;
		return store.activeSort[1] as "UP" | "DOWN" | undefined;
	}
</script>

<Menu
	conf={menuConfig || {
		width: "190px",
		right: "0px",
		top: "calc(100% + 6px)",
	}}
>
	<h5>Sort Order</h5>
	{#each sortOptions as opt (opt.id)}
		{@const sortState = getSortState(opt.id)}
		<button
			type="button"
			class="plain sort-btn"
			class:active={Boolean(sortState)}
			onclick={() => sortClicked(opt.id)}
		>
			<span class="sort-label">{opt.label}</span>
			{#if sortState === "DOWN"}
				<span class="sort-icon-badge" title="Descending">
					<Icon i="arrow" wh={13} facing="down" />
				</span>
			{:else if sortState === "UP"}
				<span class="sort-icon-badge" title="Ascending">
					<Icon i="arrow" wh={13} facing="up" />
				</span>
			{/if}
		</button>
	{/each}
</Menu>

<style lang="scss">
	button.sort-btn {
		display: flex !important;
		align-items: center;
		justify-content: space-between;
		width: 100% !important;
		padding: 7px 10px;
		font-size: 13px;
		font-weight: 500;
		color: $text-color;
		border-radius: 7px;
		border: none;
		background: transparent;
		transition: all 130ms ease;

		&:hover {
			background-color: var(--accent-color, rgba(255, 255, 255, 0.08));
			color: #ffffff;
		}

		&.active {
			color: #fbbf24;
			font-weight: 600;
			background-color: rgba(245, 158, 11, 0.12);

			.sort-icon-badge {
				color: #fbbf24;
			}
		}

		.sort-label {
			border: none !important;
			margin: 0 !important;
			padding: 0 !important;
			font-size: 13px !important;
			text-align: left !important;
		}

		.sort-icon-badge {
			display: flex;
			align-items: center;
			justify-content: center;
			color: $text-color-accent;
			border: none !important;
			margin: 0 !important;
			padding: 0 !important;
		}
	}
</style>
