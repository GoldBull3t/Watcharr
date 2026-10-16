<script lang="ts">
	import { MediaVideoType, type MediaVideo } from "@/types";
	import Icon from "@/lib/Icon.svelte";
	import { onMount } from "svelte";

	export interface LightboxItem {
		id: string;
		type: "video" | "backdrop" | "poster";
		title: string;
		url?: string;
		embedUrl?: string;
		thumbUrl?: string;
		badge?: string;
	}

	interface Props {
		videos?: MediaVideo[];
		backdropPath?: string;
		posterPath?: string;
		mediaTitle?: string;
		initialItemId?: string;
		initialType?: "video" | "backdrop" | "poster";
		onclose: () => void;
	}

	let {
		videos = [],
		backdropPath,
		posterPath,
		mediaTitle = "Galeria",
		initialItemId,
		initialType,
		onclose,
	}: Props = $props();

	// Build items array
	const items = $derived.by(() => {
		const list: LightboxItem[] = [];

		if (videos && videos.length > 0) {
			const sortedVideos = [...videos].sort((a, b) => {
				if (a.best === b.best) {
					if (a.type === b.type) return 0;
					return a.type === MediaVideoType.trailer ? -1 : 1;
				}
				return a.best ? -1 : 1;
			});

			for (const v of sortedVideos) {
				if (v.id) {
					list.push({
						id: v.id,
						type: "video",
						title: v.name || "Trailer Oficial",
						embedUrl: `https://www.youtube-nocookie.com/embed/${v.id}?autoplay=1&rel=0&modestbranding=1`,
						thumbUrl: `https://img.youtube.com/vi/${v.id}/mqdefault.jpg`,
						badge: v.best
							? "Principal"
							: v.type === MediaVideoType.trailer
								? "Trailer"
								: "Vídeo",
					});
				}
			}
		}

		if (backdropPath) {
			list.push({
				id: "backdrop",
				type: "backdrop",
				title: "Backdrop Oficial (4K)",
				url: `https://image.tmdb.org/t/p/original${backdropPath}`,
				thumbUrl: `https://image.tmdb.org/t/p/w300${backdropPath}`,
				badge: "Backdrop 4K",
			});
		}

		if (posterPath) {
			list.push({
				id: "poster",
				type: "poster",
				title: "Poster Oficial HD",
				url: `https://image.tmdb.org/t/p/original${posterPath}`,
				thumbUrl: `https://image.tmdb.org/t/p/w185${posterPath}`,
				badge: "Poster HD",
			});
		}

		return list;
	});

	// Find initial index
	let activeIndex = $state(0);

	onMount(() => {
		if (items.length > 0) {
			if (initialItemId) {
				const found = items.findIndex((it) => it.id === initialItemId);
				if (found !== -1) activeIndex = found;
			} else if (initialType) {
				const found = items.findIndex((it) => it.type === initialType);
				if (found !== -1) activeIndex = found;
			}
		}

		// Prevent background scrolling while modal is open
		const originalOverflow = document.body.style.overflow;
		document.body.style.overflow = "hidden";

		function handleKeyDown(e: KeyboardEvent) {
			if (e.key === "Escape") {
				onclose();
			} else if (e.key === "ArrowRight") {
				nextItem();
			} else if (e.key === "ArrowLeft") {
				prevItem();
			}
		}

		window.addEventListener("keydown", handleKeyDown);

		return () => {
			document.body.style.overflow = originalOverflow;
			window.removeEventListener("keydown", handleKeyDown);
		};
	});

	let currentItem = $derived(items[activeIndex] || items[0]);

	function nextItem() {
		if (items.length <= 1) return;
		activeIndex = (activeIndex + 1) % items.length;
	}

	function prevItem() {
		if (items.length <= 1) return;
		activeIndex = (activeIndex - 1 + items.length) % items.length;
	}

	function selectItem(index: number) {
		activeIndex = index;
	}

	let videoItems = $derived(items.filter((i) => i.type === "video"));
	let hasBackdrop = $derived(items.some((i) => i.type === "backdrop"));
	let hasPoster = $derived(items.some((i) => i.type === "poster"));
</script>

<div
	class="lightbox-overlay"
	role="dialog"
	aria-modal="true"
	aria-label="Reprodutor e Galeria de Mídia"
	tabindex="-1"
