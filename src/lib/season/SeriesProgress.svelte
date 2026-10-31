<script lang="ts">
	import type { MediaSeason, Watched } from "@/types";

	interface Props {
		seasons: MediaSeason[];
		watchedItem: Watched | undefined;
	}

	let { seasons, watchedItem }: Props = $props();

	const stats = $derived.by(() => {
		if (!seasons || seasons.length === 0) return null;

		const regularSeasons = seasons.filter((s) => s.number > 0);
		const totalEpisodes = regularSeasons.reduce(
			(acc, s) => acc + (s.episodeCount || 0),
			0,
		);

		if (totalEpisodes === 0) return null;

		const watchedEpisodes = (watchedItem?.watchedEpisodes || []).filter(
			(we) => we.seasonNumber > 0 && we.status === "FINISHED",
		);

		const watchedCount = watchedEpisodes.length;
		const percentage = Math.min(
			100,
			Math.round((watchedCount / totalEpisodes) * 100),
		);

		return {
			totalEpisodes,
			watchedCount,
			percentage,
			isCompleted: watchedCount >= totalEpisodes && totalEpisodes > 0,
		};
	});
</script>

{#if stats}
	<div class="series-progress-card">
		<div class="progress-top">
			<div class="progress-info">
				<span class="progress-label">Series Progress</span>
				<strong class="progress-fraction">
					{stats.watchedCount} of {stats.totalEpisodes} episodes
				</strong>
			</div>
			<div class="progress-percent">
				{#if stats.isCompleted}
					<span class="completed-pill">Completed</span>
				{:else}
					<span class="percent-text">{stats.percentage}%</span>
				{/if}
			</div>
		</div>

		<div class="progress-track">
			<div
				class="progress-fill"
				class:completed={stats.isCompleted}
				style={`width: ${stats.percentage}%`}
			></div>
		</div>
	</div>
{/if}

<style lang="scss">
	.series-progress-card {
		width: 100%;
		background: rgba(22, 27, 38, 0.65);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 14px;
		padding: 16px 20px;
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.35);
		backdrop-filter: blur(16px);
		-webkit-backdrop-filter: blur(16px);
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	.progress-top {
		display: flex;
		align-items: center;
		justify-content: space-between;

		.progress-info {
			display: flex;
			align-items: baseline;
			gap: 10px;
			flex-wrap: wrap;

			.progress-label {
				font-size: 12px;
				font-weight: 700;
				letter-spacing: 0.06em;
				text-transform: uppercase;
				color: rgba(255, 255, 255, 0.5);
			}

			.progress-fraction {
				font-size: 14.5px;
				font-weight: 600;
				color: #ffffff;
			}
		}

		.progress-percent {
			.percent-text {
				font-size: 14px;
				font-weight: 700;
				color: #818cf8;
			}

			.completed-pill {
				font-size: 12px;
				font-weight: 700;
				color: #4ade80;
				background: rgba(34, 197, 94, 0.15);
				border: 1px solid rgba(34, 197, 94, 0.3);
				padding: 2px 10px;
				border-radius: 999px;
				text-transform: uppercase;
				letter-spacing: 0.04em;
			}
		}
	}

	.progress-track {
		width: 100%;
		height: 8px;
		background: rgba(255, 255, 255, 0.08);
		border-radius: 999px;
		overflow: hidden;

		.progress-fill {
			height: 100%;
			background: #6366f1;
			border-radius: 999px;
			transition: width 300ms cubic-bezier(0.4, 0, 0.2, 1);

			&.completed {
				background: $success;
			}
		}
	}
</style>
