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
	let buttonEl: HTMLButtonElement | undefined = $state();
	let openUpward = $state(false);

	function toggleMenu(ev: MouseEvent) {
		ev.stopPropagation();
		if (!ratingsShown && buttonEl) {
			const rect = buttonEl.getBoundingClientRect();
			const spaceBelow = window.innerHeight - rect.bottom;
			openUpward = spaceBelow < 240 && rect.top > spaceBelow;
		}
		ratingsShown = !ratingsShown;
	}

	function handleWindowClick(e: MouseEvent) {
		if (ratingsShown && buttonEl && !buttonEl.contains(e.target as Node)) {
			ratingsShown = false;
		}
	}

	function handleWindowKeyDown(e: KeyboardEvent) {
		if (e.key === "Escape" && ratingsShown) {
			ratingsShown = false;
		}
	}

	// let settings = $derived($userSettings);
	let isUsingThumbs = $derived(
		store.userSettings &&
			store.userSettings.ratingSystem === RatingSystem.Thumbs,
	);
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
				"small-scrollbar",
				openUpward ? "open-upward" : direction,
				isUsingThumbs ? "is-using-thumbs" : "",
				minimal ? "is-minimal" : "",
			].join(" ")}
			role="menu"
		>
			{#if isUsingThumbs}
				<button
					type="button"
					onclick={(ev) => {
						ev.stopPropagation();
						ratingsShown = false;
						handleStarClick(1);
					}}
					class="plain{rating && rating > 0 && rating < 5 ? ' active' : ''}"
					role="menuitem"
					aria-label="Dislike"
					style="display: flex; justify-content: center; align-items: center;"
				>
					<i style="display: flex; width: 35px; justify-content: center;"><Icon i="thumb-down" /></i>
				</button>
				<button
					type="button"
					onclick={(ev) => {
						ev.stopPropagation();
						ratingsShown = false;
						handleStarClick(5);
					}}
					class="plain{rating && rating > 4 && rating < 9 ? ' active' : ''}"
					role="menuitem"
					aria-label="Neutral"
					style="display: flex; justify-content: center; align-items: center;"
				>
					<span class="neutral-dash" aria-hidden="true">
						<svg width="14" height="3" viewBox="0 0 14 3">
							<rect width="14" height="3" rx="1.5" />
						</svg>
					</span>
				</button>
				<button
					type="button"
					onclick={(ev) => {
						ev.stopPropagation();
						ratingsShown = false;
						handleStarClick(9);
					}}
					class="plain{rating && rating > 8 ? ' active' : ''}"
					role="menuitem"
					aria-label="Like"
					style="display: flex; justify-content: center; align-items: center;"
				>
					<i style="display: flex; width: 35px; justify-content: center;"><Icon i="thumb-up" /></i>
				</button>
			{:else}
				{@const stars =
					store.userSettings?.ratingSystem == RatingSystem.OutOf5
						? [5, 4, 3, 2, 1]
						: [10, 9, 8, 7, 6, 5, 4, 3, 2, 1]}
				{#each stars as v (v)}
					<button
						type="button"
						class="plain{rating === v ? ' active' : ''}"
						role="menuitem"
						onclick={(ev) => {
							ev.stopPropagation();
							handleStarClick(
								store.userSettings?.ratingSystem === RatingSystem.OutOf5
									? v * 2
									: v,
							);
							ratingsShown = false;
						}}
					>
						{#if store.userSettings?.ratingSystem === RatingSystem.OutOf100}
							{v * 10}
						{:else if store.userSettings?.ratingSystem === RatingSystem.OutOf5}
							{v}
						{:else}
							{v}
						{/if}
					</button>
				{/each}
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

		&.is-minimal {
			.ratings-dropdown {
				height: unset;
				max-height: 240px;
				overflow-y: auto;
				overflow-x: hidden;
			}
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
		flex-flow: column;
		position: absolute;
		width: 100%;
		min-width: 48px;
		right: 0;
		left: auto;
		height: 200px;
		background: rgba(15, 19, 27, 0.96);
		border: 1px solid rgba(255, 255, 255, 0.1);
		list-style: none;
		border-radius: 10px;
		overflow-y: auto;
		overflow-x: hidden;
		scrollbar-width: thin;
		z-index: 100;
		box-shadow: 0 16px 36px -6px rgba(0, 0, 0, 0.75);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);

		&::before {
			content: "";
			position: absolute;
			left: 0;
			right: 0;
			height: 10px;
		}

		&:not(.shown) {
			display: none;
		}

		&.top,
		&.open-upward {
			bottom: calc(100% + 6px);
			top: auto;

			&::before {
				bottom: -8px;
			}
		}

		&.bot:not(.open-upward) {
			top: calc(100% + 6px);
			bottom: auto;

			&::before {
				top: -8px;
			}
		}

		button.plain {
			width: 100%;
			color: $text-color;
			fill: $text-color;
			font-size: 13.5px;
			font-weight: 600;
			padding: 6px 4px;
			font-family: inherit;
			border-radius: 6px;
			transition: background-color 120ms ease, color 120ms ease;

			&.active {
				background-color: rgba(245, 158, 11, 0.2);
				color: #fbbf24;
				font-weight: 700;
			}

			& :global(svg) {
				width: 100%;
				padding: 0 4.5px;
			}

			&:hover,
			&:focus-visible {
				background-color: rgba(255, 255, 255, 0.08);
			}

			&.active:hover {
				background-color: rgba(245, 158, 11, 0.28);
			}
		}

		&.is-using-thumbs {
			height: 150px;

			button.plain {
				min-height: 40px;
				overflow: hidden;

				span {
					color: $text-color;
					font-family: inherit;
				}
			}
		}
	}
</style>
