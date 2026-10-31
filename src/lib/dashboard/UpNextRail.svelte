<script lang="ts">
	import { resolve } from "$app/paths";
	import Icon from "@/lib/Icon.svelte";
	import SpinnerTiny from "@/lib/SpinnerTiny.svelte";
	import tooltip from "@/lib/actions/tooltip";
	import { req } from "@/lib/util/api";
	import { notify } from "@/lib/util/notify";
	import type { UpNextItem } from "@/types";

	interface Props {
		items: UpNextItem[];
		onEpisodeWatched?: () => void;
	}

	let { items = $bindable([]), onEpisodeWatched }: Props = $props();

	let updatingWatchedId: number | null = $state(null);

	async function markEpisodeWatched(item: UpNextItem, e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();

		updatingWatchedId = item.watchedId;
		const label = `S${String(item.seasonNumber).padStart(2, "0")}E${String(item.episodeNumber).padStart(2, "0")}`;
		const nid = notify({
			type: "loading",
			text: `A marcar ${item.showTitle} ${label} como visto...`,
		});

		try {
			await req.post("/watched/episode", {
				watchedId: item.watchedId,
				seasonNumber: item.seasonNumber,
				episodeNumber: item.episodeNumber,
				status: "FINISHED",
			});

			notify({
				id: nid,
				type: "success",
				text: `${item.showTitle} ${label} marcado como visto!`,
			});

			// Optimistic local update: advance episode or remove card if season finished
			if (item.episodesRemaining && item.episodesRemaining > 1) {
				item.episodeNumber += 1;
				item.episodesRemaining -= 1;
				item.episodeName = `Episódio ${item.episodeNumber}`;
			} else {
				items = items.filter((it) => it.watchedId !== item.watchedId);
			}

			if (onEpisodeWatched) {
				onEpisodeWatched();
			}
		} catch (err: any) {
			console.error("Failed to mark episode as watched:", err);
			notify({
				id: nid,
				type: "error",
				text: err?.message || "Falha ao marcar episódio como visto.",
			});
		} finally {
			updatingWatchedId = null;
		}
	}

	function getBackdropUrl(item: UpNextItem): string | null {
		if (item.episodeStill) {
			return `https://image.tmdb.org/t/p/w780${item.episodeStill}`;
		}
		if (item.showBackdrop) {
			return `https://image.tmdb.org/t/p/w780${item.showBackdrop}`;
		}
		if (item.showPoster) {
			return `https://image.tmdb.org/t/p/w500${item.showPoster}`;
		}
		return null;
	}

	function formatAirDate(d?: string) {
		if (!d) return "";
		const date = new Date(d);
		if (isNaN(date.getTime())) return "";
		return date.toLocaleDateString(undefined, {
			day: "numeric",
			month: "short",
		});
	}
</script>

