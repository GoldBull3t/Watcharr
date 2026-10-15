<script lang="ts">
	import { store } from "@/store.svelte";
	import SpinnerTiny from "./SpinnerTiny.svelte";
	import { unNotify } from "./util/notify";
	import Icon from "./Icon.svelte";
</script>

<div id="notifications">
	{#each store.notifications as n (n.id)}
		<div class={`${n.type} notif`}>
			{#if n.type === "loading"}
				<SpinnerTiny />
			{/if}
			<!-- only comes from our strings (which may have html) -->
			<!-- eslint-disable-next-line -->
			<span>{@html n.text}</span>
			<button
				class="plain"
				onclick={() => {
					unNotify(n.id);
				}}
			>
				<Icon i="close" />
			</button>
		</div>
	{/each}
</div>

<style lang="scss">
	#notifications {
		display: flex;
		flex-flow: column;
		gap: 10px;
		position: fixed;
		bottom: 24px;
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
			min-width: 240px;
			max-width: 480px;
			color: #f8fafc;
			background-color: rgba(21, 26, 38, 0.95);
			border-radius: 10px;
			border: 1px solid var(--card-border, rgba(255, 255, 255, 0.12));
			box-shadow: 0 12px 32px rgba(0, 0, 0, 0.45);
			backdrop-filter: blur(12px);
			-webkit-backdrop-filter: blur(12px);
			animation: toastIn 200ms cubic-bezier(0.16, 1, 0.3, 1) forwards;
			position: relative;
			padding: 4px;

			&.loading {
				border-left: 3px solid $brand-color;
				padding-left: 12px;
			}

			@keyframes toastIn {
				from {
					opacity: 0;
					transform: translateY(12px) scale(0.96);
				}
				to {
					opacity: 1;
					transform: translateY(0) scale(1);
				}
			}

			&.error {
				border-left: 3px solid #ef4444;
				background-color: rgba(35, 18, 22, 0.95);
			}

			&.success {
				border-left: 3px solid #10b981;
				background-color: rgba(18, 35, 26, 0.95);
			}

			span {
				width: 100%;
				padding: 10px 14px;
				font-size: 13.5px;
				font-weight: 500;
				line-height: 1.4;

				:global(a) {
					color: #818cf8;
					text-decoration: underline;

					&:hover {
						color: #a5b4fc;
					}
				}
			}

			button {
				display: flex;
				align-items: center;
				justify-content: center;
				margin-right: 6px;
				width: 26px;
				height: 26px;
				border-radius: 6px;
				color: #94a3b8;
				background: transparent;
				border: none;
				cursor: pointer;
				transition: all 120ms ease;

				&:hover {
					color: #ffffff;
					background-color: rgba(255, 255, 255, 0.1);
				}
			}
		}
	}
</style>
