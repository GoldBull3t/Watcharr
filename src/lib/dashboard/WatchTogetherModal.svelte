<script lang="ts">
	import Modal from "@/lib/Modal.svelte";
	import Icon from "@/lib/Icon.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import UserAvatar from "@/lib/img/UserAvatar.svelte";
	import { req } from "@/lib/util/api";
	import { notify } from "@/lib/util/notify";
	import type { WatchTogetherItem, WatchTogetherResponse, WatchTogetherUser } from "@/types";
	import { onMount } from "svelte";
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";

	interface Props {
		initialTargetUserId?: number;
		onClose: () => void;
	}

	let { initialTargetUserId = undefined, onClose }: Props = $props();

	let users: WatchTogetherUser[] = $state([]);
	let selectedUser: WatchTogetherUser | undefined = $state();
	let mode: "common" | "unwatched" | "recommended" = $state("common");
	let mediaFilter: "all" | "movie" | "tv" = $state("all");
	let searchQuery: string = $state("");

	let loadingUsers = $state(true);
	let loadingComparison = $state(false);
	let comparison: WatchTogetherResponse | undefined = $state();

	async function loadUsers() {
		try {
			loadingUsers = true;
			const res = await req.get<WatchTogetherUser[]>("/watched/together/users");
			if (res && res.length > 0) {
				users = res;
				if (initialTargetUserId) {
					selectedUser = users.find((u) => u.id === initialTargetUserId) || users[0];
				} else {
					selectedUser = users[0];
				}
				loadComparison();
			}
		} catch (err) {
			console.error("loadUsers error:", err);
			notify({ type: "error", text: "Falha ao carregar utilizadores da instância" });
		} finally {
			loadingUsers = false;
		}
	}

	async function loadComparison() {
		if (!selectedUser) return;
		try {
			loadingComparison = true;
			const res = await req.get<WatchTogetherResponse>(
				`/watched/together/${selectedUser.id}?mode=${mode}`,
			);
			comparison = res;
		} catch (err) {
			console.error("loadComparison error:", err);
			notify({ type: "error", text: "Erro ao comparar listas com o utilizador" });
		} finally {
			loadingComparison = false;
		}
	}

	function selectUser(user: WatchTogetherUser) {
		if (selectedUser?.id === user.id) return;
		selectedUser = user;
		loadComparison();
	}

	function selectMode(newMode: "common" | "unwatched" | "recommended") {
		if (mode === newMode) return;
		mode = newMode;
		loadComparison();
	}

	let filteredItems = $derived.by(() => {
		if (!comparison?.items) return [];
		return comparison.items.filter((item) => {
			if (!item.content) return false;
			if (mediaFilter !== "all" && item.content.type !== mediaFilter) {
				return false;
			}
			if (searchQuery.trim()) {
				const q = searchQuery.toLowerCase().trim();
				return item.content.title.toLowerCase().includes(q);
			}
			return true;
		});
	});

	function getMediaUrl(item: WatchTogetherItem): string {
		if (!item.content) return "#";
		const prefix = item.content.type === "movie" ? "movie" : "tv";
		return resolve(`/${prefix}/${item.content.tmdbId}` as any);
	}

	onMount(() => {
		loadUsers();
	});
</script>

<Modal
	title="🍿 Ver em Conjunto (Watch Together)"
	desc="Cruza a tua Watchlist com amigos ou parceiros para escolher o filme ou série ideal sem discussão."
	{onClose}
	maxWidth="1100px"
