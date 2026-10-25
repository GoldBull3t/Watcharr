<script lang="ts">
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import { req } from "@/lib/util/api";
	import type { FranchiseSummary } from "@/types";
	import Spinner from "@/lib/Spinner.svelte";
	import Icon from "@/lib/Icon.svelte";
	import { onMount } from "svelte";

	let franchises = $state<FranchiseSummary[]>([]);
	let loading = $state(true);
	let error = $state<string | undefined>();
	let searchQuery = $state("");
	let activeFilter = $state<"all" | "in_progress" | "completed" | "unwatched">("all");

	async function loadFranchises() {
		loading = true;
		error = undefined;
		try {
			const data = await req.get<FranchiseSummary[]>("/franchise");
			franchises = data || [];
		} catch (err: any) {
			console.error("Failed to load franchises:", err);
			error = err?.message || "Erro ao carregar lista de franquias.";
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadFranchises();
	});

	// Global statistics derived
	let totalFranchises = $derived(franchises.length);
	let completedFranchises = $derived(
		franchises.filter((f) => f.totalItems > 0 && f.watchedCount === f.totalItems).length,
	);
	let inProgressFranchises = $derived(
		franchises.filter((f) => f.watchedCount > 0 && f.watchedCount < f.totalItems).length,
	);
	let totalWatchedItems = $derived(
		franchises.reduce((acc, f) => acc + f.watchedCount, 0),
	);
	let totalPossibleItems = $derived(
		franchises.reduce((acc, f) => acc + f.totalItems, 0),
	);
	let globalPct = $derived(
		totalPossibleItems > 0 ? Math.round((totalWatchedItems / totalPossibleItems) * 100) : 0,
	);

	// Filtered list
	let filteredFranchises = $derived.by(() => {
		let list = franchises;

		// Tab filter
		if (activeFilter === "in_progress") {
			list = list.filter((f) => f.watchedCount > 0 && f.watchedCount < f.totalItems);
		} else if (activeFilter === "completed") {
			list = list.filter((f) => f.totalItems > 0 && f.watchedCount === f.totalItems);
		} else if (activeFilter === "unwatched") {
			list = list.filter((f) => f.watchedCount === 0);
		}

		// Search query filter
		const q = searchQuery.trim().toLowerCase();
		if (q) {
			list = list.filter(
				(f) =>
					f.name.toLowerCase().includes(q) ||
					(f.description && f.description.toLowerCase().includes(q)),
			);
		}

		return list;
	});
</script>

<svelte:head>
	<title>Sagas & Franquias | Watcharr</title>
</svelte:head>

