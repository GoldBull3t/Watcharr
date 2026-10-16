<script lang="ts">
	import { store } from "@/store.svelte";
	import tooltip from "../actions/tooltip";
	import { RatingSystem } from "@/types";
	import Icon from "../Icon.svelte";
	import { toShowableRating, toWhichThumb } from "../rating/helpers";

	interface Props {
		rating?: number | undefined;
		handleStarClick: (rating: number) => void;
		disableInteraction?: boolean;
		/**
		 * When not minimal, we will use user settings to
		 * display ratings as they want.
		 */
		minimal?: boolean;
		direction?: "top" | "bot";
		btnTooltip?: string;
		hideStarWhenRated?: boolean;
	}

	let {
		rating = undefined,
		handleStarClick,
		disableInteraction = false,
		minimal = false,
		direction = "top",
		btnTooltip = "",
		hideStarWhenRated = false,
	}: Props = $props();

	let ratingsShown = $state(false);
	let hoveredRating: number | undefined = $state();
	let isDragging = $state(false);
	let buttonEl: HTMLButtonElement | undefined = $state();
	let trackEl: HTMLDivElement | undefined = $state();
	let openUpward = $state(false);
	let alignRight = $state(false);

	const ratingDesc = [
		"Appalling",
		"Horrible",
		"Very Bad",
		"Bad",
		"Average",
		"Fine",
		"Good",
		"Very Good",
		"Great",
		"Masterpiece",
	];

	function toggleMenu(ev: MouseEvent) {
		ev.stopPropagation();
		if (!ratingsShown && buttonEl) {
			const rect = buttonEl.getBoundingClientRect();
			const spaceBelow = window.innerHeight - rect.bottom;
			const spaceAbove = rect.top;
			openUpward = spaceBelow < 220 && spaceAbove > spaceBelow;
			alignRight = rect.left > window.innerWidth / 2;
		}
		ratingsShown = !ratingsShown;
		if (!ratingsShown) {
			hoveredRating = undefined;
			isDragging = false;
		}
	}

	function handleWindowClick(e: MouseEvent) {
		if (ratingsShown && buttonEl && !buttonEl.contains(e.target as Node)) {
			ratingsShown = false;
			hoveredRating = undefined;
			isDragging = false;
		}
	}

	let isUsingThumbs = $derived(
		store.userSettings &&
			store.userSettings.ratingSystem === RatingSystem.Thumbs,
	);

	let isOutOf5 = $derived(
		store.userSettings &&
			store.userSettings.ratingSystem === RatingSystem.OutOf5,
	);

	let isOutOf100 = $derived(
		store.userSettings &&
			store.userSettings.ratingSystem === RatingSystem.OutOf100,
	);

	let activeDbRating = $derived(hoveredRating ?? rating ?? 0);

	let filledStarsOutOf5 = $derived(activeDbRating / 2);

	let moodLabel = $derived.by(() => {
		if (!activeDbRating || activeDbRating <= 0) return "Select Rating";
		const idx = Math.max(
			0,
			Math.min(ratingDesc.length - 1, Math.ceil(activeDbRating) - 1),
		);
		return ratingDesc[idx] ?? "";
	});

	let displayScoreStr = $derived.by(() => {
		if (!activeDbRating || activeDbRating <= 0) return "";
		if (isOutOf5) {
			return `${activeDbRating / 2} / 5`;
		}
		if (isOutOf100) {
			return `${activeDbRating * 10} / 100`;
		}
		return `${activeDbRating} / 10`;
	});

	function commitRating(val?: number) {
		const target = val ?? hoveredRating ?? rating;
		if (typeof target === "number" && target > 0) {
			handleStarClick(target);
		}
		ratingsShown = false;
		hoveredRating = undefined;
		isDragging = false;
	}

	function updateScoreFromPointer(ev: PointerEvent) {
		if (!trackEl) return;
		const rect = trackEl.getBoundingClientRect();
		if (rect.width <= 0) return;
		const relX = Math.max(0, Math.min(rect.width, ev.clientX - rect.left));
		const ratio = relX / rect.width;

		let steps = Math.ceil(ratio * 10);
		if (steps < 1) steps = 1;
		if (steps > 10) steps = 10;

		hoveredRating = steps;
	}

	function handleTrackPointerDown(ev: PointerEvent) {
		ev.stopPropagation();
		ev.preventDefault();
		try {
			(ev.currentTarget as HTMLElement).setPointerCapture(ev.pointerId);
		} catch {}
		isDragging = true;
		updateScoreFromPointer(ev);
	}

	function handleTrackPointerMove(ev: PointerEvent) {
		ev.stopPropagation();
		if (isDragging || ev.buttons === 1) {
			updateScoreFromPointer(ev);
		} else {
			updateScoreFromPointer(ev);
		}
	}

	function handleTrackPointerUp(ev: PointerEvent) {
		ev.stopPropagation();
		if (isDragging) {
			isDragging = false;
			try {
				(ev.currentTarget as HTMLElement).releasePointerCapture(ev.pointerId);
			} catch {}
			commitRating();
		}
	}

	function handleTrackPointerLeave() {
		if (!isDragging) {
			hoveredRating = undefined;
		}
	}

	function stepRating(delta: number) {
		const cur = activeDbRating || (delta > 0 ? 0 : 10);
		const next = Math.max(1, Math.min(10, cur + delta));
		hoveredRating = next;
	}

	function handleWindowKeyDown(e: KeyboardEvent) {
		if (!ratingsShown) return;

		if (e.key === "Escape") {
			e.preventDefault();
			ratingsShown = false;
			hoveredRating = undefined;
			isDragging = false;
			return;
		}

		if (e.key === "Enter") {
			e.preventDefault();
			if (hoveredRating !== undefined) {
				commitRating(hoveredRating);
			} else {
				ratingsShown = false;
			}
			return;
		}

		if (e.key >= "1" && e.key <= "9") {
			e.preventDefault();
			const num = parseInt(e.key, 10);
			if (isUsingThumbs) {
				if (num <= 3) commitRating(1);
				else if (num <= 7) commitRating(5);
				else commitRating(9);
			} else if (isOutOf5) {
				if (num <= 5) commitRating(num * 2);
			} else {
				commitRating(num);
			}
			return;
		}

		if (e.key === "0") {
			e.preventDefault();
			if (isUsingThumbs) {
				commitRating(9);
			} else {
				commitRating(10);
			}
			return;
		}

		if (e.key === "ArrowLeft") {
			e.preventDefault();
			stepRating(-1);
			return;
		}

		if (e.key === "ArrowRight") {
			e.preventDefault();
			stepRating(1);
			return;
		}
	}
