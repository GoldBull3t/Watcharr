<script lang="ts">
	import Modal from "@/lib/Modal.svelte";
	import Icon from "@/lib/Icon.svelte";
	import SpinnerTiny from "@/lib/SpinnerTiny.svelte";
	import Checkbox from "@/lib/Checkbox.svelte";
	import { store } from "@/store.svelte";
	import { req, updateUserSetting } from "@/lib/util/api";
	import { notify } from "@/lib/util/notify";
	import { ReqerError } from "@/lib/util/fetch";

	interface Props {
		open?: boolean;
		onClose?: () => void;
	}

	let { open = $bindable(false), onClose = undefined }: Props = $props();

	let settings = $derived(store.userSettings);

	let webhookUrl = $state(store.userSettings?.calendarWebhookUrl ?? "");
	let isEnabled = $state(store.userSettings?.calendarWebhookEnabled ?? false);
	let filterType: "all" | "movie" | "tv" = $state(
		(store.userSettings?.calendarWebhookFilter as "all" | "movie" | "tv") ?? "all",
	);

	let isTesting = $state(false);
	let isSending = $state(false);
	let isSaving = $state(false);
	let statusMessage = $state<{ type: "success" | "error"; text: string } | null>(null);

	// Sync local states if store changes
	$effect(() => {
		if (settings?.calendarWebhookUrl !== undefined && !isSaving) {
			webhookUrl = settings.calendarWebhookUrl ?? "";
		}
		if (settings?.calendarWebhookEnabled !== undefined) {
			isEnabled = settings.calendarWebhookEnabled ?? false;
		}
		if (settings?.calendarWebhookFilter !== undefined) {
			filterType = (settings.calendarWebhookFilter as "all" | "movie" | "tv") ?? "all";
		}
	});

	let detectedPlatform = $derived.by(() => {
		const u = webhookUrl.toLowerCase().trim();
		if (u.includes("discord.com") || u.includes("discordapp.com")) return "discord";
		if (u.includes("slack.com")) return "slack";
		if (u.includes("ntfy.sh")) return "ntfy";
		if (u.startsWith("http://") || u.startsWith("https://")) return "generic";
		return "unknown";
	});

	function handleClose() {
		open = false;
		if (onClose) onClose();
	}

	function saveUrl() {
		const trimmed = webhookUrl.trim();
		isSaving = true;
		updateUserSetting("calendarWebhookUrl", trimmed, () => {
			isSaving = false;
		});
	}

	function handleToggleEnabled(val: boolean) {
		isEnabled = val;
		updateUserSetting("calendarWebhookEnabled", val);
	}

	function handleFilterChange(f: "all" | "movie" | "tv") {
		filterType = f;
		updateUserSetting("calendarWebhookFilter", f);
	}

	async function testWebhook() {
		const target = webhookUrl.trim();
		if (!target) {
			notify({
				type: "error",
				text: "Por favor insere um Webhook URL antes de testar.",
				time: 4000,
			});
			return;
		}

		try {
			isTesting = true;
			statusMessage = null;
			saveUrl();

			const res = await req.post<{ success: boolean; message: string }>(
				"/calendar/webhook/test",
				{ url: target },
			);

			statusMessage = {
				type: "success",
				text: res.message || "Notificação de teste enviada com sucesso!",
			};
			notify({
				type: "success",
				text: "Teste enviado com sucesso!",
				time: 4000,
			});
		} catch (err) {
			const errMsg = ReqerError.getMsg(err, "Falha ao enviar notificação de teste.");
			statusMessage = {
				type: "error",
				text: errMsg,
			};
			notify({
				type: "error",
				text: errMsg,
				time: 6000,
			});
		} finally {
			isTesting = false;
		}
	}

	async function sendTodayReleases() {
		const target = webhookUrl.trim();
		if (!target) {
			notify({
				type: "error",
				text: "Configura primeiro o Webhook URL.",
				time: 4000,
			});
			return;
		}

		try {
			isSending = true;
			statusMessage = null;
			saveUrl();

			const res = await req.post<{ success: boolean; count: number; message: string }>(
				"/calendar/webhook/send",
				{},
			);

			const count = res.count ?? 0;
			const msg =
				count > 0
					? `Sucesso! ${count} lançamento(s) de hoje enviado(s) para o webhook!`
					: "Nenhuma estreia encontrada na tua biblioteca para o dia de hoje.";

			statusMessage = {
				type: "success",
				text: msg,
			};
			notify({
				type: "success",
				text: msg,
				time: 5000,
			});
		} catch (err) {
			const errMsg = ReqerError.getMsg(err, "Falha ao verificar/enviar lançamentos.");
			statusMessage = {
				type: "error",
				text: errMsg,
			};
			notify({
				type: "error",
				text: errMsg,
				time: 6000,
			});
		} finally {
			isSending = false;
		}
	}
