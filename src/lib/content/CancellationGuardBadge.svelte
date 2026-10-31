<script lang="ts">
	import Icon from "../Icon.svelte";
	import tooltip from "../actions/tooltip";
	import { req } from "../util/api";
	import { notify } from "../util/notify";
	import type { CancellationInfo, CancellationStatus } from "@/types";

	interface Props {
		showId?: number;
		status?: string;
		cancellation?: CancellationInfo;
		onOverrideChanged?: (info: CancellationInfo) => void;
	}

	let { showId, status, cancellation, onOverrideChanged }: Props = $props();

	let overrideOpen = $state(false);
	let saving = $state(false);

	const activeStatus: CancellationStatus | "unknown" = $derived(
		cancellation?.status ?? "unknown",
	);

	const badgeConfig = $derived.by(() => {
		switch (activeStatus) {
			case "cliffhanger":
				return {
					label: cancellation?.badgeText || "Cancelada sem Desfecho",
					type: "cliffhanger",
					icon: "close-circle" as const,
					bg: "rgba(225, 29, 72, 0.2)",
					border: "rgba(225, 29, 72, 0.45)",
					color: "#fb7185",
				};
			case "canceled":
				return {
					label: cancellation?.badgeText || "Cancelada",
					type: "canceled",
					icon: "close-circle" as const,
					bg: "rgba(239, 68, 68, 0.16)",
					border: "rgba(239, 68, 68, 0.35)",
					color: "#f87171",
				};
			case "hiatus":
				return {
					label: cancellation?.badgeText || "Hiato Prolongado (+2 anos)",
					type: "hiatus",
					icon: "clock" as const,
					bg: "rgba(245, 158, 11, 0.18)",
					border: "rgba(245, 158, 11, 0.4)",
					color: "#fbbf24",
				};
			case "concluded":
				return {
					label: cancellation?.badgeText || "História Concluída",
					type: "concluded",
					icon: "check" as const,
					bg: "rgba(16, 185, 129, 0.18)",
					border: "rgba(16, 185, 129, 0.4)",
					color: "#34d399",
				};
			case "returning":
				return {
					label: cancellation?.badgeText || "Em Exibição",
					type: "returning",
					icon: "sparkles" as const,
					bg: "rgba(99, 102, 241, 0.15)",
					border: "rgba(99, 102, 241, 0.3)",
					color: "#c7d2fe",
				};
			default:
				return {
					label: status || "Desconhecido",
					type: "unknown",
					icon: undefined,
					bg: "rgba(255, 255, 255, 0.08)",
					border: "rgba(255, 255, 255, 0.15)",
					color: "#cbd5e1",
				};
		}
	});

	async function setEndingType(type: string) {
		if (!showId || saving) return;
		saving = true;
		try {
			await req.post(`/content/tv/${showId}/ending-type`, {
				endingType: type,
			});
			notify({
				text: "Classificação de encerramento atualizada com sucesso!",
				type: "success",
			});
			overrideOpen = false;

			let updated: CancellationInfo;
			if (type === "cliffhanger") {
				updated = {
					status: "cliffhanger",
					badgeText: "Cancelada sem Desfecho",
					description:
						"Marcada manualmente como cancelada sem conclusão narrativa.",
					warning: "Atenção: Série com gancho em aberto (cliffhanger).",
					hasCliffhanger: true,
				};
			} else if (type === "conclusive") {
				updated = {
					status: "concluded",
					badgeText: "História Concluída",
					description:
						"Marcada manualmente como história concluída de forma planeada.",
					hasCliffhanger: false,
				};
			} else if (type === "canceled") {
				updated = {
					status: "canceled",
					badgeText: "Cancelada pela Produção",
					description: "Marcada manualmente como cancelada pela produção.",
					hasCliffhanger: false,
				};
			} else if (type === "hiatus") {
				updated = {
					status: "hiatus",
					badgeText: "Hiato Prolongado",
					description: "Marcada manualmente como em hiato prolongado.",
					hasCliffhanger: false,
				};
			} else {
				updated = {
					status: "unknown",
					badgeText: status || "Automático",
					description: "Baseado em heurística automática do TMDB.",
					hasCliffhanger: false,
				};
			}
			onOverrideChanged?.(updated);
		} catch (e) {
			notify({
				text: "Erro ao atualizar a classificação de encerramento.",
				type: "error",
			});
		} finally {
			saving = false;
		}
	}