</script>

<svelte:window onclick={handleWindowClick} onkeydown={handleWindowKeyDown} />

<div
	class={[
		"rating-wrap",
		minimal ? "is-minimal" : "",
		ratingsShown ? "open" : "",
	].join(" ")}
>
	<button
		bind:this={buttonEl}
		type="button"
		class={[
			"rating-btn",
			minimal ? (!rating ? "minimal" : "minimal-space") : "",
			disableInteraction ? "interaction-disabled" : "",
			minimal ? "is-minimal" : "",
			ratingsShown ? "open" : "",
		].join(" ")}
		onclick={toggleMenu}
		aria-haspopup="true"
		aria-expanded={ratingsShown}
		use:tooltip={{
			text: btnTooltip,
			pos: "top",
			condition: !!btnTooltip && !ratingsShown,
		}}
	>
		{#if !isUsingThumbs || (isUsingThumbs && minimal && !rating)}
			<span
				class="star-icon"
				style={hideStarWhenRated && rating ? "display: none" : ""}
			>
				<Icon i="star" wh={minimal ? 12 : 14} />
			</span>
		{/if}
		<span
			class={[
				!rating && disableInteraction ? "unrated-text" : "",
				"rating-text",
			].join(" ")}
		>
			{#if rating}
				{#if isUsingThumbs}
					{@const r = toWhichThumb(rating)}
					{#if r === -1}
						<Icon i="thumb-down" />
					{:else if r === 0}
						<span class="neutral-dash" aria-hidden="true">
							<svg width="14" height="3" viewBox="0 0 14 3">
								<rect width="14" height="3" rx="1.5" />
							</svg>
						</span>
					{:else if r === 1}
						<Icon i="thumb-up" />
					{/if}
				{:else}
					{toShowableRating(rating)}
				{/if}
			{:else if minimal}
				<!-- eslint-disable-next-line svelte/no-useless-mustaches -->
				{""}
			{:else if disableInteraction}
				Unrated
			{:else}
				Rate
			{/if}
		</span>
	</button>

	{#if ratingsShown}
		<div
			class={[
				"ratings-dropdown",
				"shown",
				openUpward ? "open-upward" : direction,
				alignRight ? "align-right" : "",
				isUsingThumbs ? "is-using-thumbs" : "",
				minimal ? "is-minimal" : "",
			].join(" ")}
			role="dialog"
			aria-label="Rate content"
		>
			{#if isUsingThumbs}
				<div class="thumbs-grid">
					<button
						type="button"
						onclick={(ev) => {
							ev.stopPropagation();
							commitRating(1);
						}}
						class="thumb-action-btn dislike"
						class:active={rating && rating > 0 && rating < 5}
						aria-label="Dislike"
					>
						<Icon i="thumb-down" wh={20} />
						<span>Dislike</span>
					</button>
					<button
						type="button"
						onclick={(ev) => {
							ev.stopPropagation();
							commitRating(5);
						}}
						class="thumb-action-btn neutral"
						class:active={rating && rating >= 4 && rating <= 7}
						aria-label="Neutral"
					>
						<span class="neutral-dash-large" aria-hidden="true">
							<svg width="18" height="3" viewBox="0 0 18 3">
								<rect width="18" height="3" rx="1.5" />
							</svg>
						</span>
						<span>Neutral</span>
					</button>
					<button
						type="button"
						onclick={(ev) => {
							ev.stopPropagation();
							commitRating(9);
						}}
						class="thumb-action-btn like"
						class:active={rating && rating >= 8}
						aria-label="Like"
					>
						<Icon i="thumb-up" wh={20} />
						<span>Like</span>
					</button>
				</div>
				<div class="keyboard-hint">
					<span>Keys 1, 2, 3 • Esc to close</span>
				</div>
			{:else}
				<!-- Modern Interactive Star Bar with Half-Star Drag and Quick Chips -->
				<header class="rating-header">
					<div class="score-tag">
						<span class="star-badge-icon">★</span>
						<span class="score-num">{displayScoreStr || "Rate"}</span>
					</div>
					<span class="mood-label">{moodLabel}</span>
				</header>

				<!-- Interactive 5-Star Drag Track -->
				<div
					bind:this={trackEl}
					class="stars-drag-track"
					class:dragging={isDragging}
					role="slider"
					tabindex="0"
					aria-label="Star rating slider"
					aria-valuemin="1"
					aria-valuemax={isOutOf5 ? 5 : 10}
					aria-valuenow={activeDbRating}
					onpointerdown={handleTrackPointerDown}
					onpointermove={handleTrackPointerMove}
					onpointerup={handleTrackPointerUp}
					onpointerleave={handleTrackPointerLeave}
				>
					{#each [1, 2, 3, 4, 5] as starIdx (starIdx)}
						{@const fill =
							filledStarsOutOf5 >= starIdx
								? 1
								: filledStarsOutOf5 >= starIdx - 0.5
									? 0.5
									: 0}
						<div
							class="star-node"
							class:full={fill === 1}
							class:half={fill === 0.5}
						>
							<svg class="star-bg" viewBox="0 0 24 24">
								<path
									d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"
								/>
							</svg>
							<div class="star-fg" style="width: {fill * 100}%;">
								<svg class="star-gold" viewBox="0 0 24 24">
									<path
										d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"
									/>
								</svg>
							</div>
						</div>
					{/each}
				</div>

				<!-- Quick Direct Number Chips -->
				<div class="quick-chips">
					{#if isOutOf5}
						{#each [1, 2, 3, 4, 5] as n (n)}
							<button
								type="button"
								class="chip-btn"
								class:active={activeDbRating === n * 2}
								onclick={(ev) => {
									ev.stopPropagation();
									commitRating(n * 2);
								}}
								title={`${n} / 5`}
							>
								{n}
							</button>
						{/each}
					{:else}
						{#each [1, 2, 3, 4, 5, 6, 7, 8, 9, 10] as n (n)}
							<button
								type="button"
								class="chip-btn"
								class:active={activeDbRating === n}
								onclick={(ev) => {
									ev.stopPropagation();
									commitRating(n);
								}}
								title={`${n} / 10`}
							>
								{n}
							</button>
						{/each}
					{/if}
				</div>

				<div class="keyboard-hint">
					<span>Drag stars • Keys 1-9, 0</span>
				</div>
			{/if}
		</div>
	{/if}
</div>

<style lang="scss">
	.rating-wrap {
		position: relative;
		width: 100%;
		height: 100%;
		display: flex;

		&.open {
			z-index: 60 !important;
		}
	}

	button.rating-btn {
		padding: 3px 6px;
		position: relative;
		font-family: inherit;
		font-weight: 700;
		width: 100%;
		height: 100%;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 4px;

		&.is-minimal {
			padding: 3px 8px;
		}

		&.interaction-disabled {
			pointer-events: none;
			cursor: default;
			background-color: transparent;
			border: unset;
			fill: white;
			color: white;

			span {
				color: white !important;
			}

			.unrated-text {
				display: flex;
				align-items: center;
				font-size: 13px !important;
				font-weight: 600;
			}
		}

		span {
			&.star-icon {
				color: $rating-color;
				display: inline-flex;
				align-items: center;
				justify-content: center;
				line-height: 1;
				flex-shrink: 0;

				:global(svg) {
					fill: $rating-color;
					color: $rating-color;
				}
			}

			&.rating-text {
				color: $text-color;
				font-size: 14px;
				font-weight: 700;
				display: inline-flex;
				align-items: center;
				line-height: 1;

				& :global(svg) {
					height: 100%;
					padding: 2px;
				}
			}
		}

		&:hover span,
		&:focus-visible span {
			color: $rating-color;
			fill: $rating-color;
		}

		&.open {
			z-index: 60 !important;
		}
	}

	.neutral-dash {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 20px;
		height: 20px;

		svg {
			rect {
				fill: currentColor;
			}
		}
	}

	.ratings-dropdown {
		display: flex;
		flex-direction: column;
		position: absolute;
		width: 260px;
		max-width: calc(100vw - 28px);
		padding: 12px 14px;
		gap: 10px;
		background: rgba(13, 17, 28, 0.96);
		border: 1px solid rgba(245, 158, 11, 0.28);
		border-radius: 16px;
		z-index: 100;
		box-shadow:
			0 18px 40px -6px rgba(0, 0, 0, 0.8),
			0 0 24px rgba(245, 158, 11, 0.12);
		backdrop-filter: blur(24px) saturate(180%);
		-webkit-backdrop-filter: blur(24px) saturate(180%);
		animation: ratingPopIn 150ms cubic-bezier(0.16, 1, 0.3, 1);
		user-select: none;

		@keyframes ratingPopIn {
			0% {
				opacity: 0;
				transform: scale(0.92) translateY(4px);
			}
			100% {
				opacity: 1;
				transform: scale(1) translateY(0);
			}
		}

		&.open-upward {
			bottom: calc(100% + 8px);
			top: auto;
		}

		&:not(.open-upward) {
			top: calc(100% + 8px);
			bottom: auto;
		}

		&.align-right {
			right: 0;
			left: auto;
		}

		&:not(.align-right) {
			left: 0;
			right: auto;
		}

		.rating-header {
			display: flex;
			align-items: center;
			justify-content: space-between;
			gap: 8px;

			.score-tag {
				display: inline-flex;
				align-items: center;
				gap: 5px;
				font-size: 13px;
				font-weight: 700;
				color: #ffd060;
				background: rgba(245, 158, 11, 0.12);
				border: 1px solid rgba(245, 158, 11, 0.28);
				padding: 2px 8px;
				border-radius: 6px;

				.star-badge-icon {
					font-size: 12px;
				}
			}

			.mood-label {
				font-size: 12px;
				font-weight: 600;
				color: rgba(255, 255, 255, 0.7);
				white-space: nowrap;
				overflow: hidden;
				text-overflow: ellipsis;
			}
		}

		.stars-drag-track {
			display: flex;
			align-items: center;
			justify-content: space-between;
			padding: 6px 4px;
			cursor: pointer;
			touch-action: none;
			user-select: none;
			border-radius: 10px;
			background: rgba(255, 255, 255, 0.03);
			border: 1px solid rgba(255, 255, 255, 0.05);
			transition: background 120ms ease;

			&:hover,
			&.dragging {
				background: rgba(245, 158, 11, 0.06);
				border-color: rgba(245, 158, 11, 0.2);
			}

			.star-node {
				position: relative;
				width: 32px;
				height: 32px;
				flex-shrink: 0;
				transition: transform 120ms cubic-bezier(0.34, 1.56, 0.64, 1);

				&:hover,
				&.full,
				&.half {
					transform: scale(1.1);
				}

				.star-bg {
					width: 100%;
					height: 100%;
					fill: rgba(255, 255, 255, 0.18);
					stroke: rgba(255, 255, 255, 0.08);
					stroke-width: 0.75;
				}

				.star-fg {
					position: absolute;
					top: 0;
					left: 0;
					height: 100%;
					overflow: hidden;
					pointer-events: none;
					transition: width 80ms ease;

					.star-gold {
						width: 32px;
						height: 32px;
						fill: #fbbf24;
						filter: drop-shadow(0 0 6px rgba(251, 191, 36, 0.6));
					}
				}
			}
		}

		.quick-chips {
			display: flex;
			align-items: center;
			justify-content: space-between;
			gap: 3px;

			.chip-btn {
				flex: 1;
				height: 24px;
				min-width: 0;
				display: inline-flex;
				align-items: center;
				justify-content: center;
				font-size: 11px;
				font-weight: 700;
				border-radius: 6px;
				background: rgba(255, 255, 255, 0.06);
				border: 1px solid rgba(255, 255, 255, 0.08);
				color: rgba(255, 255, 255, 0.7);
				cursor: pointer;
				padding: 0;
				transition:
					background 120ms ease,
					color 120ms ease,
					transform 120ms ease;

				&:hover {
					background: rgba(255, 255, 255, 0.16);
					color: #ffffff;
					transform: translateY(-1px);
				}

				&.active {
					background: rgba(245, 158, 11, 0.28);
					border-color: rgba(245, 158, 11, 0.55);
					color: #ffd060;
				}
			}
		}

		.keyboard-hint {
			display: flex;
			align-items: center;
			justify-content: center;
			gap: 6px;
			font-size: 11px;
			font-weight: 500;
			color: rgba(255, 255, 255, 0.4);
			text-align: center;
		}

		/* Thumbs Mode */
		.thumbs-grid {
			display: grid;
			grid-template-columns: repeat(3, 1fr);
			gap: 8px;
			width: 100%;

			.thumb-action-btn {
				display: flex;
				flex-direction: column;
				align-items: center;
				justify-content: center;
				gap: 6px;
				padding: 10px 4px;
				border-radius: 10px;
				background: rgba(255, 255, 255, 0.06);
				border: 1px solid rgba(255, 255, 255, 0.08);
				color: #ffffff;
				font-size: 12px;
				font-weight: 600;
				cursor: pointer;
				transition:
					background 140ms ease,
					border-color 140ms ease,
					transform 140ms ease;

				&:hover {
					transform: translateY(-2px);
				}

				&.dislike:hover,
				&.dislike.active {
					background: rgba(244, 63, 94, 0.2);
					border-color: rgba(244, 63, 94, 0.5);
					color: #fb7185;
				}

				&.neutral:hover,
				&.neutral.active {
					background: rgba(245, 158, 11, 0.2);
					border-color: rgba(245, 158, 11, 0.5);
					color: #fbbf24;
				}

				&.like:hover,
				&.like.active {
					background: rgba(16, 185, 129, 0.2);
					border-color: rgba(16, 185, 129, 0.5);
					color: #34d399;
				}

				.neutral-dash-large {
					display: inline-flex;
					align-items: center;
					justify-content: center;
					width: 20px;
					height: 20px;

					svg rect {
						fill: currentColor;
					}
				}
			}
		}
	}
</style>
