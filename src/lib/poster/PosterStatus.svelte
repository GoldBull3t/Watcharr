<script lang="ts">
	import type { WatchedStatus } from "@/types";
	import Icon from "../Icon.svelte";
	import tooltip from "../actions/tooltip";
	import { toUnderstandableStatus, watchedStatuses } from "../util/helpers";

	interface Props {
		status?: WatchedStatus | undefined;
		handleStatusClick: (status: WatchedStatus | "DELETE") => void;
		direction?: "top" | "bot";
		minimal?: boolean;
		btnTooltip?: string;
		disableInteraction?: boolean;
		isForGame?: boolean;
	}

	let {
		status = undefined,
		handleStatusClick,
		direction = "top",
		minimal = false,
		btnTooltip = "",
		disableInteraction = false,
		isForGame = false,
	}: Props = $props();

	let statusesShown = $state(false);
	let buttonEl: HTMLButtonElement | undefined = $state();
	let openUpward = $state(false);

	function toggleMenu(ev: MouseEvent) {
		ev.stopPropagation();
		if (!statusesShown && buttonEl) {
			const rect = buttonEl.getBoundingClientRect();
			const spaceBelow = window.innerHeight - rect.bottom;
			openUpward = spaceBelow < 240 && rect.top > spaceBelow;
		}
		statusesShown = !statusesShown;
	}

	function handleWindowClick(e: MouseEvent) {
		if (statusesShown && buttonEl && !buttonEl.contains(e.target as Node)) {
			statusesShown = false;
		}
	}

	function handleWindowKeyDown(e: KeyboardEvent) {
		if (e.key === "Escape" && statusesShown) {
			statusesShown = false;
		}
	}
</script>

<svelte:window onclick={handleWindowClick} onkeydown={handleWindowKeyDown} />

<div
	class={[
		"status-wrap",
		minimal ? "minimal" : "",
		statusesShown ? "open" : "",
	].join(" ")}
>
	<button
		bind:this={buttonEl}
		type="button"
		class={[
			"status-btn",
			disableInteraction ? "interaction-disabled" : "",
			statusesShown ? "open" : "",
		].join(" ")}
		class:minimal
		onclick={toggleMenu}
		aria-haspopup="true"
		aria-expanded={statusesShown}
		use:tooltip={{
			text: btnTooltip,
			pos: "top",
			condition: !!btnTooltip && !statusesShown,
		}}
	>
		{#if status}
			<Icon i={watchedStatuses[status]} wh={minimal ? 15 : 17} />
		{:else}
			<span class={["no-icon", minimal ? "minimal" : ""].join(" ")}>
				<Icon i="add" wh={minimal ? 14 : 16} />
			</span>
		{/if}
	</button>

	{#if statusesShown}
		<div
			class={[
				"status-dropdown",
				"shown",
				"small-scrollbar",
				status ? "has-status" : "",
				openUpward ? "open-upward" : direction,
			].join(" ")}
			role="menu"
		>
			{#each Object.entries(watchedStatuses) as [statusName, icon] (statusName)}
				<button
					type="button"
					class="plain{status === statusName ? ' active' : status ? ' not-active' : ''}"
					role="menuitem"
					onclick={(ev) => {
						ev.stopPropagation();
						statusesShown = false;
						handleStatusClick(statusName as WatchedStatus);
					}}
					use:tooltip={{
						text: toUnderstandableStatus(statusName as WatchedStatus, isForGame),
						pos: "left",
					}}
				>
					<Icon i={icon} wh={17} />
				</button>
			{/each}
			{#if status}
				<button
					type="button"
					class="plain delete-btn"
					role="menuitem"
					onclick={(ev) => {
						ev.stopPropagation();
						statusesShown = false;
						handleStatusClick("DELETE");
					}}
					use:tooltip={{ text: "Delete", pos: "left" }}
				>
					<Icon i="trash" wh={16} />
				</button>
			{/if}
		</div>
	{/if}
</div>

<style lang="scss">
	.status-wrap {
		position: relative;
		height: 100%;
		width: 40%;
		display: flex;

		&.open {
			z-index: 60 !important;
		}

		&.minimal {
			width: 100%;

			.status-dropdown {
				height: unset;
				max-height: 240px;
				overflow-y: auto;
				overflow-x: hidden;
			}
		}
	}

	button.status-btn {
		padding: 3px 6px;
		position: relative;
		font-family: inherit;
		font-weight: 600;
		height: 100%;
		width: 100%;
		display: inline-flex;
		align-items: center;
		justify-content: center;

		&.open {
			z-index: 60 !important;
		}

		&.interaction-disabled {
			pointer-events: none;
			cursor: default;
			background-color: transparent;
			border: unset;
			fill: white;
			color: white;

			span {
				color: white !important;
			}
		}

		.no-icon {
			color: $text-color;
			display: inline-flex;
			align-items: center;
			justify-content: center;
			height: 100%;

			&.minimal {
				height: 24px;
			}
		}

		&:hover .no-icon,
		&:focus-visible .no-icon {
			color: $rating-color;
			fill: $rating-color;
		}
	}

	.status-dropdown {
		display: flex;
		flex-flow: column;
		position: absolute;
		width: 100%;
		min-width: 48px;
		right: 0;
		left: auto;
		height: 200px;
		background: rgba(15, 19, 27, 0.96);
		border: 1px solid rgba(255, 255, 255, 0.1);
		list-style: none;
		border-radius: 10px;
		overflow-y: auto;
		overflow-x: hidden;
		scrollbar-width: thin;
		z-index: 100;
		box-shadow: 0 16px 36px -6px rgba(0, 0, 0, 0.75);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);

		&::before {
			content: "";
			position: absolute;
			left: 0;
			right: 0;
			height: 10px;
		}

		&:not(.shown) {
			display: none;
		}

		&.top,
		&.open-upward {
			bottom: calc(100% + 6px);
			top: auto;

			&::before {
				bottom: -8px;
			}
		}

		&.bot:not(.open-upward) {
			top: calc(100% + 6px);
			bottom: auto;

			&::before {
				top: -8px;
			}
		}

		button {
			width: 100%;
			color: $text-color;
			fill: $text-color;
			padding: 7px 4px;
			border-radius: 6px;
			display: flex;
			align-items: center;
			justify-content: center;
			transition: background-color 120ms ease, color 120ms ease;

			&.active {
				background-color: rgba(245, 158, 11, 0.2);
				color: #fbbf24;
				fill: #fbbf24;
			}

			& :global(svg) {
				padding: 0 2px;
			}

			&:hover,
			&:focus-visible {
				background-color: rgba(255, 255, 255, 0.08);
			}

			&.active:hover {
				background-color: rgba(245, 158, 11, 0.28);
			}

			&.delete-btn:hover {
				background-color: rgba(239, 68, 68, 0.2);
				color: #f87171;
				fill: #f87171;
			}
		}

		&.has-status :global(svg) {
			padding: 0 4.5px;
		}
	}
</style>
