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
			<span class="zoom-pill">Ampliar Arte</span>
		</div>
	</button>
{:else}
	<img {src} {alt} loading="lazy" decoding="async" />
{/if}

<style lang="scss">
	img {
		width: 240px;
		max-width: 100%;
		aspect-ratio: 2 / 3;
		object-fit: cover;
		box-shadow:
			0 12px 32px rgba(0, 0, 0, 0.65),
			0 0 0 1px rgba(255, 255, 255, 0.1);
		border-radius: 10px;
		flex-shrink: 0;
		transition:
			transform 180ms ease,
			box-shadow 180ms ease;

		@media screen and (max-width: 1120px) {
			width: 200px;
		}

		&:hover {
			transform: translateY(-2px);
			box-shadow:
				0 16px 36px rgba(0, 0, 0, 0.75),
				0 0 0 1px rgba(255, 255, 255, 0.22);
		}
	}

	.poster-interactive-wrap {
		position: relative;
		display: inline-block;
		padding: 0;
		background: transparent;
		border: none;
		cursor: pointer;
		border-radius: 10px;

		img {
			display: block;
		}

		.poster-zoom-overlay {
			position: absolute;
			inset: 0;
			border-radius: 10px;
			background: rgba(0, 0, 0, 0.25);
			display: flex;
			align-items: flex-end;
			justify-content: center;
			padding-bottom: 12px;
			opacity: 0;
			pointer-events: none;
			transition:
				opacity 160ms ease,
				background 160ms ease;

			.zoom-pill {
				display: inline-flex;
				align-items: center;
				gap: 4px;
				padding: 4px 10px;
				border-radius: 6px;
				background: rgba(15, 19, 28, 0.9);
				border: 1px solid rgba(255, 255, 255, 0.18);
				color: #ffffff;
				font-size: 11px;
				font-weight: 600;
				letter-spacing: 0.03em;
				text-transform: uppercase;
				transform: translateY(4px);
				transition: transform 160ms ease;
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
