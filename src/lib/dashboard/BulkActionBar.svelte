<script lang="ts">
	import type { WatchedStatus } from "@/types";
	import { RatingSystem } from "@/types";
	import { store } from "@/store.svelte";
	import { t } from "@/lib/i18n";
	import Icon from "../Icon.svelte";

	interface Props {
		selectedCount: number;
		totalCount?: number;
		onSelectAll: () => void;
		onSelectAllEverywhere?: () => void;
		onClear: () => void;
		onUpdateStatus: (status: WatchedStatus) => void;
		onUpdateRating?: (rating: number) => void;
		onDelete: () => void;
		onClose?: () => void;
	}

	let {
		selectedCount,
		totalCount,
		onSelectAll,
		onSelectAllEverywhere,
		onClear,
		onUpdateStatus,
		onUpdateRating,
		onDelete,
		onClose,
	}: Props = $props();

	let statusMenuOpen = $state(false);
	let rateMenuOpen = $state(false);
	let isConfirmingDelete = $state(false);
	let hoveredScore = $state<number | undefined>();
	let statusWrapEl: HTMLDivElement | undefined = $state();
	let rateWrapEl: HTMLDivElement | undefined = $state();

	function handleWindowClick(e: MouseEvent) {
		const target = e.target as Node;
		if (statusMenuOpen && statusWrapEl && !statusWrapEl.contains(target)) {
			statusMenuOpen = false;
		}
		if (rateMenuOpen && rateWrapEl && !rateWrapEl.contains(target)) {
			rateMenuOpen = false;
		}
		if (isConfirmingDelete) {
			isConfirmingDelete = false;
		}
	}

	function handleWindowKeyDown(e: KeyboardEvent) {
		if (e.key === "Escape") {
			if (isConfirmingDelete) {
				isConfirmingDelete = false;
				return;
			}
			if (statusMenuOpen || rateMenuOpen) {
				statusMenuOpen = false;
				rateMenuOpen = false;
			} else if (onClose) {
				onClose();
			}
		}
	}

	function handleDeleteClick(e: MouseEvent) {
		e.stopPropagation();
		if (!isConfirmingDelete) {
			isConfirmingDelete = true;
			return;
		}
		isConfirmingDelete = false;
		onDelete();
	}

	function selectStatus(s: WatchedStatus) {
		statusMenuOpen = false;
		onUpdateStatus(s);
	}

	function applyRating(r: number) {
		rateMenuOpen = false;
		hoveredScore = undefined;
		onUpdateRating?.(r);
	}

	const isThumbs = $derived(
		store.userSettings?.ratingSystem === RatingSystem.Thumbs,
	);
	const isOutOf5 = $derived(
		store.userSettings?.ratingSystem === RatingSystem.OutOf5,
	);
	const isOutOf100 = $derived(
		store.userSettings?.ratingSystem === RatingSystem.OutOf100,
	);

	const starCount = $derived(isOutOf5 ? 5 : 10);
	const starList = $derived(Array.from({ length: starCount }, (_, i) => i + 1));
</script>

<svelte:window onclick={handleWindowClick} onkeydown={handleWindowKeyDown} />

