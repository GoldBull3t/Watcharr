<script lang="ts">
	import type { MediaCollection } from "@/types";
	import { req } from "@/lib/util/api";
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
	let error: unknown | undefined = $state();

	$effect(() => {
		if (collectionId) {
			loading = true;
			error = undefined;
			req
				.get<MediaCollection>(`/content/collection/${collectionId}`)
				.then((resp) => {
					collection = resp;
					loading = false;
				})
				.catch((err) => {
					console.error("Failed to load collection:", err);
					error = err;
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
							<Icon i="reel" wh={14} />
							Movie Franchise
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
		margin: 28px 0;
		border-radius: 18px;
		overflow: hidden;
		background: rgba(22, 27, 38, 0.7);
		border: 1px solid rgba(255, 255, 255, 0.1);
		box-shadow: 0 12px 32px -4px rgba(0, 0, 0, 0.5);
		backdrop-filter: blur(16px);
		-webkit-backdrop-filter: blur(16px);

		.collection-bg {
			position: absolute;
			inset: 0;
			background-size: cover;
			background-position: center 25%;
			opacity: 0.45;
			pointer-events: none;
			z-index: 0;
		}

		.collection-inner {
			position: relative;
			z-index: 1;
			padding: 24px 28px;

			@media screen and (max-width: 600px) {
				padding: 16px;
			}
		}

		.collection-header {
			display: flex;
			flex-direction: column;
			gap: 12px;
			margin-bottom: 20px;

			.title-row {
				display: flex;
				align-items: flex-start;
				justify-content: space-between;
				gap: 16px;
				flex-wrap: wrap;

				.title-wrap {
					display: flex;
					flex-direction: column;
					gap: 6px;

					.series-tag {
						display: inline-flex;
						align-items: center;
						gap: 6px;
						font-size: 11.5px;
						font-weight: 700;
						letter-spacing: 0.08em;
						text-transform: uppercase;
						color: $brand-color;
					}

					.collection-name {
						margin: 0;
						font-size: 24px;
						font-weight: 800;
						color: #ffffff;
						line-height: 1.25;

						@media screen and (max-width: 600px) {
							font-size: 19px;
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
						font-size: 12px;
						font-weight: 600;
						padding: 4px 10px;
						border-radius: 9999px;
						border: 1px solid transparent;

						&.count-badge {
							background: rgba(255, 255, 255, 0.08);
							color: rgba(255, 255, 255, 0.85);
							border-color: rgba(255, 255, 255, 0.12);
						}

						&.progress-badge {
							background: rgba(59, 130, 246, 0.15);
							color: #93c5fd;
							border-color: rgba(59, 130, 246, 0.35);
						}

						&.completed-badge {
							background: rgba(16, 185, 129, 0.15);
							color: #6ee7b7;
							border-color: rgba(16, 185, 129, 0.35);
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
				margin-top: 4px;
				font-size: 14px;
				line-height: 1.5;
				color: rgba(255, 255, 255, 0.7);
			}
		}

		.collection-parts-scroll {
			display: flex;
			flex-flow: row;
			gap: 16px;
			overflow-x: auto;
			padding-bottom: 8px;
			padding-top: 6px;

			.part-card-wrap {
				position: relative;
				flex-shrink: 0;
				border-radius: 12px;
				transition:
					transform 180ms ease,
					box-shadow 180ms ease;

				&.is-current {
					border: 2px solid $brand-color;
					box-shadow: 0 0 16px rgba(99, 102, 241, 0.4);
					transform: scale(1.02);

					.current-badge {
						position: absolute;
						top: -10px;
						left: 50%;
						transform: translateX(-50%);
						background: $brand-color;
						color: white;
						font-size: 10.5px;
						font-weight: 700;
						letter-spacing: 0.04em;
						text-transform: uppercase;
						padding: 2px 8px;
						border-radius: 9999px;
						box-shadow: 0 2px 8px rgba(0, 0, 0, 0.5);
						z-index: 10;
						white-space: nowrap;
					}
				}
			}
		}
	}
</style>
