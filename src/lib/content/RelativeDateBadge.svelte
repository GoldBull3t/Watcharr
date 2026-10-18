<!--
  Semantic Relative Date Urgency Badge
  Displays dynamic countdown indicators (Hoje, Amanhã, Em X dias)
  with semantic color accents (Emerald, Amber, Indigo, Slate).
-->
<script lang="ts">
	import { getRelativeDateInfo, type DateUrgency } from "@/lib/util/helpers";
	import Icon from "@/lib/Icon.svelte";

	interface Props {
		date: Date | string | number | undefined | null;
		prefix?: string;
		showIcon?: boolean;
		showDot?: boolean;
		size?: "sm" | "md" | "lg";
	}

	let {
		date,
		prefix,
		showIcon = true,
		showDot = true,
		size = "md",
	}: Props = $props();

	let info = $derived(getRelativeDateInfo(date));
</script>

{#if info.label}
	<span
		class="relative-date-badge size-{size} urgency-{info.urgency}"
		title={info.formattedDate
			? `${prefix ? prefix + ": " : ""}${info.formattedDate}`
			: undefined}
	>
		{#if showDot && (info.urgency === "today" || info.urgency === "tomorrow")}
			<span class="status-pulse-dot" class:is-today={info.urgency === "today"}
			></span>
		{:else if showIcon}
			<Icon
				i={info.urgency === "today" ? "clock" : "calendar"}
				wh={size === "sm" ? 11 : 12}
			/>
		{/if}

		{#if prefix}
			<span class="prefix">{prefix}</span>
		{/if}

		<span class="label">{info.label}</span>
	</span>
{/if}

<style lang="scss">
	.relative-date-badge {
		display: inline-flex;
		align-items: center;
		gap: 5px;
		line-height: 1;
		font-weight: 700;
		letter-spacing: 0.3px;
		border-radius: 999px;
		white-space: nowrap;
		user-select: none;
		transition:
			background 140ms ease,
			border-color 140ms ease,
			color 140ms ease,
			transform 140ms ease;

		.prefix {
			font-weight: 500;
			opacity: 0.75;
		}

		.status-pulse-dot {
			width: 6.5px;
			height: 6.5px;
			border-radius: 50%;
			flex-shrink: 0;

			&.is-today {
				background: #34d399;
				box-shadow: 0 0 8px #34d399;
				animation: pulseToday 2s infinite cubic-bezier(0.4, 0, 0.6, 1);
			}

			&:not(.is-today) {
				background: #fbbf24;
				box-shadow: 0 0 6px #fbbf24;
			}
		}

		@keyframes pulseToday {
			0%,
			100% {
				opacity: 1;
				transform: scale(1);
			}
			50% {
				opacity: 0.45;
				transform: scale(1.25);
			}
		}

		/* Sizes */
		&.size-sm {
			padding: 2.5px 7px;
			font-size: 10.5px;
		}

		&.size-md {
			padding: 4px 10px;
			font-size: 12px;
		}

		&.size-lg {
			padding: 5px 12px;
			font-size: 13px;
		}

		/* Urgency Color Palettes */
		&.urgency-today {
			background: rgba(16, 185, 129, 0.2);
			border: 1px solid rgba(52, 211, 153, 0.45);
			color: #6ee7b7;
			box-shadow: 0 0 12px rgba(52, 211, 153, 0.15);

			&:hover {
				background: rgba(16, 185, 129, 0.28);
				border-color: rgba(52, 211, 153, 0.6);
			}
		}

		&.urgency-tomorrow {
			background: rgba(245, 158, 11, 0.18);
			border: 1px solid rgba(251, 191, 36, 0.45);
			color: #fbbf24;
			box-shadow: 0 0 10px rgba(245, 158, 11, 0.12);

			&:hover {
				background: rgba(245, 158, 11, 0.26);
				border-color: rgba(251, 191, 36, 0.6);
			}
		}

		&.urgency-this-week {
			background: rgba(99, 102, 241, 0.16);
			border: 1px solid rgba(129, 140, 248, 0.35);
			color: #a5b4fc;

			&:hover {
				background: rgba(99, 102, 241, 0.24);
				border-color: rgba(129, 140, 248, 0.5);
			}
		}

		&.urgency-upcoming {
			background: rgba(14, 165, 233, 0.14);
			border: 1px solid rgba(56, 189, 248, 0.28);
			color: #7dd3fc;

			&:hover {
				background: rgba(14, 165, 233, 0.22);
				border-color: rgba(56, 189, 248, 0.4);
			}
		}

		&.urgency-yesterday {
			background: rgba(244, 63, 94, 0.14);
			border: 1px solid rgba(251, 113, 133, 0.3);
			color: #fda4af;
		}

		&.urgency-past,
		&.urgency-unknown {
			background: rgba(255, 255, 255, 0.06);
			border: 1px solid rgba(255, 255, 255, 0.1);
			color: rgba(255, 255, 255, 0.65);

			&:hover {
				background: rgba(255, 255, 255, 0.1);
				color: #ffffff;
			}
		}
	}
</style>
