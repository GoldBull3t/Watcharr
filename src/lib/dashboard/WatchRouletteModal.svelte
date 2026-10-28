<script lang="ts">
	import Modal from "@/lib/Modal.svelte";
	import Icon from "@/lib/Icon.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import Poster from "@/lib/poster/Poster.svelte";
	import { req, updateWatched } from "@/lib/util/api";
	import { notify } from "@/lib/util/notify";
	import type { Media, RoulettePickRequest, RoulettePickResponse } from "@/types";
	import { onMount } from "svelte";
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";

	interface Props {
		onClose: () => void;
	}

	let { onClose }: Props = $props();

	// Filter state
	let mediaType: "movie" | "tv" | "game" | "all" = $state("movie");
	let maxRuntime: number = $state(0); // 0 = any
	let minRating: number = $state(0); // 0 = any
	let selectedGenre: string = $state("");
	let statusFilter: "planned" | "watching" | "all" = $state("planned");

	// Roulette state
	let loadingPool = $state(true);
	let totalPool = $state(0);
	let candidatePool: Media[] = $state([]);
	let spinning = $state(false);
	let displayedCandidate: Media | undefined = $state();
	let winner: Media | undefined = $state();
	let isMarkingWatching = $state(false);
	let hasMarkedWatching = $state(false);

	const genreOptions = [
		{ id: "", label: "Qualquer Género" },
		{ id: "ação", label: "💥 Ação" },
		{ id: "comédia", label: "😂 Comédia" },
		{ id: "ficção científica", label: "🚀 Sci-Fi" },
		{ id: "terror", label: "👻 Terror" },
		{ id: "drama", label: "🎭 Drama" },
		{ id: "animação", label: "🎨 Animação" },
		{ id: "suspense", label: "🔍 Suspense" },
		{ id: "romance", label: "❤️ Romance" },
		{ id: "fantasia", label: "🧙 Fantasia" },
	];

	const runtimeOptions = [
		{ value: 0, label: "Qualquer duração" },
		{ value: 90, label: "⚡ < 90 min" },
		{ value: 105, label: "🍿 < 1h45" },
		{ value: 120, label: "🎬 < 2h" },
	];

	const ratingOptions = [
		{ value: 0, label: "Qualquer nota" },
		{ value: 7, label: "⭐ 7.0+" },
		{ value: 8, label: "🔥 8.0+" },
	];

	async function fetchPool(triggerSpinOnLoad = false) {
		loadingPool = true;
		winner = undefined;
		hasMarkedWatching = false;
		try {
			const reqBody: RoulettePickRequest = {
				type: mediaType,
				status: statusFilter,
				maxRuntime: maxRuntime > 0 ? maxRuntime : undefined,
				minRating: minRating > 0 ? minRating : undefined,
				genre: selectedGenre.trim() ? selectedGenre.trim() : undefined,
				count: 10,
			};
			const res = await req.post<RoulettePickResponse>("/watched/pick", reqBody);
			if (res) {
				totalPool = res.totalPool;
				candidatePool = res.items || [];
				if (candidatePool.length > 0) {
					displayedCandidate = candidatePool[0];
				} else {
					displayedCandidate = undefined;
				}

				if (triggerSpinOnLoad && candidatePool.length > 0) {
					startSpin();
				}
			}
		} catch (err) {
			console.error("fetchPool error:", err);
			notify({ type: "error", text: "Erro ao consultar catálogo para a roleta." });
		} finally {
			loadingPool = false;
		}
	}

	function startSpin() {
		if (candidatePool.length === 0 || spinning) return;
		spinning = true;
		winner = undefined;
		hasMarkedWatching = false;

		// Slot machine reel deceleration animation
		const totalTicks = 22 + Math.floor(Math.random() * 8);
		let tick = 0;
		let delay = 50;

		function step() {
			tick++;
			const randIdx = Math.floor(Math.random() * candidatePool.length);
			displayedCandidate = candidatePool[randIdx];

			if (tick >= totalTicks) {
				// Finish spin, reveal final pick
				spinning = false;
				winner = displayedCandidate;
			} else {
				// Deceleration curve
				if (tick > totalTicks - 8) {
					delay += 40;
				} else if (tick > totalTicks - 14) {
					delay += 15;
				}
				setTimeout(step, delay);
			}
		}

		step();
	}

	function resetFilters() {
		mediaType = "movie";
		maxRuntime = 0;
		minRating = 0;
		selectedGenre = "";
		statusFilter = "planned";
		fetchPool();
	}

	function formatRuntime(mins?: number) {
		if (!mins || mins <= 0) return "";
		const h = Math.floor(mins / 60);
		const m = mins % 60;
		if (h > 0 && m > 0) return `${h}h ${m}m`;
		if (h > 0) return `${h}h`;
		return `${m}m`;
	}

	function getMediaUrl(item: Media): string {
		const type = item.type === "movie" ? "movie" : item.type === "tv" ? "tv" : "game";
		const id = item.ids?.tmdbId || item.ids?.igdbId || item.ids?.id;
		return resolve(`/${type}/${id}` as any);
	}

	async function markAsWatching() {
		if (!winner || isMarkingWatching) return;
		isMarkingWatching = true;
		try {
			if (winner.watched?.id) {
				await updateWatched(winner.watched, {
					contentId: winner.ids?.tmdbId || (winner.ids?.id as number),
					contentType: winner.type as any,
					status: "WATCHING",
				});
				hasMarkedWatching = true;
				notify({ type: "success", text: `"${winner.name}" marcado como a ver!` });
			} else {
				// No existing watched entry, add new
				await req.post("/watched", {
					contentId: winner.ids?.tmdbId || (winner.ids?.id as number),
					contentType: winner.type,
					status: "WATCHING",
				});
				hasMarkedWatching = true;
				notify({ type: "success", text: `"${winner.name}" adicionado e a ver!` });
			}
		} catch (err) {
			console.error("markAsWatching error:", err);
			notify({ type: "error", text: "Erro ao atualizar estado do título." });
		} finally {
			isMarkingWatching = false;
		}
	}

	onMount(() => {
		fetchPool();
	});
