<!-- Floating menu container for dropdowns and popovers -->
<script lang="ts">
	import stayInView from "./actions/stayInView";

	export interface MenuConfig {
		width?: string;
		top?: string;
		right?: string;
		left?: string;
		arrowRight?: string;
		arrowLeft?: string;
		arrowColor?: string;
	}

	interface Props {
		children: import("svelte").Snippet;
		conf?: MenuConfig;
	}

	let { children, conf }: Props = $props();
</script>

<div
	class="floating-menu"
	style={`--w: ${conf?.width || "180px"}; --r: ${
		conf?.right
			? conf.right
			: conf?.left && conf.left !== "unset"
				? "unset"
				: "0px"
	}; --l: ${conf?.left || "unset"}; --t: ${conf?.top || "calc(100% + 6px)"}; --origin: ${
		conf?.left && conf.left !== "unset" ? "top left" : "top right"
	};`}
	use:stayInView={{}}
>
	<div class="menu-content small-scrollbar">
		{@render children?.()}
	</div>
</div>

<style lang="scss">
	.floating-menu {
		display: flex;
		flex-flow: column;
		position: absolute;
		right: var(--r);
		left: var(--l);
		top: var(--t);
		width: var(--w);
		min-width: 160px;
		max-width: 90vw;
		border: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
		border-radius: 12px;
		background: rgba(15, 19, 27, 0.96);
		box-shadow:
			0 20px 48px -8px rgba(0, 0, 0, 0.75),
			0 0 0 1px rgba(255, 255, 255, 0.04);
		backdrop-filter: blur(24px);
		-webkit-backdrop-filter: blur(24px);
		list-style: none;
		z-index: 1000;
		transform-origin: var(--origin, top right);
		animation: popoverFadeIn 140ms cubic-bezier(0.16, 1, 0.3, 1) forwards;

		@keyframes popoverFadeIn {
			from {
				opacity: 0;
				transform: translateY(-4px) scale(0.97);
			}
			to {
				opacity: 1;
				transform: translateY(0) scale(1);
			}
		}

		.menu-content {
			display: flex;
			flex-flow: column;
			gap: 3px;
			padding: 6px;
			width: 100%;
			max-height: calc(100dvh - 80px);
			overflow-y: auto;

			:global {
				h5 {
					margin: 0;
					padding: 6px 10px 8px 10px;
					font-size: 11px;
					font-weight: 700;
					color: $text-color-accent;
					letter-spacing: 0.06em;
					text-transform: uppercase;
					white-space: nowrap;
					overflow: hidden;
					text-overflow: ellipsis;
					cursor: default;
					border-bottom: 1px solid var(--card-border, rgba(255, 255, 255, 0.07));
				}

				button,
				a {
					font-size: 12.5px;
					font-weight: 500;
					padding: 7px 10px;
					text-align: left;
					border-radius: 7px;
					cursor: pointer;
					color: $text-color;
					border: none;
					background: transparent;
					width: 100% !important;
					transition:
						background-color 130ms ease,
						color 130ms ease;

					&:hover,
					&:focus-visible {
						background-color: var(--accent-color, rgba(255, 255, 255, 0.08));
						color: #ffffff;
					}

					&.active {
						color: #fbbf24;
						background-color: rgba(245, 158, 11, 0.12);
						font-weight: 600;
					}
				}

				.menu-footer-wrap {
					display: flex;
					align-items: center;
					justify-content: center;
					gap: 6px;
					margin-top: 6px;
					padding-top: 6px;
					font-size: 11px;
					color: $text-color-accent;
					border-top: 1px solid var(--card-border, rgba(255, 255, 255, 0.07));

					a.menu-footer,
					button.menu-footer {
						font-size: 11px;
						padding: 2px 4px;
						width: auto !important;
						color: $text-color-accent;

						&:hover {
							color: #ffffff;
							text-decoration: underline;
							background: transparent;
						}
					}
				}
			}
		}
	}
</style>
