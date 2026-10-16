<script lang="ts">
	import { resolve } from "$app/paths";
	import Icon from "@/lib/Icon.svelte";
	import type { Media } from "@/types";
	import { getMediaLink, getPosterImage } from "./mediaHelpers";
	import { formatPublicRating } from "@/lib/rating/helpers";

	interface Props {
		items: Media[];
	}

	let { items }: Props = $props();
</script>

<section class="continue-watching-section" aria-label="Continue watching">
	<div class="section-header">
		<div class="title-cluster">
			<span class="header-icon-wrap">
				<Icon i="play" wh={12} />
			</span>
			<h3 class="norm">Continue Watching</h3>
			<span class="count-badge">{items.length}</span>
		</div>
		<span class="section-subtitle">
			In progress across your collection
		</span>
	</div>
	<div class="continue-rail small-scrollbar">
		{#each items.slice(0, 16) as item (item.ids.tmdb || item.ids.igdb || item.name)}
			{@const img = getPosterImage(item)}
			{@const link = getMediaLink(item)}
			<a href={resolve(link)} class="continue-card">
				<div class="card-thumb">
					{#if img}
						<img src={img} alt={item.name} loading="lazy" />
					{:else}
						<div class="card-thumb-fallback">
							<Icon i="film" wh={24} />
						</div>
					{/if}

					{#if formatPublicRating(item.rating)}
						{@const pr = formatPublicRating(item.rating)}
						<div
							class="card-rating"
							title={`Rating: ${pr} out of 10${item.ratingCount ? ` (${item.ratingCount} votes)` : ""}`}
						>
							<Icon i="star" wh={10} />
							<span>{pr}</span>
						</div>
					{/if}

					<!-- Play overlay on hover -->
					<div class="play-overlay">
						<span class="play-circle">
							<Icon i="play" wh={18} />
						</span>
					</div>

					<!-- Progress bar / indicator along card thumb bottom -->
					{#if item.watched?.watchingSeason}
						<div class="thumb-progress-bar">
							<span class="progress-fill"></span>
						</div>
					{/if}
				</div>

				<div class="card-meta">
					<strong class="title" title={item.name}>{item.name}</strong>
					{#if item.watched?.watchingSeason}
						<span class="progress-pill">{item.watched.watchingSeason}</span>
					{:else}
						<span class="status-pill">Watching</span>
					{/if}
				</div>
			</a>
		{/each}
	</div>
</section>

<style lang="scss">
	.continue-watching-section {
		position: relative;
		z-index: 1;
		margin-bottom: 24px;
		padding: 16px 18px;
		background: rgba(14, 18, 26, 0.75);
		border: 1px solid var(--card-border, rgba(255, 255, 255, 0.07));
		border-radius: 14px;
		backdrop-filter: blur(12px);
		-webkit-backdrop-filter: blur(12px);

		.section-header {
			display: flex;
			align-items: center;
			justify-content: space-between;
			margin-bottom: 12px;
			gap: 8px;
			flex-wrap: wrap;

			.title-cluster {
				display: flex;
				align-items: center;
				gap: 8px;

				.header-icon-wrap {
					display: flex;
					align-items: center;
					justify-content: center;
					width: 22px;
					height: 22px;
					border-radius: 6px;
					background: rgba(234, 179, 8, 0.15);
					color: #eab308;
				}

				h3 {
					font-size: 15px;
					font-weight: 700;
					color: $text-color;
					letter-spacing: 0.01em;
				}

				.count-badge {
					font-size: 11px;
					font-weight: 600;
					color: $text-color-accent;
					background: rgba(255, 255, 255, 0.06);
					padding: 2px 7px;
					border-radius: 9999px;
					border: 1px solid rgba(255, 255, 255, 0.08);
				}
			}

			.section-subtitle {
				font-size: 11.5px;
				color: $text-color-accent;
				font-weight: 500;
			}
		}

		.continue-rail {
			display: flex;
			gap: 12px;
			overflow-x: auto;
			padding-bottom: 4px;
			scroll-behavior: smooth;
			-webkit-overflow-scrolling: touch;
		}

		.continue-card {
			display: flex;
			flex-flow: column;
			width: 135px;
			min-width: 135px;
			border-radius: 10px;
			overflow: hidden;
			background-color: rgba(20, 25, 36, 0.85);
			border: 1px solid var(--card-border, rgba(255, 255, 255, 0.07));
			box-shadow: 0 4px 12px rgba(0, 0, 0, 0.25);
			text-decoration: none;
			transition:
				transform 160ms cubic-bezier(0.16, 1, 0.3, 1),
				box-shadow 160ms ease,
				border-color 160ms ease;

			&:hover {
				transform: translateY(-3px);
				box-shadow: 0 8px 24px rgba(0, 0, 0, 0.45);
				border-color: rgba(255, 255, 255, 0.15);

				.play-overlay {
					opacity: 1;
				}

				.play-circle {
					transform: scale(1.08);
				}
			}

			.card-thumb {
				position: relative;
				width: 100%;
				aspect-ratio: 2/3;
				background-color: #12151e;
				overflow: hidden;

				img {
					width: 100%;
					height: 100%;
					object-fit: cover;
					display: block;
				}

				.card-thumb-fallback {
					display: flex;
					align-items: center;
					justify-content: center;
					width: 100%;
					height: 100%;
					color: $placeholder-color;
				}

				.card-rating {
					position: absolute;
					top: 6px;
					right: 6px;
					display: inline-flex;
					align-items: center;
					gap: 3px;
					padding: 2px 5px;
					border-radius: 5px;
					background: rgba(12, 16, 23, 0.88);
					border: 1px solid rgba(234, 179, 8, 0.35);
					color: #eab308;
					font-size: 10.5px;
					font-weight: 700;
					line-height: 1;
					box-shadow: 0 2px 6px rgba(0, 0, 0, 0.4);
					backdrop-filter: blur(4px);
					-webkit-backdrop-filter: blur(4px);
					z-index: 2;
					pointer-events: none;

					:global(svg) {
						fill: #eab308;
						color: #eab308;
						flex-shrink: 0;
					}
				}

				.play-overlay {
					position: absolute;
					inset: 0;
					display: flex;
					align-items: center;
					justify-content: center;
					background-color: rgba(0, 0, 0, 0.4);
					backdrop-filter: blur(2px);
					opacity: 0;
					transition: opacity 150ms ease;

					.play-circle {
						display: flex;
						align-items: center;
						justify-content: center;
						width: 36px;
						height: 36px;
						border-radius: 50%;
						background-color: rgba(234, 179, 8, 0.9);
						color: #0f1218;
						box-shadow: 0 2px 10px rgba(0, 0, 0, 0.5);
						transition: transform 150ms cubic-bezier(0.16, 1, 0.3, 1);
					}
				}

				.thumb-progress-bar {
					position: absolute;
					bottom: 0;
					left: 0;
					right: 0;
					height: 3px;
					background: rgba(0, 0, 0, 0.5);
					z-index: 3;

					.progress-fill {
						display: block;
						height: 100%;
						width: 60%;
						background: #eab308;
						box-shadow: 0 0 6px rgba(234, 179, 8, 0.6);
					}
				}
			}

			.card-meta {
				display: flex;
				flex-flow: column;
				gap: 4px;
				padding: 8px 10px 10px 10px;

				.title {
					font-size: 12.5px;
					font-weight: 600;
					color: $text-color;
					white-space: nowrap;
					overflow: hidden;
					text-overflow: ellipsis;
				}

				.progress-pill {
					font-size: 11px;
					font-weight: 600;
					color: #eab308;
					background-color: rgba(234, 179, 8, 0.12);
					border: 1px solid rgba(234, 179, 8, 0.2);
					padding: 1px 6px;
					border-radius: 4px;
					width: max-content;
				}

				.status-pill {
					font-size: 10px;
					font-weight: 600;
					text-transform: uppercase;
					padding: 1px 6px;
					border-radius: 4px;
					width: max-content;
					background-color: rgba(234, 179, 8, 0.12);
					color: #eab308;
				}
			}
		}
	}
</style>