<div class="franchises-container">
	<!-- Header & Hero -->
	<div class="hero-section">
		<div class="hero-content">
			<div class="badge-pill">
				<Icon i="reel" wh={14} />
				<span>Coleções & Sagas TMDB</span>
			</div>
			<h1 class="hero-title">Sagas & Linhas Cronológicas</h1>
			<p class="hero-subtitle">
				Acompanha as maiores sagas do cinema e televisão, descobre a ordem cronológica correta e vê o teu progresso em tempo real.
			</p>

			{#if !loading && franchises.length > 0}
				<div class="stats-ribbon">
					<div class="stat-pill">
						<span class="stat-val">{totalFranchises}</span>
						<span class="stat-lbl">Franquias</span>
					</div>
					<div class="stat-sep">•</div>
					<div class="stat-pill">
						<span class="stat-val highlight">{completedFranchises}</span>
						<span class="stat-lbl">Completas</span>
					</div>
					<div class="stat-sep">•</div>
					<div class="stat-pill">
						<span class="stat-val">{inProgressFranchises}</span>
						<span class="stat-lbl">Em Progresso</span>
					</div>
					<div class="stat-sep">•</div>
					<div class="stat-pill">
						<span class="stat-val">{globalPct}%</span>
						<span class="stat-lbl">Concluído</span>
					</div>
				</div>
			{/if}
		</div>
	</div>

	<!-- Controls: Search & Tabs -->
	<div class="controls-bar">
		<div class="search-box">
			<Icon i="search" wh={16} />
			<input
				type="text"
				placeholder="Pesquisar saga (ex: Marvel, Dune, Star Wars, Batman)..."
				bind:value={searchQuery}
			/>
			{#if searchQuery}
				<button type="button" class="clear-btn" onclick={() => (searchQuery = "")}>
					<Icon i="close" wh={14} />
				</button>
			{/if}
		</div>

		<div class="filter-tabs">
			<button
				type="button"
				class="tab {activeFilter === 'all' ? 'active' : ''}"
				onclick={() => (activeFilter = "all")}
			>
				Todas ({totalFranchises})
			</button>
			<button
				type="button"
				class="tab {activeFilter === 'in_progress' ? 'active' : ''}"
				onclick={() => (activeFilter = "in_progress")}
			>
				Em Progresso ({inProgressFranchises})
			</button>
			<button
				type="button"
				class="tab {activeFilter === 'completed' ? 'active' : ''}"
				onclick={() => (activeFilter = "completed")}
			>
				Completas ({completedFranchises})
			</button>
			<button
				type="button"
				class="tab {activeFilter === 'unwatched' ? 'active' : ''}"
				onclick={() => (activeFilter = "unwatched")}
			>
				Por Iniciar
			</button>
		</div>
	</div>

	<!-- Content Area -->
	{#if loading}
		<div class="loading-wrap">
			<Spinner />
			<p>A carregar sagas e coleções...</p>
		</div>
	{:else if error}
		<div class="error-wrap">
			<Icon i="alert-triangle" wh={32} />
			<p>{error}</p>
			<button type="button" class="retry-btn" onclick={loadFranchises}>Tentar Novamente</button>
		</div>
	{:else if filteredFranchises.length === 0}
		<div class="empty-wrap">
			<Icon i="reel" wh={42} />
			<h3>Nenhuma saga encontrada</h3>
			<p>Tenta ajustar os termos de pesquisa ou os filtros selecionados.</p>
			{#if searchQuery || activeFilter !== "all"}
				<button
					type="button"
					class="reset-btn"
					onclick={() => {
						searchQuery = "";
						activeFilter = "all";
					}}
				>
					Repor Filtros
				</button>
			{/if}
		</div>
	{:else}
		<div class="franchise-grid">
			{#each filteredFranchises as f (f.id)}
				<a
					href={resolve(`/franchise/${f.id}`)}
					class="franchise-card"
					style={`--card-bg-image: url('https://image.tmdb.org/t/p/w780${f.backdropPath || f.posterPath || ""}');`}
				>
					<div class="card-bg"></div>
					<div class="card-overlay"></div>

					<div class="card-body">
						<div class="card-top">
							<span class="media-type-badge">
								<Icon i="reel" wh={12} />
								{f.itemTypes || "Saga"}
							</span>

							{#if f.watchedCount === f.totalItems && f.totalItems > 0}
								<span class="status-badge complete">
									<Icon i="check" wh={12} />
									Completa
								</span>
							{:else if f.watchedCount > 0}
								<span class="status-badge progress">
									{f.watchedCount}/{f.totalItems} Vistos
								</span>
							{:else}
								<span class="status-badge unwatched">
									{f.totalItems} Títulos
								</span>
							{/if}
						</div>

						<div class="card-bottom">
							<h3 class="franchise-title">{f.name}</h3>
							{#if f.description}
								<p class="franchise-desc">{f.description}</p>
							{/if}

							<!-- Progress Bar -->
							<div class="progress-container">
								<div class="progress-bar">
									<div
										class="progress-fill {f.percentWatched === 100 ? 'done' : ''}"
										style={`width: ${f.percentWatched}%;`}
									></div>
								</div>
								<span class="progress-label">{f.percentWatched}%</span>
							</div>
						</div>
					</div>
				</a>
			{/each}
		</div>
	{/if}
</div>

<style lang="scss">
	.franchises-container {
		width: 100%;
		max-width: 1380px;
		margin: 0 auto;
		padding: 24px 20px 80px;
		display: flex;
		flex-direction: column;
		gap: 28px;
	}

	.hero-section {
		position: relative;
		padding: 40px 32px;
		border-radius: 20px;
		background: radial-gradient(120% 120% at 50% 0%, rgba(99, 102, 241, 0.15) 0%, rgba(20, 20, 22, 0.6) 100%),
			rgba(255, 255, 255, 0.02);
		border: 1px solid rgba(255, 255, 255, 0.08);
		overflow: hidden;
	}

	.hero-content {
		max-width: 760px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.badge-pill {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 12px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.6px;
		color: #a5b4fc;
		background: rgba(99, 102, 241, 0.18);
		padding: 4px 12px;
		border-radius: 20px;
		border: 1px solid rgba(99, 102, 241, 0.3);
		width: fit-content;
	}

	.hero-title {
		font-size: 2.2rem;
		font-weight: 700;
		color: #fff;
		margin: 0;
		letter-spacing: -0.5px;
	}

	.hero-subtitle {
		font-size: 1rem;
		color: rgba(255, 255, 255, 0.7);
		margin: 0;
		line-height: 1.5;
	}

	.stats-ribbon {
		display: flex;
		align-items: center;
		gap: 14px;
		margin-top: 10px;
		flex-wrap: wrap;
	}

	.stat-pill {
		display: flex;
		align-items: baseline;
		gap: 6px;
	}

	.stat-val {
		font-size: 1.25rem;
		font-weight: 700;
		color: #fff;

		&.highlight {
			color: #4ade80;
		}
	}

	.stat-lbl {
		font-size: 0.85rem;
		color: rgba(255, 255, 255, 0.55);
	}

	.stat-sep {
		color: rgba(255, 255, 255, 0.25);
	}

	.controls-bar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		flex-wrap: wrap;
	}

	.search-box {
		position: relative;
		display: flex;
		align-items: center;
		gap: 10px;
		background: rgba(255, 255, 255, 0.05);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 12px;
		padding: 10px 16px;
		flex: 1;
		min-width: 280px;
		max-width: 480px;
		transition: border-color 0.2s;

		&:focus-within {
			border-color: #818cf8;
			background: rgba(255, 255, 255, 0.08);
		}

		input {
			border: none;
			background: transparent;
			color: #fff;
			font-size: 0.92rem;
			width: 100%;
			outline: none;

			&::placeholder {
				color: rgba(255, 255, 255, 0.4);
			}
		}

		.clear-btn {
			background: transparent;
			border: none;
			color: rgba(255, 255, 255, 0.5);
			cursor: pointer;
			padding: 2px;
			display: flex;

			&:hover {
				color: #fff;
			}
		}
	}

	.filter-tabs {
		display: flex;
		align-items: center;
		gap: 6px;
		background: rgba(255, 255, 255, 0.04);
		padding: 4px;
		border-radius: 12px;
		border: 1px solid rgba(255, 255, 255, 0.06);

		.tab {
			background: transparent;
			border: none;
			color: rgba(255, 255, 255, 0.65);
			font-size: 0.85rem;
			font-weight: 500;
			padding: 8px 14px;
			border-radius: 8px;
			cursor: pointer;
			transition: all 0.18s;

			&:hover {
				color: #fff;
			}

			&.active {
				background: rgba(99, 102, 241, 0.25);
				color: #fff;
				font-weight: 600;
			}
		}
	}

	.franchise-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
		gap: 20px;

		@media screen and (max-width: 640px) {
			grid-template-columns: 1fr;
		}
	}

	.franchise-card {
		position: relative;
		height: 240px;
		border-radius: 16px;
		overflow: hidden;
		text-decoration: none;
		border: 1px solid rgba(255, 255, 255, 0.09);
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.35);
		transition: transform 0.22s cubic-bezier(0.16, 1, 0.3, 1), border-color 0.22s, box-shadow 0.22s;
		display: flex;
		flex-direction: column;

		&:hover {
			transform: translateY(-4px);
			border-color: rgba(99, 102, 241, 0.6);
			box-shadow: 0 12px 30px rgba(99, 102, 241, 0.22);

			.card-bg {
				transform: scale(1.05);
			}
		}
	}

	.card-bg {
		position: absolute;
		inset: 0;
		background-image: var(--card-bg-image);
		background-size: cover;
		background-position: center;
		transition: transform 0.4s ease-out;
		z-index: 1;
	}

	.card-overlay {
		position: absolute;
		inset: 0;
		background: linear-gradient(180deg, rgba(14, 14, 17, 0.35) 0%, rgba(14, 14, 17, 0.94) 85%);
		z-index: 2;
	}

	.card-body {
		position: relative;
		z-index: 3;
		padding: 18px 20px;
		display: flex;
		flex-direction: column;
		justify-content: space-between;
		height: 100%;
	}

	.card-top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
	}

	.media-type-badge {
		display: inline-flex;
		align-items: center;
		gap: 5px;
		font-size: 11px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		color: rgba(255, 255, 255, 0.85);
		background: rgba(0, 0, 0, 0.55);
		backdrop-filter: blur(8px);
		padding: 3px 10px;
		border-radius: 12px;
		border: 1px solid rgba(255, 255, 255, 0.12);
	}

	.status-badge {
		font-size: 11px;
		font-weight: 600;
		padding: 3px 10px;
		border-radius: 12px;
		display: inline-flex;
		align-items: center;
		gap: 4px;

		&.complete {
			background: rgba(34, 197, 94, 0.22);
			color: #4ade80;
			border: 1px solid rgba(74, 222, 128, 0.4);
		}

		&.progress {
			background: rgba(99, 102, 241, 0.24);
			color: #a5b4fc;
			border: 1px solid rgba(165, 180, 252, 0.4);
		}

		&.unwatched {
			background: rgba(255, 255, 255, 0.1);
			color: rgba(255, 255, 255, 0.7);
			border: 1px solid rgba(255, 255, 255, 0.15);
		}
	}

	.card-bottom {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.franchise-title {
		font-size: 1.25rem;
		font-weight: 700;
		color: #fff;
		margin: 0;
		line-height: 1.3;
		text-shadow: 0 2px 8px rgba(0, 0, 0, 0.6);
	}

	.franchise-desc {
		font-size: 0.82rem;
		color: rgba(255, 255, 255, 0.7);
		margin: 0;
		line-height: 1.4;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	.progress-container {
		display: flex;
		align-items: center;
		gap: 10px;
		margin-top: 4px;
	}

	.progress-bar {
		flex: 1;
		height: 6px;
		background: rgba(255, 255, 255, 0.14);
		border-radius: 6px;
		overflow: hidden;
	}

	.progress-fill {
		height: 100%;
		background: linear-gradient(90deg, #6366f1 0%, #818cf8 100%);
		border-radius: 6px;
		transition: width 0.3s ease;

		&.done {
			background: linear-gradient(90deg, #22c55e 0%, #4ade80 100%);
		}
	}

	.progress-label {
		font-size: 0.75rem;
		font-weight: 700;
		color: rgba(255, 255, 255, 0.85);
		min-width: 32px;
		text-align: right;
	}

	.loading-wrap,
	.empty-wrap,
	.error-wrap {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 12px;
		padding: 80px 20px;
		color: rgba(255, 255, 255, 0.6);
		text-align: center;
	}

	.reset-btn,
	.retry-btn {
		background: rgba(99, 102, 241, 0.2);
		border: 1px solid rgba(99, 102, 241, 0.4);
		color: #fff;
		padding: 8px 18px;
		border-radius: 8px;
		cursor: pointer;
		font-size: 0.88rem;
		font-weight: 600;
		margin-top: 6px;
		transition: background 0.2s;

		&:hover {
			background: rgba(99, 102, 241, 0.35);
		}
	}
</style>
