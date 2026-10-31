<script lang="ts">
	import type { GameCompletionStatus, Media, Watched } from "@/types";
	import Icon from "@/lib/Icon.svelte";

	interface Props {
		game: Media;
		watched?: Watched;
		onHoursChanged: (hours: number) => Promise<boolean>;
		onPlatformChanged: (platform: string) => Promise<boolean>;
		onCompletionStatusChanged: (
			status: GameCompletionStatus,
		) => Promise<boolean>;
		onStartTracking: () => Promise<boolean>;
	}

	let {
		game,
		watched,
		onHoursChanged,
		onPlatformChanged,
		onCompletionStatusChanged,
		onStartTracking,
	}: Props = $props();

	// Hours state
	// eslint-disable-next-line svelte/prefer-writable-derived
	let hoursInput = $state("");
	let isSavingHours = $state(false);

	$effect(() => {
		hoursInput =
			typeof watched?.hoursPlayed === "number"
				? watched.hoursPlayed.toString()
				: "";
	});

	// Time-to-Beat targets (converted from minutes to hours)
	let mainStoryHours = $derived.by(() => {
		if (game.timeToBeat?.normally && game.timeToBeat.normally > 0) {
			return Math.round(game.timeToBeat.normally / 60);
		}
		if (game.runtime && game.runtime > 0) {
			return Math.round(game.runtime / 60);
		}
		return undefined;
	});

	let completionistHours = $derived.by(() => {
		if (game.timeToBeat?.completely && game.timeToBeat.completely > 0) {
			return Math.round(game.timeToBeat.completely / 60);
		}
		return undefined;
	});

	let currentHours = $derived(watched?.hoursPlayed || 0);

	let activeTargetHours = $derived.by(() => {
		if (
			watched?.gameCompletionStatus === "COMPLETIONIST" &&
			completionistHours
		) {
			return completionistHours;
		}
		return mainStoryHours;
	});

	let progressPercent = $derived.by(() => {
		if (!activeTargetHours || activeTargetHours <= 0) return 0;
		return Math.min(100, Math.round((currentHours / activeTargetHours) * 100));
	});

	// Clean platforms list
	const defaultPlatforms = [
		"PC / Steam",
		"PlayStation 5",
		"Xbox Series X|S",
		"Nintendo Switch",
		"PlayStation 4",
		"Steam Deck",
	];

	let availablePlatforms = $derived.by(() => {
		if (game.platforms && game.platforms.length > 0) {
			return Array.from(
				new Set(
					game.platforms.map((p) => {
						if (p.includes("Microsoft Windows") || p.toLowerCase() === "pc")
							return "PC / Steam";
						if (p.includes("PlayStation 5")) return "PlayStation 5";
						if (p.includes("PlayStation 4")) return "PlayStation 4";
						if (p.includes("Xbox Series")) return "Xbox Series X|S";
						if (p.includes("Xbox One")) return "Xbox One";
						if (p.includes("Switch")) return "Nintendo Switch";
						return p;
					}),
				),
			).slice(0, 8);
		}
		return defaultPlatforms;
	});

	const completionOptions: {
		id: GameCompletionStatus;
		label: string;
		icon: string;
		desc: string;
	}[] = [
		{
			id: "MAIN_STORY",
			label: "História Principal",
			icon: "🎯",
			desc: "Foco na campanha",
		},
		{
			id: "MAIN_EXTRA",
			label: "História + Extras",
			icon: "⚡",
			desc: "Missões secundárias",
		},
		{
			id: "COMPLETIONIST",
			label: "100% / Platina",
			icon: "🏆",
			desc: "Todos os troféus",
		},
		{
			id: "DROPPED",
			label: "Abandonado",
			icon: "🛑",
			desc: "Pausado ou largado",
		},
	];

	async function adjustHours(delta: number) {
		const base =
			typeof watched?.hoursPlayed === "number" ? watched.hoursPlayed : 0;
		const next = Math.max(0, parseFloat((base + delta).toFixed(1)));
		hoursInput = next.toString();
		isSavingHours = true;
		try {
			await onHoursChanged(next);
		} finally {
			isSavingHours = false;
		}
	}

	async function handleHoursBlur() {
		const parsed = parseFloat(hoursInput);
		if (isNaN(parsed) || parsed < 0) {
			hoursInput = (watched?.hoursPlayed || 0).toString();
			return;
		}
		const rounded = parseFloat(parsed.toFixed(1));
		if (rounded !== watched?.hoursPlayed) {
			isSavingHours = true;
			try {
				await onHoursChanged(rounded);
			} finally {
				isSavingHours = false;
			}
		}
	}

	function handleHoursKeydown(e: KeyboardEvent) {
		if (e.key === "Enter") {
			(e.target as HTMLInputElement).blur();
		}
	}
