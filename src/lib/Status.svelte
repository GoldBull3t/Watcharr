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
</div>

<style lang="scss">
	.status {
		display: flex;
		flex-flow: row;
		gap: 10px;
		width: 100%;
		height: 100%;
		container-type: inline-size;

		button {
			font-size: 10px;
			padding: 5px 10px;
			height: 54px;
			flex: 1;
			display: inline-flex;
			align-items: center;
			justify-content: center;
			border-radius: 12px;
			border: 1.5px solid var(--card-border, rgba(255, 255, 255, 0.1));
			background-color: var(--card-bg, rgba(255, 255, 255, 0.04));
			color: $text-color;
			transition: all 160ms ease;
			cursor: pointer;

			&:hover {
				opacity: 1 !important;
				transform: translateY(-2px);
				border-color: rgba(255, 255, 255, 0.25);
			}

			&.untracked {
				opacity: 0.6;
			}

			&.not-active {
				opacity: 0.35;
			}

			&.active {
				opacity: 1;
				transform: translateY(-1px);

				&.planned {
					border-color: #f59e0b;
					background-color: rgba(245, 158, 11, 0.18);
					color: #fbbf24;
					box-shadow: 0 0 14px rgba(245, 158, 11, 0.25);
				}
				&.watching {
					border-color: #3b82f6;
					background-color: rgba(59, 130, 246, 0.18);
					color: #60a5fa;
					box-shadow: 0 0 14px rgba(59, 130, 246, 0.25);
				}
				&.finished {
					border-color: #10b981;
					background-color: rgba(16, 185, 129, 0.18);
					color: #34d399;
					box-shadow: 0 0 14px rgba(16, 185, 129, 0.25);
				}
				&.hold {
					border-color: #f97316;
					background-color: rgba(249, 115, 22, 0.18);
					color: #fb923c;
					box-shadow: 0 0 14px rgba(249, 115, 22, 0.25);
				}
				&.dropped {
					border-color: #ef4444;
					background-color: rgba(239, 68, 68, 0.18);
					color: #f87171;
					box-shadow: 0 0 14px rgba(239, 68, 68, 0.25);
				}
			}
		}

		/* When container starts to shrink make button height responsive */
		@container (width < 380px) {
			button {
				height: unset;
				padding: 8px 6px;
			}
		}
	}
</style>
