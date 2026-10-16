<script module>
	let activeModalCount = 0;
	let originalBodyOverflow = "";
</script>

<script lang="ts">
	import Icon from "./Icon.svelte";
	import { onMount } from "svelte";

	interface Props {
		title: string;
		desc?: string | undefined;
		onClose?: (() => void) | undefined;
		maxWidth?: string;
		error?: string | undefined;
		children?: import("svelte").Snippet;
	}

	let {
		title,
		desc = undefined,
		onClose = undefined,
		maxWidth = "1000px",
		error = undefined,
		children,
	}: Props = $props();

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === "Escape" && onClose) {
			onClose();
		}
	}

	onMount(() => {
		if (activeModalCount === 0) {
			originalBodyOverflow = document.body.style.overflow;
			document.body.style.overflow = "hidden";
		}
		activeModalCount++;
		window.addEventListener("keydown", handleKeydown);

		return () => {
			window.removeEventListener("keydown", handleKeydown);
			activeModalCount = Math.max(0, activeModalCount - 1);
			if (activeModalCount === 0) {
				document.body.style.overflow = originalBodyOverflow;
			}
		};
	});
</script>

<div
	class="backdrop"
	onclick={onClose}
	role="presentation"
></div>
<div class="modal" role="dialog" aria-modal="true" aria-label={title}>
	<div class="modal-card" style="max-width:{maxWidth};">
		{#if typeof onClose !== "undefined"}
			<button
				type="button"
				class="close"
				onclick={onClose}
				aria-label="Close dialog"
			>
				<Icon i="close" wh="18" />
			</button>
		{/if}
		<h3 class="norm">{title}</h3>
		{#if desc}
			<h5 class="norm">{desc}</h5>
		{/if}
		{#if error}
			<span class="error">{error}</span>
		{/if}
		{@render children?.()}
	</div>
</div>

<style lang="scss">
	.backdrop {
		position: fixed;
		top: 0;
		left: 0;
		width: 100dvw;
		height: 100dvh;
		backdrop-filter: blur(12px) saturate(160%);
		-webkit-backdrop-filter: blur(12px) saturate(160%);
		background-color: rgba(0, 0, 0, 0.72);
		z-index: 99998;
		animation: fadeIn 180ms cubic-bezier(0.16, 1, 0.3, 1) forwards;
	}

	@keyframes fadeIn {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	@keyframes scaleUp {
		from {
			opacity: 0;
			transform: scale(0.96) translateY(10px);
		}
		to {
			opacity: 1;
			transform: scale(1) translateY(0);
		}
	}

	.modal {
		display: flex;
		justify-content: center;
		align-items: center;
		width: 100dvw;
		height: 100dvh;
		top: 0;
		left: 0;
		position: fixed;
		z-index: 99998;
		pointer-events: none;

		.modal-card {
			pointer-events: auto;
			position: relative;
			display: flex;
			flex-flow: column;
			min-width: 300px;
			width: 100%;
			max-height: calc(100% - 40px);
			color: $text-color;
			background: var(--glass-modal-bg, rgba(16, 21, 30, 0.94));
			border: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
			border-radius: 16px;
			padding: 24px;
			margin: 20px 40px;
			box-shadow:
				0 24px 64px -12px rgba(0, 0, 0, 0.8),
				0 0 0 1px rgba(255, 255, 255, 0.05);
			backdrop-filter: blur(24px);
			-webkit-backdrop-filter: blur(24px);
			overflow-y: auto;
			animation: scaleUp 220ms cubic-bezier(0.16, 1, 0.3, 1) forwards;

			h3 {
				font-size: 18px;
				font-weight: 700;
				letter-spacing: -0.015em;
				margin-bottom: 4px;
				color: $text-color;
			}

			h5 {
				font-size: 13px;
				font-weight: 400;
				color: $text-color-accent;
				margin-bottom: 16px;
			}

			button.close {
				position: absolute;
				top: 14px;
				right: 14px;
				display: inline-flex;
				align-items: center;
				justify-content: center;
				width: 30px;
				height: 30px;
				padding: 0;
				border-radius: 8px;
				border: 1px solid transparent;
				background-color: rgba(255, 255, 255, 0.04);
				color: $text-color-accent;
				cursor: pointer;
				transition: all 140ms ease;

				&:hover {
					background-color: var(--accent-color, rgba(255, 255, 255, 0.1));
					color: #ffffff;
					border-color: rgba(255, 255, 255, 0.12);
					transform: scale(1.05);
				}
			}

			@media screen and (max-width: 680px) {
				margin: 16px;
				padding: 18px;
			}

			@media screen and (max-width: 360px) {
				min-width: unset;
				margin: 8px;
				padding: 14px;
				width: calc(100% - 16px);
			}

			.error {
				position: sticky;
				top: 0;
				display: flex;
				justify-content: center;
				width: 100%;
				padding: 10px 14px;
				border-radius: 8px;
				background-color: rgba(239, 68, 68, 0.15);
				border: 1px solid rgba(239, 68, 68, 0.35);
				text-transform: capitalize;
				color: #fca5a5;
				font-size: 13px;
				font-weight: 500;
				margin-bottom: 16px;
			}
		}
	}
</style>
