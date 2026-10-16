<script lang="ts">
	import { resolve } from "$app/paths";
	import Icon from "@/lib/Icon.svelte";
	import { MediaTypeE, type Media } from "@/types";
	import { getMediaLink, getPosterImage, getReleaseYear } from "./mediaHelpers";
	import { formatPublicRating, formatVoteCount } from "@/lib/rating/helpers";
	import { onDestroy, onMount } from "svelte";

	interface Props {
		items: Media[];
		onItemUpdated?: () => void;
	}

	let { items = [], onItemUpdated }: Props = $props();

	let activeIndex = $state(0);
	let isHovered = $state(false);
	let autoPlayInterval: number | undefined;
	let isCollapsed = $state(false);

	const spotlightItems = $derived.by(() => {
		if (!items || items.length === 0) return [];
		return items.slice(0, 6);
	});

	const activeItem: Media | undefined = $derived(spotlightItems[activeIndex]);

	function nextSlide() {
		if (spotlightItems.length <= 1) return;
		activeIndex = (activeIndex + 1) % spotlightItems.length;
	}

	function prevSlide() {
		if (spotlightItems.length <= 1) return;
		activeIndex = (activeIndex - 1 + spotlightItems.length) % spotlightItems.length;
	}

	function goToSlide(idx: number) {
		activeIndex = idx;
	}

	function startAutoPlay() {
		stopAutoPlay();
		if (spotlightItems.length <= 1) return;
		autoPlayInterval = window.setInterval(() => {
			if (!isHovered && !isCollapsed) {
				nextSlide();
			}
		}, 7000);
	}

	function stopAutoPlay() {
		if (autoPlayInterval) {
			clearInterval(autoPlayInterval);
			autoPlayInterval = undefined;
		}
	}

	function toggleCollapse() {
		isCollapsed = !isCollapsed;
		localStorage.setItem(
			"watcharr_hero_spotlight_collapsed",
			isCollapsed ? "true" : "false",
		);
	}

	onMount(() => {
		const saved = localStorage.getItem("watcharr_hero_spotlight_collapsed");
		if (saved === "true") {
			isCollapsed = true;
		}
		startAutoPlay();
	});

	onDestroy(() => {
		stopAutoPlay();
	});

	$effect(() => {
		if (spotlightItems.length > 0 && activeIndex >= spotlightItems.length) {
			activeIndex = 0;
		}
	});

	function getBackdropUrl(m: Media): string | undefined {
		if (!m) return undefined;
		if (m.extBackdropPath) {
			return `https://image.tmdb.org/t/p/w1280${m.extBackdropPath}`;
		}
		// Fallback to poster if backdrop doesn't exist
		return getPosterImage(m);
	}
</script>

