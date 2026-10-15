<script lang="ts">
	interface Props {
		count?: number;
	}

	let { count = 12 }: Props = $props();
</script>

<div class="skeleton-grid" aria-hidden="true">
	{#each Array(count) as _, i (i)}
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
		grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
		gap: 24px 18px;
		width: 100%;
		max-width: 1440px;
		padding: 0 4px;
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
	}

	.skeleton-card {
		display: flex;
		flex-flow: column;
		width: 100%;
		max-width: 220px;
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
