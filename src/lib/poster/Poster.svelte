<script lang="ts">
	import {
		type WatchedStatus,
		type Watched,
		type Media,
		MediaTypeE,
		type SupportedMedia,
	} from "@/types";
	import {
		isTouch,
		mouseOverEl,
	} from "@/lib/util/helpers";
	import { goto } from "$app/navigation";
	import { baseURL, removeWatched, updateWatched } from "../util/api";
	import { notify, notifyUndo } from "../util/notify";
	import { onMount, onDestroy } from "svelte";
	import PosterStatus from "./PosterStatus.svelte";
	import PosterRating from "./PosterRating.svelte";
	import ExtraDetails from "./ExtraDetails.svelte";
	import { buildExtraDetails } from "./lib";
	import { decode } from "blurhash";
	import { resolve } from "$app/paths";
	import Icon from "../Icon.svelte";
	import { formatPublicRating, formatVoteCount } from "../rating/helpers";

	interface Props {
		media: Media;
		/**
		 * If this content is on our watched list,
		 * the entry should be provided in full.
		 */
		watched?: Watched;
		small?: boolean;
		disableInteraction?: boolean;
		hideButtons?: boolean;
		hideRating?: boolean;
		fluidSize?: boolean;
		/**
		 * If the poster should be hidden if not on users watched list (no `id`).
		 * Doing it this way so we can quickly hide posters with css and avoid
		 * triggering the #each block again where we create poster lists,
		 * which makes this functionality more performant (because we don't have
		 * support for virtual lists yet, we are re-creating all posters in places).
		 * Notably 'On my list' feature (eg on person page).
		 */
		hideIfNotOnList?: boolean;
		// When provided, default click handlers will instead run this callback.
		onClick?: (() => void) | undefined;
		/**
		 * Ran when watched item is updated via poster.
		 */
		onUpdated?: (() => void) | undefined;
		children?: import("svelte").Snippet;
	}

	let {
		media,
		// The `watched` prop is bindable so that when we update
		// or add content to our list, we can let the update flow
		// back to our parent that passed it in (so state is in sync
		// between parent and this component).
		watched = $bindable(undefined),
		small = false,
		disableInteraction = false,
		hideButtons = false,
		hideRating = false,
		fluidSize = false,
		hideIfNotOnList = false,
		onClick = undefined,
		onUpdated = undefined,
		children,
	}: Props = $props();

	// If poster is active (scaled up)
	let posterActive = $state(false);
	// If mouse in on poster. Added to fix #656.
	let mouseOverPoster = $state(false);
	// If the image is loaded or failed. -1 = fail, 0 = false, 1 = true
	let posterImgLoaded = $state(0);

	let containerEl: HTMLDivElement | undefined = $state();
	let bhCanvas: HTMLCanvasElement | undefined = $state();

	// If the item was just deleted from watched list (via this poster.)
	let justDeletedFromWatcheds = $state(false);

	const meta:
		| {
				id: number | undefined;
				type: SupportedMedia;
		  }
		| undefined = $derived.by(() => {
		let id: number | undefined;
		let type: SupportedMedia;
		switch (media.type) {
			case MediaTypeE.tmdbMovie:
				id = media.ids.tmdb;
				type = "movie";
				break;
			case MediaTypeE.tmdbShow:
				id = media.ids.tmdb;
				type = "tv";
				break;
			case MediaTypeE.igdbGame:
				id = media.ids.igdb;
				type = "game";
				break;
			default:
				return;
		}
		return {
			id,
			type,
		};
	});
	const poster = $derived.by(() => {
		if (media.poster?.path) {
			return `${baseURL}/${media.poster.path}`;
		}

		// Logic below uses `extPosterPath`, so check here first.
		// If it doesn't exist, return undefined.
		if (!media.extPosterPath) {
			return;
		}

		if (
			media.type == MediaTypeE.tmdbMovie ||
			media.type == MediaTypeE.tmdbShow
		) {
			if (watched) {
				// For now, if the content is on watched list, we can assume we have a local
				// cached image. Could be improved, since we could have a cached image for
				// show not on someone elses watched list.
				return `${baseURL}/img${media.extPosterPath}`;
			} else {
				return `https://image.tmdb.org/t/p/w500${media.extPosterPath}`;
			}
		} else if (media.type == MediaTypeE.igdbGame) {
			return `https://images.igdb.com/igdb/image/upload/t_cover_big/${media.extPosterPath}.jpg`;
		}
	});
	const link = $derived<`${`/${SupportedMedia}/${string}`}` | undefined>(
		meta?.id ? `/${meta.type}/${meta.id}` : undefined,
	);
	const year = $derived.by(() => {
		if (!media.releaseDate) return undefined;
		const y = new Date(media.releaseDate).getFullYear();
		return isNaN(y) ? undefined : y;
	});
	const publicRating = $derived(
		!hideRating ? formatPublicRating(media.rating) : undefined,
	);
	const formattedRatingCount = $derived(formatVoteCount(media.ratingCount));

	function updateWatchedVar(w: Watched | undefined) {
		watched = w;
		justDeletedFromWatcheds = typeof w === "undefined";
	}

	function handleStarClick(r: number) {
		if (r == watched?.rating || !meta?.id) return;
		updateWatched(watched, {
			contentId: meta.id,
			contentType: meta.type,
			rating: r,
		}).then((w) => {
			if (typeof onUpdated === "function") {
				onUpdated();
				runPosterMouseLeaveIfNeeded();
			}
			// If watched was just added, we need to assign
			// it to our `watched` var to get the update.
			updateWatchedVar(w);
		});
	}

	function handleStatusClick(type: WatchedStatus | "DELETE") {
		if (type === "DELETE") {
			if (!watched) {
				notify({
					text: "Content has no watched list entry, can't delete.",
					type: "error",
				});
				return;
			}
			const prevWatched = watched;
			const mediaTitle = media.name;
			updateWatchedVar(undefined);
			if (typeof onUpdated === "function") {
				onUpdated();
				runPosterMouseLeaveIfNeeded();
			}

			notifyUndo({
				text: `<b>${mediaTitle}</b> removido da lista`,
				undoLabel: "Desfazer",
				time: 6000,
				onUndo: () => {
					updateWatchedVar(prevWatched);
					if (typeof onUpdated === "function") {
						onUpdated();
					}
				},
				onDismiss: () => {
					removeWatched(prevWatched.id);
				},
			});
			return;
		}
		if (type == watched?.status || !meta?.id) return;
		updateWatched(watched, {
			contentId: meta.id,
			contentType: meta.type,
			status: type,
		}).then((w) => {
			if (typeof onUpdated === "function") {
				onUpdated();
				runPosterMouseLeaveIfNeeded();
			}
			// If watched was just added, we need to assign
			// it to our `watched` var to get the update.
			updateWatchedVar(w);
		});
	}

	function handleInnerKeyUp(e: KeyboardEvent) {
		if (
			e.key === "Enter" &&
			(e.target as HTMLElement)?.closest(".poster-body")
		) {
			if (typeof onClick !== "undefined") {
				onClick();
				return;
			}
			if (link) {
				goto(resolve(link));
			}
		}
	}

	/**
	 * Manual way to check if we need to run the mouseleave
	 * event for the poster containerEl.
	 */
	function runPosterMouseLeaveIfNeeded() {
		// Timeout to give enough time for the element to
		// actually move if it needs to (which can happen if
		// certain filters/sorts are applied).
		setTimeout(() => {
			if (!mouseOverEl(containerEl)) {
				posterOnMouseLeave();
			}
		}, 100);
	}

	// 3D Parallax Tilt & Specular Light Sheen
	let tiltX = $state(0);
	let tiltY = $state(0);
	let sheenX = $state(50);
	let sheenY = $state(50);
	let sheenOpacity = $state(0);
	let isTilting = $state(false);
	let tiltRaf: number | undefined;

	function handleMouseMove(e: MouseEvent) {
		if (isTouch() || disableInteraction || !containerEl) return;
		if (
			typeof window !== "undefined" &&
			window.matchMedia("(prefers-reduced-motion: reduce)").matches
		) {
			return;
		}

		if (tiltRaf) cancelAnimationFrame(tiltRaf);
		tiltRaf = requestAnimationFrame(() => {
			if (!containerEl) return;
			const rect = containerEl.getBoundingClientRect();
			if (!rect.width || !rect.height) return;

			const x = (e.clientX - rect.left) / rect.width;
			const y = (e.clientY - rect.top) / rect.height;

			const clampedX = Math.max(0, Math.min(1, x));
			const clampedY = Math.max(0, Math.min(1, y));

			// Subtle 10-12 deg maximum tilt for high-end feel
			tiltY = Number(((clampedX - 0.5) * 12).toFixed(2));
			tiltX = Number(((0.5 - clampedY) * 12).toFixed(2));

			sheenX = Number((clampedX * 100).toFixed(1));
			sheenY = Number((clampedY * 100).toFixed(1));
			sheenOpacity = 1;
			isTilting = true;
		});
	}

	function resetTilt() {
		if (tiltRaf) cancelAnimationFrame(tiltRaf);
		tiltX = 0;
		tiltY = 0;
		sheenOpacity = 0;
		isTilting = false;
	}

	function posterOnMouseLeave() {
		mouseOverPoster = false;
		posterActive = false;
		resetTilt();
		const ae = document.activeElement;
		if (
			ae &&
			ae instanceof HTMLElement &&
			ae.closest(".poster-body")
		) {
			// Stops the poster being re-focused after the browser window
			// loses focus, then regains it (ex: you middle click the poster,
			// go to the opened tab (or lose browser window focus, then when
			// you come back the poster is sent `focusin` and stuck activated
			// until mouseleave again).
			ae.blur();
		}
	}

	onMount(() => {
		if (containerEl) {
			if (small) {
				containerEl.classList.add("small");
			}
			if (fluidSize) {
				containerEl.classList.add("fluid-size");
			}
		}

		// Show blurhash if needed (optimized 32x48 decode with smooth GPU upscaling)
		if (
			media.poster?.path &&
			media.poster?.blurHash &&
			bhCanvas &&
			posterImgLoaded === 0
		) {
			requestAnimationFrame(() => {
				if (posterImgLoaded === 1 || !bhCanvas) return;
				try {
					const pixels = decode(media.poster!.blurHash, 32, 48);
					const ctx = bhCanvas.getContext("2d");
					if (ctx) {
						const imageData = ctx.createImageData(32, 48);
						imageData.data.set(pixels);
						ctx.putImageData(imageData, 0, 0);
					}
				} catch (e) {
					console.warn("Poster: Failed to decode blurhash:", e);
				}
			});
		}
	});

	onDestroy(() => {
		if (tiltRaf) cancelAnimationFrame(tiltRaf);
	});
