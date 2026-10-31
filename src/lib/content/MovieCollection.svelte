<script lang="ts">
	import type { MediaCollection } from "@/types";
	import { req } from "@/lib/util/api";
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import Poster from "@/lib/poster/Poster.svelte";
	import Icon from "@/lib/Icon.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import ExpandableText from "./ExpandableText.svelte";

	interface Props {
		collectionId: number;
		currentMovieId?: number;
	}

	let { collectionId, currentMovieId }: Props = $props();

	let collection: MediaCollection | undefined = $state();
	let loading = $state(false);

	$effect(() => {
		if (collectionId) {
			loading = true;
			req
				.get<MediaCollection>(`/content/collection/${collectionId}`)
				.then((resp) => {
					collection = resp;
					loading = false;
				})
				.catch((err) => {
					console.error("Failed to load collection:", err);
					loading = false;
				});
		}
	});

	const watchedCount = $derived(
		collection?.parts?.filter((p) => p.watched?.status === "FINISHED").length ||
			0,
	);
	const totalCount = $derived(collection?.parts?.length || 0);
	const pctWatched = $derived(
		totalCount > 0 ? Math.round((watchedCount / totalCount) * 100) : 0,
	);
</script>

{#if loading}
	<div class="collection-loading">
		<Spinner />
	</div>
{:else if collection && collection.parts && collection.parts.length > 0}
	<div class="movie-collection-section">
		{#if collection.backdropPath}
			<div
				class="collection-bg"
				style={`background-image: linear-gradient(180deg, rgba(13, 17, 23, 0.85) 0%, rgba(13, 17, 23, 0.98) 100%), url('https://image.tmdb.org/t/p/w1920_and_h800_multi_faces${collection.backdropPath}');`}
			></div>
		{/if}

		<div class="collection-inner">
			<div class="collection-header">
				<div class="title-row">
					<div class="title-wrap">
						<span class="series-tag">
							<Icon i="film" wh={13} />
							Movie Franchise & Collection
						</span>
						<h2 class="collection-name">{collection.name}</h2>
					</div>

					<div class="collection-meta">
						<span class="badge count-badge">
							{totalCount}
							{totalCount === 1 ? "Movie" : "Movies"}
						</span>
						{#if watchedCount > 0}
							{#if watchedCount === totalCount}
								<span class="badge completed-badge">
									<Icon i="check" wh={13} />
									Completed ({watchedCount}/{totalCount})
								</span>
							{:else}
								<span class="badge progress-badge">
									{watchedCount}/{totalCount} Watched
								</span>
							{/if}
						{/if}
						<button
							type="button"
							class="roadmap-btn"
							onclick={() => goto(resolve(`/franchise/tmdb-${collectionId}`))}
						>
							<Icon i="clock" wh={13} />
							<span>Timeline & Roadmap</span>
							<span>→</span>
						</button>
					</div>
				</div>

				{#if watchedCount > 0}
					<div class="collection-progress-bar">
						<div class="progress-fill" style={`width: ${pctWatched}%`}></div>
					</div>
				{/if}

				{#if collection.overview}
					<div class="collection-overview">
						<ExpandableText text={collection.overview} />
					</div>
				{/if}
			</div>

			<div class="collection-parts-scroll small-scrollbar">
				{#each collection.parts as part, i (part.ids.tmdb)}
					{@const isCurrent = part.ids.tmdb === currentMovieId}
					<div class="part-card-wrap" class:is-current={isCurrent}>
						{#if isCurrent}
							<div class="current-badge">
								<span>Viewing</span>
							</div>
						{/if}
						<Poster
							media={part}
							small={true}
							bind:watched={collection.parts[i].watched}
						/>
					</div>
				{/each}
			</div>
		</div>
	</div>
{/if}

<style lang="scss">
	.collection-loading {
		display: flex;
		justify-content: center;
		padding: 24px;
	}

	.movie-collection-section {
		position: relative;
		width: 100%;
		margin: 0;
		border-radius: 10px;
		overflow: hidden;
		background: rgba(15, 19, 28, 0.86);
		border: 1px solid rgba(255, 255, 255, 0.08);
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);

		.collection-bg {
			position: absolute;
			inset: 0;
			background-size: cover;
			background-position: center 25%;
			opacity: 0.35;
			pointer-events: none;
			z-index: 0;
		}

		.collection-inner {
			position: relative;
			z-index: 1;
			padding: 22px 24px;

			@media screen and (max-width: 600px) {
				padding: 16px;
			}
		}

		.collection-header {
			display: flex;
			flex-direction: column;
			gap: 10px;
			margin-bottom: 18px;

			.title-row {
				display: flex;
				align-items: flex-start;
				justify-content: space-between;
				gap: 16px;
				flex-wrap: wrap;

				.title-wrap {
					display: flex;
					flex-direction: column;
					gap: 4px;

					.series-tag {
						display: inline-flex;
						align-items: center;
						gap: 6px;
						font-size: 11px;
						font-weight: 700;
						letter-spacing: 0.06em;
						text-transform: uppercase;
						color: $brand-color;
					}

					.collection-name {
						margin: 0;
						font-size: 20px;
						font-weight: 700;
						color: #ffffff;
						line-height: 1.25;

						@media screen and (max-width: 600px) {
							font-size: 17px;
						}
					}
				}

				.collection-meta {
					display: flex;
					align-items: center;
					gap: 8px;
					flex-wrap: wrap;

					.badge {
						display: inline-flex;
						align-items: center;
						gap: 5px;
						font-size: 11.5px;
						font-weight: 600;
						padding: 4px 10px;
						border-radius: 6px;
						border: 1px solid transparent;

						&.count-badge {
							background: rgba(255, 255, 255, 0.06);
							color: rgba(255, 255, 255, 0.85);
							border-color: rgba(255, 255, 255, 0.1);
						}

						&.progress-badge {
							background: rgba(59, 130, 246, 0.12);
							color: #93c5fd;
							border-color: rgba(59, 130, 246, 0.3);
						}

						&.completed-badge {
							background: rgba(16, 185, 129, 0.12);
							color: #6ee7b7;
							border-color: rgba(16, 185, 129, 0.3);
						}
					}

					.roadmap-btn {
						display: inline-flex;
						align-items: center;
						gap: 6px;
						padding: 5px 12px;
						border-radius: 6px;
						background: rgba(245, 158, 11, 0.12);
						border: 1px solid rgba(245, 158, 11, 0.35);
						color: #fbbf24;
						font-size: 12px;
						font-weight: 600;
						cursor: pointer;
						transition:
							background 150ms ease,
							border-color 150ms ease,
							color 150ms ease;

						&:hover {
							background: $brand-color;
							border-color: $brand-color;
							color: #0b0f19;
						}
					}
				}
			}

			.collection-progress-bar {
				width: 100%;
				height: 4px;
				background: rgba(255, 255, 255, 0.08);
				border-radius: 9999px;
				overflow: hidden;

				.progress-fill {
					height: 100%;
					background: $brand-color;
					border-radius: 9999px;
					transition: width 300ms ease;
				}
			}

			.collection-overview {
				margin-top: 2px;
				font-size: 13.5px;
				line-height: 1.5;
				color: rgba(255, 255, 255, 0.68);
			}
		}

		.collection-parts-scroll {
			display: flex;
			flex-flow: row;
			gap: 14px;
			overflow-x: auto;
			padding-bottom: 6px;
			padding-top: 8px;

			.part-card-wrap {
				position: relative;
				flex-shrink: 0;
				border-radius: 10px;
				transition: border-color 160ms ease;

				&.is-current {
					border: 2px solid $brand-color;

					.current-badge {
						position: absolute;
						top: -9px;
						left: 50%;
						transform: translateX(-50%);
						background: $brand-color;
						color: #0b0f19;
						font-size: 10px;
						font-weight: 700;
						letter-spacing: 0.04em;
						text-transform: uppercase;
						padding: 2px 8px;
						border-radius: 6px;
						z-index: 10;
						white-space: nowrap;
					}
				}
			}
		}
	}
</style>