{#if items && items.length > 0}
	<section class="up-next-section" aria-label="A Seguir (Up Next)">
		<div class="section-header">
			<div class="title-cluster">
				<span class="header-icon-wrap">
					<Icon i="play" wh={14} />
				</span>
				<h3 class="norm">A Seguir (Up Next)</h3>
				<span class="count-badge">{items.length}</span>
			</div>
			<span class="section-subtitle">
				Próximos episódios prontos a ver com 1 clique
			</span>
		</div>

		<div class="up-next-rail small-scrollbar">
			{#each items as item (item.watchedId)}
				{@const bgUrl = getBackdropUrl(item)}
				{@const isUpdating = updatingWatchedId === item.watchedId}

				<a href={resolve(`/tv/${item.showTmdbId}`)} class="up-next-card">
					<div class="card-media">
						{#if bgUrl}
							<img
								src={bgUrl}
								alt={item.episodeName || item.showTitle}
								loading="lazy"
							/>
						{:else}
							<div class="media-fallback">
								<Icon i="tv" wh={32} />
							</div>
						{/if}

						<div class="gradient-overlay"></div>

						<!-- Top Badges -->
						<div class="top-row">
							<span class="ep-pill">
								S{item.seasonNumber} • E{item.episodeNumber}
							</span>

							{#if item.airDate && !item.isAired}
								<span class="status-pill upcoming">
									Estreia {formatAirDate(item.airDate)}
								</span>
							{:else if item.episodesRemaining > 1}
								<span class="status-pill remaining">
									Restam {item.episodesRemaining} eps
								</span>
							{/if}
						</div>

						<!-- 1-Click Fast Mark As Watched Checkmark -->
						<button
							class="quick-watch-btn"
							class:active={isUpdating}
							disabled={isUpdating}
							onclick={(e) => markEpisodeWatched(item, e)}
							use:tooltip={{
								text: `Marcar S${item.seasonNumber}E${item.episodeNumber} como visto`,
								pos: "top",
							}}
						>
							{#if isUpdating}
								<SpinnerTiny />
							{:else}
								<Icon i="check" wh={16} />
							{/if}
						</button>

						<!-- Bottom Meta Details -->
						<div class="card-bottom">
							<span class="show-title" title={item.showTitle}
								>{item.showTitle}</span
							>
							<strong
								class="ep-title"
								title={item.episodeName || `Episódio ${item.episodeNumber}`}
							>
								{item.episodeName || `Episódio ${item.episodeNumber}`}
							</strong>
						</div>
					</div>
				</a>
			{/each}
		</div>
	</section>
{/if}

<style lang="scss">
	.up-next-section {
		position: relative;
		z-index: 1;
		margin-bottom: 24px;
		padding: 18px 20px;
		background: rgba(14, 18, 26, 0.78);
		border: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
		border-radius: 16px;
		backdrop-filter: blur(14px);
		-webkit-backdrop-filter: blur(14px);

		.section-header {
			display: flex;
			align-items: center;
			justify-content: space-between;
			margin-bottom: 14px;
			gap: 8px;
			flex-wrap: wrap;

			.title-cluster {
				display: flex;
				align-items: center;
				gap: 10px;

				.header-icon-wrap {
					display: flex;
					align-items: center;
					justify-content: center;
					width: 26px;
					height: 26px;
					border-radius: 8px;
					background: rgba($accent-color, 0.2);
					color: $accent-color;
				}

				h3 {
					font-size: 17px;
					font-weight: 700;
					color: #fff;
					margin: 0;
					letter-spacing: -0.01em;
				}

				.count-badge {
					display: inline-flex;
					align-items: center;
					justify-content: center;
					min-width: 20px;
					height: 20px;
					padding: 0 6px;
					border-radius: 10px;
					background: rgba(255, 255, 255, 0.08);
					border: 1px solid rgba(255, 255, 255, 0.12);
					color: rgba(255, 255, 255, 0.85);
					font-size: 11px;
					font-weight: 700;
				}
			}

			.section-subtitle {
				font-size: 12px;
				color: rgba(255, 255, 255, 0.5);
			}
		}

		.up-next-rail {
			display: flex;
			gap: 14px;
			overflow-x: auto;
			padding-bottom: 8px;
			scroll-behavior: smooth;
			scrollbar-width: thin;
		}

		.up-next-card {
			display: block;
			flex: 0 0 290px;
			text-decoration: none;
			color: inherit;
			border-radius: 12px;
			overflow: hidden;
			background: rgba(255, 255, 255, 0.03);
			border: 1px solid rgba(255, 255, 255, 0.08);
			transition: all 200ms cubic-bezier(0.16, 1, 0.3, 1);

			&:hover {
				transform: translateY(-3px);
				border-color: rgba($accent-color, 0.5);
				box-shadow: 0 10px 24px rgba(0, 0, 0, 0.4);

				.card-media img {
					transform: scale(1.04);
				}

				.quick-watch-btn {
					opacity: 1;
					transform: scale(1);
				}
			}

			.card-media {
				position: relative;
				width: 100%;
				height: 164px;
				overflow: hidden;
				display: flex;
				flex-direction: column;
				justify-content: space-between;
				padding: 12px;

				img {
					position: absolute;
					top: 0;
					left: 0;
					width: 100%;
					height: 100%;
					object-fit: cover;
					transition: transform 300ms ease;
					z-index: 0;
				}

				.media-fallback {
					position: absolute;
					top: 0;
					left: 0;
					width: 100%;
					height: 100%;
					background: #181d28;
					display: flex;
					align-items: center;
					justify-content: center;
					color: rgba(255, 255, 255, 0.2);
					z-index: 0;
				}

				.gradient-overlay {
					position: absolute;
					top: 0;
					left: 0;
					width: 100%;
					height: 100%;
					background: linear-gradient(
						180deg,
						rgba(10, 14, 22, 0.65) 0%,
						rgba(10, 14, 22, 0.1) 40%,
						rgba(10, 14, 22, 0.95) 100%
					);
					z-index: 1;
				}

				.top-row {
					position: relative;
					z-index: 2;
					display: flex;
					align-items: center;
					gap: 6px;
					max-width: calc(100% - 40px);
					flex-wrap: wrap;

					.ep-pill {
						background: rgba(0, 0, 0, 0.65);
						backdrop-filter: blur(8px);
						border: 1px solid rgba(255, 255, 255, 0.15);
						color: #fff;
						font-size: 11px;
						font-weight: 700;
						padding: 3px 8px;
						border-radius: 6px;
						letter-spacing: 0.02em;
					}

					.status-pill {
						font-size: 10px;
						font-weight: 600;
						padding: 3px 7px;
						border-radius: 6px;
						backdrop-filter: blur(8px);

						&.remaining {
							background: rgba(255, 255, 255, 0.12);
							color: rgba(255, 255, 255, 0.85);
						}

						&.upcoming {
							background: rgba(255, 179, 0, 0.25);
							color: #ffca28;
							border: 1px solid rgba(255, 179, 0, 0.35);
						}
					}
				}

				.quick-watch-btn {
					position: absolute;
					top: 10px;
					right: 10px;
					z-index: 3;
					display: flex;
					align-items: center;
					justify-content: center;
					width: 32px;
					height: 32px;
					border-radius: 50%;
					background: rgba(0, 0, 0, 0.65);
					backdrop-filter: blur(8px);
					border: 1px solid rgba(255, 255, 255, 0.2);
					color: #fff;
					cursor: pointer;
					transition: all 180ms ease;
					opacity: 0.85;

					&:hover:not(:disabled) {
						background: #4caf50;
						border-color: #81c784;
						transform: scale(1.12);
						color: #fff;
					}

					&:disabled {
						opacity: 0.6;
						cursor: not-allowed;
					}
				}

				.card-bottom {
					position: relative;
					z-index: 2;
					display: flex;
					flex-direction: column;
					gap: 2px;
					overflow: hidden;

					.show-title {
						font-size: 11px;
						font-weight: 600;
						color: rgba(255, 255, 255, 0.7);
						text-transform: uppercase;
						letter-spacing: 0.04em;
						white-space: nowrap;
						overflow: hidden;
						text-overflow: ellipsis;
					}

					.ep-title {
						font-size: 13.5px;
						font-weight: 700;
						color: #fff;
						white-space: nowrap;
						overflow: hidden;
						text-overflow: ellipsis;
						line-height: 1.3;
					}
				}
			}
		}
	}
</style>
