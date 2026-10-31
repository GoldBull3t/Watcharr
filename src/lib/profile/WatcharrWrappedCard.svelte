<script lang="ts">
	import type { Profile } from "@/types";
	import Icon from "@/lib/Icon.svelte";

	interface Props {
		profile: Profile;
	}

	let { profile }: Props = $props();

	// Convert total minutes into Days, Hours, Minutes
	let runtimeDays = $derived(
		Math.floor((profile.totalRuntime || 0) / (24 * 60)),
	);
	let runtimeHours = $derived(
		Math.floor(((profile.totalRuntime || 0) % (24 * 60)) / 60),
	);

	let topGenre = $derived.by(() => {
		if (profile.genreDistribution && profile.genreDistribution.length > 0) {
			return profile.genreDistribution[0];
		}
		return undefined;
	});

	let topNetwork = $derived.by(() => {
		if (profile.topNetworks && profile.topNetworks.length > 0) {
			return profile.topNetworks[0];
		}
		return undefined;
	});

	let totalGamingHours = $derived(profile.totalGamingHours || 0);

	// Dynamic Gamification Badges
	let badges = $derived.by(() => {
		const list: {
			id: string;
			icon: string;
			title: string;
			desc: string;
			unlocked: boolean;
		}[] = [];

		list.push({
			id: "cinephile",
			icon: "🍿",
			title: "Cinéfilo Ativo",
			desc: `${profile.moviesWatched}/25 filmes vistos`,
			unlocked: profile.moviesWatched >= 25,
		});

		list.push({
			id: "binger",
			icon: "📺",
			title: "Mestre das Maratonas",
			desc: `${profile.episodesWatched}/50 episódios vistos`,
			unlocked: profile.episodesWatched >= 50,
		});

		list.push({
			id: "gamer",
			icon: "🎮",
			title: "Gamer Consagrado",
			desc: `${totalGamingHours.toFixed(0)}/20 horas jogadas`,
			unlocked: totalGamingHours >= 20 || profile.gamesPlayed >= 5,
		});

		list.push({
			id: "critic",
			icon: "⭐",
			title: "Crítico Exigente",
			desc: `${profile.ratingStats?.count || 0}/15 avaliações dadas`,
			unlocked: (profile.ratingStats?.count || 0) >= 15,
		});

		list.push({
			id: "collector",
			icon: "🏆",
			title: "Grande Colecionador",
			desc: "Biblioteca com +50 títulos",
			unlocked:
				profile.moviesWatched + profile.showsWatched + profile.gamesPlayed >=
				50,
		});

		return list;
	});
</script>

