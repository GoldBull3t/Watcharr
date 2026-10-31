<script lang="ts">
	interface Props {
		type?: "movie" | "tv" | "game" | "person";
	}

	let { type = "movie" }: Props = $props();
</script>

<div class="media-detail-skeleton {type}" aria-hidden="true" role="status">
	<div class="skeleton-backdrop shimmer"></div>
	<div class="skeleton-container">
		<div class="skeleton-layout">
			<div class="skeleton-poster shimmer"></div>
			<div class="skeleton-details">
				<div class="skeleton-title shimmer"></div>
				<div class="skeleton-badges">
					<div class="skeleton-badge shimmer"></div>
					<div class="skeleton-badge shimmer"></div>
					<div class="skeleton-badge shimmer"></div>
				</div>
				<div class="skeleton-actions">
					<div class="skeleton-btn shimmer"></div>
					<div class="skeleton-btn secondary shimmer"></div>
				</div>
				<div class="skeleton-overview">
					<div class="skeleton-line shimmer"></div>
					<div class="skeleton-line shimmer"></div>
					<div class="skeleton-line short shimmer"></div>
				</div>
			</div>
			{#if type !== "person"}
				<div class="skeleton-sidebar">
					<div class="skeleton-line short shimmer"></div>
					<div class="skeleton-stars shimmer"></div>
					<div class="skeleton-status-row shimmer"></div>
					<div class="skeleton-thoughts shimmer"></div>
				</div>
			{/if}
		</div>
	</div>
</div>

<style lang="scss">
	.media-detail-skeleton {
		width: 100%;
		position: relative;
		overflow: hidden;
		min-height: 440px;
	}

	.skeleton-backdrop {
		width: 100%;
		height: 380px;
		position: absolute;
		top: 0;
		left: 0;
		opacity: 0.35;
		mask-image: linear-gradient(to bottom, black 50%, transparent 100%);
		-webkit-mask-image: linear-gradient(to bottom, black 50%, transparent 100%);
	}

	.skeleton-container {
		width: 100%;
		max-width: 1360px;
		margin: 0 auto;
		padding: 36px 32px 24px;
		position: relative;
		z-index: 1;

		@media screen and (max-width: 1120px) {
			padding: 28px 24px 20px;
		}

		@media screen and (max-width: 680px) {
			padding: 20px 16px;
		}
	}

	.skeleton-layout {
		display: grid;
		grid-template-columns: 240px minmax(0, 1fr) 340px;
		gap: 24px;
		align-items: start;

		.person & {
			grid-template-columns: 240px minmax(0, 1fr);
		}

		@media screen and (max-width: 1120px) {
			grid-template-columns: 200px minmax(0, 1fr);
			gap: 20px;

			.skeleton-sidebar {
				grid-column: 1 / -1;
			}
		}

		@media screen and (max-width: 680px) {
			grid-template-columns: minmax(0, 1fr);
			justify-items: center;
			gap: 16px;
		}
	}

	.skeleton-poster {
		width: 240px;
		aspect-ratio: 2 / 3;
		border-radius: 10px;
		flex-shrink: 0;
		border: 1px solid rgba(255, 255, 255, 0.08);

		@media screen and (max-width: 1120px) {
			width: 200px;
		}
	}

	.skeleton-details,
	.skeleton-sidebar {
		display: flex;
		flex-direction: column;
		gap: 16px;
		width: 100%;
		background: rgba(15, 19, 28, 0.86);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 10px;
		padding: 24px;
	}

	.skeleton-stars {
		width: 100%;
		height: 38px;
		border-radius: 8px;
	}

	.skeleton-status-row {
		width: 100%;
		height: 42px;
		border-radius: 8px;
	}

	.skeleton-thoughts {
		width: 100%;
		height: 68px;
		border-radius: 8px;
	}

	.skeleton-title {
		height: 36px;
		width: 60%;
		border-radius: 8px;

		@media screen and (max-width: 768px) {
			width: 80%;
		}
	}

	.skeleton-badges {
		display: flex;
		gap: 10px;

		.skeleton-badge {
			height: 22px;
			width: 65px;
			border-radius: 9999px;

			&:nth-child(2) {
				width: 85px;
			}
			&:nth-child(3) {
				width: 50px;
			}
		}
	}

	.skeleton-actions {
		display: flex;
		gap: 12px;
		margin-top: 8px;

		.skeleton-btn {
			height: 38px;
			width: 130px;
			border-radius: 8px;

			&.secondary {
				width: 90px;
			}
		}
	}

	.skeleton-overview {
		display: flex;
		flex-direction: column;
		gap: 10px;
		margin-top: 14px;
		width: 100%;
		max-width: 620px;

		.skeleton-line {
			height: 14px;
			width: 100%;
			border-radius: 4px;

			&.short {
				width: 65%;
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
			rgba(20, 24, 36, 0.75) 0%,
			rgba(38, 46, 68, 0.9) 50%,
			rgba(20, 24, 36, 0.75) 100%
		);
		background-size: 200% 100%;
		animation: shimmer 1.8s infinite ease-in-out;
	}
</style>