</script>

<Modal
	title="🎲 Roleta Inteligente — O Que Ver Hoje?"
	desc="Não sabes o que ver? Define os teus filtros de tempo e mood e deixa a roleta escolher o título ideal da tua Watchlist!"
	{onClose}
	maxWidth="900px"
>
	<div class="roulette-container">
		<!-- 1. Filters & Constraints Header -->
		<div class="filters-panel">
			<!-- Tipo de Media -->
			<div class="filter-group">
				<span class="filter-label">Tipo:</span>
				<div class="chip-bar">
					<button
						type="button"
						class="chip-btn"
						class:active={mediaType === "movie"}
						onclick={() => {
							mediaType = "movie";
							fetchPool();
						}}
					>
						🎬 Filmes
					</button>
					<button
						type="button"
						class="chip-btn"
						class:active={mediaType === "tv"}
						onclick={() => {
							mediaType = "tv";
							fetchPool();
						}}
					>
						📺 Séries
					</button>
					<button
						type="button"
						class="chip-btn"
						class:active={mediaType === "game"}
						onclick={() => {
							mediaType = "game";
							fetchPool();
						}}
					>
						🎮 Jogos
					</button>
					<button
						type="button"
						class="chip-btn"
						class:active={mediaType === "all"}
						onclick={() => {
							mediaType = "all";
							fetchPool();
						}}
					>
						🌐 Tudo
					</button>
				</div>
			</div>

			<!-- Tempo Livre (Duração Máxima) -->
			{#if mediaType === "movie" || mediaType === "all"}
				<div class="filter-group">
					<span class="filter-label">Tempo Disponível:</span>
					<div class="chip-bar">
						{#each runtimeOptions as opt}
							<button
								type="button"
								class="chip-btn"
								class:active={maxRuntime === opt.value}
								onclick={() => {
									maxRuntime = opt.value;
									fetchPool();
								}}
							>
								{opt.label}
							</button>
						{/each}
					</div>
				</div>
			{/if}

			<!-- Classificação Mínima -->
			<div class="filter-group">
				<span class="filter-label">Qualidade:</span>
				<div class="chip-bar">
					{#each ratingOptions as opt}
						<button
							type="button"
							class="chip-btn"
							class:active={minRating === opt.value}
							onclick={() => {
								minRating = opt.value;
								fetchPool();
							}}
						>
							{opt.label}
						</button>
					{/each}
				</div>
			</div>

			<!-- Mood / Género -->
			<div class="filter-group">
				<span class="filter-label">Mood / Género:</span>
				<div class="chip-bar genre-scroll">
					{#each genreOptions as g}
						<button
							type="button"
							class="chip-btn"
							class:active={selectedGenre === g.id}
							onclick={() => {
								selectedGenre = g.id;
								fetchPool();
							}}
						>
							{g.label}
						</button>
					{/each}
				</div>
			</div>

			<!-- Origem dos Dados (Watchlist status) -->
			<div class="filter-group">
				<span class="filter-label">Origem:</span>
				<div class="chip-bar">
					<button
						type="button"
						class="chip-btn"
						class:active={statusFilter === "planned"}
						onclick={() => {
							statusFilter = "planned";
							fetchPool();
						}}
					>
						📋 Na Minha Watchlist (Planeados)
					</button>
					<button
						type="button"
						class="chip-btn"
						class:active={statusFilter === "watching"}
						onclick={() => {
							statusFilter = "watching";
							fetchPool();
						}}
					>
						⏱️ Em Progresso
					</button>
					<button
						type="button"
						class="chip-btn"
						class:active={statusFilter === "all"}
						onclick={() => {
							statusFilter = "all";
							fetchPool();
						}}
					>
						📚 Toda a Biblioteca
					</button>
				</div>
			</div>
		</div>

		<!-- 2. Dynamic Pool Status Bar & Spin Trigger -->
		<div class="spin-action-bar">
			<div class="pool-stat">
				{#if loadingPool}
					<Spinner />
					<span>A calcular títulos elegíveis...</span>
				{:else if totalPool > 0}
					<span class="pool-badge">🎯 {totalPool} títulos elegíveis</span>
					<span class="pool-subtitle">Prontos a serem sorteados da tua biblioteca</span>
				{:else}
					<span class="pool-empty">⚠️ Nenhum título com estes filtros</span>
				{/if}
			</div>

			<div class="spin-btn-wrap">
				{#if totalPool > 0}
					<button
						type="button"
						class="btn-spin"
						disabled={spinning || loadingPool}
						onclick={startSpin}
					>
						<span class="dice-icon" class:shake={spinning}>🎲</span>
						<span>{spinning ? "A Sortear..." : winner ? "Girar Novamente" : "Girar Roleta!"}</span>
					</button>
				{:else}
					<button
						type="button"
						class="btn-reset"
						onclick={resetFilters}
					>
						<Icon i="close" wh={14} />
						<span>Limpar Filtros</span>
					</button>
				{/if}
			</div>
		</div>

		<!-- 3. Stage Display (Reel Animation or Revealed Winner) -->
		<div class="stage-section" class:has-winner={!!winner}>
			{#if spinning && displayedCandidate}
				<!-- Slot Machine Reel Effect -->
				<div class="slot-machine-reel">
					<div class="reel-card-shadow"></div>
					<div class="reel-card">
						<div class="reel-poster">
							<Poster
								media={displayedCandidate}
								fluidSize={true}
								disableInteraction={true}
								hideButtons={true}
								hideRating={true}
							/>
						</div>
						<div class="reel-title-overlay">
							<span class="reel-spinner-tag">A Procurar...</span>
							<h3>{displayedCandidate.name}</h3>
						</div>
					</div>
				</div>
			{:else if winner}
				<!-- Winner Presentation Card -->
				<div class="winner-card">
					<div class="winner-glow"></div>
					
					<div class="winner-poster-wrap">
						<Poster
							media={winner}
							fluidSize={true}
							disableInteraction={true}
							hideButtons={true}
						/>
					</div>

					<div class="winner-details">
						<div class="winner-header">
							<div class="winner-meta-pills">
								<span class="pill type-pill">
									{winner.type === "movie" ? "🎬 FILME" : winner.type === "tv" ? "📺 SÉRIE" : "🎮 JOGO"}
								</span>
								{#if winner.releaseDate}
									<span class="pill year-pill">
										{new Date(winner.releaseDate).getFullYear()}
									</span>
								{/if}
								{#if winner.rating}
									<span class="pill rating-pill">
										⭐ {(winner.rating / 10).toFixed(1)}/10
									</span>
								{/if}
								{#if winner.type === "movie" && winner.watched?.runtime}
									<span class="pill runtime-pill">
										⏱️ {formatRuntime(winner.watched.runtime)}
									</span>
								{/if}
							</div>
							
							<h2 class="winner-title">{winner.name}</h2>
						</div>

						{#if winner.summary}
							<p class="winner-summary">{winner.summary}</p>
						{/if}

						<div class="winner-actions">
							{#if !hasMarkedWatching}
								<button
									type="button"
									class="btn-action btn-watch"
									disabled={isMarkingWatching}
									onclick={markAsWatching}
								>
									{#if isMarkingWatching}
										<Spinner />
										<span>A Atualizar...</span>
									{:else}
										<Icon i="clock" wh={16} />
										<span>Começar a Ver Agora</span>
									{/if}
								</button>
							{:else}
								<div class="btn-action btn-watching-success">
									<Icon i="check" wh={16} />
									<span>A Ver Atualmente!</span>
								</div>
							{/if}

							<a
								href={getMediaUrl(winner)}
								class="btn-action btn-details"
								target="_blank"
							>
								<Icon i="eye" wh={16} />
								<span>Ver Detalhes</span>
							</a>

							<button
								type="button"
								class="btn-action btn-reroll"
								disabled={spinning}
								onclick={startSpin}
								title="Girar novamente com os mesmos filtros"
							>
								<span style="font-size: 15px;">🎲</span>
								<span>Outra Opção</span>
							</button>
						</div>
					</div>
				</div>
			{:else if candidatePool.length > 0 && !loadingPool}
				<!-- Ready to Spin State -->
				<div class="ready-hero">
					<div class="dice-preview">🎲</div>
					<h3>Tudo pronto para sortear!</h3>
					<p>Temos <strong>{totalPool}</strong> títulos à tua espera com estes critérios. Clica em <strong>Girar Roleta</strong> para escolher a tua sessão de hoje.</p>
					<button
						type="button"
						class="btn-spin-large"
						onclick={startSpin}
					>
						<span>Girar Roleta 🎲</span>
					</button>
				</div>
			{:else if !loadingPool}
				<!-- Empty Pool State -->
				<div class="empty-pool-hero">
					<span class="empty-icon">📭</span>
					<h3>Nenhum título corresponde aos teus filtros</h3>
					<p>Experimenta aumentar a duração máxima, remover o filtro de género ou incluir títulos de toda a biblioteca.</p>
					<button
						type="button"
						class="btn-reset-hero"
						onclick={resetFilters}
					>
						Limpar Filtros e Tentar Novamente
					</button>
				</div>
			{/if}
		</div>
	</div>
</Modal>

<style lang="scss">
	.roulette-container {
		display: flex;
		flex-direction: column;
		gap: 20px;
		width: 100%;
		max-height: 82vh;
		overflow-y: auto;
		padding-right: 4px;
	}

	.filters-panel {
		display: flex;
		flex-direction: column;
		gap: 12px;
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid rgba(255, 255, 255, 0.07);
		border-radius: 12px;
		padding: 16px;

		.filter-group {
			display: flex;
			align-items: center;
			gap: 12px;
			flex-wrap: wrap;

			.filter-label {
				font-size: 12px;
				font-weight: 700;
				text-transform: uppercase;
				letter-spacing: 0.5px;
				color: $text-muted;
				min-width: 110px;
			}

			.chip-bar {
				display: flex;
				align-items: center;
				gap: 8px;
				flex-wrap: wrap;

				&.genre-scroll {
					overflow-x: auto;
					padding-bottom: 4px;
					max-width: 100%;
				}

				.chip-btn {
					background: rgba(255, 255, 255, 0.06);
					border: 1px solid rgba(255, 255, 255, 0.1);
					color: $text-color;
					font-size: 13px;
					font-weight: 500;
					padding: 6px 12px;
					border-radius: 20px;
					cursor: pointer;
					white-space: nowrap;
					transition: all 0.2s ease;

					&:hover {
						background: rgba(255, 255, 255, 0.12);
						border-color: rgba(255, 255, 255, 0.2);
						transform: translateY(-1px);
					}

					&.active {
						background: $accent-color;
						color: #000;
						border-color: $accent-color;
						font-weight: 700;
						box-shadow: 0 0 12px rgba(255, 183, 3, 0.35);
					}
				}
			}
		}
	}

	.spin-action-bar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 14px 20px;
		background: rgba(255, 255, 255, 0.04);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 12px;
		flex-wrap: wrap;
		gap: 14px;

		.pool-stat {
			display: flex;
			flex-direction: column;
			gap: 3px;

			.pool-badge {
				font-size: 15px;
				font-weight: 700;
				color: $text-color;
			}

			.pool-subtitle {
				font-size: 12px;
				color: $text-muted;
			}

			.pool-empty {
				font-size: 14px;
				color: #ef4444;
				font-weight: 600;
			}
		}

		.btn-spin {
			display: flex;
			align-items: center;
			gap: 10px;
			background: linear-gradient(135deg, #ffb703 0%, #fb8500 100%);
			color: #000;
			font-size: 15px;
			font-weight: 800;
			padding: 10px 24px;
			border-radius: 30px;
			border: none;
			cursor: pointer;
			box-shadow: 0 4px 15px rgba(251, 133, 0, 0.35);
			transition: all 0.25s ease;

			&:hover:not(:disabled) {
				transform: translateY(-2px) scale(1.02);
				box-shadow: 0 6px 20px rgba(251, 133, 0, 0.5);
			}

			&:active:not(:disabled) {
				transform: translateY(0);
			}

			&:disabled {
				opacity: 0.6;
				cursor: not-allowed;
			}

			.dice-icon {
				font-size: 18px;

				&.shake {
					animation: diceShake 0.4s infinite ease-in-out;
				}
			}
		}

		.btn-reset {
			display: flex;
			align-items: center;
			gap: 6px;
			background: rgba(255, 255, 255, 0.08);
			border: 1px solid rgba(255, 255, 255, 0.15);
			color: $text-color;
			padding: 8px 16px;
			border-radius: 20px;
			font-size: 13px;
			font-weight: 600;
			cursor: pointer;
			transition: all 0.2s ease;

			&:hover {
				background: rgba(255, 255, 255, 0.15);
			}
		}
	}

	.stage-section {
		display: flex;
		justify-content: center;
		align-items: center;
		min-height: 330px;
		background: rgba(0, 0, 0, 0.25);
		border: 1px solid rgba(255, 255, 255, 0.06);
		border-radius: 16px;
		padding: 24px;
		position: relative;
		overflow: hidden;

		&.has-winner {
			padding: 20px;
		}
	}

	// Slot Machine Reel effect
	.slot-machine-reel {
		display: flex;
		flex-direction: column;
		align-items: center;
		position: relative;

		.reel-card {
			width: 220px;
			background: rgba(255, 255, 255, 0.05);
			border-radius: 14px;
			border: 2px solid $accent-color;
			box-shadow: 0 0 30px rgba(255, 183, 3, 0.35);
			overflow: hidden;
			animation: pulseGlow 1s infinite alternate ease-in-out;

			.reel-poster {
				width: 100%;
				height: 310px;
				overflow: hidden;
			}

			.reel-title-overlay {
				padding: 10px 14px;
				background: rgba(0, 0, 0, 0.85);
				text-align: center;

				.reel-spinner-tag {
					font-size: 11px;
					color: $accent-color;
					font-weight: 700;
					text-transform: uppercase;
					letter-spacing: 0.5px;
				}

				h3 {
					font-size: 14px;
					font-weight: 700;
					color: #fff;
					margin: 4px 0 0 0;
					white-space: nowrap;
					overflow: hidden;
					text-overflow: ellipsis;
				}
			}
		}
	}

	// Winner Card Presentation
	.winner-card {
		display: flex;
		gap: 28px;
		width: 100%;
		position: relative;
		align-items: flex-start;

		.winner-glow {
			position: absolute;
			top: -40px;
			left: -40px;
			right: -40px;
			bottom: -40px;
			background: radial-gradient(circle at 30% 40%, rgba(255, 183, 3, 0.15) 0%, transparent 70%);
			pointer-events: none;
			z-index: 0;
		}

		.winner-poster-wrap {
			flex-shrink: 0;
			width: 200px;
			border-radius: 12px;
			overflow: hidden;
			border: 2px solid rgba(255, 183, 3, 0.5);
			box-shadow: 0 10px 30px rgba(0, 0, 0, 0.6);
			position: relative;
			z-index: 1;
		}

		.winner-details {
			flex: 1;
			display: flex;
			flex-direction: column;
			gap: 16px;
			position: relative;
			z-index: 1;

			.winner-header {
				display: flex;
				flex-direction: column;
				gap: 8px;

				.winner-meta-pills {
					display: flex;
					gap: 8px;
					flex-wrap: wrap;

					.pill {
						font-size: 11px;
						font-weight: 700;
						padding: 4px 10px;
						border-radius: 12px;
						letter-spacing: 0.4px;

						&.type-pill {
							background: rgba(255, 183, 3, 0.15);
							color: $accent-color;
							border: 1px solid rgba(255, 183, 3, 0.3);
						}

						&.year-pill {
							background: rgba(255, 255, 255, 0.08);
							color: $text-color;
						}

						&.rating-pill {
							background: rgba(234, 179, 8, 0.15);
							color: #facc15;
						}

						&.runtime-pill {
							background: rgba(59, 130, 246, 0.15);
							color: #60a5fa;
						}
					}
				}

				.winner-title {
					font-size: 26px;
					font-weight: 800;
					color: #fff;
					margin: 0;
					line-height: 1.25;
				}
			}

			.winner-summary {
				font-size: 14px;
				color: $text-muted;
				line-height: 1.5;
				margin: 0;
				display: -webkit-box;
				-webkit-line-clamp: 4;
				-webkit-box-orient: vertical;
				overflow: hidden;
			}

			.winner-actions {
				display: flex;
				gap: 12px;
				flex-wrap: wrap;
				margin-top: 8px;

				.btn-action {
					display: flex;
					align-items: center;
					gap: 8px;
					padding: 10px 18px;
					border-radius: 8px;
					font-size: 14px;
					font-weight: 700;
					cursor: pointer;
					text-decoration: none;
					transition: all 0.2s ease;
					border: none;

					&.btn-watch {
						background: #10b981;
						color: #fff;

						&:hover:not(:disabled) {
							background: #059669;
							transform: translateY(-1px);
						}

						&:disabled {
							opacity: 0.7;
							cursor: not-allowed;
						}
					}

					&.btn-watching-success {
						background: rgba(16, 185, 129, 0.2);
						color: #34d399;
						border: 1px solid rgba(16, 185, 129, 0.4);
					}

					&.btn-details {
						background: rgba(255, 255, 255, 0.1);
						color: $text-color;

						&:hover {
							background: rgba(255, 255, 255, 0.18);
							transform: translateY(-1px);
						}
					}

					&.btn-reroll {
						background: rgba(255, 255, 255, 0.06);
						border: 1px solid rgba(255, 255, 255, 0.12);
						color: $text-color;

						&:hover {
							background: rgba(255, 255, 255, 0.12);
							transform: translateY(-1px);
						}
					}
				}
			}
		}
	}

	// Ready Hero state
	.ready-hero {
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
		gap: 12px;
		max-width: 480px;

		.dice-preview {
			font-size: 54px;
			filter: drop-shadow(0 0 20px rgba(255, 183, 3, 0.4));
		}

		h3 {
			font-size: 20px;
			font-weight: 700;
			color: #fff;
			margin: 0;
		}

		p {
			font-size: 14px;
			color: $text-muted;
			margin: 0;
			line-height: 1.5;
		}

		.btn-spin-large {
			margin-top: 10px;
			background: linear-gradient(135deg, #ffb703 0%, #fb8500 100%);
			color: #000;
			font-size: 16px;
			font-weight: 800;
			padding: 12px 36px;
			border-radius: 30px;
			border: none;
			cursor: pointer;
			box-shadow: 0 4px 20px rgba(251, 133, 0, 0.4);
			transition: all 0.25s ease;

			&:hover {
				transform: translateY(-2px) scale(1.03);
				box-shadow: 0 8px 25px rgba(251, 133, 0, 0.6);
			}
		}
	}

	// Empty Pool Hero state
	.empty-pool-hero {
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
		gap: 10px;
		max-width: 420px;

		.empty-icon {
			font-size: 42px;
		}

		h3 {
			font-size: 18px;
			font-weight: 700;
			color: #fff;
			margin: 0;
		}

		p {
			font-size: 13px;
			color: $text-muted;
			margin: 0;
			line-height: 1.4;
		}

		.btn-reset-hero {
			margin-top: 10px;
			background: rgba(255, 255, 255, 0.1);
			color: #fff;
			border: 1px solid rgba(255, 255, 255, 0.2);
			font-size: 13px;
			font-weight: 600;
			padding: 8px 20px;
			border-radius: 20px;
			cursor: pointer;
			transition: all 0.2s ease;

			&:hover {
				background: rgba(255, 255, 255, 0.18);
			}
		}
	}

	@keyframes diceShake {
		0%, 100% { transform: rotate(0deg) scale(1); }
		25% { transform: rotate(-15deg) scale(1.15); }
		75% { transform: rotate(15deg) scale(1.15); }
	}

	@keyframes pulseGlow {
		0% { box-shadow: 0 0 15px rgba(255, 183, 3, 0.2); }
		100% { box-shadow: 0 0 35px rgba(255, 183, 3, 0.6); }
	}

	@media (max-width: 650px) {
		.winner-card {
			flex-direction: column;
			align-items: center;
			text-align: center;

			.winner-poster-wrap {
				width: 160px;
			}

			.winner-details {
				align-items: center;

				.winner-actions {
					justify-content: center;
				}
			}
		}
	}
</style>
