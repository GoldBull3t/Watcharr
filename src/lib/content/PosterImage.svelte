<script lang="ts">
	interface Props {
		src: string;
		alt?: string;
		onclick?: () => void;
	}

	let { src, alt = "Media poster", onclick }: Props = $props();
</script>

{#if onclick}
	<button
		type="button"
		class="poster-interactive-wrap"
		{onclick}
		title="Ampliar arte e ver em alta resolução"
	>
		<img {src} {alt} loading="lazy" decoding="async" />
		<div class="poster-zoom-overlay">
			<span class="zoom-pill">🔍 Ampliar</span>
		</div>
	</button>
{:else}
	<img {src} {alt} loading="lazy" decoding="async" />
{/if}

<style lang="scss">
	img {
		view-transition-name: active-poster;
		width: 260px;
		max-width: 100%;
		aspect-ratio: 2 / 3;
		object-fit: cover;
		box-shadow:
			0 20px 48px -10px rgba(0, 0, 0, 0.75),
			0 0 0 1px rgba(255, 255, 255, 0.1);
		border-radius: 18px;
		flex-shrink: 0;
		transition:
			transform 200ms ease,
			box-shadow 200ms ease;

		&:hover {
			transform: scale(1.015);
			box-shadow:
				0 24px 54px -10px rgba(0, 0, 0, 0.85),
				0 0 20px rgba(99, 102, 241, 0.2),
				0 0 0 1px rgba(255, 255, 255, 0.16);
		}
	}

	.poster-interactive-wrap {
		position: relative;
		display: inline-block;
		padding: 0;
		background: transparent;
		border: none;
		cursor: pointer;
		border-radius: 18px;

		img {
			display: block;
		}

		.poster-zoom-overlay {
			position: absolute;
			inset: 0;
			border-radius: 18px;
			background: rgba(0, 0, 0, 0.2);
			display: flex;
			align-items: flex-end;
			justify-content: center;
			padding-bottom: 14px;
			opacity: 0;
			pointer-events: none;
			transition:
				opacity 180ms ease,
				background 180ms ease;

			.zoom-pill {
				display: inline-flex;
				align-items: center;
				gap: 4px;
				padding: 5px 12px;
				border-radius: 20px;
				background: rgba(15, 23, 42, 0.8);
				border: 1px solid rgba(255, 255, 255, 0.2);
				color: #ffffff;
				font-size: 11.5px;
				font-weight: 600;
				box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
				backdrop-filter: blur(8px);
				transform: translateY(4px);
				transition: transform 180ms ease;
			}
		}

		&:hover {
			img {
				transform: scale(1.02);
			}

			.poster-zoom-overlay {
				opacity: 1;

				.zoom-pill {
					transform: translateY(0);
				}
			}
		}
	}
</style>
