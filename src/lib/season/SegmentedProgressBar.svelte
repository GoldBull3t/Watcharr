<script lang="ts">
	import type {
		TMDBSeasonDetailsEpisode,
		Watched,
		WatchedStatus,
	} from "@/types";
	import Icon from "@/lib/Icon.svelte";

	interface Props {
		seasonNumber: number;
		episodes: TMDBSeasonDetailsEpisode[];
		watchedItem: Watched | undefined;
	}

	let { seasonNumber, episodes, watchedItem }: Props = $props();

	let watchedMap = $derived.by(() => {
		const map = new Map<number, WatchedStatus>();
		if (!watchedItem?.watchedEpisodes) return map;
		for (const we of watchedItem.watchedEpisodes) {
			if (we.seasonNumber === seasonNumber) {
				map.set(we.episodeNumber, we.status);
			}
		}
		return map;
	});

	let totalEpisodes = $derived(episodes?.length || 0);

	let watchedCount = $derived.by(() => {
		if (!episodes || episodes.length === 0) return 0;
		let count = 0;
		for (const ep of episodes) {
			if (watchedMap.get(ep.episode_number) === "FINISHED") {
				count++;
			}
		}
		return count;
	});

	let percentage = $derived(
		totalEpisodes > 0
			? Math.min(100, Math.round((watchedCount / totalEpisodes) * 100))
			: 0,
	);

	let nextUpEpisode = $derived.by(() => {
		if (!episodes || episodes.length === 0) return undefined;
		return episodes.find(
			(ep) => watchedMap.get(ep.episode_number) !== "FINISHED",
		);
	});

	let activeHoveredEp = $state<TMDBSeasonDetailsEpisode | null>(null);

	function scrollToEpisode(epNum: number) {
		if (typeof document === "undefined") return;
		const el = document.getElementById(`ep-${seasonNumber}-${epNum}`);
		if (el) {
			el.scrollIntoView({ behavior: "smooth", block: "center" });
			el.classList.add("ep-highlight-pulse");
			setTimeout(() => {
				el.classList.remove("ep-highlight-pulse");
			}, 1400);
		}
	}
</script>

<div
	class="segmented-timeline-dock"
	aria-label="Season episode progress timeline"
