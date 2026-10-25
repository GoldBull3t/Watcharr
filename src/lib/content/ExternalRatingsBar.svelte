<script lang="ts">
	import type { Media } from "@/types";

	interface Props {
		media: Media;
	}

	let { media }: Props = $props();

	let ext = $derived(media.externalRatings);
	let imdbUrl = $derived(
		media.ids?.imdb ? `https://www.imdb.com/title/${media.ids.imdb}/` : undefined,
	);
	let malUrl = $derived(
		ext?.malId ? `https://myanimelist.net/anime/${ext.malId}` : undefined,
	);

	function formatCurrency(val?: number): string {
		if (!val || val <= 0) return "";
		if (val >= 1_000_000_000) {
			return `$${(val / 1_000_000_000).toFixed(2)}B`;
		}
		if (val >= 1_000_000) {
			return `$${(val / 1_000_000).toFixed(1)}M`;
		}
		return `$${val.toLocaleString()}`;
	}

	function formatMinutesToHours(mins?: number): string {
		if (!mins || mins <= 0) return "";
		const h = Math.floor(mins / 60);
		const m = mins % 60;
		if (h > 0 && m > 0) return `${h}h ${m}m`;
		if (h > 0) return `${h}h`;
		return `${m}m`;
	}

	let hasAnyScore = $derived(
		Boolean(
			ext?.imdbRating ||
				ext?.rottenTomatoes ||
				ext?.rottenTomatoesAudience ||
				ext?.metacritic ||
				ext?.letterboxd ||
				ext?.trakt ||
				ext?.tvmaze ||
				ext?.anilist ||
				media.criticRating ||
				media.ids?.imdb,
		),
	);

	let hasMetadataChips = $derived(
		Boolean(
			(media.networks && media.networks.length > 0) ||
				(media.developers && media.developers.length > 0) ||
				(media.publishers && media.publishers.length > 0) ||
				(media.platforms && media.platforms.length > 0) ||
				media.timeToBeat?.normally ||
				media.timeToBeat?.completely ||
				ext?.awards ||
				ext?.boxOffice ||
				media.budget ||
				media.revenue ||
				ext?.animationStudio ||
				ext?.networkSchedule,
		),
	);
</script>

