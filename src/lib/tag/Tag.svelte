<script lang="ts">
	import type { Tag } from "@/types";
	import {
		splitTagEmoji,
		getContrastTextColor,
		hexToRgba,
	} from "./tagUtil";

	interface Props {
		tag: Tag | { name: string; color?: string; bgColor?: string; id?: number };
		onClick?: (() => void) | undefined;
		removable?: boolean;
		onRemove?: (() => void) | undefined;
		size?: "xs" | "sm" | "md" | "lg";
		clickable?: boolean;
		active?: boolean;
		title?: string;
	}

	let {
		tag,
		onClick = undefined,
		removable = false,
		onRemove = undefined,
		size = "md",
		clickable = true,
		active = false,
		title = undefined,
	}: Props = $props();

	let parsed = $derived(splitTagEmoji(tag?.name));
	let tagBg = $derived(tag?.bgColor || "#36BA98");
	let tagColor = $derived(tag?.color || getContrastTextColor(tagBg));
	let borderColor = $derived(hexToRgba(tagBg, 0.45));
	let glowColor = $derived(hexToRgba(tagBg, 0.3));
	let isInteractive = $derived(clickable && (typeof onClick === "function" || removable));
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<button
	type="button"
	class="plain watcharr-tag-pill size-{size} {active ? 'active' : ''} {isInteractive ? 'interactive' : 'static'}"
	style:background={tagBg}
	style:color={tagColor}
	style:--tag-border={borderColor}
	style:--tag-glow={glowColor}
	title={title || tag?.name}
	onclick={() => {
		if (typeof onClick === "function") {
			onClick();
		}
	}}
>
	{#if parsed.emoji}
		<span class="tag-emoji" aria-hidden="true">{parsed.emoji}</span>
	{/if}
	<span class="tag-label">{parsed.label}</span>

	{#if removable && onRemove}
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<span
			role="button"
			tabindex="0"
			class="tag-remove-btn"
			title="Remove tag"
			aria-label={`Remove tag ${tag.name}`}
			onclick={(e) => {
				e.stopPropagation();
				onRemove();
			}}
		>
			<svg
				xmlns="http://www.w3.org/2000/svg"
				width="12"
				height="12"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2.5"
				stroke-linecap="round"
				stroke-linejoin="round"
			>
				<line x1="18" y1="6" x2="6" y2="18"></line>
				<line x1="6" y1="6" x2="18" y2="18"></line>
			</svg>
		</span>
	{/if}
</button>

<style lang="scss">
	.watcharr-tag-pill {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: max-content;
		max-width: 100%;
		border-radius: 9999px;
		border: 1px solid var(--tag-border, rgba(255, 255, 255, 0.2));
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2), 0 0 8px var(--tag-glow, transparent);
		font-weight: 600;
		line-height: 1.2;
		white-space: nowrap;
		cursor: default;
		transition: all 180ms cubic-bezier(0.16, 1, 0.3, 1);
		user-select: none;
		-webkit-user-select: none;

		&.size-xs {
			padding: 2px 7px;
			font-size: 11px;
			gap: 3px;

			.tag-emoji {
				font-size: 11px;
			}
		}

		&.size-sm {
			padding: 3px 9px;
			font-size: 12px;
			gap: 4px;

			.tag-emoji {
				font-size: 12px;
			}
		}

		&.size-md {
			padding: 5px 12px;
			font-size: 13px;
			gap: 6px;

			.tag-emoji {
				font-size: 13px;
			}
		}

		&.size-lg {
			padding: 7px 16px;
			font-size: 14px;
			gap: 8px;

			.tag-emoji {
				font-size: 15px;
			}
		}

		&.interactive {
			cursor: pointer;

			&:hover {
				transform: translateY(-1px);
				filter: brightness(1.08);
				box-shadow: 0 3px 10px rgba(0, 0, 0, 0.25), 0 0 12px var(--tag-glow, transparent);
				border-color: rgba(255, 255, 255, 0.4);
			}

			&:active {
				transform: translateY(0) scale(0.98);
			}
		}

		&.active {
			outline: 2px solid #ffffff;
			outline-offset: 1px;
		}

		.tag-emoji {
			display: inline-flex;
			align-items: center;
			justify-content: center;
			line-height: 1;
		}

		.tag-label {
			overflow: hidden;
			text-overflow: ellipsis;
			letter-spacing: 0.2px;
		}

		.tag-remove-btn {
			display: inline-flex;
			align-items: center;
			justify-content: center;
			width: 16px;
			height: 16px;
			margin-left: 2px;
			border-radius: 50%;
			background: rgba(0, 0, 0, 0.18);
			color: inherit;
			cursor: pointer;
			opacity: 0.75;
			transition: all 140ms ease;

			&:hover {
				opacity: 1;
				background: rgba(0, 0, 0, 0.35);
				transform: scale(1.1);
			}

			&:active {
				transform: scale(0.95);
			}
		}
	}
</style>
