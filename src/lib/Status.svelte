<script lang="ts">
	import Icon from "./Icon.svelte";
	import type { WatchedStatus } from "../types";
	import tooltip from "./actions/tooltip";
	import { toUnderstandableStatus } from "./util/helpers";

	interface Props {
		status: WatchedStatus | undefined;
		isForGame?: boolean;
		onChange: (newStatus: WatchedStatus) => void;
	}

	let { status, isForGame = false, onChange }: Props = $props();

	function handleStatusClick(s: WatchedStatus) {
		if (s === status) return;
		onChange(s);
	}
</script>

<div class="status">
	<button
		class={`status-btn ${status === "PLANNED" ? "active planned" : status ? "not-active" : "untracked"}`}
		onclick={() => handleStatusClick("PLANNED")}
		use:tooltip={{
			text: toUnderstandableStatus("PLANNED", isForGame),
			pos: "top",
		}}
		aria-label={toUnderstandableStatus("PLANNED", isForGame)}
	>
		<Icon i="calendar" wh={20} />
	</button>
	<button
		class={`status-btn ${status === "WATCHING" ? "active watching" : status ? "not-active" : "untracked"}`}
		onclick={() => handleStatusClick("WATCHING")}
		use:tooltip={{
			text: toUnderstandableStatus("WATCHING", isForGame),
			pos: "top",
		}}
		aria-label={toUnderstandableStatus("WATCHING", isForGame)}
	>
		<Icon i="clock" wh={20} />
	</button>
	<button
		class={`status-btn ${status === "FINISHED" ? "active finished" : status ? "not-active" : "untracked"}`}
		onclick={() => handleStatusClick("FINISHED")}
		use:tooltip={{
			text: toUnderstandableStatus("FINISHED", isForGame),
			pos: "top",
		}}
		aria-label={toUnderstandableStatus("FINISHED", isForGame)}
	>
		<Icon i="check" wh={20} />
	</button>
	<button
		class={`status-btn ${status === "HOLD" ? "active hold" : status ? "not-active" : "untracked"}`}
		onclick={() => handleStatusClick("HOLD")}
		use:tooltip={{
			text: toUnderstandableStatus("HOLD", isForGame),
			pos: "top",
		}}
		aria-label={toUnderstandableStatus("HOLD", isForGame)}
	>
		<Icon i="pause" wh={20} />
	</button>
	<button
		class={`status-btn ${status === "DROPPED" ? "active dropped" : status ? "not-active" : "untracked"}`}
		onclick={() => handleStatusClick("DROPPED")}
		use:tooltip={{
			text: toUnderstandableStatus("DROPPED", isForGame),
			pos: "top",
		}}
		aria-label={toUnderstandableStatus("DROPPED", isForGame)}
	>
		<Icon i="thumb-down" wh={20} />
	</button>
	<button
		class={`status-btn ${status === "CANCELLED" ? "active cancelled" : status ? "not-active" : "untracked"}`}
		onclick={() => handleStatusClick("CANCELLED")}
		use:tooltip={{
			text: toUnderstandableStatus("CANCELLED", isForGame),
			pos: "top",
		}}
		aria-label={toUnderstandableStatus("CANCELLED", isForGame)}
	>
		<Icon i="close-circle" wh={20} />
	</button>
</div>

<style lang="scss">
	.status {
		display: flex;
		flex-flow: row;
		gap: 6px;
		width: 100%;
		height: 100%;
		container-type: inline-size;

		button {
			font-size: 10px;
			padding: 6px 8px;
			height: 42px;
			flex: 1;
			display: inline-flex;
			align-items: center;
			justify-content: center;
			border-radius: 8px;
			border: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
			background-color: var(--card-bg, rgba(255, 255, 255, 0.04));
			color: $text-color;
			transition:
				border-color 140ms ease,
				background-color 140ms ease,
				opacity 140ms ease;
			cursor: pointer;

			&:hover {
				opacity: 1 !important;
				border-color: rgba(255, 255, 255, 0.25);
				background-color: rgba(255, 255, 255, 0.08);
			}

			&.untracked {
				opacity: 0.65;
			}

			&.not-active {
				opacity: 0.4;
			}

			&.active {
				opacity: 1;

				&.planned {
					border-color: rgba(245, 158, 11, 0.55);
					background-color: rgba(245, 158, 11, 0.16);
					color: #fbbf24;
				}
				&.watching {
					border-color: rgba(59, 130, 246, 0.55);
					background-color: rgba(59, 130, 246, 0.16);
					color: #60a5fa;
				}
				&.finished {
					border-color: rgba(16, 185, 129, 0.55);
					background-color: rgba(16, 185, 129, 0.16);
					color: #34d399;
				}
				&.hold {
					border-color: rgba(249, 115, 22, 0.55);
					background-color: rgba(249, 115, 22, 0.16);
					color: #fb923c;
				}
				&.dropped {
					border-color: rgba(239, 68, 68, 0.55);
					background-color: rgba(239, 68, 68, 0.16);
					color: #f87171;
				}
				&.cancelled {
					border-color: rgba(220, 38, 38, 0.55);
					background-color: rgba(220, 38, 38, 0.16);
					color: #fca5a5;
				}
			}
		}

		/* When container starts to shrink make button height responsive */
		@container (width < 320px) {
			button {
				height: 38px;
				padding: 6px 4px;
			}
		}
	}
</style>