</script>

<!-- HACK: disabled this issue for now, it should probably be fixed properly -->
<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<li
	onmouseenter={() => {
		mouseOverPoster = true;
		if (!isTouch()) {
			posterActive = true;
		}
	}}
	onmousemove={handleMouseMove}
	onfocusin={() => {
		if (!isTouch()) {
			posterActive = true;
		}
	}}
	onfocusout={() => {
		if (!isTouch() && !mouseOverPoster) {
			// Only on !isTouch (to match focusin) to avoid breaking a tap and hold on link on mobile.
			// and only if mouse isn't still over the poster, fixes focusout on click of rating/status
			// poster buttons causing poster to shrink until refocused with click/mouse out & in again.
			posterActive = false;
			resetTilt();
		}
	}}
	onmouseleave={posterOnMouseLeave}
	onclick={() => (posterActive = true)}
	onkeyup={(e) => {
		if (e.key === "Tab") {
			e.currentTarget.scrollIntoView({ block: "center" });
		}
	}}
	onkeypress={() => console.log("on kpress")}
	class={`${posterActive ? "active " : ""}${watched?.pinned ? "pinned " : ""}${hideIfNotOnList && !watched ? "hidden " : ""}`}
	class:just-deleted={justDeletedFromWatcheds}
>
	<div
		class={`container${!poster || posterImgLoaded == -1 ? " details-shown" : ""}`}
		class:is-tilting={isTilting}
		style={posterActive && (tiltX !== 0 || tiltY !== 0)
			? `transform: perspective(800px) rotateX(${tiltX}deg) rotateY(${tiltY}deg) translateY(-5px) scale3d(1.025, 1.025, 1.025);`
			: undefined}
		bind:this={containerEl}
	>
		{@render children?.()}

		<!-- 3D Specular Sheen (Dynamic Glare/Reflection) -->
		<div
			class="specular-sheen"
			style="--sheen-x: {sheenX}%; --sheen-y: {sheenY}%; --sheen-opacity: {sheenOpacity};"
			aria-hidden="true"
		></div>
		{#if poster}
			<canvas
				width="32"
				height="48"
				bind:this={bhCanvas}
				class="img-loader"
				class:fade-out={posterImgLoaded === 1}
			></canvas>
			{#if !media?.poster?.blurHash && posterImgLoaded === 0}
				<div class="img-loader"></div>
			{/if}
			<img
				loading="lazy"
				src={poster}
				alt=""
				class:loaded={posterImgLoaded === 1}
				onload={() => {
					posterImgLoaded = 1;
				}}
				onerror={() => {
					posterImgLoaded = -1;
				}}
			/>
		{/if}
		{#if publicRating && !posterActive}
			<div
				class="public-rating"
				title={`Rating: ${publicRating} out of 10${formattedRatingCount ? ` (based on ${media.ratingCount} votes)` : ""}`}
			>
				<Icon i="star" wh={11} />
				<span>{publicRating}</span>
			</div>
		{/if}
		{#if watched && meta && !posterActive}
			<!-- Must be on watched list, and poster not hovered -->
			<ExtraDetails {...buildExtraDetails(meta.type, watched)} />
		{/if}
		<div
			onclick={(e) => {
				if (typeof onClick !== "undefined") {
					onClick();
					// Prevent the link inside this div from being clicked in this case.
					e.preventDefault();
					return;
				}
				if (posterActive && link) goto(resolve(link));
			}}
			onkeyup={handleInnerKeyUp}
			class="inner poster-body"
			role="button"
			tabindex="-1"
		>
			<a
				data-sveltekit-preload-data="tap"
				href={link ? resolve(link) : undefined}
				class="small-scrollbar"
			>
				<h2>
					{media.name}
					{#if year}
						<time>{year}</time>
					{/if}
					{#if media.status === "Canceled"}
						<span class="canceled-tag" title="Série cancelada pela emissora/produtores">Cancelada</span>
					{/if}
				</h2>
				{#if media.status === "Canceled"}
					<div class="card-canceled-alert" title="Série marcada como cancelada pela emissora ou produtores">
						<Icon i="close-circle" wh={12} />
						<span>Cancelada pela produção</span>
					</div>
				{/if}
				{#if publicRating}
					<div
						class="inner-rating"
						title={`Rating: ${publicRating} out of 10${formattedRatingCount ? ` (based on ${media.ratingCount} votes)` : ""}`}
					>
						<Icon i="star" wh={12} />
						<span>{publicRating}</span>
						{#if formattedRatingCount}
							<span class="rating-count">({formattedRatingCount})</span>
						{/if}
					</div>
				{/if}
				<span class="summary">{media.summary}</span>
			</a>

			{#if !hideButtons}
				<div class="buttons">
					<PosterRating
						rating={watched?.rating}
						{handleStarClick}
						{disableInteraction}
					/>
					<PosterStatus
						status={watched?.status}
						{handleStatusClick}
						{disableInteraction}
					/>
				</div>
			{/if}
		</div>
	</div>
</li>

<style lang="scss">
	li.hidden {
		display: none;
	}

	li.active {
		cursor: pointer;
	}

	li.pinned:not(.active) .container {
		box-shadow:
			0 0 0 2px rgba(234, 179, 8, 0.85),
			0 8px 24px rgba(234, 179, 8, 0.25);
	}

	li {
		perspective: 900px;

		&:not(.active) {
			.container .inner,
			.container .inner .buttons {
				pointer-events: none !important;
			}
		}
	}

	li.just-deleted:not(.active) {
		filter: grayscale(0.8) blur(1px);
	}

	.container {
		display: flex;
		flex-flow: column;
		background-color: var(--card-bg, #181c25);
		overflow: hidden;
		flex: 1 1;
		border-radius: 10px;
		min-width: 170px;
		width: 170px;
		position: relative;
		aspect-ratio: 170000/256367;
		border: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
		box-shadow: 0 4px 14px rgba(0, 0, 0, 0.25);
		transform-style: preserve-3d;
		transition:
			transform 240ms cubic-bezier(0.16, 1, 0.3, 1),
			box-shadow 240ms cubic-bezier(0.16, 1, 0.3, 1);

		&.is-tilting {
			transition:
				transform 60ms linear,
				box-shadow 180ms ease;
		}

		.specular-sheen {
			position: absolute;
			inset: 0;
			pointer-events: none;
			border-radius: inherit;
			background: radial-gradient(
				circle 220px at var(--sheen-x, 50%) var(--sheen-y, 50%),
				rgba(255, 255, 255, 0.24) 0%,
				rgba(255, 255, 255, 0.07) 45%,
				transparent 85%
			);
			opacity: var(--sheen-opacity, 0);
			mix-blend-mode: overlay;
			transition: opacity 220ms ease;
			z-index: 5;
		}

		&.fluid-size {
			height: 100%;
			width: 100%;
		}

		img {
			width: 100%;
			height: 100%;
			object-fit: cover;
			opacity: 0;
			transition: opacity 220ms ease, filter 180ms ease;

			&.loaded {
				opacity: 1;
			}
		}

		&.details-shown .img-loader {
			display: none;
		}

		.img-loader {
			position: absolute;
			width: 100%;
			height: 100%;
			background: linear-gradient(359deg, #242b38, #181d27, #0e1117);
			background-size: 400% 400%;
			animation: imgloader 4s ease infinite;
			transition: opacity 280ms ease;

			&.fade-out {
				opacity: 0;
				pointer-events: none;
			}

			@keyframes imgloader {
				0% {
					background-position: 50% 0%;
				}
				50% {
					background-position: 50% 100%;
				}
				100% {
					background-position: 50% 0%;
				}
			}
		}

		.public-rating {
			position: absolute;
			top: 8px;
			right: 8px;
			display: inline-flex;
			align-items: center;
			gap: 3.5px;
			padding: 2.5px 6px;
			border-radius: 6px;
			background: rgba(10, 13, 20, 0.82);
			border: 1px solid rgba(245, 158, 11, 0.3);
			color: #fbbf24;
			font-size: 11px;
			font-weight: 700;
			line-height: 1;
			letter-spacing: 0.02em;
			box-shadow: 0 4px 12px rgba(0, 0, 0, 0.45);
			backdrop-filter: blur(8px);
			-webkit-backdrop-filter: blur(8px);
			transform: translateZ(16px);
			backface-visibility: hidden;
			z-index: 2;
			pointer-events: none;
			user-select: none;

			:global(svg) {
				fill: #f59e0b;
				color: #f59e0b;
				flex-shrink: 0;
			}
		}

		&.small .public-rating {
			top: 6px;
			right: 6px;
			padding: 2px 4.5px;
			font-size: 10px;
			gap: 2.5px;
		}

		.inner {
			position: absolute;
			opacity: 0;
			display: flex;
			flex-flow: column;
			top: 0;
			height: 100%;
			width: 100%;
			padding: 12px;
			background: linear-gradient(
				to top,
				rgba(8, 10, 16, 0.96) 0%,
				rgba(8, 10, 16, 0.88) 55%,
				rgba(8, 10, 16, 0.65) 100%
			);
			backdrop-filter: blur(6px);
			-webkit-backdrop-filter: blur(6px);
			transform: translateZ(12px);
			backface-visibility: hidden;
			transition: opacity 180ms cubic-bezier(0.16, 1, 0.3, 1);

			& > a {
				display: flex;
				flex-flow: column;
				height: 100%;
				overflow: hidden;
				text-decoration: none;
			}

			h2 {
				font-family: inherit;
				font-size: 14.5px;
				font-weight: 700;
				color: white;
				word-wrap: break-word;
				line-height: 1.25;

				time {
					font-size: 11.5px;
					font-weight: 500;
					color: rgba(255, 255, 255, 0.65);
					margin-left: 4px;
				}

				.canceled-tag {
					display: inline-block;
					background: rgba(225, 29, 72, 0.22);
					border: 1px solid rgba(225, 29, 72, 0.45);
					color: #fb7185;
					font-size: 9.5px;
					font-weight: 700;
					padding: 1px 5px;
					border-radius: 4px;
					margin-left: 6px;
					vertical-align: middle;
					text-transform: uppercase;
					letter-spacing: 0.03em;
				}
			}

			.card-canceled-alert {
				display: inline-flex;
				align-items: center;
				gap: 4px;
				margin: 3px 0 2px 0;
				padding: 2px 6px;
				border-radius: 4px;
				background: rgba(225, 29, 72, 0.22);
				border: 1px solid rgba(225, 29, 72, 0.45);
				color: #fb7185;
				font-size: 10px;
				font-weight: 600;
				letter-spacing: 0.02em;
				line-height: 1.2;
				width: fit-content;

				:global(svg) {
					color: #fb7185;
					flex-shrink: 0;
				}
			}

			.inner-rating {
				display: inline-flex;
				align-items: center;
				gap: 4px;
				margin-top: 4px;
				margin-bottom: 2px;
				font-size: 12px;
				font-weight: 700;
				color: #fbbf24;

				:global(svg) {
					fill: #f59e0b;
					color: #f59e0b;
					flex-shrink: 0;
				}

				span {
					color: #fbbf24;
					margin: 0;
					font-size: 12px;
					line-height: 1;
					display: inline;
					-webkit-line-clamp: unset;
					line-clamp: unset;
				}

				.rating-count {
					font-size: 11px;
					font-weight: 400;
					color: rgba(255, 255, 255, 0.6);
				}
			}

			span.summary,
			span:not(.inner-rating *) {
				color: rgba(255, 255, 255, 0.85);
				margin: 6px 0;
				font-size: 11px;
				line-height: 1.4;
				display: -webkit-box;
				-webkit-line-clamp: 4;
				line-clamp: 4;
				-webkit-box-orient: vertical;
				hyphens: auto;
				overflow: hidden;
				text-overflow: ellipsis;
			}

			.buttons {
				display: flex;
				flex-flow: row;
				margin-top: auto;
				gap: 8px;
				height: 34px;
				flex-shrink: 0;
			}
		}

		&.small .inner span {
			font-size: 11px;
		}

		.active & {
			transform: translateY(-4px);
			box-shadow: 0 16px 36px -4px rgba(0, 0, 0, 0.7), 0 0 0 1px rgba(255, 255, 255, 0.14);
			z-index: 10;
		}

		.active &.small {
			transform: translateY(-3px);
		}

		.active &,
		&:global(.details-shown) {
			img {
				filter: blur(2px) brightness(0.65);
			}

			.inner {
				color: white;
				opacity: 1;
			}
		}

		@media (prefers-reduced-motion: reduce) {
			transform: none !important;
			transition: none !important;

			.specular-sheen {
				display: none !important;
			}
		}
	}
</style>