<div class="wrapped-showcase-card">
	<div class="wrapped-header">
		<div class="title-with-badge">
			<span class="sparkle-badge">✨</span>
			<div>
				<h3 class="wrapped-title">Watcharr Wrapped & Milestones</h3>
				<span class="wrapped-subtitle"
					>A tua jornada cinéfila e gamer em números e conquistas</span
				>
			</div>
		</div>

		{#if topGenre}
			<div
				class="favorite-genre-chip"
				title="O teu género cinematográfico predileto"
			>
				<span class="chip-label">Género Favorito:</span>
				<strong class="chip-val"
					>👑 {topGenre.name || (topGenre as any).Name}</strong
				>
			</div>
		{/if}
	</div>

	<!-- Highlights Matrix -->
	<div class="metrics-matrix">
		<!-- Total Watch Time -->
		<div class="matrix-card">
			<div class="matrix-icon-wrap time-icon">
				<Icon i="clock" wh={20} />
			</div>
			<div class="matrix-data">
				<span class="matrix-val">
					<strong>{runtimeDays}</strong>d <strong>{runtimeHours}</strong>h
				</span>
				<span class="matrix-label">Tempo em Filmes e Séries</span>
			</div>
		</div>

		<!-- Gaming Hours -->
		<div class="matrix-card">
			<div class="matrix-icon-wrap game-icon">
				<span style="font-size: 20px;">🎮</span>
			</div>
			<div class="matrix-data">
				<span class="matrix-val">
					<strong>{totalGamingHours.toFixed(1)}</strong>h
				</span>
				<span class="matrix-label">{profile.gamesPlayed} jogos terminados</span>
			</div>
		</div>

		<!-- Completed Media -->
		<div class="matrix-card">
			<div class="matrix-icon-wrap media-icon">
				<Icon i="film" wh={20} />
			</div>
			<div class="matrix-data">
				<span class="matrix-val">
					<strong>{profile.moviesWatched}</strong> filmes •
					<strong>{profile.episodesWatched}</strong> eps
				</span>
				<span class="matrix-label"
					>{profile.showsWatched} séries concluídas</span
				>
			</div>
		</div>

		<!-- Average Rating -->
		<div class="matrix-card">
			<div class="matrix-icon-wrap rating-icon">
				<Icon i="star" wh={20} />
			</div>
			<div class="matrix-data">
				<span class="matrix-val">
					<strong
						>{profile.ratingStats?.average
							? profile.ratingStats.average.toFixed(1)
							: "-"}</strong
					> ★
				</span>
				<span class="matrix-label"
					>Nota Média ({profile.ratingStats?.count || 0} avaliações)</span
				>
			</div>
		</div>
	</div>

	<!-- Badges & Achievements Bar -->
	<div class="badges-shelf">
		<h5 class="shelf-title">Conquistas Desbloqueadas:</h5>
		<div class="badges-row">
			{#each badges as badge (badge.id)}
				<div
					class="achievement-pill"
					class:unlocked={badge.unlocked}
					title={badge.desc}
				>
					<span class="badge-emoji">{badge.icon}</span>
					<div class="badge-text">
						<strong class="badge-name">{badge.title}</strong>
						<span class="badge-sub">{badge.desc}</span>
					</div>
					{#if badge.unlocked}
						<span class="unlocked-check">✓</span>
					{/if}
				</div>
			{/each}
		</div>
	</div>
</div>

<style lang="scss">
	.wrapped-showcase-card {
		width: 100%;
		border-radius: 12px;
		background: rgba(22, 27, 38, 0.75);
		border: 1px solid rgba(255, 255, 255, 0.08);
		box-shadow: 0 8px 30px -4px rgba(0, 0, 0, 0.5);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		padding: 24px;
		margin-bottom: 24px;
		box-sizing: border-box;

		@media screen and (max-width: 768px) {
			padding: 18px 16px;
			border-radius: 12px;
		}

		.wrapped-header {
			display: flex;
			align-items: center;
			justify-content: space-between;
			flex-wrap: wrap;
			gap: 14px;
			padding-bottom: 18px;
			border-bottom: 1px solid rgba(255, 255, 255, 0.08);

			.title-with-badge {
				display: flex;
				align-items: center;
				gap: 12px;

				.sparkle-badge {
					font-size: 24px;
					display: flex;
					align-items: center;
					justify-content: center;
					width: 40px;
					height: 40px;
					border-radius: 8px;
					background: rgba(245, 158, 11, 0.12);
					border: 1px solid rgba(245, 158, 11, 0.25);
				}

				.wrapped-title {
					margin: 0;
					font-size: 19px;
					font-weight: 700;
					color: #ffffff;
					letter-spacing: -0.01em;
				}

				.wrapped-subtitle {
					font-size: 12px;
					color: rgba(255, 255, 255, 0.6);
				}
			}

			.favorite-genre-chip {
				display: inline-flex;
				align-items: center;
				gap: 6px;
				padding: 6px 14px;
				border-radius: 9999px;
				background: rgba(245, 158, 11, 0.14);
				border: 1px solid rgba(245, 158, 11, 0.35);

				.chip-label {
					font-size: 11px;
					color: rgba(255, 255, 255, 0.65);
					text-transform: uppercase;
					letter-spacing: 0.04em;
				}

				.chip-val {
					font-size: 12.5px;
					color: #fbbf24;
				}
			}
		}

		.metrics-matrix {
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(210px, 1fr));
			gap: 14px;
			margin: 20px 0;

			.matrix-card {
				display: flex;
				align-items: center;
				gap: 14px;
				padding: 14px 16px;
				border-radius: 14px;
				background: rgba(0, 0, 0, 0.28);
				border: 1px solid rgba(255, 255, 255, 0.06);
				transition:
					transform 140ms ease,
					border-color 140ms ease;

				&:hover {
					border-color: rgba(255, 255, 255, 0.15);
					transform: translateY(-1px);
				}

				.matrix-icon-wrap {
					display: flex;
					align-items: center;
					justify-content: center;
					width: 40px;
					height: 40px;
					border-radius: 10px;
					flex-shrink: 0;

					&.time-icon {
						background: rgba(99, 102, 241, 0.16);
						color: #818cf8;
					}

					&.game-icon {
						background: rgba(245, 158, 11, 0.16);
						color: #fbbf24;
					}

					&.media-icon {
						background: rgba(16, 185, 129, 0.16);
						color: #34d399;
					}

					&.rating-icon {
						background: rgba(236, 72, 153, 0.16);
						color: #f472b6;
					}
				}

				.matrix-data {
					display: flex;
					flex-direction: column;
					gap: 2px;
					overflow: hidden;

					.matrix-val {
						font-size: 15px;
						color: #ffffff;
						white-space: nowrap;
						text-overflow: ellipsis;
						overflow: hidden;

						strong {
							font-size: 18px;
							font-weight: 700;
							color: #fbbf24;
						}
					}

					.matrix-label {
						font-size: 11px;
						color: rgba(255, 255, 255, 0.55);
					}
				}
			}
		}

		.badges-shelf {
			display: flex;
			flex-direction: column;
			gap: 10px;
			padding-top: 16px;
			border-top: 1px solid rgba(255, 255, 255, 0.08);

			.shelf-title {
				margin: 0;
				font-size: 11.5px;
				font-weight: 700;
				color: rgba(255, 255, 255, 0.6);
				text-transform: uppercase;
				letter-spacing: 0.05em;
			}

			.badges-row {
				display: flex;
				flex-wrap: wrap;
				gap: 10px;

				.achievement-pill {
					display: flex;
					align-items: center;
					gap: 8px;
					padding: 6px 12px;
					border-radius: 10px;
					background: rgba(255, 255, 255, 0.035);
					border: 1px solid rgba(255, 255, 255, 0.08);
					opacity: 0.5;
					filter: grayscale(0.8);
					transition: all 140ms ease;

					&.unlocked {
						opacity: 1;
						filter: none;
						background: rgba(245, 158, 11, 0.1);
						border-color: rgba(245, 158, 11, 0.35);

						.badge-name {
							color: #fbbf24;
						}
					}

					.badge-emoji {
						font-size: 18px;
						flex-shrink: 0;
					}

					.badge-text {
						display: flex;
						flex-direction: column;

						.badge-name {
							font-size: 12px;
							font-weight: 600;
							color: #ffffff;
						}

						.badge-sub {
							font-size: 9.5px;
							color: rgba(255, 255, 255, 0.5);
						}
					}

					.unlocked-check {
						font-size: 11px;
						font-weight: 800;
						color: #10b981;
						margin-left: 2px;
					}
				}
			}
		}
	}
</style>
