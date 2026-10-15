<script lang="ts">
	import { store } from "@/store.svelte";
	import type { WLDetailedViewOption } from "@/types";
	import Menu, { type MenuConfig } from "../Menu.svelte";
	import Icon from "../Icon.svelte";

	interface Props {
		menuConfig?: MenuConfig;
	}

	let { menuConfig }: Props = $props();

	const detailOptions = [
		{ id: "statusRating" as const, label: "Status & Rating" },
		{ id: "lastWatched" as const, label: "Watching Season" },
		{ id: "dateAdded" as const, label: "Date Added" },
		{ id: "dateModified" as const, label: "Date Modified" },
	];

	function detailClicked(d: WLDetailedViewOption) {
		if (store.wlDetailedView.includes(d)) {
			store.wlDetailedView = store.wlDetailedView.filter((a) => a !== d);
		} else {
			store.wlDetailedView.push(d);
			store.wlDetailedView = store.wlDetailedView;
		}
	}
</script>

<Menu
	conf={menuConfig || {
		width: "205px",
		right: "0px",
		top: "calc(100% + 6px)",
	}}
>
	<h5>Shown Details</h5>
	{#each detailOptions as opt (opt.id)}
		{@const isOn = store.wlDetailedView?.includes(opt.id)}
		<button
			type="button"
			class="plain detail-btn"
			class:active={isOn}
			onclick={() => detailClicked(opt.id)}
		>
			<span class="detail-label">{opt.label}</span>
			<span class="detail-check-box" class:checked={isOn}>
				{#if isOn}
					<Icon i="check" wh={11} />
				{/if}
			</span>
		</button>
	{/each}
</Menu>

<style lang="scss">
	button.detail-btn {
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
			color: #ffffff;
		}

		.detail-label {
			border: none !important;
			margin: 0 !important;
			padding: 0 !important;
			font-size: 13px !important;
			text-align: left !important;
		}

		.detail-check-box {
			display: flex;
			align-items: center;
			justify-content: center;
			width: 16px;
			height: 16px;
			border-radius: 4px;
			border: 1px solid var(--card-border, rgba(255, 255, 255, 0.2));
			background: rgba(0, 0, 0, 0.25);
			transition: all 130ms ease;
			border: none !important;
			margin: 0 !important;
			padding: 0 !important;

			&.checked {
				background-color: #f59e0b !important;
				border-color: #f59e0b !important;
				color: #0b0e14 !important;
			}
		}
	}
</style>