>
	<div class="watch-together-container">
		<!-- 1. Seletor de Parceiro / Amigo -->
		<div class="user-selector-section">
			<span class="section-label">Escolhe com quem queres ver:</span>
			{#if loadingUsers}
				<div class="loading-users">
					<Spinner />
					<span>A carregar utilizadores...</span>
				</div>
			{:else if users.length === 0}
				<div class="no-users-msg">
					Nenhum outro utilizador público encontrado na instância. Segue outros utilizadores para poderes cruzar as vossas listas!
				</div>
			{:else}
				<div class="users-pill-bar">
					{#each users as u (u.id)}
						<button
							class="user-pill"
							class:active={selectedUser?.id === u.id}
							onclick={() => selectUser(u)}
						>
							<div class="avatar-wrap">
								<UserAvatar img={u.avatar} />
							</div>
							<span class="user-name">{u.username}</span>
							{#if u.isFollowed}
								<span class="follow-badge">A Seguir</span>
							{/if}
						</button>
					{/each}
				</div>
			{/if}
		</div>

		<!-- 2. Modos de Cruzamento & Filtros -->
		{#if selectedUser}
			<div class="modes-toolbar">
				<div class="segmented-modes">
					<button
						class="mode-btn"
						class:active={mode === "common"}
						onclick={() => selectMode("common")}
					>
						<span class="icon">🎯</span>
						<span>Ambos Queremos Ver</span>
						{#if comparison?.mode === "common" && !loadingComparison}
							<span class="count-tag">{comparison.totalCount}</span>
						{/if}
					</button>

					<button
						class="mode-btn"
						class:active={mode === "unwatched"}
						onclick={() => selectMode("unwatched")}
					>
						<span class="icon">🍿</span>
						<span>Na Minha Lista (Ele/Ela Não Viu)</span>
						{#if comparison?.mode === "unwatched" && !loadingComparison}
							<span class="count-tag">{comparison.totalCount}</span>
						{/if}
					</button>

					<button
						class="mode-btn"
						class:active={mode === "recommended"}
						onclick={() => selectMode("recommended")}
					>
						<span class="icon">⭐</span>
						<span>Recomendações Dele/Dela (Nota &ge; 8)</span>
						{#if comparison?.mode === "recommended" && !loadingComparison}
							<span class="count-tag">{comparison.totalCount}</span>
						{/if}
					</button>
				</div>

				<div class="filters-row">
					<div class="media-type-pills">
						<button
							class="pill-btn"
							class:active={mediaFilter === "all"}
							onclick={() => (mediaFilter = "all")}
						>
							Todos
						</button>
						<button
							class="pill-btn"
							class:active={mediaFilter === "movie"}
							onclick={() => (mediaFilter = "movie")}
						>
							Filmes
						</button>
						<button
							class="pill-btn"
							class:active={mediaFilter === "tv"}
							onclick={() => (mediaFilter = "tv")}
						>
							Séries
						</button>
					</div>

					<div class="search-box">
						<Icon i="search" wh={14} />
						<input
							type="text"
							placeholder="Filtrar por título..."
							bind:value={searchQuery}
						/>
						{#if searchQuery}
							<button class="clear-search" onclick={() => (searchQuery = "")}>
								<Icon i="close" wh={12} />
							</button>
						{/if}
					</div>
				</div>
			</div>

			<!-- 3. Grelha de Filmes e Séries Cruzados -->
			<div class="results-area">
				{#if loadingComparison}
					<div class="loading-state">
						<Spinner />
						<p>A cruzar listas com @{selectedUser.username}...</p>
					</div>
				{:else if filteredItems.length === 0}
					<div class="empty-state">
						<div class="empty-icon">
							{#if mode === "common"}
								🎯
							{:else if mode === "unwatched"}
								🍿
							{:else}
								⭐
							{/if}
						</div>
						<h4>Nenhum resultado encontrado</h4>
						<p>
							{#if mode === "common"}
								Tu e o(a) @{selectedUser.username} ainda não têm filmes ou séries em comum marcados como "Por Ver" na Watchlist.
							{:else if mode === "unwatched"}
								Nenhum item da tua watchlist pendente que o(a) @{selectedUser.username} ainda não tenha começado ou visto.
							{:else}
								O(A) @{selectedUser.username} ainda não tem avaliações $\ge 8$ de obras que tu não tenhas visto.
							{/if}
						</p>
					</div>
				{:else}
					<div class="posters-grid">
						{#each filteredItems as item (item.contentId)}
							{@const content = item.content}
							{#if content}
								{@const pPath = (content as any).posterPath || (content as any).poster_path}
								{@const rDate = (content as any).releaseDate || (content as any).release_date}
								<a
									class="matched-card"
									href={getMediaUrl(item)}
									onclick={() => onClose()}
								>
									<div class="poster-container">
										<img
											src={pPath
												? pPath.startsWith("http")
													? pPath
													: `https://image.tmdb.org/t/p/w500${pPath}`
												: "https://image.tmdb.org/t/p/w500/wwemzKWzjKYJFfCeiB57q3r4Bcm.png"}
											alt={content.title}
											loading="lazy"
										/>
										<div class="card-overlay-badge">
											{#if mode === "common"}
												<span class="badge common-badge">🎯 Match Watchlist</span>
											{:else if mode === "unwatched"}
												<span class="badge unwatched-badge">🍿 Por Descobrir</span>
											{:else if mode === "recommended"}
												<span class="badge recommended-badge">★ {item.targetUserRating} ({selectedUser.username})</span>
											{/if}
										</div>

										<div class="type-badge">
											{content.type === "movie" ? "Filme" : "Série"}
										</div>
									</div>

									<div class="card-meta">
										<span class="card-title" title={content.title}>{content.title}</span>
										{#if rDate}
											<span class="card-year">
												{new Date(rDate).getFullYear()}
											</span>
										{/if}
									</div>
								</a>
							{/if}
						{/each}
					</div>
				{/if}
			</div>
		{/if}
	</div>
</Modal>

<style lang="scss">
	.watch-together-container {
		display: flex;
		flex-direction: column;
		gap: 20px;
		width: 100%;
		max-height: 80vh;
	}

	.user-selector-section {
		display: flex;
		flex-direction: column;
		gap: 10px;

		.section-label {
			font-size: 13px;
			font-weight: 700;
			color: $text-color;
			text-transform: uppercase;
			letter-spacing: 0.5px;
		}

		.loading-users {
			display: flex;
			align-items: center;
			gap: 10px;
			color: $text-muted;
			font-size: 14px;
		}

		.no-users-msg {
			padding: 12px;
			background: rgba(255, 255, 255, 0.05);
			border-radius: 8px;
			font-size: 13px;
			color: $text-muted;
		}

		.users-pill-bar {
			display: flex;
			gap: 10px;
			overflow-x: auto;
			padding-bottom: 6px;
			scroll-snap-type: x mandatory;

			.user-pill {
				display: flex;
				align-items: center;
				gap: 8px;
				padding: 6px 14px 6px 8px;
				border-radius: 30px;
				background: rgba(255, 255, 255, 0.06);
				border: 1px solid rgba(255, 255, 255, 0.1);
				cursor: pointer;
				white-space: nowrap;
				transition: all 0.2s ease;
				color: $text-color;

				.avatar-wrap {
					width: 28px;
					height: 28px;
					border-radius: 50%;
					overflow: hidden;
					display: flex;
					align-items: center;
					justify-content: center;
					background: #27272a;
				}

				.user-name {
					font-weight: 600;
					font-size: 14px;
				}

				.follow-badge {
					font-size: 10px;
					font-weight: 700;
					background: rgba(14, 165, 233, 0.2);
					color: #38bdf8;
					padding: 2px 6px;
					border-radius: 10px;
				}

				&:hover {
					background: rgba(255, 255, 255, 0.12);
					border-color: rgba(255, 255, 255, 0.25);
				}

				&.active {
					background: #e50914;
					border-color: #e50914;
					color: white;

					.follow-badge {
						background: rgba(0, 0, 0, 0.3);
						color: white;
					}
				}
			}
		}
	}

	.modes-toolbar {
		display: flex;
		flex-direction: column;
		gap: 12px;
		background: rgba(0, 0, 0, 0.25);
		padding: 12px;
		border-radius: 12px;
		border: 1px solid rgba(255, 255, 255, 0.06);

		.segmented-modes {
			display: flex;
			gap: 8px;
			flex-wrap: wrap;

			.mode-btn {
				display: flex;
				align-items: center;
				gap: 8px;
				padding: 8px 16px;
				border-radius: 8px;
				font-size: 13px;
				font-weight: 700;
				background: rgba(255, 255, 255, 0.05);
				color: $text-muted;
				border: 1px solid transparent;
				cursor: pointer;
				transition: all 0.2s ease;

				.count-tag {
					background: rgba(255, 255, 255, 0.15);
					color: white;
					padding: 1px 7px;
					border-radius: 10px;
					font-size: 11px;
				}

				&:hover {
					color: $text-color;
					background: rgba(255, 255, 255, 0.08);
				}

				&.active {
					background: rgba(255, 255, 255, 0.15);
					color: white;
					border-color: rgba(255, 255, 255, 0.25);

					.count-tag {
						background: #e50914;
					}
				}
			}
		}

		.filters-row {
			display: flex;
			align-items: center;
			justify-content: space-between;
			gap: 12px;
			flex-wrap: wrap;

			.media-type-pills {
				display: flex;
				gap: 6px;

				.pill-btn {
					padding: 4px 12px;
					border-radius: 6px;
					font-size: 12px;
					font-weight: 600;
					background: transparent;
					color: $text-muted;
					border: 1px solid rgba(255, 255, 255, 0.1);
					cursor: pointer;

					&.active {
						background: rgba(255, 255, 255, 0.2);
						color: white;
						border-color: rgba(255, 255, 255, 0.3);
					}
				}
			}

			.search-box {
				display: flex;
				align-items: center;
				gap: 8px;
				background: rgba(255, 255, 255, 0.06);
				border: 1px solid rgba(255, 255, 255, 0.1);
				border-radius: 6px;
				padding: 4px 10px;
				min-width: 220px;

				input {
					background: transparent;
					border: none;
					outline: none;
					color: white;
					font-size: 13px;
					width: 100%;

					&::placeholder {
						color: $text-muted;
					}
				}

				.clear-search {
					background: transparent;
					border: none;
					color: $text-muted;
					cursor: pointer;
					display: flex;
					align-items: center;
					padding: 0;

					&:hover {
						color: white;
					}
				}
			}
		}
	}

	.results-area {
		overflow-y: auto;
		max-height: 52vh;
		padding-right: 4px;

		.loading-state {
			display: flex;
			flex-direction: column;
			align-items: center;
			justify-content: center;
			gap: 12px;
			padding: 60px 0;
			color: $text-muted;
		}

		.empty-state {
			display: flex;
			flex-direction: column;
			align-items: center;
			justify-content: center;
			text-align: center;
			padding: 50px 20px;
			gap: 8px;

			.empty-icon {
				font-size: 40px;
				margin-bottom: 6px;
			}

			h4 {
				margin: 0;
				font-size: 18px;
				font-weight: 700;
				color: white;
			}

			p {
				margin: 0;
				font-size: 14px;
				color: $text-muted;
				max-width: 480px;
				line-height: 1.5;
			}
		}

		.posters-grid {
			display: grid;
			grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
			gap: 14px;

			.matched-card {
				display: flex;
				flex-direction: column;
				gap: 6px;
				text-decoration: none;
				color: inherit;
				transition: transform 0.2s ease;

				&:hover {
					transform: translateY(-4px);

					.poster-container {
						box-shadow: 0 10px 24px rgba(0, 0, 0, 0.7);
						border-color: rgba(255, 255, 255, 0.3);
					}
				}

				.poster-container {
					position: relative;
					width: 100%;
					aspect-ratio: 2 / 3;
					border-radius: 8px;
					overflow: hidden;
					background: #18181b;
					border: 1px solid rgba(255, 255, 255, 0.08);
					box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
					transition: all 0.2s ease;

					img {
						width: 100%;
						height: 100%;
						object-fit: cover;
					}

					.card-overlay-badge {
						position: absolute;
						bottom: 0;
						left: 0;
						right: 0;
						background: linear-gradient(to top, rgba(0, 0, 0, 0.95) 20%, transparent);
						padding: 8px 6px 6px 6px;

						.badge {
							display: inline-block;
							font-size: 9px;
							font-weight: 800;
							padding: 2px 5px;
							border-radius: 4px;
							text-transform: uppercase;
							letter-spacing: 0.3px;

							&.common-badge {
								background: #e50914;
								color: white;
							}

							&.unwatched-badge {
								background: #0284c7;
								color: white;
							}

							&.recommended-badge {
								background: #eab308;
								color: #000;
							}
						}
					}

					.type-badge {
						position: absolute;
						top: 6px;
						right: 6px;
						background: rgba(0, 0, 0, 0.75);
						color: #e4e4e7;
						font-size: 9px;
						font-weight: 700;
						padding: 2px 6px;
						border-radius: 4px;
						backdrop-filter: blur(4px);
					}
				}

				.card-meta {
					display: flex;
					flex-direction: column;
					gap: 2px;

					.card-title {
						font-size: 13px;
						font-weight: 700;
						color: #f4f4f5;
						white-space: nowrap;
						overflow: hidden;
						text-overflow: ellipsis;
					}

					.card-year {
						font-size: 11px;
						color: $text-muted;
					}
				}
			}
		}
	}
</style>