</script>

<section class="gaming-tracker-card" aria-label="Progresso de Jogo">
	<div class="card-header">
		<div class="header-title-row">
			<span class="game-badge-icon">🎮</span>
			<div>
				<h3 class="card-title">Acompanhamento Gamer</h3>
				<span class="card-subtitle"
					>Horas jogadas, plataforma e tempo médio de conclusão</span
				>
			</div>
		</div>

		{#if !watched}
			<button
				type="button"
				class="start-tracking-btn"
				onclick={onStartTracking}
			>
				<span>+ Começar a Acompanhar</span>
			</button>
		{/if}
	</div>

	<div class="tracker-body">
		<!-- Hours Counter -->
		<div class="tracker-section hours-section">
			<div class="section-label">
				<Icon i="clock" wh={14} />
				<span>Horas Jogadas</span>
			</div>

			<div class="hours-stepper-cluster">
				<button
					type="button"
					class="step-btn"
					onclick={() => adjustHours(-5)}
					disabled={!watched || currentHours < 5 || isSavingHours}
					title="-5 horas"
				>
					-5h
				</button>
				<button
					type="button"
					class="step-btn"
					onclick={() => adjustHours(-1)}
					disabled={!watched || currentHours < 1 || isSavingHours}
					title="-1 hora"
				>
					-1h
				</button>

				<div class="hours-input-wrapper">
					<input
						type="number"
						step="0.5"
						min="0"
						class="hours-input"
						value={hoursInput}
						oninput={(e) => (hoursInput = (e.target as HTMLInputElement).value)}
						onblur={handleHoursBlur}
						onkeydown={handleHoursKeydown}
						disabled={!watched || isSavingHours}
						placeholder="0"
					/>
					<span class="hours-suffix">h</span>
				</div>

				<button
					type="button"
					class="step-btn"
					onclick={() => adjustHours(1)}
					disabled={!watched || isSavingHours}
					title="+1 hora"
				>
					+1h
				</button>
				<button
					type="button"
					class="step-btn"
					onclick={() => adjustHours(5)}
					disabled={!watched || isSavingHours}
					title="+5 horas"
				>
					+5h
				</button>
			</div>
		</div>

		<!-- Time to Beat Comparison Bar -->
		{#if mainStoryHours || completionistHours}
			<div class="tracker-section time-to-beat-section">
				<div class="ttb-header">
					<div class="section-label">
						<Icon i="reel" wh={14} />
						<span>Tempo Médio (Time-to-Beat)</span>
					</div>

					<div class="ttb-targets-row">
						{#if mainStoryHours}
							<span
								class="target-pill"
								class:active-target={watched?.gameCompletionStatus !==
									"COMPLETIONIST"}
							>
								🎯 História: <strong>{mainStoryHours}h</strong>
							</span>
						{/if}
						{#if completionistHours}
							<span
								class="target-pill"
								class:active-target={watched?.gameCompletionStatus ===
									"COMPLETIONIST"}
							>
								🏆 100%: <strong>{completionistHours}h</strong>
							</span>
						{/if}
					</div>
				</div>

				<div
					class="progress-track"
					role="progressbar"
					aria-valuenow={progressPercent}
					aria-valuemin={0}
					aria-valuemax={100}
				>
					<div
						class="progress-fill"
						style={`width: ${progressPercent}%`}
						class:completed={progressPercent >= 100}
					></div>
				</div>

				<div class="progress-meta">
					<span class="progress-pct-label">
						{#if currentHours > 0}
							<strong>{progressPercent}%</strong> completado
							{#if activeTargetHours && currentHours < activeTargetHours}
								(faltam ~{(activeTargetHours - currentHours).toFixed(1)}h)
							{:else if activeTargetHours && currentHours >= activeTargetHours}
								(meta de duração alcançada! 🎉)
							{/if}
						{:else}
							Regista as tuas primeiras horas para acompanhar o progresso
						{/if}
					</span>
				</div>
			</div>
		{/if}

		<!-- Platform Selection Chips -->
		<div class="tracker-section">
			<div class="section-label">
				<Icon i="gamepad" wh={14} />
				<span>Plataforma</span>
			</div>

			<div class="platform-chips-grid">
				{#each availablePlatforms as p (p)}
					{@const isSelected = watched?.platform === p}
					<button
						type="button"
						class="platform-chip"
						class:selected={isSelected}
						disabled={!watched}
						onclick={() => onPlatformChanged(isSelected ? "" : p)}
					>
						<span class="platform-name">{p}</span>
						{#if isSelected}
							<Icon i="check" wh={12} />
						{/if}
					</button>
				{/each}
			</div>
		</div>

		<!-- Completion Statuses -->
		<div class="tracker-section">
			<div class="section-label">
				<Icon i="star" wh={14} />
				<span>Objetivo / Estado de Conclusão</span>
			</div>

			<div class="completion-grid">
				{#each completionOptions as opt (opt.id)}
					{@const isSelected = watched?.gameCompletionStatus === opt.id}
					<button
						type="button"
						class="completion-option-card"
						class:selected={isSelected}
						disabled={!watched}
						onclick={() =>
							onCompletionStatusChanged(
								isSelected ? ("" as GameCompletionStatus) : opt.id,
							)}
					>
						<span class="completion-icon">{opt.icon}</span>
						<div class="completion-text">
							<span class="completion-title">{opt.label}</span>
							<span class="completion-desc">{opt.desc}</span>
						</div>
						{#if isSelected}
							<div class="completion-check">
								<Icon i="check" wh={12} />
							</div>
						{/if}
					</button>
				{/each}
			</div>
		</div>
	</div>
</section>

<style lang="scss">
	.gaming-tracker-card {
		width: 100%;
		max-width: 1200px;
		margin: 20px auto;
		border-radius: 12px;
		background: rgba(15, 20, 30, 0.82);
		border: 1px solid rgba(255, 255, 255, 0.12);
		box-shadow: 0 8px 30px -4px rgba(0, 0, 0, 0.5);
		backdrop-filter: blur(24px) saturate(160%);
		-webkit-backdrop-filter: blur(24px) saturate(160%);
		padding: 24px 28px;
		box-sizing: border-box;

		@media screen and (max-width: 768px) {
			padding: 18px 16px;
			margin: 16px auto;
			border-radius: 16px;
		}

		.card-header {
			display: flex;
			align-items: center;
			justify-content: space-between;
			flex-wrap: wrap;
			gap: 14px;
			padding-bottom: 18px;
			border-bottom: 1px solid rgba(255, 255, 255, 0.08);

			.header-title-row {
				display: flex;
				align-items: center;
				gap: 12px;

				.game-badge-icon {
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

				.card-title {
					margin: 0;
					font-size: 18px;
					font-weight: 700;
					color: #ffffff;
					letter-spacing: -0.01em;
				}

				.card-subtitle {
					font-size: 12px;
					color: rgba(255, 255, 255, 0.6);
				}
			}

			.start-tracking-btn {
				display: inline-flex;
				align-items: center;
				gap: 6px;
				padding: 8px 16px;
				border-radius: 8px;
				border: none;
				background: $brand-color;
				color: #0b0e14;
				font-size: 13px;
				font-weight: 600;
				cursor: pointer;
				box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
				transition:
					background-color 140ms ease,
					opacity 140ms ease;

				&:hover {
					filter: brightness(1.05);
				}
			}
		}

		.tracker-body {
			display: grid;
			grid-template-columns: 1fr;
			gap: 22px;
			margin-top: 20px;

			.tracker-section {
				display: flex;
				flex-direction: column;
				gap: 10px;

				.section-label {
					display: flex;
					align-items: center;
					gap: 6px;
					font-size: 12px;
					font-weight: 700;
					text-transform: uppercase;
					letter-spacing: 0.05em;
					color: rgba(255, 255, 255, 0.7);
				}
			}

			/* Hours Stepper */
			.hours-stepper-cluster {
				display: flex;
				align-items: center;
				gap: 8px;
				max-width: 440px;

				.step-btn {
					padding: 6px 12px;
					height: 38px;
					border-radius: 8px;
					background: rgba(255, 255, 255, 0.06);
					border: 1px solid rgba(255, 255, 255, 0.1);
					color: #ffffff;
					font-size: 12px;
					font-weight: 600;
					cursor: pointer;
					transition: all 130ms ease;

					&:hover:not(:disabled) {
						background: rgba(255, 255, 255, 0.12);
						border-color: rgba(255, 255, 255, 0.2);
						transform: translateY(-1px);
					}

					&:disabled {
						opacity: 0.4;
						cursor: not-allowed;
					}
				}

				.hours-input-wrapper {
					position: relative;
					display: flex;
					align-items: center;
					flex: 1;

					.hours-input {
						width: 100%;
						height: 38px;
						padding: 0 28px 0 12px;
						border-radius: 8px;
						border: 1px solid rgba(255, 255, 255, 0.12);
						background: rgba(0, 0, 0, 0.4);
						color: #fbbf24;
						font-size: 17px;
						font-weight: 700;
						text-align: center;
						outline: none;
						transition: all 140ms ease;

						&:focus {
							border-color: #f59e0b;
							box-shadow: 0 0 0 2px rgba(245, 158, 11, 0.2);
							background: rgba(0, 0, 0, 0.6);
						}

						&:disabled {
							opacity: 0.5;
							cursor: not-allowed;
						}
					}

					.hours-suffix {
						position: absolute;
						right: 12px;
						font-size: 14px;
						font-weight: 700;
						color: rgba(255, 255, 255, 0.5);
						pointer-events: none;
					}
				}
			}

			/* Time to Beat Progress */
			.time-to-beat-section {
				padding: 16px;
				border-radius: 14px;
				background: rgba(0, 0, 0, 0.25);
				border: 1px solid rgba(255, 255, 255, 0.06);

				.ttb-header {
					display: flex;
					align-items: center;
					justify-content: space-between;
					flex-wrap: wrap;
					gap: 10px;

					.ttb-targets-row {
						display: flex;
						align-items: center;
						gap: 8px;

						.target-pill {
							font-size: 11.5px;
							padding: 3px 8px;
							border-radius: 9999px;
							background: rgba(255, 255, 255, 0.05);
							border: 1px solid rgba(255, 255, 255, 0.08);
							color: rgba(255, 255, 255, 0.7);

							&.active-target {
								background: rgba(245, 158, 11, 0.15);
								border-color: rgba(245, 158, 11, 0.4);
								color: #fbbf24;
							}
						}
					}
				}

				.progress-track {
					width: 100%;
					height: 10px;
					margin-top: 10px;
					border-radius: 9999px;
					background: rgba(255, 255, 255, 0.08);
					overflow: hidden;

					.progress-fill {
						height: 100%;
						border-radius: 9999px;
						background: $brand-color;
						transition: width 300ms cubic-bezier(0.16, 1, 0.3, 1);

						&.completed {
							background: $success;
						}
					}
				}

				.progress-meta {
					margin-top: 6px;
					font-size: 11.5px;
					color: rgba(255, 255, 255, 0.65);

					strong {
						color: #ffffff;
					}
				}
			}

			/* Platform Chips */
			.platform-chips-grid {
				display: flex;
				flex-wrap: wrap;
				gap: 8px;

				.platform-chip {
					display: inline-flex;
					align-items: center;
					gap: 6px;
					padding: 7px 14px;
					border-radius: 9999px;
					border: 1px solid rgba(255, 255, 255, 0.1);
					background: rgba(255, 255, 255, 0.04);
					color: rgba(255, 255, 255, 0.8);
					font-size: 12px;
					font-weight: 500;
					cursor: pointer;
					transition: all 140ms ease;

					&:hover:not(:disabled) {
						background: rgba(255, 255, 255, 0.09);
						border-color: rgba(255, 255, 255, 0.2);
						color: #ffffff;
					}

					&.selected {
						background: rgba(245, 158, 11, 0.16);
						border-color: #f59e0b;
						color: #fbbf24;
						font-weight: 600;
						box-shadow: 0 0 12px rgba(245, 158, 11, 0.2);
					}

					&:disabled {
						opacity: 0.5;
						cursor: not-allowed;
					}
				}
			}

			/* Completion Grid */
			.completion-grid {
				display: grid;
				grid-template-columns: repeat(auto-fit, minmax(210px, 1fr));
				gap: 10px;

				.completion-option-card {
					display: flex;
					align-items: center;
					gap: 10px;
					padding: 10px 14px;
					border-radius: 12px;
					border: 1px solid rgba(255, 255, 255, 0.09);
					background: rgba(255, 255, 255, 0.035);
					cursor: pointer;
					text-align: left;
					transition: all 140ms ease;

					&:hover:not(:disabled) {
						background: rgba(255, 255, 255, 0.08);
						border-color: rgba(255, 255, 255, 0.18);
						transform: translateY(-1px);
					}

					&.selected {
						background: rgba(245, 158, 11, 0.14);
						border-color: rgba(245, 158, 11, 0.45);
						box-shadow: 0 0 16px rgba(245, 158, 11, 0.15);

						.completion-title {
							color: #fbbf24;
						}
					}

					&:disabled {
						opacity: 0.5;
						cursor: not-allowed;
					}

					.completion-icon {
						font-size: 20px;
						flex-shrink: 0;
					}

					.completion-text {
						display: flex;
						flex-direction: column;
						gap: 2px;
						flex: 1;

						.completion-title {
							font-size: 13px;
							font-weight: 600;
							color: #ffffff;
							transition: color 130ms ease;
						}

						.completion-desc {
							font-size: 10.5px;
							color: rgba(255, 255, 255, 0.5);
						}
					}

					.completion-check {
						display: flex;
						align-items: center;
						justify-content: center;
						color: #fbbf24;
						flex-shrink: 0;
					}
				}
			}
		}
	}
</style>