</script>

<div class="cancellation-guard-wrapper">
	<div class="badge-row">
		<span
			class="cancellation-badge {badgeConfig.type}"
			style="background: {badgeConfig.bg}; border-color: {badgeConfig.border}; color: {badgeConfig.color};"
			use:tooltip={{
				text: cancellation?.description || badgeConfig.label,
				pos: "top",
			}}
		>
			{#if badgeConfig.icon}
				<Icon i={badgeConfig.icon} wh={13} />
			{/if}
			<span>{badgeConfig.label}</span>
		</span>

		{#if showId}
			<button
				class="override-trigger"
				onclick={() => (overrideOpen = !overrideOpen)}
				use:tooltip={{
					text: "Classificar desfecho manualmente",
					pos: "top",
				}}
			>
				<Icon i="pencil" wh={11} />
				<span>Classificar</span>
			</button>
		{/if}
	</div>

	{#if overrideOpen}
		<div class="override-popover">
			<div class="popover-header">
				<span>Classificação de Desfecho:</span>
				<button class="close-btn" onclick={() => (overrideOpen = false)}>
					<Icon i="close" wh={13} />
				</button>
			</div>
			<div class="popover-options">
				<button
					class:active={activeStatus === "concluded"}
					onclick={() => setEndingType("conclusive")}
				>
					<Icon i="check" wh={13} />
					<span>História Concluída (Conclusiva)</span>
				</button>
				<button
					class:active={activeStatus === "cliffhanger"}
					onclick={() => setEndingType("cliffhanger")}
				>
					<Icon i="close-circle" wh={13} />
					<span>Cancelada sem Desfecho (Cliffhanger)</span>
				</button>
				<button
					class:active={activeStatus === "canceled"}
					onclick={() => setEndingType("canceled")}
				>
					<Icon i="close-circle" wh={13} />
					<span>Cancelada pela Produção</span>
				</button>
				<button
					class:active={activeStatus === "hiatus"}
					onclick={() => setEndingType("hiatus")}
				>
					<Icon i="clock" wh={13} />
					<span>Hiato Prolongado (+2 anos)</span>
				</button>
				<button onclick={() => setEndingType("unknown")}>
					<Icon i="refresh" wh={13} />
					<span>Automático (Heurística TMDB)</span>
				</button>
			</div>
		</div>
	{/if}

	{#if activeStatus === "cliffhanger" || cancellation?.hasCliffhanger}
		<div class="guard-banner alert-cliffhanger">
			<div class="banner-icon">
				<Icon i="close-circle" wh={17} />
			</div>
			<div class="banner-text">
				<strong class="banner-title">
					{cancellation?.warning ||
						"Atenção: Cancelada com Gancho Aberto (Cliffhanger)"}
				</strong>
				{#if cancellation?.description}
					<p class="banner-subtext">{cancellation.description}</p>
				{/if}
			</div>
		</div>
	{:else if activeStatus === "hiatus"}
		<div class="guard-banner alert-hiatus">
			<div class="banner-icon">
				<Icon i="clock" wh={17} />
			</div>
			<div class="banner-text">
				<strong class="banner-title">
					Hiato Prolongado (+2 anos sem novos episódios)
				</strong>
				{#if cancellation?.description}
					<p class="banner-subtext">{cancellation.description}</p>
				{/if}
			</div>
		</div>
	{/if}
</div>

<style lang="scss">
	.cancellation-guard-wrapper {
		display: flex;
		flex-direction: column;
		gap: 8px;
		margin-bottom: 12px;
		width: 100%;
		max-width: 650px;
	}

	.badge-row {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 8px;

		.cancellation-badge {
			display: inline-flex;
			align-items: center;
			gap: 6px;
			font-size: 11.5px;
			font-weight: 700;
			letter-spacing: 0.03em;
			text-transform: uppercase;
			padding: 4px 10px;
			border-radius: 999px;
			border: 1px solid transparent;
			cursor: default;
			transition: all 0.2s ease;
		}

		.override-trigger {
			display: inline-flex;
			align-items: center;
			gap: 4px;
			font-size: 11px;
			font-weight: 600;
			padding: 3px 8px;
			border-radius: 6px;
			background: rgba(255, 255, 255, 0.06);
			border: 1px solid rgba(255, 255, 255, 0.12);
			color: rgba(255, 255, 255, 0.65);
			cursor: pointer;
			transition: all 0.15s ease;

			&:hover {
				background: rgba(255, 255, 255, 0.12);
				color: #fff;
			}
		}
	}

	.override-popover {
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding: 12px;
		border-radius: 12px;
		background: #181d28;
		border: 1px solid rgba(255, 255, 255, 0.12);
		box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.5);
		width: fit-content;
		min-width: 280px;
		z-index: 20;

		.popover-header {
			display: flex;
			align-items: center;
			justify-content: space-between;
			font-size: 11.5px;
			font-weight: 700;
			color: rgba(255, 255, 255, 0.7);
			text-transform: uppercase;
			letter-spacing: 0.5px;
			margin-bottom: 4px;

			.close-btn {
				background: transparent;
				border: none;
				color: rgba(255, 255, 255, 0.5);
				cursor: pointer;
				padding: 2px;

				&:hover {
					color: #fff;
				}
			}
		}

		.popover-options {
			display: flex;
			flex-direction: column;
			gap: 4px;

			button {
				display: flex;
				align-items: center;
				gap: 8px;
				padding: 7px 10px;
				font-size: 12px;
				font-weight: 500;
				color: #e2e8f0;
				background: rgba(255, 255, 255, 0.04);
				border: 1px solid rgba(255, 255, 255, 0.06);
				border-radius: 6px;
				cursor: pointer;
				text-align: left;
				transition: all 0.12s ease;

				&:hover {
					background: rgba(255, 255, 255, 0.1);
					color: #fff;
				}

				&.active {
					background: rgba(99, 102, 241, 0.2);
					border-color: rgba(99, 102, 241, 0.5);
					color: #a5b4fc;
					font-weight: 600;
				}
			}
		}
	}

	.guard-banner {
		display: flex;
		align-items: flex-start;
		gap: 10px;
		padding: 10px 14px;
		border-radius: 10px;
		width: 100%;
		box-sizing: border-box;

		.banner-icon {
			flex-shrink: 0;
			margin-top: 1px;
		}

		.banner-text {
			display: flex;
			flex-direction: column;
			gap: 3px;

			.banner-title {
				font-size: 13px;
				font-weight: 700;
				line-height: 1.3;
			}

			.banner-subtext {
				margin: 0;
				font-size: 12px;
				line-height: 1.4;
				opacity: 0.9;
			}
		}

		&.alert-cliffhanger {
			background: rgba(225, 29, 72, 0.16);
			border: 1px solid rgba(225, 29, 72, 0.4);
			color: #ffe4e6;

			.banner-icon {
				color: #fb7185;
			}

			.banner-title {
				color: #fff;
			}
		}

		&.alert-hiatus {
			background: rgba(245, 158, 11, 0.14);
			border: 1px solid rgba(245, 158, 11, 0.35);
			color: #fef3c7;

			.banner-icon {
				color: #fbbf24;
			}

			.banner-title {
				color: #fff;
			}
		}
	}
</style>
