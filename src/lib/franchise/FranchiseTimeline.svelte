<script lang="ts">
	import type { FranchiseRoadmap, FranchiseItem } from "@/types";
	import { goto } from "$app/navigation";
	import Icon from "@/lib/Icon.svelte";
	import { flip } from "svelte/animate";
	import { fade } from "svelte/transition";

	interface Props {
		roadmap: FranchiseRoadmap;
		onOrderChange?: (order: "timeline" | "release") => void;
	}

	let { roadmap, onOrderChange }: Props = $props();

	let selectedPhase = $state<string>("all");

	// Extract unique phases / eras for filter chips
	const allPhases = $derived.by(() => {
		const phases: string[] = [];
		for (const item of roadmap.items) {
			if (item.phaseOrEra && !phases.includes(item.phaseOrEra)) {
				phases.push(item.phaseOrEra);
			}
		}
		return phases;
	});

	// Filtered items based on phase filter
	const displayItems = $derived.by(() => {
		if (selectedPhase === "all") return roadmap.items;
		return roadmap.items.filter((item) => item.phaseOrEra === selectedPhase);
	});

	function handleOrderSwitch(order: "timeline" | "release") {
		if (roadmap.activeOrder !== order && onOrderChange) {
			onOrderChange(order);
		}
	}

	function navigateToItem(item: FranchiseItem) {
		const path = item.type === "tv" ? `/tv/${item.tmdbId}` : `/movie/${item.tmdbId}`;
		goto(path);
	}

	function scrollToNextUp() {
		const el = document.getElementById("next-up-node");
		if (el) {
			el.scrollIntoView({ behavior: "smooth", block: "center" });
		}
	}
</script>

