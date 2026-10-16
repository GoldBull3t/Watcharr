<script lang="ts">
	import { store } from "@/store.svelte";
	import SpinnerTiny from "./SpinnerTiny.svelte";
	import { unNotify, pauseNotify, resumeNotify } from "./util/notify";
	import Icon from "./Icon.svelte";
</script>

<div id="notifications">
	{#each store.notifications as n (n.id)}
		<div
			class={`${n.type || "info"} notif`}
			onmouseenter={() => n.id && pauseNotify(n.id)}
			onmouseleave={() => n.id && resumeNotify(n.id)}
			role="status"
			aria-live="polite"
		>
			{#if n.type === "loading"}
				<SpinnerTiny />
			{:else if n.type === "success"}
				<span class="status-icon success">
					<Icon i="check" wh={16} />
				</span>
			{:else if n.type === "error"}
				<span class="status-icon error">
					<Icon i="close" wh={16} />
				</span>
			{:else}
				<span class="status-icon info">
					<Icon i="sparkles" wh={16} />
				</span>
			{/if}

			<!-- only comes from our strings (which may have html) -->
			<!-- eslint-disable-next-line svelte/no-at-html-tags -->
			<span class="text">{@html n.text}</span>

			{#if n.action}
				<button
					type="button"
					class="undo-action-btn"
					onclick={() => n.action?.onClick()}
				>
					{#if n.action.icon}
						<Icon i={n.action.icon} wh={14} />
					{/if}
					<span>{n.action.label}</span>
				</button>
			{/if}

			<button
				class="plain close-btn"
				title="Dismiss"
				aria-label="Dismiss notification"
				onclick={() => {
					unNotify(n.id);
				}}
			>
				<Icon i="close" wh={14} />
			</button>

			{#if n.showProgress && n.time && n.type !== "loading"}
				<div
					class="progress-bar"
					style="animation-duration: {n.time}ms;"
				></div>
			{/if}
		</div>
	{/each}
</div>

<style lang="scss">
	#notifications {
		display: flex;
		flex-flow: column;
		gap: 10px;
		position: fixed;
		bottom: 28px;
		left: 50%;
		transform: translateX(-50%);
		z-index: 99999;
		pointer-events: none;
		max-width: 90vw;

		.notif {
			pointer-events: auto;
			display: flex;
			flex-flow: row;
			align-items: center;
			min-width: 280px;
			max-width: 520px;
			color: #f8fafc;
			background-color: rgba(13, 17, 28, 0.95);
			border-radius: 14px;
			border: 1px solid rgba(255, 255, 255, 0.12);
			box-shadow:
				0 16px 36px -4px rgba(0, 0, 0, 0.7),
				0 0 24px rgba(99, 102, 241, 0.14);
			backdrop-filter: blur(20px) saturate(180%);
			-webkit-backdrop-filter: blur(20px) saturate(180%);
			animation: toastSlideUp 180ms cubic-bezier(0.16, 1, 0.3, 1) forwards;
			position: relative;
			padding: 8px 12px 10px 14px;
			gap: 10px;
			overflow: hidden;
			transition: transform 120ms ease;

			&:hover {
				transform: scale(1.01);

				.progress-bar {
					animation-play-state: paused;
				}
			}

			@keyframes toastSlideUp {
				from {
					opacity: 0;
					transform: translateY(16px) scale(0.95);
				}
				to {
					opacity: 1;
					transform: translateY(0) scale(1);
				}
			}

			&.loading {
				border-left: 3.5px solid #6366f1;
			}

			&.error {
				border-left: 3.5px solid #ef4444;
				background-color: rgba(30, 14, 20, 0.95);
			}

			&.success {
				border-left: 3.5px solid #10b981;
				background-color: rgba(14, 32, 24, 0.95);
			}

			&.info {
				border-left: 3.5px solid #818cf8;
				background-color: rgba(16, 20, 36, 0.95);
			}

			.status-icon {
				display: inline-flex;
				align-items: center;
				justify-content: center;
				flex-shrink: 0;

				&.success {
					color: #34d399;
				}
				&.error {
					color: #f87171;
				}
				&.info {
					color: #818cf8;
				}
			}

			span.text {
				width: 100%;
				font-size: 13.5px;
				font-weight: 500;
				line-height: 1.4;
				color: #e2e8f0;

				:global(b) {
					color: #ffffff;
					font-weight: 700;
				}

				:global(a) {
					color: #818cf8;
					text-decoration: underline;

					&:hover {
						color: #a5b4fc;
					}
				}
			}

			.undo-action-btn {
				display: inline-flex;
				align-items: center;
				gap: 6px;
				padding: 5px 12px;
				border-radius: 8px;
				background: rgba(99, 102, 241, 0.22);
				border: 1px solid rgba(129, 140, 248, 0.45);
				color: #c7d2fe;
				font-size: 12.5px;
				font-weight: 700;
				cursor: pointer;
				white-space: nowrap;
				flex-shrink: 0;
				transition:
					background 120ms ease,
					border-color 120ms ease,
					transform 120ms ease,
					box-shadow 120ms ease;

				&:hover {
					background: rgba(99, 102, 241, 0.4);
					border-color: rgba(129, 140, 248, 0.85);
					color: #ffffff;
					transform: translateY(-1px) scale(1.03);
					box-shadow: 0 4px 14px rgba(99, 102, 241, 0.35);
				}

				&:active {
					transform: scale(0.97);
				}
			}

			.close-btn {
				display: flex;
				align-items: center;
				justify-content: center;
				width: 24px;
				height: 24px;
				border-radius: 6px;
				color: #94a3b8;
				background: transparent;
				border: none;
				cursor: pointer;
				flex-shrink: 0;
				transition: all 120ms ease;

				&:hover {
					color: #ffffff;
					background-color: rgba(255, 255, 255, 0.1);
				}
			}

			.progress-bar {
				position: absolute;
				bottom: 0;
				left: 0;
				height: 2.5px;
				background: linear-gradient(90deg, #6366f1, #818cf8, #38bdf8);
				box-shadow: 0 0 8px rgba(99, 102, 241, 0.8);
				animation-name: progressDrain;
				animation-timing-function: linear;
				animation-fill-mode: forwards;
			}

			@keyframes progressDrain {
				from {
					width: 100%;
				}
				to {
					width: 0%;
				}
			}
		}
	}
</style>
