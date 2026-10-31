<script lang="ts">
	import { page } from "$app/state";
	import { goto } from "$app/navigation";
	import { req } from "@/lib/util/api";
	import type { FranchiseRoadmap } from "@/types";
	import Error from "@/lib/Error.svelte";
	import FranchiseTimeline from "@/lib/franchise/FranchiseTimeline.svelte";
	import MediaDetailSkeleton from "@/lib/content/MediaDetailSkeleton.svelte";
	import Icon from "@/lib/Icon.svelte";

	let franchiseId = $derived(page.params.id);
	let activeOrder = $state<"timeline" | "release">("timeline");
	let roadmap = $state<FranchiseRoadmap | undefined>();
	let loading = $state(true);
	let error = $state<string | undefined>();

	async function loadRoadmap(
		order: "timeline" | "release",
		signal?: AbortSignal,
	) {
		if (!franchiseId) return;
		try {
			loading = true;
			error = undefined;
			const data = await req.get<FranchiseRoadmap>(
				`/franchise/${franchiseId}?order=${order}`,
				{
					signal,
				},
			);
			if (signal?.aborted) return;
			roadmap = data;
			activeOrder = data?.activeOrder || order;
		} catch (err: any) {
			if (signal?.aborted) return;
			console.error("Failed to load franchise roadmap:", err);
			error =
				err?.message || "Não foi possível carregar a rota desta franquia.";
		} finally {
			if (!signal?.aborted) {
				loading = false;
			}
		}
	}

	$effect(() => {
		const controller = new AbortController();
		const orderParam = page.url.searchParams.get("order");
		const currentOrder = orderParam === "release" ? "release" : "timeline";
		activeOrder = currentOrder;
		loadRoadmap(currentOrder, controller.signal);
		return () => controller.abort();
	});

	function handleOrderChange(newOrder: "timeline" | "release") {
		activeOrder = newOrder;
		const url = new URL(window.location.href);
		url.searchParams.set("order", newOrder);
		goto(url.toString(), { replaceState: true, noScroll: true });
	}
</script>

<svelte:head>
	<title
		>{roadmap ? `${roadmap.name} • Linha Cronológica` : "Franquia"} | Watcharr</title
	>
</svelte:head>

<div class="franchise-detail-page">
	<div class="back-nav-bar">
		<button
			type="button"
			class="back-link"
			onclick={() => window.history.back()}
		>
			<span class="back-arrow">←</span>
			<span>Voltar</span>
		</button>
	</div>

	{#if loading && !roadmap}
		<MediaDetailSkeleton type="movie" />
	{:else if error}
		<div class="error-state">
			<Error
				pretty="Falha ao carregar linha cronológica"
				{error}
				onRetry={() => {
					loadRoadmap(activeOrder);
				}}
			/>
		</div>
	{:else if roadmap}
		<!-- Universe Hero Banner -->
		<div class="universe-hero">
			{#if roadmap.backdropPath}
				<div
					class="hero-bg"
					style={`background-image: linear-gradient(180deg, rgba(13, 17, 23, 0.4) 0%, rgba(13, 17, 23, 0.96) 100%), url('https://image.tmdb.org/t/p/w1920_and_h800_multi_faces${roadmap.backdropPath}');`}
				></div>
			{/if}

			<div class="hero-inner">
				{#if roadmap.posterPath}
					<div class="hero-poster-wrap">
						<img
							src={`https://image.tmdb.org/t/p/w500${roadmap.posterPath}`}
							alt={roadmap.name}
							class="hero-poster"
						/>
					</div>
				{/if}

				<div class="hero-text">
					<div class="hero-badges">
						<span class="badge cosmic-badge">🌌 Saga & Universo</span>
						<span class="badge stats-badge">{roadmap.totalItems} Produções</span
						>
						{#if roadmap.watchedCount > 0}
							<span
								class="badge completion-badge"
								class:completed={roadmap.percentWatched === 100}
							>
								<Icon i="check" wh={13} />
								{roadmap.watchedCount} / {roadmap.totalItems} Vistos ({roadmap.percentWatched}%)
							</span>
						{/if}
					</div>

					<h1 class="hero-title">{roadmap.name}</h1>
					<p class="hero-desc">{roadmap.description}</p>
				</div>
			</div>
		</div>

		<!-- Interactive Timeline Tree -->
		<FranchiseTimeline {roadmap} onOrderChange={handleOrderChange} />
	{/if}
</div>

<style lang="scss">
	.franchise-detail-page {
		max-width: 1200px;
		margin: 0 auto;
		padding: 16px 20px 80px;
	}

	.back-nav-bar {
		margin-bottom: 16px;
	}

	.back-link {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		background: transparent;
		border: none;
		color: #8b949e;
		font-size: 0.88rem;
		font-weight: 600;
		cursor: pointer;
		padding: 6px 0;
		transition: color 0.2s ease;

		&:hover {
			color: #58a6ff;
			.back-arrow {
				transform: translateX(-3px);
			}
		}
	}

	.back-arrow {
		transition: transform 0.2s ease;
		display: inline-block;
	}

	.error-state {
		padding: 60px 20px;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 16px;
	}

	/* Universe Hero */
	.universe-hero {
		position: relative;
		border-radius: 20px;
		overflow: hidden;
		background: #161b22;
		border: 1px solid rgba(255, 255, 255, 0.08);
		margin-bottom: 24px;
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.4);
	}

	.hero-bg {
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		background-size: cover;
		background-position: center 25%;
		opacity: 0.35;
		pointer-events: none;
	}

	.hero-inner {
		position: relative;
		z-index: 2;
		display: flex;
		align-items: center;
		gap: 28px;
		padding: 32px;

		@media (max-width: 680px) {
			flex-direction: column;
			align-items: flex-start;
			padding: 20px;
			gap: 16px;
		}
	}

	.hero-poster-wrap {
		width: 140px;
		min-width: 140px;

		@media (max-width: 680px) {
			width: 110px;
			min-width: 110px;
		}
	}

	.hero-poster {
		width: 100%;
		aspect-ratio: 2/3;
		object-fit: cover;
		border-radius: 12px;
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
		display: block;
	}

	.hero-text {
		flex: 1;
	}

	.hero-badges {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
		margin-bottom: 10px;
	}

	.badge {
		padding: 3px 10px;
		border-radius: 999px;
		font-size: 0.78rem;
		font-weight: 600;
	}

	.cosmic-badge {
		background: rgba(137, 87, 229, 0.2);
		border: 1px solid rgba(137, 87, 229, 0.4);
		color: #d2a8ff;
	}

	.stats-badge {
		background: rgba(255, 255, 255, 0.08);
		border: 1px solid rgba(255, 255, 255, 0.12);
		color: #e6edf3;
	}

	.completion-badge {
		display: inline-flex;
		align-items: center;
		gap: 5px;
		background: rgba(31, 111, 235, 0.18);
		border: 1px solid rgba(31, 111, 235, 0.4);
		color: #58a6ff;

		&.completed {
			background: rgba(46, 160, 67, 0.2);
			border-color: rgba(46, 160, 67, 0.45);
			color: #3fb950;
		}
	}

	.hero-title {
		font-size: 2rem;
		font-weight: 800;
		color: #f0f6fc;
		margin: 0 0 10px;
		letter-spacing: -0.5px;
		line-height: 1.2;

		@media (max-width: 680px) {
			font-size: 1.5rem;
		}
	}

	.hero-desc {
		font-size: 0.95rem;
		color: #c9d1d9;
		margin: 0;
		line-height: 1.5;
		max-width: 800px;
	}
</style>