<div class="franchise-timeline-container">
	<!-- Sticky Progress Header -->
	<header class="sticky-roadmap-header">
		<div class="sticky-inner">
			<div class="sticky-left">
				<span class="roadmap-title">{roadmap.name}</span>
				<div class="progress-pill">
					<span class="progress-nums">{roadmap.watchedCount} / {roadmap.totalItems} vistos</span>
					<span class="progress-pct">({roadmap.percentWatched}%)</span>
				</div>
			</div>

			<!-- Visual progress bar -->
			<div class="progress-track" title={`${roadmap.percentWatched}% concluído`}>
				<div class="progress-fill" style={`width: ${roadmap.percentWatched}%`}></div>
			</div>

			<div class="sticky-right">
				{#if roadmap.nextUpItem}
					<button class="next-up-btn" onclick={scrollToNextUp}>
						<span class="fire-icon">🔥</span>
						<span class="next-label">A Seguir:</span>
						<strong class="next-title">{roadmap.nextUpItem.title}</strong>
					</button>
				{:else if roadmap.watchedCount === roadmap.totalItems && roadmap.totalItems > 0}
					<div class="all-done-pill">
						<Icon i="check" wh={14} />
						<span>Franquia Concluída!</span>
					</div>
				{/if}
			</div>
		</div>
	</header>

	<!-- Controls & Mode Switcher Bar -->
	<div class="timeline-controls-bar">
		<div class="mode-switcher" role="group" aria-label="Modo de Ordenação">
			<button
				type="button"
				class="mode-btn"
				class:active={roadmap.activeOrder === "timeline"}
				onclick={() => handleOrderSwitch("timeline")}
			>
				<Icon i="clock" wh={16} />
				<span>Ordem Cronológica (Lore)</span>
			</button>
			<button
				type="button"
				class="mode-btn"
				class:active={roadmap.activeOrder === "release"}
				onclick={() => handleOrderSwitch("release")}
			>
				<Icon i="calendar" wh={16} />
				<span>Ordem de Lançamento</span>
			</button>
		</div>

		<!-- Phase / Era Chips (if there are multiple) -->
		{#if allPhases.length > 1}
			<div class="phase-filters-scroll">
				<button
					type="button"
					class="phase-chip"
					class:active={selectedPhase === "all"}
					onclick={() => (selectedPhase = "all")}
				>
					Todos ({roadmap.items.length})
				</button>
				{#each allPhases as phase}
					<button
						type="button"
						class="phase-chip"
						class:active={selectedPhase === phase}
						onclick={() => (selectedPhase = phase)}
					>
						{phase}
					</button>
				{/each}
			</div>
		{/if}
	</div>

	<!-- Vertical Timeline Track -->
	<div class="timeline-tree">
		<div class="timeline-spine"></div>

		{#each displayItems as item, idx (item.type + ":" + item.tmdbId)}
			<div
				animate:flip={{ duration: 320 }}
				transition:fade={{ duration: 200 }}
				id={item.isNextUp ? "next-up-node" : undefined}
				class="timeline-row"
				class:is-watched={item.isWatched}
				class:is-next-up={item.isNextUp}
				class:is-unwatched={!item.isWatched && !item.isNextUp}
				class:row-left={idx % 2 === 0}
				class:row-right={idx % 2 !== 0}
			>
				<!-- Center Node Milestone -->
				<div class="timeline-node-wrapper">
					<div class="timeline-node" class:watched-node={item.isWatched} class:next-node={item.isNextUp}>
						{#if item.isWatched}
							<Icon i="check" wh={16} />
						{:else if item.isNextUp}
							<span class="pulsing-halo"></span>
							<Icon i="play" wh={14} />
						{:else}
							<span class="order-num">
								{roadmap.activeOrder === "release" ? item.releaseOrder : item.timelineOrder}
							</span>
						{/if}
					</div>
					<div class="node-connector"></div>
				</div>

				<!-- Content Card -->
				<div
					class="timeline-card"
					class:highlight-next={item.isNextUp}
					role="button"
					tabindex="0"
					onclick={() => navigateToItem(item)}
					onkeydown={(e) => e.key === "Enter" && navigateToItem(item)}
				>
					<!-- Poster thumbnail -->
					<div class="poster-thumb">
						{#if item.posterPath}
							<img
								src={`https://image.tmdb.org/t/p/w300${item.posterPath}`}
								alt={item.title}
								loading="lazy"
							/>
						{:else}
							<div class="poster-fallback">
								<Icon i="film" wh={28} />
							</div>
						{/if}

						<div class="type-pill" class:tv-pill={item.type === "tv"}>
							{item.type === "tv" ? "SÉRIE" : "FILME"}
						</div>
					</div>

					<!-- Details -->
					<div class="card-content">
						<div class="card-meta-top">
							{#if item.phaseOrEra}
								<span class="era-tag">{item.phaseOrEra}</span>
							{/if}
							<span class="year-tag">{item.year}</span>
						</div>

						<h3 class="card-title">{item.title}</h3>

						<div class="card-status-row">
							{#if item.isWatched}
								<span class="status-badge badge-watched">
									<Icon i="check" wh={12} />
									<span>Visto</span>
								</span>
								{#if item.watched?.rating}
									<span class="rating-badge">
										★ {item.watched.rating.toFixed(1)}
									</span>
								{/if}
							{:else if item.isNextUp}
								<span class="status-badge badge-next">
									<span class="pulse-dot"></span>
									<span>Seguinte a Ver</span>
								</span>
								<span class="jump-cta">
									Ver Página →
								</span>
							{:else}
								<span class="status-badge badge-unwatched">
									Pendente
								</span>
							{/if}
						</div>

						{#if item.overview}
							<p class="card-overview">{item.overview}</p>
						{/if}
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>

<style lang="scss">
	.franchise-timeline-container {
		width: 100%;
		max-width: 1100px;
		margin: 0 auto;
		padding-bottom: 60px;
		position: relative;
	}

	/* Sticky Header */
	.sticky-roadmap-header {
		position: sticky;
		top: 0;
		z-index: 35;
		background: rgba(13, 17, 23, 0.88);
		backdrop-filter: blur(16px);
		-webkit-backdrop-filter: blur(16px);
		border-bottom: 1px solid rgba(255, 255, 255, 0.08);
		padding: 12px 18px;
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.35);
		margin-bottom: 24px;
	}

	.sticky-inner {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		flex-wrap: wrap;
	}

	.sticky-left {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.roadmap-title {
		font-weight: 700;
		font-size: 1.05rem;
		color: #f0f6fc;
		white-space: nowrap;
	}

	.progress-pill {
		display: flex;
		align-items: center;
		gap: 6px;
		background: rgba(255, 255, 255, 0.07);
		border: 1px solid rgba(255, 255, 255, 0.1);
		padding: 3px 10px;
		border-radius: 999px;
		font-size: 0.82rem;
	}

	.progress-nums {
		color: #e6edf3;
		font-weight: 600;
	}

	.progress-pct {
		color: #388bfd;
		font-weight: 700;
	}

	.progress-track {
		flex: 1;
		min-width: 140px;
		height: 8px;
		background: rgba(255, 255, 255, 0.08);
		border-radius: 999px;
		overflow: hidden;
		position: relative;
	}

	.progress-fill {
		height: 100%;
		background: linear-gradient(90deg, #1f6feb, #2ea043);
		border-radius: 999px;
		transition: width 0.4s cubic-bezier(0.16, 1, 0.3, 1);
		box-shadow: 0 0 10px rgba(46, 160, 67, 0.5);
	}

	.sticky-right {
		display: flex;
		align-items: center;
	}

	.next-up-btn {
		display: flex;
		align-items: center;
		gap: 8px;
		background: rgba(240, 136, 62, 0.15);
		border: 1px solid rgba(240, 136, 62, 0.35);
		color: #f0883e;
		padding: 5px 12px;
		border-radius: 8px;
		cursor: pointer;
		font-size: 0.85rem;
		transition: all 0.2s ease;

		&:hover {
			background: rgba(240, 136, 62, 0.25);
			border-color: #f0883e;
			transform: translateY(-1px);
		}
	}

	.fire-icon {
		font-size: 1rem;
	}

	.next-label {
		color: #8b949e;
		font-size: 0.78rem;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.next-title {
		color: #fff;
		font-weight: 600;
		max-width: 180px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.all-done-pill {
		display: flex;
		align-items: center;
		gap: 6px;
		background: rgba(46, 160, 67, 0.18);
		border: 1px solid rgba(46, 160, 67, 0.35);
		color: #3fb950;
		padding: 5px 12px;
		border-radius: 8px;
		font-size: 0.85rem;
		font-weight: 600;
	}

	/* Controls Bar */
	.timeline-controls-bar {
		display: flex;
		flex-direction: column;
		gap: 16px;
		margin-bottom: 32px;
		padding: 0 16px;
	}

	.mode-switcher {
		display: inline-flex;
		align-self: center;
		background: rgba(22, 27, 34, 0.9);
		border: 1px solid rgba(255, 255, 255, 0.12);
		border-radius: 12px;
		padding: 4px;
		gap: 4px;
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.25);
	}

	.mode-btn {
		display: flex;
		align-items: center;
		gap: 8px;
		background: transparent;
		border: none;
		color: #8b949e;
		padding: 8px 18px;
		border-radius: 8px;
		font-size: 0.88rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;

		&:hover {
			color: #e6edf3;
			background: rgba(255, 255, 255, 0.05);
		}

		&.active {
			background: #238636;
			color: #ffffff;
			box-shadow: 0 2px 8px rgba(35, 134, 54, 0.4);
		}
	}

	.phase-filters-scroll {
		display: flex;
		gap: 8px;
		overflow-x: auto;
		padding-bottom: 6px;
		scrollbar-width: thin;

		&::-webkit-scrollbar {
			height: 4px;
		}
		&::-webkit-scrollbar-thumb {
			background: rgba(255, 255, 255, 0.1);
			border-radius: 4px;
		}
	}

	.phase-chip {
		background: rgba(255, 255, 255, 0.05);
		border: 1px solid rgba(255, 255, 255, 0.08);
		color: #8b949e;
		padding: 5px 12px;
		border-radius: 999px;
		font-size: 0.8rem;
		white-space: nowrap;
		cursor: pointer;
		transition: all 0.18s ease;

		&:hover {
			color: #e6edf3;
			background: rgba(255, 255, 255, 0.1);
		}

		&.active {
			background: rgba(56, 139, 253, 0.18);
			border-color: #388bfd;
			color: #58a6ff;
			font-weight: 600;
		}
	}

	/* Timeline Tree Structure */
	.timeline-tree {
		position: relative;
		display: flex;
		flex-direction: column;
		gap: 24px;
		padding: 0 16px;
	}

	.timeline-spine {
		position: absolute;
		left: 50%;
		top: 12px;
		bottom: 12px;
		width: 3px;
		background: linear-gradient(180deg, rgba(56, 139, 253, 0.4) 0%, rgba(46, 160, 67, 0.5) 50%, rgba(255, 255, 255, 0.1) 100%);
		transform: translateX(-50%);
		border-radius: 3px;
	}

	.timeline-row {
		display: flex;
		position: relative;
		width: 100%;
		align-items: center;

		&.row-left {
			flex-direction: row;
			.timeline-card {
				margin-right: auto;
			}
			.timeline-node-wrapper {
				order: 2;
			}
			.timeline-card {
				order: 1;
			}
		}

		&.row-right {
			flex-direction: row;
			.timeline-card {
				margin-left: auto;
			}
			.timeline-node-wrapper {
				order: 1;
			}
			.timeline-card {
				order: 2;
			}
		}

		&.is-unwatched {
			opacity: 0.65;
			transition: opacity 0.2s ease;
			&:hover {
				opacity: 0.95;
			}
		}

		&.is-next-up {
			opacity: 1;
		}
	}

	.timeline-node-wrapper {
		position: absolute;
		left: 50%;
		transform: translateX(-50%);
		z-index: 10;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.timeline-node {
		width: 36px;
		height: 36px;
		border-radius: 50%;
		background: #161b22;
		border: 2px solid rgba(255, 255, 255, 0.2);
		color: #8b949e;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 700;
		font-size: 0.8rem;
		box-shadow: 0 0 10px rgba(0, 0, 0, 0.5);
		transition: all 0.25s ease;

		&.watched-node {
			background: #238636;
			border-color: #3fb950;
			color: #ffffff;
			box-shadow: 0 0 14px rgba(46, 160, 67, 0.6);
		}

		&.next-node {
			background: #1f6feb;
			border-color: #58a6ff;
			color: #ffffff;
			box-shadow: 0 0 18px rgba(31, 111, 235, 0.8);
			position: relative;
		}
	}

	.pulsing-halo {
		position: absolute;
		width: 100%;
		height: 100%;
		border-radius: 50%;
		border: 2px solid #58a6ff;
		animation: pulse-halo 1.8s infinite cubic-bezier(0.25, 1, 0.5, 1);
	}

	@keyframes pulse-halo {
		0% {
			transform: scale(1);
			opacity: 0.9;
		}
		100% {
			transform: scale(1.8);
			opacity: 0;
		}
	}

	/* Timeline Card */
	.timeline-card {
		width: calc(50% - 42px);
		background: #161b22;
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 14px;
		display: flex;
		overflow: hidden;
		cursor: pointer;
		transition: all 0.22s cubic-bezier(0.16, 1, 0.3, 1);
		box-shadow: 0 6px 18px rgba(0, 0, 0, 0.28);

		&:hover {
			transform: translateY(-3px);
			border-color: rgba(255, 255, 255, 0.22);
			box-shadow: 0 10px 28px rgba(0, 0, 0, 0.45);
		}

		&.highlight-next {
			border: 2px solid #388bfd;
			background: linear-gradient(135deg, rgba(31, 111, 235, 0.12) 0%, rgba(22, 27, 34, 0.95) 100%);
			box-shadow: 0 0 24px rgba(56, 139, 253, 0.35);

			&:hover {
				border-color: #58a6ff;
				box-shadow: 0 0 32px rgba(56, 139, 253, 0.55);
			}
		}
	}

	.poster-thumb {
		width: 110px;
		min-width: 110px;
		position: relative;
		background: #0d1117;
		display: flex;
		align-items: center;
		justify-content: center;

		img {
			width: 100%;
			height: 100%;
			object-fit: cover;
			display: block;
			transition: transform 0.3s ease;
		}

		&:hover img {
			transform: scale(1.04);
		}
	}

	.poster-fallback {
		color: #484f58;
	}

	.type-pill {
		position: absolute;
		top: 6px;
		left: 6px;
		background: rgba(0, 0, 0, 0.75);
		backdrop-filter: blur(4px);
		border: 1px solid rgba(255, 255, 255, 0.15);
		color: #f0f6fc;
		padding: 2px 6px;
		border-radius: 4px;
		font-size: 0.65rem;
		font-weight: 700;
		letter-spacing: 0.5px;

		&.tv-pill {
			background: rgba(137, 87, 229, 0.85);
			border-color: rgba(137, 87, 229, 0.4);
		}
	}

	.card-content {
		flex: 1;
		padding: 14px 16px;
		display: flex;
		flex-direction: column;
		justify-content: center;
		gap: 6px;
		min-width: 0;
	}

	.card-meta-top {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
	}

	.era-tag {
		background: rgba(56, 139, 253, 0.12);
		color: #58a6ff;
		padding: 2px 8px;
		border-radius: 6px;
		font-size: 0.72rem;
		font-weight: 600;
	}

	.year-tag {
		color: #8b949e;
		font-size: 0.75rem;
	}

	.card-title {
		font-size: 1.02rem;
		font-weight: 700;
		color: #f0f6fc;
		margin: 0;
		line-height: 1.3;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.card-status-row {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-top: 2px;
	}

	.status-badge {
		display: inline-flex;
		align-items: center;
		gap: 5px;
		padding: 3px 8px;
		border-radius: 6px;
		font-size: 0.75rem;
		font-weight: 600;

		&.badge-watched {
			background: rgba(46, 160, 67, 0.18);
			color: #3fb950;
			border: 1px solid rgba(46, 160, 67, 0.35);
		}

		&.badge-next {
			background: rgba(31, 111, 235, 0.2);
			color: #58a6ff;
			border: 1px solid rgba(31, 111, 235, 0.45);
		}

		&.badge-unwatched {
			background: rgba(255, 255, 255, 0.05);
			color: #8b949e;
			border: 1px solid rgba(255, 255, 255, 0.08);
		}
	}

	.pulse-dot {
		width: 6px;
		height: 6px;
		background: #58a6ff;
		border-radius: 50%;
		animation: pulse-dot 1.2s infinite;
	}

	@keyframes pulse-dot {
		0%, 100% { opacity: 1; transform: scale(1); }
		50% { opacity: 0.4; transform: scale(0.8); }
	}

	.rating-badge {
		background: rgba(210, 153, 34, 0.15);
		border: 1px solid rgba(210, 153, 34, 0.35);
		color: #e3b341;
		font-weight: 700;
		font-size: 0.75rem;
		padding: 2px 7px;
		border-radius: 6px;
	}

	.jump-cta {
		font-size: 0.78rem;
		color: #58a6ff;
		font-weight: 600;
		margin-left: auto;
	}

	.card-overview {
		font-size: 0.8rem;
		color: #8b949e;
		margin: 2px 0 0;
		line-height: 1.35;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	/* Responsive for tablets and mobile */
	@media (max-width: 820px) {
		.timeline-spine {
			left: 28px;
		}

		.timeline-row {
			&.row-left,
			&.row-right {
				flex-direction: row;
				.timeline-node-wrapper {
					position: absolute;
					left: 28px;
					order: 1;
				}
				.timeline-card {
					margin-left: 64px;
					margin-right: 0;
					width: calc(100% - 64px);
					order: 2;
				}
			}
		}

		.sticky-inner {
			flex-direction: column;
			align-items: stretch;
		}

		.progress-track {
			width: 100%;
		}
	}
</style>