>
	<div class="timeline-header">
		<div class="progress-meta">
			<span class="meta-label">Season Progress</span>
			<span class="meta-fraction">
				<strong>{watchedCount}</strong> / {totalEpisodes} episodes
				<span class="meta-percent">({percentage}%)</span>
			</span>
		</div>

		<div class="timeline-actions">
			{#if percentage === 100}
				<div class="completed-pill">
					<Icon i="check" wh={12} />
					<span>Season Completed</span>
				</div>
			{:else if nextUpEpisode}
				<button
					type="button"
					class="next-up-btn"
					onclick={() => scrollToEpisode(nextUpEpisode.episode_number)}
					title={`Jump to Episode ${nextUpEpisode.episode_number}: ${nextUpEpisode.name}`}
				>
					<span class="next-label">Next Up:</span>
					<span class="next-ep">E{nextUpEpisode.episode_number}</span>
					<Icon i="arrow" facing="right" wh={11} />
				</button>
			{/if}
		</div>
	</div>

	<!-- Segmented Track -->
	<div class="segments-track" role="list" aria-label="Episodes track">
		{#each episodes as ep (ep.id || ep.episode_number)}
			{@const status = watchedMap.get(ep.episode_number)}
			{@const isFinished = status === "FINISHED"}
			{@const isNext = nextUpEpisode?.episode_number === ep.episode_number}
			{@const isHovered = activeHoveredEp?.episode_number === ep.episode_number}
			<button
				type="button"
				class="segment-pill"
				class:finished={isFinished}
				class:next-target={isNext}
				class:hovered={isHovered}
				onclick={() => scrollToEpisode(ep.episode_number)}
				onmouseenter={() => (activeHoveredEp = ep)}
				onmouseleave={() => {
					if (activeHoveredEp?.episode_number === ep.episode_number) {
						activeHoveredEp = null;
					}
				}}
				title={`E${ep.episode_number}: ${ep.name}${isFinished ? " (Watched)" : ""}`}
				aria-label={`Episode ${ep.episode_number}: ${ep.name}, ${isFinished ? "Watched" : "Unwatched"}`}
			>
				<span class="visually-hidden">E{ep.episode_number}</span>
			</button>
		{/each}
	</div>

	<!-- Active Hover Popover Tooltip -->
	{#if activeHoveredEp}
		{@const hStatus = watchedMap.get(activeHoveredEp.episode_number)}
		<div class="segment-tooltip-card">
			<div class="tip-badge">E{activeHoveredEp.episode_number}</div>
			<div class="tip-details">
				<span class="tip-title">{activeHoveredEp.name}</span>
				<div class="tip-sub">
					{#if activeHoveredEp.runtime}
						<span>{activeHoveredEp.runtime}m</span>
						<span class="dot">•</span>
					{/if}
					{#if hStatus === "FINISHED"}
						<span class="status-tag watched">Watched ✓</span>
					{:else}
						<span class="status-tag unwatched">Not watched</span>
					{/if}
				</div>
			</div>
			<span class="tip-action-hint">Click to jump</span>
		</div>
	{/if}
</div>

<style lang="scss">
	.segmented-timeline-dock {
		display: flex;
		flex-flow: column;
		gap: 12px;
		margin-bottom: 24px;
		padding: 14px 18px;
		border-radius: 14px;
		background: rgba(14, 18, 26, 0.82);
		border: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
		box-shadow: 0 8px 24px -4px rgba(0, 0, 0, 0.45);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		position: relative;
	}

	.timeline-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		flex-wrap: wrap;

		.progress-meta {
			display: flex;
			align-items: baseline;
			gap: 8px;

			.meta-label {
				font-size: 13px;
				font-weight: 700;
				color: #ffffff;
				letter-spacing: 0.01em;
			}

			.meta-fraction {
				font-size: 12.5px;
				color: $text-color-accent;

				strong {
					color: #fbbf24;
					font-weight: 700;
				}

				.meta-percent {
					color: $placeholder-color;
					font-size: 11.5px;
					margin-left: 2px;
				}
			}
		}

		.timeline-actions {
			display: flex;
			align-items: center;
			gap: 8px;

			.completed-pill {
				display: inline-flex;
				align-items: center;
				gap: 5px;
				padding: 3px 10px;
				border-radius: 9999px;
				background: rgba(16, 185, 129, 0.16);
				border: 1px solid rgba(16, 185, 129, 0.35);
				color: #34d399;
				font-size: 11px;
				font-weight: 600;
				box-shadow: 0 0 10px rgba(16, 185, 129, 0.2);
			}

			.next-up-btn {
				display: inline-flex;
				align-items: center;
				gap: 5px;
				padding: 3px 10px;
				border-radius: 9999px;
				background: rgba(245, 158, 11, 0.14);
				border: 1px solid rgba(245, 158, 11, 0.32);
				color: #fbbf24;
				font-size: 11px;
				font-weight: 600;
				cursor: pointer;
				transition: all 140ms ease;

				&:hover {
					background: rgba(245, 158, 11, 0.24);
					border-color: rgba(245, 158, 11, 0.55);
					transform: translateX(1px);
				}

				.next-label {
					color: rgba(255, 255, 255, 0.7);
					font-weight: 500;
				}

				.next-ep {
					color: #ffffff;
					font-weight: 700;
				}
			}
		}
	}

	.segments-track {
		display: flex;
		align-items: center;
		gap: 4px;
		width: 100%;
		height: 12px;
		padding: 2px;
		border-radius: 8px;
		background: rgba(0, 0, 0, 0.4);
		border: 1px solid rgba(255, 255, 255, 0.05);

		.segment-pill {
			flex: 1 1 0;
			min-width: 4px;
			height: 100%;
			border: none;
			padding: 0;
			border-radius: 4px;
			background: rgba(255, 255, 255, 0.08);
			cursor: pointer;
			position: relative;
			transition: all 140ms ease;

			&:hover {
				transform: scaleY(1.35);
				z-index: 10;
			}

			&.finished {
				background: linear-gradient(135deg, #f59e0b 0%, #fbbf24 100%);
				box-shadow: 0 0 8px rgba(245, 158, 11, 0.35);

				&:hover {
					box-shadow: 0 0 12px rgba(245, 158, 11, 0.6);
				}
			}

			&.next-target:not(.finished) {
				background: rgba(245, 158, 11, 0.28);
				border: 1px dashed rgba(245, 158, 11, 0.75);
				animation: pulseNext 2s infinite ease-in-out;
			}
		}
	}

	@keyframes pulseNext {
		0%,
		100% {
			opacity: 0.6;
		}
		50% {
			opacity: 1;
		}
	}

	.segment-tooltip-card {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 6px 12px;
		border-radius: 8px;
		background: rgba(20, 25, 36, 0.95);
		border: 1px solid rgba(255, 255, 255, 0.12);
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.5);
		font-size: 11.5px;
		animation: fadeIn 120ms ease-out;

		.tip-badge {
			padding: 2px 6px;
			border-radius: 5px;
			background: rgba(245, 158, 11, 0.2);
			border: 1px solid rgba(245, 158, 11, 0.4);
			color: #fbbf24;
			font-weight: 700;
			font-size: 10.5px;
		}

		.tip-details {
			display: flex;
			flex-flow: column;
			gap: 1px;
			min-width: 0;

			.tip-title {
				color: #ffffff;
				font-weight: 600;
				white-space: nowrap;
				overflow: hidden;
				text-overflow: ellipsis;
			}

			.tip-sub {
				display: flex;
				align-items: center;
				gap: 5px;
				color: $placeholder-color;
				font-size: 10.5px;

				.status-tag {
					font-weight: 600;

					&.watched {
						color: #34d399;
					}

					&.unwatched {
						color: $placeholder-color;
					}
				}
			}
		}

		.tip-action-hint {
			margin-left: auto;
			color: rgba(255, 255, 255, 0.4);
			font-size: 10px;
			font-style: italic;
			white-space: nowrap;
		}
	}

	@keyframes fadeIn {
		from {
			opacity: 0;
			transform: translateY(-2px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	.visually-hidden {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip: rect(0, 0, 0, 0);
		border: 0;
	}
</style>
