<script lang="ts">
	import { resolve } from "$app/paths";
	import Icon from "@/lib/Icon.svelte";
	import { MediaTypeE, type Media } from "@/types";
	import {
		getMediaLink,
		getPosterImage,
		getReleaseYear,
	} from "./mediaHelpers";
	import { formatPublicRating } from "@/lib/rating/helpers";

	interface Props {
		data: Media[];
		bulkMode: boolean;
		selectedIds: number[];
		onToggleSelect: (id: number) => void;
	}

	let { data, bulkMode, selectedIds, onToggleSelect }: Props = $props();
</script>

<div class="list-view-container">
	{#if data?.length > 0}
		<div class="media-table-wrap">
			<table class="media-table">
				<thead>
					<tr>
						{#if bulkMode}
							<th class="col-check"></th>
						{/if}
						<th class="col-poster">Poster</th>
						<th class="col-title">Title</th>
						<th class="col-type">Type</th>
						<th class="col-status">Status</th>
						<th class="col-rating">Rating</th>
						<th class="col-action"></th>
					</tr>
				</thead>
				<tbody>
					{#each data as w, i (`${i}-${w.type}`)}
						{#if w}
							{@const img = getPosterImage(w)}
							{@const year = getReleaseYear(w)}
							{@const link = getMediaLink(w)}
							<tr class="media-row">
								{#if bulkMode}
									<td class="col-check">
										{#if w.watched?.id}
											<input
												type="checkbox"
												checked={selectedIds.includes(w.watched.id)}
												onchange={() => onToggleSelect(w.watched!.id)}
											/>
										{/if}
									</td>
								{/if}
								<td class="col-poster">
									<a href={resolve(link)} class="row-thumb">
										{#if img}
											<img src={img} alt="" loading="lazy" />
										{:else}
											<div class="row-thumb-fallback">
												<Icon i="film" wh={18} />
											</div>
										{/if}
									</a>
								</td>
								<td class="col-title">
									<a href={resolve(link)} class="row-title">
										<strong>{w.name}</strong>
										{#if year}
											<span class="row-year">({year})</span>
										{/if}
									</a>
									{#if w.watched?.watchingSeason}
										<span class="row-progress">{w.watched.watchingSeason}</span>
									{/if}
								</td>
								<td class="col-type">
									<span class="type-pill">
										{w.type === MediaTypeE.tmdbMovie
											? "Movie"
											: w.type === MediaTypeE.tmdbShow
												? "Series"
												: "Game"}
									</span>
								</td>
								<td class="col-status">
									{#if w.watched?.status}
										<span
											class={`status-pill status-${w.watched.status.toLowerCase()}`}
										>
											{w.watched.status.toLowerCase()}
										</span>
									{:else}
										<span class="status-pill status-untracked">
											untracked
										</span>
									{/if}
								</td>
								<td class="col-rating">
									{#if w.watched?.rating}
										<span class="row-rating" title="Your Rating">
											<Icon i="star" wh={14} />
											{w.watched.rating}
										</span>
									{:else if formatPublicRating(w.rating)}
										{@const pr = formatPublicRating(w.rating)}
										<span
											class="row-public-rating"
											title={`Public Rating: ${pr} out of 10${w.ratingCount ? ` (${w.ratingCount} votes)` : ""}`}
										>
											<Icon i="star" wh={12} />
											{pr}
										</span>
									{:else}
										<span class="row-unrated">-</span>
									{/if}
								</td>
								<td class="col-action">
									<a
										href={resolve(link)}
										class="action-btn"
										title="View Details"
									>
										<Icon i="arrow" facing="right" wh={16} />
									</a>
								</td>
							</tr>
						{/if}
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>

<style lang="scss">
	.list-view-container {
		width: 100%;
		max-width: 1200px;
		margin: 0 auto;

		.media-table-wrap {
			width: 100%;
			overflow-x: auto;
			border-radius: 12px;
			border: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
			background-color: var(--card-bg, #12161f);
			box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
		}

		.media-table {
			width: 100%;
			border-collapse: collapse;
			text-align: left;
			font-size: 13.5px;

			@media screen and (max-width: 640px) {
				.col-type,
				.col-action {
					display: none;
				}

				th,
				td {
					padding: 8px 10px;
				}
			}

			th {
				padding: 12px 16px;
				font-size: 12px;
				font-weight: 600;
				color: $text-color-accent;
				text-transform: uppercase;
				letter-spacing: 0.04em;
				border-bottom: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
				background-color: var(--accent-color);
			}

			td {
				padding: 12px 16px;
				border-bottom: 1px solid var(--card-border, rgba(255, 255, 255, 0.04));
				vertical-align: middle;
			}

			tr.media-row {
				transition: background-color 120ms ease;

				&:hover {
					background-color: var(--accent-color);
				}
			}

			.col-check {
				width: 40px;
				text-align: center;
			}

			.col-poster {
				width: 50px;
			}

			.row-thumb {
				width: 40px;
				height: 60px;
				border-radius: 6px;
				overflow: hidden;
				background-color: #1a1e27;

				img {
					width: 100%;
					height: 100%;
					object-fit: cover;
				}

				.row-thumb-fallback {
					display: flex;
					align-items: center;
					justify-content: center;
					width: 100%;
					height: 100%;
					color: $placeholder-color;
				}
			}

			.col-title {
				.row-title {
					display: inline-flex;
					align-items: center;
					gap: 6px;
					color: $text-color;
					font-weight: 600;

					&:hover {
						color: $brand-color;
					}

					.row-year {
						font-weight: 400;
						color: $text-color-accent;
						font-size: 12.5px;
					}
				}

				.row-progress {
					display: block;
					font-size: 11px;
					color: #60a5fa;
					margin-top: 2px;
				}
			}

			.type-pill {
				display: inline-block;
				padding: 3px 8px;
				border-radius: 4px;
				font-size: 11px;
				font-weight: 600;
				background-color: var(--accent-color);
				color: $text-color-accent;
			}

			.status-pill {
				display: inline-block;
				padding: 3px 8px;
				border-radius: 9999px;
				font-size: 11px;
				font-weight: 600;
				text-transform: capitalize;

				&.status-watching {
					background-color: rgba(59, 130, 246, 0.15);
					color: #60a5fa;
				}
				&.status-finished {
					background-color: rgba(16, 185, 129, 0.15);
					color: #34d399;
				}
				&.status-planned {
					background-color: rgba(245, 158, 11, 0.15);
					color: #fbbf24;
				}
				&.status-hold {
					background-color: rgba(249, 115, 22, 0.15);
					color: #fb923c;
				}
				&.status-dropped {
					background-color: rgba(239, 68, 68, 0.15);
					color: #f87171;
				}
				&.status-untracked {
					background-color: var(--accent-color);
					color: $text-color-accent;
				}
			}

			.row-rating {
				display: inline-flex;
				align-items: center;
				gap: 4px;
				font-weight: 700;
				color: $rating-color;
			}

			.row-public-rating {
				display: inline-flex;
				align-items: center;
				gap: 4px;
				font-size: 12px;
				font-weight: 600;
				color: #ffd060;
				opacity: 0.85;

				:global(svg) {
					fill: #ffd060;
					color: #ffd060;
				}
			}

			.row-unrated {
				color: $placeholder-color;
			}

			.action-btn {
				display: inline-flex;
				align-items: center;
				justify-content: center;
				width: 32px;
				height: 32px;
				border-radius: 6px;
				color: $text-color-accent;
				transition: all 120ms ease;

				&:hover {
					background-color: var(--accent-color);
					color: $text-color;
				}
			}
		}
	}
</style>

