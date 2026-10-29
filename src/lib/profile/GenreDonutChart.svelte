<script lang="ts">
	import type { GenreStat } from "@/types";

	interface Props {
		genres: GenreStat[];
	}

	let { genres = [] }: Props = $props();

	// Color palette for the donut segments
	const colors = [
		"#f59e0b", // Amber
		"#6366f1", // Indigo
		"#10b981", // Emerald
		"#ec4899", // Rose
		"#06b6d4", // Cyan
		"#8b5cf6", // Violet
		"#f97316", // Orange
		"#64748b", // Slate
	];

	let totalCount = $derived(genres.reduce((acc, g) => acc + (g.count ?? (g as any).Count ?? 0), 0));

	let topGenres = $derived.by(() => {
		if (genres.length <= 6) return genres;
		const head = genres.slice(0, 5);
		const othersCount = genres.slice(5).reduce((acc, g) => acc + (g.count ?? (g as any).Count ?? 0), 0);
		return [...head, { name: "Outros", count: othersCount }];
	});

	// Hovered item for interactive center preview
	let hoveredIndex: number | undefined = $state();

	// Donut geometry constants
	const radius = 64;
	const circumference = 2 * Math.PI * radius; // ~402.12

	let slices = $derived.by(() => {
		let currentOffset = 0;
		return topGenres.map((g, idx) => {
			const count = g.count ?? (g as any).Count ?? 0;
			const name = g.name ?? (g as any).Name ?? "";
			const pct = totalCount > 0 ? count / totalCount : 0;
			const length = pct * circumference;
			const dasharray = `${length} ${circumference - length}`;
			const dashoffset = -currentOffset;
			currentOffset += length;
			return {
				name,
				count,
				percent: Math.round(pct * 100),
				color: colors[idx % colors.length],
				dasharray,
				dashoffset,
			};
		});
	});

	let activeSlice = $derived.by(() => {
		if (typeof hoveredIndex === "number" && slices[hoveredIndex]) {
			return slices[hoveredIndex];
		}
		return slices[0];
	});
</script>

<div class="donut-chart-container">
	<div class="donut-visual-wrap">
		<svg
			viewBox="0 0 160 160"
			class="donut-svg"
			aria-label="Gráfico de distribuição de géneros"
		>
			<!-- Base background circle -->
			<circle
				cx="80"
				cy="80"
				r={radius}
				fill="transparent"
				stroke="rgba(255, 255, 255, 0.05)"
				stroke-width="18"
			/>

			<!-- Slices -->
			{#each slices as slice, i (slice.name)}
				<circle
					cx="80"
					cy="80"
					r={radius}
					fill="transparent"
					stroke={slice.color}
					stroke-width="18"
					stroke-dasharray={slice.dasharray}
					stroke-dashoffset={slice.dashoffset}
					stroke-linecap="round"
					class="donut-segment"
					class:hovered={hoveredIndex === i}
					onmouseenter={() => (hoveredIndex = i)}
					onmouseleave={() => (hoveredIndex = undefined)}
					role="graphics-symbol"
					aria-label={`${slice.name}: ${slice.count} títulos (${slice.percent}%)`}
				/>
			{/each}
		</svg>

		<!-- Center text overlay -->
		<div class="donut-center-info">
			{#if activeSlice}
				<span class="center-pct">{activeSlice.percent}%</span>
				<span class="center-name" title={activeSlice.name}>{activeSlice.name}</span>
				<span class="center-sub">{activeSlice.count} títulos</span>
			{:else}
				<span class="center-pct">0%</span>
				<span class="center-name">Sem dados</span>
			{/if}
		</div>
	</div>

	<!-- Legend List -->
	<div class="donut-legend">
		{#each slices as slice, i (slice.name)}
			<div
				class="legend-item"
				class:active={hoveredIndex === i}
				onmouseenter={() => (hoveredIndex = i)}
				onmouseleave={() => (hoveredIndex = undefined)}
				role="button"
				tabindex="0"
			>
				<span class="legend-dot" style={`background-color: ${slice.color};`}></span>
				<span class="legend-name">{slice.name}</span>
				<span class="legend-val">{slice.percent}%</span>
			</div>
		{/each}
	</div>
</div>

<style lang="scss">
	.donut-chart-container {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 20px;
		width: 100%;
		padding: 10px 0;

		@media screen and (max-width: 600px) {
			flex-direction: column;
		}

		.donut-visual-wrap {
			position: relative;
			width: 150px;
			height: 150px;
			flex-shrink: 0;

			.donut-svg {
				width: 100%;
				height: 100%;
				transform: rotate(-90deg);

				.donut-segment {
					transition: stroke-width 160ms ease, opacity 160ms ease, filter 160ms ease;
					cursor: pointer;

					&:hover,
					&.hovered {
						stroke-width: 22;
						filter: drop-shadow(0 0 8px currentColor);
					}
				}
			}

			.donut-center-info {
				position: absolute;
				top: 50%;
				left: 50%;
				transform: translate(-50%, -50%);
				display: flex;
				flex-direction: column;
				align-items: center;
				justify-content: center;
				text-align: center;
				pointer-events: none;
				width: 90px;

				.center-pct {
					font-size: 19px;
					font-weight: 800;
					color: #ffffff;
					line-height: 1.1;
				}

				.center-name {
					font-size: 10.5px;
					font-weight: 600;
					color: #fbbf24;
					margin-top: 2px;
					max-width: 80px;
					overflow: hidden;
					text-overflow: ellipsis;
					white-space: nowrap;
				}

				.center-sub {
					font-size: 9.5px;
					color: rgba(255, 255, 255, 0.5);
				}
			}
		}

		.donut-legend {
			display: flex;
			flex-direction: column;
			gap: 6px;
			flex: 1;

			.legend-item {
				display: flex;
				align-items: center;
				gap: 8px;
				padding: 4px 8px;
				border-radius: 6px;
				cursor: pointer;
				transition: background-color 130ms ease;

				&:hover,
				&.active {
					background-color: rgba(255, 255, 255, 0.08);
				}

				.legend-dot {
					width: 8px;
					height: 8px;
					border-radius: 50%;
					flex-shrink: 0;
				}

				.legend-name {
					font-size: 12px;
					font-weight: 500;
					color: rgba(255, 255, 255, 0.85);
					flex: 1;
					overflow: hidden;
					text-overflow: ellipsis;
					white-space: nowrap;
				}

				.legend-val {
					font-size: 11.5px;
					font-weight: 700;
					color: rgba(255, 255, 255, 0.55);
				}
			}
		}
	}
</style>
