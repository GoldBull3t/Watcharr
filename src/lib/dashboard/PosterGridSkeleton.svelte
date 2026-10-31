<script lang="ts">
	export type GridDensity = "compact" | "comfortable" | "editorial";

	interface Props {
		count?: number;
		density?: GridDensity;
	}

	let { count = 12, density = "comfortable" }: Props = $props();
</script>

<div class={`skeleton-grid density-${density}`} aria-hidden="true">
	{#each Array.from({ length: count }, (__, idx) => idx) as i (i)}
		<div class="skeleton-card">
			<div class="skeleton-poster shimmer"></div>
			<div class="skeleton-info">
				<div class="skeleton-line title shimmer"></div>
				<div class="skeleton-line sub shimmer"></div>
			</div>
		</div>
	{/each}
</div>

<style lang="scss">
	.skeleton-grid {
		display: grid;
		width: 100%;
		max-width: 1440px;
		padding: 0 4px;
		transition: gap 200ms ease;

		&.density-compact {
			grid-template-columns: repeat(auto-fill, minmax(118px, 1fr));
			gap: 14px 10px;
			margin: 14px 0 30px 0;

			@media screen and (max-width: 640px) {
				grid-template-columns: repeat(auto-fill, minmax(96px, 1fr));
				gap: 10px 8px;
				padding: 0;
			}

			.skeleton-card {
				max-width: 155px;
			}
		}

		&.density-comfortable,
		&:not(.density-compact):not(.density-editorial) {
			grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
			gap: 24px 18px;
			margin: 20px 0 36px 0;

			@media screen and (max-width: 640px) {
				grid-template-columns: repeat(auto-fill, minmax(135px, 1fr));
				gap: 16px 12px;
				padding: 0;
			}

			@media screen and (max-width: 340px) {
				grid-template-columns: repeat(2, 1fr);
				gap: 12px 10px;
				padding: 0;
			}

			.skeleton-card {
				max-width: 220px;
			}
		}

		&.density-editorial {
			grid-template-columns: repeat(auto-fill, minmax(230px, 1fr));
			gap: 32px 24px;
			margin: 24px 0 44px 0;

			@media screen and (max-width: 768px) {
				grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
				gap: 22px 16px;
				padding: 0;
			}

			@media screen and (max-width: 440px) {
				grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
				gap: 16px 12px;
				padding: 0;
			}

			.skeleton-card {
				max-width: 320px;
			}
		}
	}

	.skeleton-card {
		display: flex;
		flex-flow: column;
		width: 100%;
		margin: 0 auto;
		gap: 8px;
	}

	.skeleton-poster {
		width: 100%;
		aspect-ratio: 170 / 256;
		border-radius: 10px;
		background-color: #141824;
		border: 1px solid rgba(255, 255, 255, 0.07);
		box-shadow: 0 4px 14px rgba(0, 0, 0, 0.25);
	}

	.skeleton-info {
		display: flex;
		flex-flow: column;
		gap: 6px;
		padding: 2px 0;

		.skeleton-line {
			border-radius: 4px;
			background-color: #141824;

			&.title {
				width: 75%;
				height: 13px;
			}

			&.sub {
				width: 45%;
				height: 10px;
			}
		}
	}

	@keyframes shimmer {
		0% {
			background-position: -200% 0;
		}
		100% {
			background-position: 200% 0;
		}
	}

	.shimmer {
		background: linear-gradient(
			90deg,
			rgba(20, 24, 36, 0.9) 0%,
			rgba(36, 44, 64, 0.95) 50%,
			rgba(20, 24, 36, 0.9) 100%
		);
		background-size: 200% 100%;
		animation: shimmer 1.8s infinite ease-in-out;
	}
</style>