{#if media.tagline}
	<p class="media-tagline">“{media.tagline}”</p>
{/if}

{#if media.originalTitle && media.originalTitle !== media.name}
	<div class="original-title-sub">
		<span class="orig-label">Título Original:</span>
		<span class="orig-val">{media.originalTitle}</span>
		{#if media.originalLanguage}
			<span class="lang-code">{media.originalLanguage.toUpperCase()}</span>
		{/if}
	</div>
{/if}

{#if hasAnyScore || hasMetadataChips}
	<div class="enrichment-panel">
		{#if hasAnyScore}
			<div class="scores-row">
				{#if ext?.imdbRating}
					<a
						class="score-badge imdb"
						href={imdbUrl}
						target="_blank"
						rel="noreferrer noopener external"
						title={ext.imdbVotes ? `IMDb: ${ext.imdbRating}/10 (${ext.imdbVotes} votos)` : `IMDb: ${ext.imdbRating}/10`}
					>
						<span class="brand">IMDb</span>
						<span class="val">★ {ext.imdbRating}</span>
					</a>
				{:else if imdbUrl}
					<a
						class="score-badge imdb-link"
						href={imdbUrl}
						target="_blank"
						rel="noreferrer noopener external"
						title="Abrir página oficial no IMDb"
					>
						<span class="brand">IMDb</span>
						<span class="val">↗</span>
					</a>
				{/if}

				{#if ext?.rottenTomatoes}
					<span class="score-badge rt" title="Rotten Tomatoes (Tomatometer / Crítica)">
						<span class="brand">🍅 RT</span>
						<span class="val">{ext.rottenTomatoes}</span>
					</span>
				{/if}

				{#if ext?.rottenTomatoesAudience}
					<span class="score-badge rt-aud" title="Rotten Tomatoes (Popcornmeter / Audiência)">
						<span class="brand">🍿 Aud</span>
						<span class="val">{ext.rottenTomatoesAudience}</span>
					</span>
				{/if}

				{#if ext?.metacritic}
					<span class="score-badge metacritic" title="Metacritic Metascore">
						<span class="brand">M</span>
						<span class="val">{ext.metacritic}</span>
					</span>
				{/if}

				{#if media.criticRating && media.criticRating > 0}
					<span
						class="score-badge critic"
						title={`Nota da Crítica IGDB (${media.criticRatingCount || 0} análises)`}
					>
						<span class="brand">Crítica</span>
						<span class="val">{media.criticRating}/100</span>
					</span>
				{/if}

				{#if ext?.letterboxd}
					<span class="score-badge letterboxd" title="Letterboxd Score">
						<span class="brand">LB</span>
						<span class="val">★ {ext.letterboxd}</span>
					</span>
				{/if}

				{#if ext?.trakt}
					<span class="score-badge trakt" title="Trakt Community Score">
						<span class="brand">Trakt</span>
						<span class="val">{ext.trakt}</span>
					</span>
				{/if}

				{#if ext?.tvmaze}
					<span class="score-badge tvmaze" title="TVMaze Community Rating">
						<span class="brand">TVMaze</span>
						<span class="val">★ {ext.tvmaze}</span>
					</span>
				{/if}

				{#if ext?.anilist}
					<a
						class="score-badge anilist"
						href={malUrl}
						target={malUrl ? "_blank" : undefined}
						rel="noreferrer noopener external"
						title={malUrl ? "Ver no MyAnimeList / AniList" : "AniList Mean Score"}
					>
						<span class="brand">AniList</span>
						<span class="val">{ext.anilist}</span>
					</a>
				{/if}
			</div>
		{/if}

		{#if hasMetadataChips}
			<div class="meta-chips-row">
				{#if media.networks && media.networks.length > 0}
					{#each media.networks.slice(0, 4) as net (net.name)}
						<span class="meta-chip network" title="Canal / Rede Original">
							{#if net.logoPath}
								<img
									src={`https://image.tmdb.org/t/p/w92${net.logoPath}`}
									alt={net.name}
									class="net-logo"
								/>
							{/if}
							<span>{net.name}</span>
						</span>
					{/each}
				{/if}

				{#if ext?.animationStudio}
					<span class="meta-chip studio" title="Estúdio de Animação">
						<span>🎨 {ext.animationStudio}</span>
					</span>
				{/if}

				{#if ext?.sourceMaterial}
					<span class="meta-chip source" title="Material de Origem">
						<span>📖 {ext.sourceMaterial}</span>
					</span>
				{/if}

				{#if ext?.networkSchedule}
					<span class="meta-chip schedule" title="Horário de Transmissão">
						<span>🗓️ {ext.networkSchedule}</span>
					</span>
				{/if}

				{#if media.developers && media.developers.length > 0}
					<span class="meta-chip dev" title="Estúdio / Developer">
						<span>🛠️ {media.developers.slice(0, 2).join(", ")}</span>
					</span>
				{/if}

				{#if media.publishers && media.publishers.length > 0}
					<span class="meta-chip pub" title="Editora / Publisher">
						<span>🏢 {media.publishers.slice(0, 2).join(", ")}</span>
					</span>
				{/if}

				{#if media.timeToBeat && (media.timeToBeat.normally || media.timeToBeat.completely)}
					{#if media.timeToBeat.normally}
						<span class="meta-chip ttb" title="Tempo médio para terminar a história principal">
							<span>⏱️ História: {formatMinutesToHours(media.timeToBeat.normally)}</span>
						</span>
					{/if}
					{#if media.timeToBeat.completely}
						<span class="meta-chip ttb-100" title="Tempo médio para 100% (Completionist)">
							<span>🏆 100%: {formatMinutesToHours(media.timeToBeat.completely)}</span>
						</span>
					{/if}
				{/if}

				{#if media.platforms && media.platforms.length > 0}
					<span class="meta-chip platforms" title={media.platforms.join(", ")}>
						<span>🎮 {media.platforms.slice(0, 5).join(" • ")}{media.platforms.length > 5 ? ` +${media.platforms.length - 5}` : ""}</span>
					</span>
				{/if}

				{#if ext?.boxOffice}
					<span class="meta-chip boxoffice" title="Bilheteira (Box Office)">
						<span>🎟️ Bilheteira: {ext.boxOffice}</span>
					</span>
				{:else if media.revenue && media.revenue > 0}
					<span class="meta-chip boxoffice" title="Receita Mundial">
						<span>🎟️ Receita: {formatCurrency(media.revenue)}</span>
					</span>
				{/if}

				{#if media.budget && media.budget > 0}
					<span class="meta-chip budget" title="Orçamento de Produção">
						<span>💰 Orçamento: {formatCurrency(media.budget)}</span>
					</span>
				{/if}

				{#if ext?.awards}
					<span class="meta-chip awards" title={ext.awards}>
						<span>🏆 {ext.awards}</span>
					</span>
				{/if}
			</div>
		{/if}
	</div>
{/if}

<style lang="scss">
	.media-tagline {
		font-size: 14.5px;
		font-style: italic;
		color: rgba(255, 255, 255, 0.72);
		margin: -2px 0 8px 0;
		letter-spacing: 0.01em;
	}

	.original-title-sub {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 12.5px;
		color: rgba(255, 255, 255, 0.55);
		margin-bottom: 8px;

		.orig-label {
			font-weight: 600;
			color: rgba(255, 255, 255, 0.42);
		}

		.orig-val {
			color: rgba(255, 255, 255, 0.82);
			font-weight: 600;
		}

		.lang-code {
			font-size: 10px;
			font-weight: 700;
			padding: 1px 5px;
			border-radius: 4px;
			background: rgba(255, 255, 255, 0.1);
			color: rgba(255, 255, 255, 0.75);
		}
	}

	.enrichment-panel {
		display: flex;
		flex-direction: column;
		gap: 8px;
		margin: 6px 0 14px 0;
	}

	.scores-row {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 8px;

		.score-badge {
			display: inline-flex;
			align-items: center;
			gap: 6px;
			padding: 4px 10px;
			border-radius: 8px;
			font-size: 12.5px;
			font-weight: 700;
			text-decoration: none;
			background: rgba(255, 255, 255, 0.07);
			border: 1px solid rgba(255, 255, 255, 0.14);
			color: #f8fafc;
			backdrop-filter: blur(10px);
			transition: transform 140ms ease, border-color 140ms ease, background 140ms ease;

			&:hover {
				transform: translateY(-1px);
				border-color: rgba(255, 255, 255, 0.28);
			}

			.brand {
				font-size: 11px;
				font-weight: 800;
				letter-spacing: 0.4px;
				padding: 1px 5px;
				border-radius: 4px;
			}

			&.imdb,
			&.imdb-link {
				background: rgba(245, 197, 24, 0.12);
				border-color: rgba(245, 197, 24, 0.35);
				color: #fde047;

				.brand {
					background: #f5c518;
					color: #000000;
				}
			}

			&.rt {
				background: rgba(250, 50, 10, 0.14);
				border-color: rgba(250, 50, 10, 0.35);
				color: #fca5a5;
			}

			&.rt-aud {
				background: rgba(249, 115, 22, 0.14);
				border-color: rgba(249, 115, 22, 0.35);
				color: #fdba74;
			}

			&.metacritic {
				background: rgba(34, 197, 94, 0.14);
				border-color: rgba(34, 197, 94, 0.38);
				color: #86efac;

				.brand {
					background: #22c55e;
					color: #052e16;
				}
			}

			&.critic {
				background: rgba(168, 85, 247, 0.15);
				border-color: rgba(168, 85, 247, 0.38);
				color: #d8b4fe;
			}

			&.letterboxd {
				background: rgba(16, 185, 129, 0.14);
				border-color: rgba(16, 185, 129, 0.35);
				color: #6ee7b7;
			}

			&.trakt {
				background: rgba(237, 28, 36, 0.14);
				border-color: rgba(237, 28, 36, 0.35);
				color: #fda4af;
			}

			&.tvmaze {
				background: rgba(60, 148, 139, 0.18);
				border-color: rgba(60, 148, 139, 0.4);
				color: #5eead4;
			}

			&.anilist {
				background: rgba(2, 169, 255, 0.15);
				border-color: rgba(2, 169, 255, 0.38);
				color: #7dd3fc;
			}
		}
	}

	.meta-chips-row {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 7px;

		.meta-chip {
			display: inline-flex;
			align-items: center;
			gap: 6px;
			padding: 3px 10px;
			border-radius: 999px;
			font-size: 12px;
			font-weight: 600;
			color: rgba(255, 255, 255, 0.85);
			background: rgba(255, 255, 255, 0.06);
			border: 1px solid rgba(255, 255, 255, 0.11);

			.net-logo {
				height: 13px;
				width: auto;
				max-width: 48px;
				object-fit: contain;
				filter: brightness(0) invert(1);
			}

			&.awards {
				background: rgba(234, 179, 8, 0.12);
				border-color: rgba(234, 179, 8, 0.3);
				color: #fef08a;
			}

			&.ttb,
			&.ttb-100 {
				background: rgba(56, 189, 248, 0.12);
				border-color: rgba(56, 189, 248, 0.28);
				color: #bae6fd;
			}
		}
	}
</style>