>
	<!-- Top Bar -->
	<header class="lightbox-header">
		<div class="header-left">
			<span class="media-title">{mediaTitle}</span>
			{#if currentItem}
				<span class="item-separator">•</span>
				<span class="item-title">{currentItem.title}</span>
				{#if currentItem.badge}
					<span class="item-badge">{currentItem.badge}</span>
				{/if}
			{/if}
		</div>

		<!-- Category Switcher Tabs -->
		<div class="category-tabs">
			{#if videoItems.length > 0}
				<button
					type="button"
					class="tab-btn"
					class:active={currentItem?.type === "video"}
					onclick={() => {
						const idx = items.findIndex((i) => i.type === "video");
						if (idx !== -1) selectItem(idx);
					}}
				>
					<Icon i="play" wh={12} />
					<span>Trailers ({videoItems.length})</span>
				</button>
			{/if}

			{#if hasBackdrop}
				<button
					type="button"
					class="tab-btn"
					class:active={currentItem?.type === "backdrop"}
					onclick={() => {
						const idx = items.findIndex((i) => i.type === "backdrop");
						if (idx !== -1) selectItem(idx);
					}}
				>
					<span>🖼️ Backdrop</span>
				</button>
			{/if}

			{#if hasPoster}
				<button
					type="button"
					class="tab-btn"
					class:active={currentItem?.type === "poster"}
					onclick={() => {
						const idx = items.findIndex((i) => i.type === "poster");
						if (idx !== -1) selectItem(idx);
					}}
				>
					<span>🎨 Poster</span>
				</button>
			{/if}
		</div>

		<div class="header-right">
			<span class="counter-badge">{activeIndex + 1} / {items.length}</span>
			<button
				type="button"
				class="close-btn"
				onclick={onclose}
				title="Fechar (Esc)"
			>
				<span>Fechar</span>
				<span class="kbd">Esc</span>
			</button>
		</div>
	</header>

	<!-- Main Stage Area -->
	<div class="lightbox-stage">
		<!-- Prev Arrow -->
		{#if items.length > 1}
			<button
				type="button"
				class="nav-arrow nav-prev"
				onclick={prevItem}
				title="Anterior (Seta Esquerda)"
				aria-label="Item anterior"
			>
				‹
			</button>
		{/if}

		<!-- Stage Content -->
		<div class="stage-content">
			{#if currentItem}
				{#if currentItem.type === "video"}
					<div class="video-container">
						<iframe
							title={currentItem.title}
							src={currentItem.embedUrl}
							frameborder="0"
							width="100%"
							height="100%"
							allow="autoplay; fullscreen; encrypted-media; picture-in-picture"
							allowfullscreen
						></iframe>
					</div>
				{:else if currentItem.type === "backdrop" || currentItem.type === "poster"}
					<div
						class="image-container"
						class:is-poster={currentItem.type === "poster"}
					>
						<img
							src={currentItem.url}
							alt={currentItem.title}
							loading="eager"
							decoding="async"
						/>
						<div class="image-actions">
							<a
								href={currentItem.url}
								target="_blank"
								rel="noreferrer noopener"
								class="view-original-btn"
							>
								<span>Abrir Resolução Máxima (Original)</span>
								<span class="ext">↗</span>
							</a>
						</div>
					</div>
				{/if}
			{/if}
		</div>

		<!-- Next Arrow -->
		{#if items.length > 1}
			<button
				type="button"
				class="nav-arrow nav-next"
				onclick={nextItem}
				title="Seguinte (Seta Direita)"
				aria-label="Item seguinte"
			>
				›
			</button>
		{/if}
	</div>

	<!-- Bottom Thumbnails Strip -->
	{#if items.length > 1}
		<footer class="lightbox-footer">
			<div class="thumbs-strip">
				{#each items as item, idx (item.id + idx)}
					<button
						type="button"
						class="thumb-item"
						class:active={idx === activeIndex}
						onclick={() => selectItem(idx)}
						title={item.title}
					>
						<div class="thumb-media">
							{#if item.thumbUrl}
								<img
									src={item.thumbUrl}
									alt=""
									loading="lazy"
								/>
							{:else}
								<div class="thumb-placeholder">🎬</div>
							{/if}
							{#if item.type === "video"}
								<span class="thumb-play-icon">▶</span>
							{/if}
						</div>
						<span class="thumb-label">{item.badge || item.title}</span>
					</button>
				{/each}
			</div>
		</footer>
	{/if}
</div>

<style lang="scss">
	.lightbox-overlay {
		position: fixed;
		top: 0;
		left: 0;
		width: 100dvw;
		height: 100dvh;
		background: rgba(6, 9, 15, 0.94);
		backdrop-filter: blur(28px) saturate(180%);
		-webkit-backdrop-filter: blur(28px) saturate(180%);
		z-index: 999999;
		display: flex;
		flex-flow: column;
		justify-content: space-between;
		user-select: none;
		outline: none;
		animation: overlayFadeIn 180ms cubic-bezier(0.16, 1, 0.3, 1);

		@keyframes overlayFadeIn {
			from {
				opacity: 0;
				transform: scale(0.99);
			}
			to {
				opacity: 1;
				transform: scale(1);
			}
		}

		.lightbox-header {
			display: flex;
			align-items: center;
			justify-content: space-between;
			gap: 16px;
			padding: 16px 24px;
			border-bottom: 1px solid rgba(255, 255, 255, 0.08);
			background: rgba(14, 18, 27, 0.6);
			flex-shrink: 0;

			.header-left {
				display: flex;
				align-items: center;
				gap: 10px;
				overflow: hidden;

				.media-title {
					font-size: 15px;
					font-weight: 700;
					color: #ffffff;
					white-space: nowrap;
				}

				.item-separator {
					color: rgba(255, 255, 255, 0.3);
				}

				.item-title {
					font-size: 14px;
					color: rgba(255, 255, 255, 0.7);
					white-space: nowrap;
					overflow: hidden;
					text-overflow: ellipsis;
				}

				.item-badge {
					font-size: 10px;
					font-weight: 700;
					text-transform: uppercase;
					letter-spacing: 0.5px;
					padding: 2px 7px;
					border-radius: 4px;
					background: rgba(99, 102, 241, 0.25);
					border: 1px solid rgba(99, 102, 241, 0.4);
					color: #a5b4fc;
				}
			}

			.category-tabs {
				display: flex;
				align-items: center;
				gap: 6px;

				.tab-btn {
					display: inline-flex;
					align-items: center;
					gap: 6px;
					padding: 6px 14px;
					border-radius: 20px;
					background: rgba(255, 255, 255, 0.06);
					border: 1px solid rgba(255, 255, 255, 0.1);
					color: rgba(255, 255, 255, 0.7);
					font-size: 12px;
					font-weight: 600;
					cursor: pointer;
					transition: all 140ms ease;

					&:hover {
						background: rgba(255, 255, 255, 0.12);
						color: #ffffff;
					}

					&.active {
						background: rgba(99, 102, 241, 0.28);
						border-color: #818cf8;
						color: #ffffff;
						box-shadow: 0 0 14px rgba(99, 102, 241, 0.3);
					}
				}
			}

			.header-right {
				display: flex;
				align-items: center;
				gap: 12px;

				.counter-badge {
					font-size: 12px;
					font-weight: 600;
					color: rgba(255, 255, 255, 0.5);
				}

				.close-btn {
					display: inline-flex;
					align-items: center;
					gap: 6px;
					padding: 6px 14px;
					border-radius: 10px;
					background: rgba(255, 255, 255, 0.08);
					border: 1px solid rgba(255, 255, 255, 0.12);
					color: #ffffff;
					font-size: 12.5px;
					font-weight: 600;
					cursor: pointer;
					transition: all 140ms ease;

					.kbd {
						padding: 1px 5px;
						border-radius: 4px;
						background: rgba(0, 0, 0, 0.3);
						border: 1px solid rgba(255, 255, 255, 0.15);
						font-size: 10px;
						font-family: monospace;
						color: rgba(255, 255, 255, 0.6);
					}

					&:hover {
						background: rgba(239, 68, 68, 0.2);
						border-color: rgba(239, 68, 68, 0.4);
						color: #fca5a5;
						transform: scale(1.02);
					}
				}
			}
		}

		.lightbox-stage {
			position: relative;
			flex: 1;
			display: flex;
			align-items: center;
			justify-content: center;
			padding: 20px 70px;
			overflow: hidden;

			.nav-arrow {
				position: absolute;
				top: 50%;
				transform: translateY(-50%);
				width: 48px;
				height: 48px;
				border-radius: 50%;
				background: rgba(255, 255, 255, 0.08);
				border: 1px solid rgba(255, 255, 255, 0.14);
				color: #ffffff;
				font-size: 32px;
				line-height: 1;
				display: flex;
				align-items: center;
				justify-content: center;
				cursor: pointer;
				z-index: 10;
				transition: all 160ms cubic-bezier(0.16, 1, 0.3, 1);
				backdrop-filter: blur(8px);

				&:hover {
					background: rgba(255, 255, 255, 0.2);
					border-color: rgba(255, 255, 255, 0.35);
					transform: translateY(-50%) scale(1.08);
					box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
				}

				&.nav-prev {
					left: 20px;
				}

				&.nav-next {
					right: 20px;
				}
			}

			.stage-content {
				width: 100%;
				height: 100%;
				display: flex;
				align-items: center;
				justify-content: center;

				.video-container {
					width: 100%;
					max-width: 1050px;
					aspect-ratio: 16 / 9;
					border-radius: 16px;
					overflow: hidden;
					background: #000000;
					box-shadow:
						0 20px 60px rgba(0, 0, 0, 0.8),
						0 0 0 1px rgba(255, 255, 255, 0.1);
				}

				.image-container {
					position: relative;
					display: flex;
					flex-flow: column;
					align-items: center;
					justify-content: center;
					max-width: 90vw;
					max-height: 72vh;

					img {
						max-width: 100%;
						max-height: 68vh;
						object-fit: contain;
						border-radius: 14px;
						box-shadow:
							0 24px 60px rgba(0, 0, 0, 0.85),
							0 0 0 1px rgba(255, 255, 255, 0.12);
					}

					&.is-poster img {
						max-height: 68vh;
						aspect-ratio: 2 / 3;
					}

					.image-actions {
						margin-top: 10px;

						.view-original-btn {
							display: inline-flex;
							align-items: center;
							gap: 6px;
							padding: 5px 12px;
							border-radius: 8px;
							background: rgba(255, 255, 255, 0.08);
							border: 1px solid rgba(255, 255, 255, 0.12);
							color: #cbd5e1;
							font-size: 11.5px;
							font-weight: 600;
							text-decoration: none;
							transition: all 140ms ease;

							.ext {
								font-size: 10px;
								opacity: 0.7;
							}

							&:hover {
								background: rgba(255, 255, 255, 0.18);
								border-color: rgba(255, 255, 255, 0.3);
								color: #ffffff;
								transform: translateY(-1px);
							}
						}
					}
				}
			}
		}

		.lightbox-footer {
			padding: 12px 24px 18px;
			background: rgba(14, 18, 27, 0.6);
			border-top: 1px solid rgba(255, 255, 255, 0.08);
			flex-shrink: 0;

			.thumbs-strip {
				display: flex;
				align-items: center;
				justify-content: center;
				gap: 10px;
				overflow-x: auto;
				padding: 4px 0;

				&::-webkit-scrollbar {
					height: 4px;
				}
				&::-webkit-scrollbar-thumb {
					background: rgba(255, 255, 255, 0.15);
					border-radius: 4px;
				}

				.thumb-item {
					display: flex;
					flex-flow: column;
					align-items: center;
					gap: 4px;
					background: transparent;
					border: none;
					cursor: pointer;
					padding: 0;
					opacity: 0.55;
					transition: all 140ms ease;

					.thumb-media {
						position: relative;
						width: 84px;
						height: 48px;
						border-radius: 8px;
						overflow: hidden;
						background: rgba(255, 255, 255, 0.05);
						border: 2px solid transparent;
						transition: all 140ms ease;

						img {
							width: 100%;
							height: 100%;
							object-fit: cover;
						}

						.thumb-play-icon {
							position: absolute;
							inset: 0;
							margin: auto;
							width: 20px;
							height: 20px;
							border-radius: 50%;
							background: rgba(0, 0, 0, 0.6);
							color: #ffffff;
							font-size: 8px;
							display: flex;
							align-items: center;
							justify-content: center;
							padding-left: 2px;
						}
					}

					.thumb-label {
						font-size: 10.5px;
						font-weight: 600;
						color: rgba(255, 255, 255, 0.7);
						max-width: 84px;
						white-space: nowrap;
						overflow: hidden;
						text-overflow: ellipsis;
					}

					&:hover {
						opacity: 0.85;

						.thumb-media {
							transform: translateY(-2px);
						}
					}

					&.active {
						opacity: 1;

						.thumb-media {
							border-color: #818cf8;
							box-shadow: 0 0 12px rgba(99, 102, 241, 0.4);
							transform: translateY(-2px);
						}

						.thumb-label {
							color: #a5b4fc;
							font-weight: 700;
						}
					}
				}
			}
		}
	}

	@media screen and (max-width: 768px) {
		.lightbox-overlay {
			.lightbox-header {
				padding: 12px 14px;
				flex-wrap: wrap;

				.category-tabs {
					order: 3;
					width: 100%;
					justify-content: center;
				}
			}

			.lightbox-stage {
				padding: 10px 14px;

				.nav-arrow {
					display: none;
				}
			}
		}
	}
</style>