{#if selectedCount > 0}
	<aside class="dock-shell" aria-label="Bulk actions dock">
		<div class="dock-island">
			<!-- Selection Count & Quick Toggle -->
			<div class="dock-group selection-group">
				<div class="count-tag">
					<span class="num">{selectedCount}</span>
					<span class="lbl"
						>{selectedCount === 1
							? t("dashboard.bulk_selected_one")
							: t("dashboard.bulk_selected_many")}</span
					>
				</div>
				<button
					class="ghost-btn"
					onclick={onSelectAll}
					title={t("dashboard.bulk_loaded_title")}
				>
					{t("dashboard.bulk_loaded")}
				</button>
				{#if totalCount && totalCount > selectedCount && onSelectAllEverywhere}
					<span class="bullet">•</span>
					<button
						class="ghost-btn all-bib-btn"
						onclick={onSelectAllEverywhere}
						title={t("dashboard.bulk_all_title", { count: totalCount })}
					>
						{t("dashboard.bulk_all", { count: totalCount })}
					</button>
				{/if}
				<span class="bullet">•</span>
				<button
					class="ghost-btn"
					onclick={onClear}
					title={t("dashboard.bulk_clear_title")}
				>
					{t("dashboard.bulk_clear")}
				</button>
			</div>

			<div class="dock-sep"></div>

			<!-- Primary Actions -->
			<div class="dock-group actions-group">
				<!-- Status Popover Anchor -->
				<div class="popover-anchor" bind:this={statusWrapEl}>
					<button
						class="dock-pill-btn"
						class:active={statusMenuOpen}
						onclick={(e) => {
							e.stopPropagation();
							rateMenuOpen = false;
							statusMenuOpen = !statusMenuOpen;
						}}
						aria-haspopup="true"
						aria-expanded={statusMenuOpen}
					>
						<Icon i="clock" wh={13} />
						<span>{t("dashboard.bulk_status")}</span>
						<Icon i="chevron" facing={statusMenuOpen ? "up" : "down"} wh={10} />
					</button>

					{#if statusMenuOpen}
						<div class="dock-flyout status-flyout" role="menu">
							<button
								class="flyout-row row-watching"
								onclick={() => selectStatus("WATCHING")}
								role="menuitem"
							>
								<span class="dot-indicator watching"></span>
								<Icon i="clock" wh={13} />
								<span>{t("status.watching")}</span>
							</button>
							<button
								class="flyout-row row-finished"
								onclick={() => selectStatus("FINISHED")}
								role="menuitem"
							>
								<span class="dot-indicator finished"></span>
								<Icon i="check" wh={13} />
								<span>{t("status.finished")}</span>
							</button>
							<button
								class="flyout-row row-planned"
								onclick={() => selectStatus("PLANNED")}
								role="menuitem"
							>
								<span class="dot-indicator planned"></span>
								<Icon i="calendar" wh={13} />
								<span>{t("status.planned")}</span>
							</button>
							<button
								class="flyout-row row-hold"
								onclick={() => selectStatus("HOLD")}
								role="menuitem"
							>
								<span class="dot-indicator hold"></span>
								<Icon i="pause" wh={13} />
								<span>{t("status.hold")}</span>
							</button>
							<button
								class="flyout-row row-dropped"
								onclick={() => selectStatus("DROPPED")}
								role="menuitem"
							>
								<span class="dot-indicator dropped"></span>
								<Icon i="thumb-down" wh={13} />
								<span>{t("status.dropped")}</span>
							</button>
							<button
								class="flyout-row row-cancelled"
								onclick={() => selectStatus("CANCELLED")}
								role="menuitem"
							>
								<span class="dot-indicator cancelled"></span>
								<Icon i="close-circle" wh={13} />
								<span>{t("status.cancelled")}</span>
							</button>
						</div>
					{/if}
				</div>

				<!-- Rating Popover Anchor -->
				{#if onUpdateRating}
					<div class="popover-anchor" bind:this={rateWrapEl}>
						<button
							class="dock-pill-btn btn-rate"
							class:active={rateMenuOpen}
							onclick={(e) => {
								e.stopPropagation();
								statusMenuOpen = false;
								rateMenuOpen = !rateMenuOpen;
							}}
							aria-haspopup="true"
							aria-expanded={rateMenuOpen}
						>
							<Icon i="star" wh={13} />
							<span>{t("dashboard.bulk_rate")}</span>
						</button>

						{#if rateMenuOpen}
							<div
								class="dock-flyout rate-flyout"
								role="dialog"
								aria-label={t("dashboard.bulk_rate_items", { count: selectedCount })}
							>
								<div class="rate-popover-header">
									<span>{t("dashboard.bulk_rate_items", { count: selectedCount })}</span>
									{#if hoveredScore}
										<span class="live-score">
											★ {isOutOf100 ? hoveredScore * 10 : hoveredScore}
										</span>
									{/if}
								</div>

								{#if isThumbs}
									<div class="thumbs-row">
										<button
											class="thumb-btn dislike"
											onclick={() => applyRating(1)}
											title="Dislike"
										>
											<Icon i="thumb-down" wh={18} />
											<span>Dislike</span>
										</button>
										<button
											class="thumb-btn meh"
											onclick={() => applyRating(5)}
											title="Mediocre"
										>
											<span class="dash-icon">-</span>
											<span>Mediocre</span>
										</button>
										<button
											class="thumb-btn like"
											onclick={() => applyRating(9)}
											title="Like"
										>
											<Icon i="thumb-up" wh={18} />
											<span>Like</span>
										</button>
									</div>
								{:else}
									<div class="stars-picker-row">
										{#each starList as s (s)}
											{@const val = isOutOf5 ? s * 2 : s}
											<button
												class="star-btn"
												class:highlighted={hoveredScore
													? s <= hoveredScore
													: false}
												onmouseenter={() => (hoveredScore = s)}
												onmouseleave={() => (hoveredScore = undefined)}
												onclick={() => applyRating(val)}
												title={`Score: ${isOutOf100 ? s * 10 : s}`}
												aria-label={`Rate ${s}`}
											>
												<Icon i="star" wh={16} />
												<span class="star-num">{isOutOf100 ? s * 10 : s}</span>
											</button>
										{/each}
									</div>
								{/if}
							</div>
						{/if}
					</div>
				{/if}

				<!-- Remove Button -->
				<button
					class="dock-pill-btn btn-delete"
					class:confirming={isConfirmingDelete}
					onclick={handleDeleteClick}
					title={isConfirmingDelete
						? t("dashboard.bulk_delete_confirm_title")
						: t("action.delete")}
				>
					<Icon i="trash" wh={13} />
					<span
						>{isConfirmingDelete
							? t("dashboard.bulk_delete_confirm", { count: selectedCount })
							: t("action.delete")}</span
					>
				</button>
			</div>

			<!-- Exit Bulk Mode Button -->
			{#if onClose}
				<div class="dock-sep"></div>
				<button
					class="dock-close-btn"
					onclick={onClose}
					title="Exit selection mode (Esc)"
					aria-label="Exit selection mode"
				>
					<Icon i="close" wh={14} />
				</button>
			{/if}
		</div>
	</aside>
{/if}

<style lang="scss">
	@keyframes dockSlide {
		from {
			opacity: 0;
			transform: translate(-50%, 14px) scale(0.97);
		}
		to {
			opacity: 1;
			transform: translate(-50%, 0) scale(1);
		}
	}

	@keyframes popoverSpring {
		from {
			opacity: 0;
			transform: translate(-50%, 8px) scale(0.95);
		}
		to {
			opacity: 1;
			transform: translate(-50%, 0) scale(1);
		}
	}

	@keyframes pulseDanger {
		0%, 100% {
			box-shadow: 0 0 0 0 rgba(239, 68, 68, 0.6);
		}
		50% {
			box-shadow: 0 0 0 5px rgba(239, 68, 68, 0);
		}
	}

	.dock-shell {
		position: fixed;
		bottom: 24px;
		left: 50%;
		transform: translate(-50%, 0);
		z-index: 9999;
		pointer-events: auto;
		animation: dockSlide 240ms cubic-bezier(0.16, 1, 0.3, 1) forwards;

		@media screen and (max-width: 640px) {
			bottom: calc(72px + env(safe-area-inset-bottom));
			width: calc(100vw - 24px);
			max-width: 440px;
		}
	}

	.dock-island {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 5px 8px 5px 14px;
		background: rgba(13, 17, 24, 0.88);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 9999px;
		box-shadow:
			0 16px 36px -4px rgba(0, 0, 0, 0.65),
			0 4px 12px rgba(0, 0, 0, 0.4),
			inset 0 1px 0 rgba(255, 255, 255, 0.1);
		backdrop-filter: blur(20px) saturate(180%);
		-webkit-backdrop-filter: blur(20px) saturate(180%);
		width: max-content;
		max-width: 100%;

		@media screen and (max-width: 640px) {
			border-radius: 16px;
			padding: 8px 10px;
			justify-content: space-between;
			width: 100%;
			flex-wrap: wrap;
			gap: 6px;
		}

		.dock-group {
			display: flex;
			align-items: center;
			gap: 6px;
		}

		.selection-group {
			.count-tag {
				display: inline-flex;
				align-items: center;
				gap: 5px;
				margin-right: 4px;

				.num {
					background: rgba(99, 102, 241, 0.2);
					color: #c7d2fe;
					border: 1px solid rgba(99, 102, 241, 0.35);
					font-size: 11px;
					font-weight: 700;
					min-width: 18px;
					height: 18px;
					padding: 0 5px;
					border-radius: 9999px;
					display: inline-flex;
					align-items: center;
					justify-content: center;
					font-variant-numeric: tabular-nums;
				}

				.lbl {
					font-size: 12px;
					font-weight: 500;
					color: #94a3b8;
					white-space: nowrap;
				}
			}

			.ghost-btn {
				background: transparent;
				border: none;
				padding: 2px 4px;
				color: #64748b;
				font-size: 11.5px;
				font-weight: 600;
				cursor: pointer;
				transition: color 120ms ease;

				&:hover {
					color: #ffffff;
				}

				&.all-bib-btn {
					color: #818cf8;
					font-weight: 700;
					&:hover {
						color: #a5b4fc;
						text-decoration: underline;
					}
				}
			}

			.bullet {
				color: rgba(255, 255, 255, 0.2);
				font-size: 9px;
			}
		}

		.dock-sep {
			width: 1px;
			height: 18px;
			background: rgba(255, 255, 255, 0.08);
			flex-shrink: 0;

			@media screen and (max-width: 640px) {
				display: none;
			}
		}

		.actions-group {
			display: flex;
			align-items: center;
			gap: 5px;
		}

		.dock-pill-btn {
			display: inline-flex;
			align-items: center;
			gap: 5px;
			height: 30px;
			padding: 0 11px;
			border-radius: 9999px;
			font-size: 12px;
			font-weight: 500;
			cursor: pointer;
			border: 1px solid rgba(255, 255, 255, 0.08);
			background: rgba(255, 255, 255, 0.04);
			color: #e2e8f0;
			transition: all 130ms cubic-bezier(0.16, 1, 0.3, 1);
			white-space: nowrap;
			width: auto;

			&:hover {
				background: rgba(255, 255, 255, 0.1);
				color: #ffffff;
				border-color: rgba(255, 255, 255, 0.16);
				transform: translateY(-1px);
			}

			&:active {
				transform: scale(0.97);
			}

			&.active {
				background: rgba(255, 255, 255, 0.14);
				border-color: rgba(255, 255, 255, 0.25);
				color: #ffffff;
			}

			&.btn-rate {
				color: #fde047;
				background: rgba(234, 179, 8, 0.08);
				border-color: rgba(234, 179, 8, 0.2);

				&:hover,
				&.active {
					background: rgba(234, 179, 8, 0.2);
					border-color: rgba(234, 179, 8, 0.4);
					color: #fef08a;
				}
			}

			&.btn-delete {
				color: #fca5a5;
				background: rgba(239, 68, 68, 0.08);
				border-color: rgba(239, 68, 68, 0.2);

				&:hover {
					background: #ef4444;
					border-color: #ef4444;
					color: #ffffff;
				}

				&.confirming {
					background: #dc2626 !important;
					border-color: #f87171 !important;
					color: #ffffff !important;
					font-weight: 700 !important;
					animation: pulseDanger 1.2s infinite ease-in-out;
				}
			}
		}

		.popover-anchor {
			position: relative;
		}

		.dock-flyout {
			position: absolute;
			bottom: calc(100% + 10px);
			left: 50%;
			transform: translate(-50%, 0);
			background: rgba(13, 17, 24, 0.94);
			border: 1px solid rgba(255, 255, 255, 0.12);
			border-radius: 14px;
			box-shadow:
				0 20px 40px -4px rgba(0, 0, 0, 0.7),
				0 0 0 1px rgba(255, 255, 255, 0.05);
			backdrop-filter: blur(24px) saturate(180%);
			-webkit-backdrop-filter: blur(24px) saturate(180%);
			z-index: 10001;
			animation: popoverSpring 180ms cubic-bezier(0.16, 1, 0.3, 1) forwards;
		}

		.status-flyout {
			padding: 4px;
			display: flex;
			flex-direction: column;
			gap: 1px;
			min-width: 155px;

			.flyout-row {
				display: flex;
				align-items: center;
				gap: 8px;
				padding: 7px 10px;
				border-radius: 8px;
				border: none;
				background: transparent;
				color: #cbd5e1;
				font-size: 12.5px;
				font-weight: 500;
				text-align: left;
				cursor: pointer;
				width: 100%;
				transition: all 110ms ease;

				.dot-indicator {
					width: 6px;
					height: 6px;
					border-radius: 50%;
					flex-shrink: 0;

					&.watching {
						background: #60a5fa;
					}
					&.finished {
						background: #34d399;
					}
					&.planned {
						background: #fbbf24;
					}
					&.hold {
						background: #c084fc;
					}
					&.dropped {
						background: #94a3b8;
					}
					&.cancelled {
						background: #f87171;
					}
				}

				&:hover {
					background: rgba(255, 255, 255, 0.08);
					color: #ffffff;
					transform: translateX(2px);
				}
			}
		}

		.rate-flyout {
			padding: 10px 14px;
			display: flex;
			flex-direction: column;
			align-items: center;
			gap: 8px;
			width: max-content;

			.rate-popover-header {
				display: flex;
				align-items: center;
				justify-content: space-between;
				width: 100%;
				gap: 12px;
				font-size: 11px;
				font-weight: 600;
				color: #94a3b8;
				text-transform: uppercase;
				letter-spacing: 0.05em;

				.live-score {
					color: #fde047;
					font-weight: 700;
				}
			}

			.thumbs-row {
				display: flex;
				align-items: center;
				gap: 8px;

				.thumb-btn {
					display: flex;
					flex-direction: column;
					align-items: center;
					justify-content: center;
					gap: 4px;
					width: 64px;
					height: 52px;
					border-radius: 10px;
					border: 1px solid rgba(255, 255, 255, 0.08);
					background: rgba(255, 255, 255, 0.04);
					color: #cbd5e1;
					cursor: pointer;
					transition: all 130ms ease;

					span {
						font-size: 11px;
						font-weight: 600;
					}

					.dash-icon {
						font-size: 24px;
						line-height: 14px;
						font-family: "Shrikhand", cursive;
					}

					&:hover {
						transform: translateY(-2px);
					}

					&.dislike:hover {
						background: rgba(239, 68, 68, 0.15);
						border-color: rgba(239, 68, 68, 0.35);
						color: #fca5a5;
					}
					&.meh:hover {
						background: rgba(148, 163, 184, 0.15);
						border-color: rgba(148, 163, 184, 0.35);
						color: #ffffff;
					}
					&.like:hover {
						background: rgba(16, 185, 129, 0.15);
						border-color: rgba(16, 185, 129, 0.35);
						color: #6ee7b7;
					}
				}
			}

			.stars-picker-row {
				display: flex;
				align-items: center;
				gap: 2px;

				.star-btn {
					display: flex;
					flex-direction: column;
					align-items: center;
					gap: 2px;
					padding: 6px 4px;
					border-radius: 6px;
					border: none;
					background: transparent;
					color: rgba(255, 255, 255, 0.3);
					cursor: pointer;
					transition: all 110ms ease;

					.star-num {
						font-size: 9.5px;
						font-weight: 700;
						color: rgba(255, 255, 255, 0.45);
						font-variant-numeric: tabular-nums;
					}

					&.highlighted,
					&:hover {
						color: #fde047;
						transform: scale(1.15);

						.star-num {
							color: #fef08a;
						}
					}
				}
			}
		}

		.dock-close-btn {
			display: inline-flex;
			align-items: center;
			justify-content: center;
			width: 24px;
			height: 24px;
			border-radius: 50%;
			background: transparent;
			border: none;
			color: #64748b;
			cursor: pointer;
			transition: all 120ms ease;
			flex-shrink: 0;

			&:hover {
				color: #ffffff;
				background: rgba(255, 255, 255, 0.08);
			}
		}
	}
</style>