</script>

{#if open}
	<Modal
		title="Notificações do Calendário Webhook"
		desc="Recebe alertas automáticos de novos episódios e estreias da tua lista no Discord, Slack ou Ntfy."
		onClose={handleClose}
		maxWidth="640px"
	>
		<div class="webhook-modal-body">
			<!-- Enable Toggle Row -->
			<div class="setting-row">
				<div class="setting-info">
					<span class="setting-title">Alertas Automáticos Diários</span>
					<span class="setting-desc">
						O Watcharr verifica todas as manhãs se há episódios ou filmes a estrear hoje na tua biblioteca e envia uma notificação.
					</span>
				</div>
				<Checkbox
					name="calendar_webhook_enabled"
					value={isEnabled}
					toggled={handleToggleEnabled}
				/>
			</div>

			<!-- Webhook URL Input -->
			<div class="field-block">
				<div class="label-row">
					<label for="webhook_url_input">Webhook URL</label>
					{#if detectedPlatform === "discord"}
						<span class="platform-badge discord">
							<Icon i="discord" wh={14} /> Discord Webhook
						</span>
					{:else if detectedPlatform === "slack"}
						<span class="platform-badge slack">Slack Webhook</span>
					{:else if detectedPlatform === "ntfy"}
						<span class="platform-badge ntfy">Ntfy Webhook</span>
					{:else if detectedPlatform === "generic"}
						<span class="platform-badge generic">JSON Webhook</span>
					{/if}
				</div>
				<div class="input-wrapper">
					<input
						id="webhook_url_input"
						type="url"
						placeholder="https://discord.com/api/webhooks/..."
						bind:value={webhookUrl}
						onblur={saveUrl}
						onkeydown={(e) => e.key === "Enter" && saveUrl()}
					/>
				</div>
				<span class="field-hint">
					No Discord: vai a <strong>Definições do Servidor &gt; Integrações &gt; Webhooks</strong> e copia o URL do webhook.
				</span>
			</div>

			<!-- Scope Filter -->
			<div class="field-block">
				<span class="field-label">Notificar sobre:</span>
				<div class="filter-chips">
					<button
						type="button"
						class="chip-btn"
						class:active={filterType === "all"}
						onclick={() => handleFilterChange("all")}
					>
						Filmes & Séries
					</button>
					<button
						type="button"
						class="chip-btn"
						class:active={filterType === "tv"}
						onclick={() => handleFilterChange("tv")}
					>
						Apenas Séries / Episódios
					</button>
					<button
						type="button"
						class="chip-btn"
						class:active={filterType === "movie"}
						onclick={() => handleFilterChange("movie")}
					>
						Apenas Filmes
					</button>
				</div>
			</div>

			<!-- Status Feedback Message -->
			{#if statusMessage}
				<div class="status-alert {statusMessage.type}">
					<Icon i={statusMessage.type === "success" ? "check" : "close-circle"} wh={18} />
					<span>{statusMessage.text}</span>
				</div>
			{/if}

			<!-- Action Buttons -->
			<div class="modal-actions">
				<button
					type="button"
					class="btn-test"
					disabled={isTesting || !webhookUrl.trim()}
					onclick={testWebhook}
				>
					{#if isTesting}
						<SpinnerTiny />
						<span>A testar...</span>
					{:else}
						<Icon i="sparkles" wh={16} />
						<span>Testar Webhook</span>
					{/if}
				</button>

				<button
					type="button"
					class="btn-send-now"
					disabled={isSending || !webhookUrl.trim()}
					onclick={sendTodayReleases}
				>
					{#if isSending}
						<SpinnerTiny />
						<span>A verificar...</span>
					{:else}
						<Icon i="calendar" wh={16} />
						<span>Enviar Estreias de Hoje</span>
					{/if}
				</button>
			</div>
		</div>
	</Modal>
{/if}

<style lang="scss">
	.webhook-modal-body {
		display: flex;
		flex-direction: column;
		gap: 24px;
		padding: 8px 0 4px 0;
	}

	.setting-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		padding: 14px 16px;
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid rgba(255, 255, 255, 0.07);
		border-radius: 12px;

		.setting-info {
			display: flex;
			flex-direction: column;
			gap: 4px;

			.setting-title {
				font-size: 14px;
				font-weight: 600;
				color: $text-color;
			}

			.setting-desc {
				font-size: 12px;
				color: $text-muted;
				line-height: 1.4;
			}
		}
	}

	.field-block {
		display: flex;
		flex-direction: column;
		gap: 8px;

		.label-row {
			display: flex;
			align-items: center;
			justify-content: space-between;

			label {
				font-size: 13px;
				font-weight: 600;
				color: $text-color;
			}
		}

		.field-label {
			font-size: 13px;
			font-weight: 600;
			color: $text-color;
		}

		.platform-badge {
			display: inline-flex;
			align-items: center;
			gap: 6px;
			font-size: 11px;
			font-weight: 600;
			padding: 2px 8px;
			border-radius: 6px;

			&.discord {
				background: rgba(88, 101, 242, 0.2);
				color: #7289da;
				border: 1px solid rgba(88, 101, 242, 0.4);
				fill: #7289da;
			}

			&.slack {
				background: rgba(224, 30, 90, 0.15);
				color: #e01e5a;
				border: 1px solid rgba(224, 30, 90, 0.3);
			}

			&.ntfy {
				background: rgba(54, 186, 152, 0.15);
				color: #36ba98;
				border: 1px solid rgba(54, 186, 152, 0.3);
			}

			&.generic {
				background: rgba(255, 255, 255, 0.08);
				color: $text-muted;
			}
		}

		.input-wrapper {
			input {
				width: 100%;
				padding: 12px 14px;
				font-size: 13px;
				background: rgba(0, 0, 0, 0.25);
				border: 1px solid rgba(255, 255, 255, 0.12);
				border-radius: 10px;
				color: $text-color;
				outline: none;
				transition: border-color 0.2s ease, box-shadow 0.2s ease;

				&:focus {
					border-color: #36ba98;
					box-shadow: 0 0 0 3px rgba(54, 186, 152, 0.18);
				}

				&::placeholder {
					color: rgba(255, 255, 255, 0.3);
				}
			}
		}

		.field-hint {
			font-size: 11px;
			color: $text-muted;
			line-height: 1.4;

			strong {
				color: rgba(255, 255, 255, 0.85);
			}
		}
	}

	.filter-chips {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;

		.chip-btn {
			padding: 8px 14px;
			font-size: 12px;
			font-weight: 500;
			color: $text-muted;
			background: rgba(255, 255, 255, 0.04);
			border: 1px solid rgba(255, 255, 255, 0.08);
			border-radius: 8px;
			cursor: pointer;
			transition: all 0.18s ease;

			&:hover {
				background: rgba(255, 255, 255, 0.08);
				color: $text-color;
			}

			&.active {
				background: rgba(54, 186, 152, 0.15);
				color: #36ba98;
				border-color: rgba(54, 186, 152, 0.4);
				font-weight: 600;
			}
		}
	}

	.status-alert {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 14px;
		border-radius: 8px;
		font-size: 12px;
		line-height: 1.4;

		&.success {
			background: rgba(54, 186, 152, 0.12);
			border: 1px solid rgba(54, 186, 152, 0.3);
			color: #36ba98;
			fill: #36ba98;
		}

		&.error {
			background: rgba(239, 68, 68, 0.12);
			border: 1px solid rgba(239, 68, 68, 0.3);
			color: #f87171;
			fill: #f87171;
		}
	}

	.modal-actions {
		display: flex;
		gap: 12px;
		margin-top: 4px;

		button {
			flex: 1;
			display: flex;
			align-items: center;
			justify-content: center;
			gap: 8px;
			padding: 12px 16px;
			font-size: 13px;
			font-weight: 600;
			border-radius: 10px;
			cursor: pointer;
			transition: all 0.2s ease;
			user-select: none;

			&:disabled {
				opacity: 0.5;
				cursor: not-allowed;
			}
		}

		.btn-test {
			background: rgba(255, 255, 255, 0.05);
			color: $text-color;
			fill: $text-color;
			border: 1px solid rgba(255, 255, 255, 0.12);

			&:hover:not(:disabled) {
				background: rgba(255, 255, 255, 0.09);
				border-color: rgba(255, 255, 255, 0.2);
				transform: translateY(-1px);
			}
		}

		.btn-send-now {
			background: rgba(54, 186, 152, 0.15);
			color: #36ba98;
			fill: #36ba98;
			border: 1px solid rgba(54, 186, 152, 0.35);

			&:hover:not(:disabled) {
				background: rgba(54, 186, 152, 0.25);
				border-color: rgba(54, 186, 152, 0.55);
				transform: translateY(-1px);
			}
		}
	}
</style>