{#if spotlightItems.length > 0}
	<section
		class="hero-spotlight-container"
		class:collapsed={isCollapsed}
		aria-label="Spotlight and continue watching"
		onmouseenter={() => (isHovered = true)}
		onmouseleave={() => (isHovered = false)}
	>
		<!-- Top Bar with Status and Collapse Toggle -->
		<div class="spotlight-top-strip">
			<div class="strip-badge-wrap">
				<span class="live-dot"></span>
				<span class="strip-label">DESTAQUES EM ANDAMENTO</span>
				<span class="strip-count">({spotlightItems.length})</span>
			</div>
			<button
				class="collapse-toggle-btn"
				onclick={toggleCollapse}
				title={isCollapsed ? "Expandir Spotlight" : "Minimizar Spotlight"}
				aria-label={isCollapsed ? "Expandir Spotlight" : "Minimizar Spotlight"}
			>
				<span class="toggle-text">{isCollapsed ? "Mostrar Destaques" : "Minimizar"}</span>
				<span class="toggle-icon" class:rotated={isCollapsed}>
					<Icon i="chevron" wh={14} />
				</span>
			</button>
		</div>

		{#if !isCollapsed && activeItem}
			{@const current = activeItem}
			{@const backdrop = getBackdropUrl(current)}
			{@const poster = getPosterImage(current)}
			{@const link = getMediaLink(current)}
			{@const year = getReleaseYear(current)}
			{@const rating = formatPublicRating(current.rating)}
			{@const voteCount = formatVoteCount(current.ratingCount)}
			{@const typeLabel =
				current.type === MediaTypeE.tmdbMovie
					? "Filme"
					: current.type === MediaTypeE.tmdbShow
						? "Série"
						: "Jogo"}

			<div class="spotlight-stage">
				<!-- Panoramic Backdrop Layer -->
				<div class="stage-backdrop-wrap">
					{#each spotlightItems as item, idx (item.ids.tmdb || item.ids.igdb || item.name || idx)}
						{@const bg = getBackdropUrl(item)}
						{#if bg}
							<div
								class="backdrop-layer"
								class:active={idx === activeIndex}
								style={`background-image: url('${bg}');`}
							></div>
						{/if}
					{/each}
					<!-- Cinematic Multi-direction Gradient Scrims -->
					<div class="scrim-left"></div>
					<div class="scrim-bottom"></div>
					<div class="scrim-top"></div>
				</div>

				<!-- Content Details Card (Left Overlaid) -->
				<div class="stage-content">
					<div class="meta-pills">
						<span class="type-pill">
							{#if current.type === MediaTypeE.tmdbShow}
								<Icon i="tv" wh={12} />
							{:else if current.type === MediaTypeE.igdbGame}
								<Icon i="gamepad" wh={12} />
							{:else}
								<Icon i="film" wh={12} />
							{/if}
							{typeLabel}
						</span>

						{#if current.watched?.watchingSeason}
							<span class="progress-pill">
								<Icon i="clock" wh={12} />
								{current.watched.watchingSeason}
							</span>
						{:else}
							<span class="progress-pill in-progress">
								<Icon i="play" wh={12} />
								A Assistir
							</span>
						{/if}

						{#if rating}
							<span class="rating-pill" title={`${rating} de 10 (${voteCount})`}>
								<Icon i="star" wh={12} />
								{rating}
							</span>
						{/if}

						{#if year}
							<span class="year-pill">{year}</span>
						{/if}
					</div>

					<h2 class="spotlight-title">
						<a href={resolve(link)}>{current.name}</a>
					</h2>

					{#if current.summary}
						<p class="spotlight-summary">{current.summary}</p>
					{/if}

					<!-- Call to Action Buttons -->
					<div class="spotlight-actions">
						<a href={resolve(link)} class="btn-primary">
							<Icon i="play" wh={16} />
							<span>Continuar a Ver</span>
						</a>
						<a href={resolve(link)} class="btn-secondary">
							<Icon i="arrow" wh={14} />
							<span>Ver Detalhes</span>
						</a>
					</div>
				</div>

				<!-- Navigation Chevrons (Hover) -->
				{#if spotlightItems.length > 1}
					<button
						class="nav-arrow prev"
						onclick={prevSlide}
						aria-label="Item anterior"
					>
						<span class="arrow-rotate-left">
							<Icon i="arrow" wh={16} />
						</span>
					</button>
					<button
						class="nav-arrow next"
						onclick={nextSlide}
						aria-label="Próximo item"
					>
						<Icon i="arrow" wh={16} />
					</button>
				{/if}

				<!-- Slide Indicator Bars (Bottom Right) -->
				{#if spotlightItems.length > 1}
					<div class="spotlight-indicators">
						{#each spotlightItems as item, idx}
							<button
								class="indicator-pill"
								class:active={idx === activeIndex}
								onclick={() => goToSlide(idx)}
								title={item.name}
								aria-label={`Slide ${idx + 1}: ${item.name}`}
							>
								{#if idx === activeIndex && !isHovered}
									<span class="indicator-timer"></span>
								{/if}
							</button>
						{/each}
					</div>
				{/if}
			</div>
		{/if}
	</section>
{/if}

<style lang="scss">
	.hero-spotlight-container {
		position: relative;
		margin-bottom: 24px;
		border-radius: 16px;
		background: rgba(13, 17, 26, 0.85);
		border: 1px solid rgba(255, 255, 255, 0.09);
		box-shadow:
			0 20px 50px -10px rgba(0, 0, 0, 0.75),
			0 0 0 1px rgba(255, 255, 255, 0.04);
		overflow: hidden;
		transition: all 260ms cubic-bezier(0.16, 1, 0.3, 1);

		&.collapsed {
			margin-bottom: 16px;
			box-shadow: 0 4px 14px rgba(0, 0, 0, 0.3);
		}
	}

	.spotlight-top-strip {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 10px 18px;
		background: rgba(9, 12, 19, 0.75);
		border-bottom: 1px solid rgba(255, 255, 255, 0.06);
		backdrop-filter: blur(10px);

		.strip-badge-wrap {
			display: flex;
			align-items: center;
			gap: 8px;

			.live-dot {
				width: 7px;
				height: 7px;
				border-radius: 50%;
				background: #6366f1;
				box-shadow: 0 0 10px #6366f1;
				animation: livePulse 2s infinite ease-in-out;
			}

			@keyframes livePulse {
				0%,
				100% {
					opacity: 1;
					transform: scale(1);
				}
				50% {
					opacity: 0.4;
					transform: scale(0.85);
				}
			}

			.strip-label {
				font-size: 11px;
				font-weight: 800;
				letter-spacing: 0.08em;
				color: #c7d2fe;
				text-transform: uppercase;
			}

			.strip-count {
				font-size: 11px;
				font-weight: 600;
				color: rgba(255, 255, 255, 0.4);
			}
		}

		.collapse-toggle-btn {
			display: inline-flex;
			align-items: center;
			gap: 6px;
			background: transparent;
			border: none;
			color: rgba(255, 255, 255, 0.55);
			font-size: 11.5px;
			font-weight: 600;
			cursor: pointer;
			padding: 2px 8px;
			border-radius: 6px;
			transition: all 120ms ease;

			&:hover {
				color: #ffffff;
				background: rgba(255, 255, 255, 0.07);
			}

			.toggle-icon {
				display: inline-flex;
				transition: transform 200ms ease;

				&.rotated {
					transform: rotate(180deg);
				}
			}
		}
	}

	.spotlight-stage {
		position: relative;
		min-height: 340px;
		height: 380px;
		display: flex;
		align-items: center;
		overflow: hidden;

		@media (max-width: 768px) {
			height: 320px;
			min-height: 300px;
		}

		@media (max-width: 500px) {
			height: 290px;
		}
	}

	.stage-backdrop-wrap {
		position: absolute;
		inset: 0;
		overflow: hidden;
		pointer-events: none;

		.backdrop-layer {
			position: absolute;
			inset: 0;
			background-size: cover;
			background-position: center 25%;
			opacity: 0;
			transform: scale(1.02);
			transition:
				opacity 700ms cubic-bezier(0.4, 0, 0.2, 1),
				transform 7000ms linear;

			&.active {
				opacity: 1;
				transform: scale(1.05);
			}
		}

		.scrim-left {
			position: absolute;
			inset: 0;
			background: linear-gradient(
				to right,
				rgba(11, 14, 22, 0.98) 0%,
				rgba(11, 14, 22, 0.92) 35%,
				rgba(11, 14, 22, 0.6) 65%,
				rgba(11, 14, 22, 0.15) 100%
			);

			@media (max-width: 768px) {
				background: linear-gradient(
					to right,
					rgba(11, 14, 22, 0.98) 0%,
					rgba(11, 14, 22, 0.88) 70%,
					rgba(11, 14, 22, 0.5) 100%
				);
			}
		}

		.scrim-bottom {
			position: absolute;
			bottom: 0;
			left: 0;
			right: 0;
			height: 140px;
			background: linear-gradient(to top, rgba(11, 14, 22, 1) 0%, rgba(11, 14, 22, 0) 100%);
		}

		.scrim-top {
			position: absolute;
			top: 0;
			left: 0;
			right: 0;
			height: 60px;
			background: linear-gradient(to bottom, rgba(11, 14, 22, 0.6) 0%, rgba(11, 14, 22, 0) 100%);
		}
	}

	.stage-content {
		position: relative;
		z-index: 2;
		padding: 32px 42px;
		max-width: 620px;
		display: flex;
		flex-direction: column;
		gap: 12px;

		@media (max-width: 768px) {
			padding: 24px;
			max-width: 100%;
		}

		@media (max-width: 500px) {
			padding: 18px;
			gap: 10px;
		}

		.meta-pills {
			display: flex;
			align-items: center;
			gap: 8px;
			flex-wrap: wrap;

			span {
				display: inline-flex;
				align-items: center;
				gap: 4px;
				padding: 3px 8px;
				border-radius: 6px;
				font-size: 11px;
				font-weight: 700;
				line-height: 1;
				letter-spacing: 0.03em;
				backdrop-filter: blur(6px);
			}

			.type-pill {
				background: rgba(99, 102, 241, 0.22);
				border: 1px solid rgba(99, 102, 241, 0.45);
				color: #c7d2fe;
				text-transform: uppercase;
			}

			.progress-pill {
				background: rgba(16, 185, 129, 0.2);
				border: 1px solid rgba(16, 185, 129, 0.4);
				color: #6ee7b7;

				&.in-progress {
					background: rgba(56, 189, 248, 0.18);
					border-color: rgba(56, 189, 248, 0.38);
					color: #7dd3fc;
				}
			}

			.rating-pill {
				background: rgba(245, 158, 11, 0.18);
				border: 1px solid rgba(245, 158, 11, 0.35);
				color: #fbbf24;

				:global(svg) {
					fill: #f59e0b;
				}
			}

			.year-pill {
				background: rgba(255, 255, 255, 0.08);
				border: 1px solid rgba(255, 255, 255, 0.12);
				color: rgba(255, 255, 255, 0.7);
			}
		}

		.spotlight-title {
			margin: 0;
			font-size: 32px;
			font-weight: 800;
			letter-spacing: -0.025em;
			line-height: 1.18;
			text-shadow: 0 2px 14px rgba(0, 0, 0, 0.8);

			@media (max-width: 768px) {
				font-size: 26px;
			}

			@media (max-width: 500px) {
				font-size: 22px;
			}

			a {
				color: #ffffff;
				text-decoration: none;
				transition: color 140ms ease;

				&:hover {
					color: #c7d2fe;
				}
			}
		}

		.spotlight-summary {
			margin: 0;
			color: rgba(255, 255, 255, 0.78);
			font-size: 13.5px;
			line-height: 1.5;
			display: -webkit-box;
			-webkit-line-clamp: 3;
			line-clamp: 3;
			-webkit-box-orient: vertical;
			overflow: hidden;
			text-shadow: 0 1px 8px rgba(0, 0, 0, 0.7);

			@media (max-width: 500px) {
				-webkit-line-clamp: 2;
				line-clamp: 2;
				font-size: 12.5px;
			}
		}

		.spotlight-actions {
			display: flex;
			align-items: center;
			gap: 10px;
			margin-top: 6px;

			.btn-primary {
				display: inline-flex;
				align-items: center;
				gap: 8px;
				padding: 10px 20px;
				border-radius: 9px;
				background: linear-gradient(135deg, #6366f1 0%, #4f46e5 100%);
				color: #ffffff;
				font-size: 13.5px;
				font-weight: 700;
				letter-spacing: 0.01em;
				text-decoration: none;
				box-shadow:
					0 4px 16px rgba(99, 102, 241, 0.45),
					0 0 0 1px rgba(255, 255, 255, 0.2) inset;
				transition:
					transform 140ms ease,
					box-shadow 140ms ease,
					filter 140ms ease;

				&:hover {
					transform: translateY(-2px);
					filter: brightness(1.1);
					box-shadow:
						0 8px 24px rgba(99, 102, 241, 0.6),
						0 0 0 1px rgba(255, 255, 255, 0.3) inset;
				}
			}

			.btn-secondary {
				display: inline-flex;
				align-items: center;
				gap: 6px;
				padding: 10px 16px;
				border-radius: 9px;
				background: rgba(255, 255, 255, 0.08);
				border: 1px solid rgba(255, 255, 255, 0.14);
				color: #ffffff;
				font-size: 13.5px;
				font-weight: 600;
				text-decoration: none;
				backdrop-filter: blur(8px);
				transition: all 140ms ease;

				&:hover {
					background: rgba(255, 255, 255, 0.16);
					border-color: rgba(255, 255, 255, 0.28);
				}
			}
		}
	}

	.nav-arrow {
		position: absolute;
		top: 50%;
		transform: translateY(-50%);
		z-index: 5;
		width: 38px;
		height: 38px;
		border-radius: 50%;
		background: rgba(13, 17, 26, 0.6);
		border: 1px solid rgba(255, 255, 255, 0.12);
		color: #ffffff;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		backdrop-filter: blur(10px);
		box-shadow: 0 4px 14px rgba(0, 0, 0, 0.4);
		transition: all 140ms ease;

		&:hover {
			background: rgba(99, 102, 241, 0.8);
			border-color: rgba(255, 255, 255, 0.3);
			transform: translateY(-50%) scale(1.08);
		}

		&.prev {
			left: 14px;
		}

		&.next {
			right: 14px;
		}

		.arrow-rotate-left {
			display: inline-flex;
			transform: rotate(180deg);
		}

		@media (max-width: 600px) {
			display: none;
		}
	}

	.spotlight-indicators {
		position: absolute;
		bottom: 18px;
		right: 24px;
		z-index: 5;
		display: flex;
		align-items: center;
		gap: 6px;

		@media (max-width: 500px) {
			bottom: 12px;
			right: 14px;
		}

		.indicator-pill {
			position: relative;
			height: 5px;
			width: 24px;
			border-radius: 999px;
			background: rgba(255, 255, 255, 0.2);
			border: none;
			padding: 0;
			cursor: pointer;
			overflow: hidden;
			transition: all 200ms ease;

			&.active {
				width: 44px;
				background: rgba(255, 255, 255, 0.35);

				.indicator-timer {
					display: block;
					height: 100%;
					width: 100%;
					background: #6366f1;
					animation: indicatorProgress 7000ms linear forwards;
				}
			}

			&:hover:not(.active) {
				background: rgba(255, 255, 255, 0.45);
			}
		}

		@keyframes indicatorProgress {
			from {
				transform: translateX(-100%);
			}
			to {
				transform: translateX(0%);
			}
		}
	}
</style>

