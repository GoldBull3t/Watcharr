<script lang="ts">
	import Modal from "@/lib/Modal.svelte";
	import Icon from "@/lib/Icon.svelte";
	import { req, baseURL, updateUserSetting } from "@/lib/util/api";
	import { store } from "@/store.svelte";
	import { notify } from "@/lib/util/notify";
	import SpinnerTiny from "@/lib/SpinnerTiny.svelte";

	interface Props {
		open: boolean;
	}

	let { open = $bindable(false) }: Props = $props();

	let apiKey: string = $state("");
	let loading: boolean = $state(true);
	let regenerating: boolean = $state(false);
	let activeTab: "homepage" | "ha_sensor" | "ha_svg" | "ical" | "webhook" = $state("homepage");

	let webhookUrl = $state(store.userSettings?.activityWebhookUrl ?? "");
	let webhookEnabled = $state(store.userSettings?.activityWebhookEnabled ?? false);
	let webhookFilter = $state(store.userSettings?.activityWebhookFilter ?? "all");
	let isTestingWebhook = $state(false);

	$effect(() => {
		if (store.userSettings) {
			webhookUrl = store.userSettings.activityWebhookUrl ?? "";
			webhookEnabled = store.userSettings.activityWebhookEnabled ?? false;
			webhookFilter = store.userSettings.activityWebhookFilter ?? "all";
		}
	});

	function saveWebhookSettings() {
		updateUserSetting("activityWebhookUrl", webhookUrl.trim());
		updateUserSetting("activityWebhookEnabled", webhookEnabled);
		updateUserSetting("activityWebhookFilter", webhookFilter);
	}

	function toggleWebhookEnabled(val: boolean) {
		webhookEnabled = val;
		updateUserSetting("activityWebhookEnabled", val);
	}

	async function testWebhook() {
		const target = webhookUrl.trim();
		if (!target) {
			notify({ text: "Insere um Webhook URL antes de testar.", type: "error" });
			return;
		}
		isTestingWebhook = true;
		try {
			saveWebhookSettings();
			await req.post("/integrations/webhook/activity/test", { url: target });
			notify({ text: "Webhook de teste enviado com sucesso!", type: "success" });
		} catch (e: any) {
			notify({ text: e?.message || "Erro ao disparar webhook de teste", type: "error" });
		} finally {
			isTestingWebhook = false;
		}
	}
	let copiedKey: boolean = $state(false);
	let copiedSnippet: boolean = $state(false);

	$effect(() => {
		if (open) {
			loadApiKey();
		}
	});

	async function loadApiKey() {
		loading = true;
		try {
			const res = await req.get<{ apiKey: string }>("/user/api-key");
			if (res?.apiKey) {
				apiKey = res.apiKey;
			}
		} catch (e) {
			notify({ text: "Falha ao carregar chave de API", type: "error" });
		} finally {
			loading = false;
		}
	}

	async function regenerateApiKey() {
		if (
			!confirm(
				"Tens a certeza de que queres regenerar a tua chave de API? Todas as integrações existentes deixarão de funcionar até atualizares a chave.",
			)
		) {
			return;
		}
		regenerating = true;
		try {
			const res = await req.post<{ apiKey: string }>("/user/api-key/regenerate");
			if (res?.apiKey) {
				apiKey = res.apiKey;
				notify({ text: "Nova chave de API gerada com sucesso!", type: "success" });
			}
		} catch (e) {
			notify({ text: "Erro ao regenerar chave de API", type: "error" });
		} finally {
			regenerating = false;
		}
	}

	function copyToClipboard(text: string, isSnippet = false) {
		navigator.clipboard.writeText(text);
		if (isSnippet) {
			copiedSnippet = true;
			setTimeout(() => (copiedSnippet = false), 2000);
		} else {
			copiedKey = true;
			setTimeout(() => (copiedKey = false), 2000);
		}
		notify({ text: "Copiado para a área de transferência!", type: "success" });
	}

	const origin = $derived(
		typeof window !== "undefined" ? window.location.origin : "http://localhost:3080",
	);

	const homepageSnippet = $derived(`- Watcharr:
    icon: watcharr.png
    href: ${origin}
    description: Estatísticas da Biblioteca
    widget:
      type: customapi
      url: ${origin}/api/v1/integrations/homepage?apiKey=${apiKey || "A_TUA_CHAVE"}
      mappings:
        - field: watchedMovies
          label: Filmes
        - field: watchedEpisodes
          label: Episódios
        - field: hoursWatched
          label: Horas Vistas
          format: number
        - field: watchlistCount
          label: Na Lista`);

	const haSensorSnippet = $derived(`sensor:
  - platform: rest
    name: Watcharr
    resource: ${origin}/api/v1/integrations/ha/sensor?apiKey=${apiKey || "A_TUA_CHAVE"}
    scan_interval: 60
    value_template: "{{ value_json.state }}"
    json_attributes_path: "$.attributes"
    json_attributes:
      - watchedMovies
      - watchedEpisodes
      - hoursWatched
      - watchlistCount
      - lastWatched
      - nextUpTitle
      - nextUpSeason
      - nextUpEpisode
      - currentlyWatching
      - watchlist
      - upcoming`);

	const haAutomationSnippet = $derived(`automation:
  - alias: "Watcharr: Automação em Tempo Real (Luzes/Cinema)"
    trigger:
      - platform: webhook
        webhook_id: watcharr_event
        allowed_methods:
          - POST
        local_only: false
    action:
      - choose:
          # Quando terminas um filme ou episódio:
          - conditions:
              - condition: template
                value_template: "{{ trigger.json.status == 'FINISHED' }}"
            sequence:
              - service: light.turn_on
                target:
                  entity_id: light.sala
              - service: notify.notify
                data:
                  title: "Filme/Série Concluído! 🎬"
                  message: "{{ trigger.json.user }} acabou de ver {{ trigger.json.title }}!"
          # Quando começas a ver (Modo Cinema):
          - conditions:
              - condition: template
                value_template: "{{ trigger.json.status == 'WATCHING' }}"
            sequence:
              - service: light.turn_off
                target:
                  entity_id: light.sala`);

	const haSvgSnippet = $derived(`type: picture
image: ${origin}/api/v1/integrations/ha/now-next.svg?apiKey=${apiKey || "A_TUA_CHAVE"}
tap_action:
  action: url
  url_path: ${origin}`);

	const icalUrl = $derived(
		`${origin}/api/v1/integrations/calendar.ics?apiKey=${apiKey || "A_TUA_CHAVE"}`,
	);
</script>

{#if open}
	<Modal
		title="Integrações & Dashboards Homelab"
		maxWidth="780px"
		onClose={() => (open = false)}
	>
		<div class="modal-body">
			<!-- API Key Section -->
			<div class="key-box">
				<div class="key-label">
					<Icon i="lock-closed" wh={14} />
					<span>Chave de API Pessoal (Token Estático)</span>
				</div>
				<div class="key-row">
					<div class="key-display">
						{#if loading}
							<SpinnerTiny />
							<span>A carregar chave...</span>
						{:else}
							<code>{apiKey}</code>
						{/if}
					</div>
					<button
						class="btn-copy"
						onclick={() => copyToClipboard(apiKey)}
						disabled={loading || !apiKey}
					>
						<Icon i="document" wh={13} />
						{copiedKey ? "Copiado!" : "Copiar"}
					</button>
					<button
						class="btn-regen"
						onclick={regenerateApiKey}
						disabled={loading || regenerating}
					>
						<Icon i="refresh" wh={13} />
						{regenerating ? "A gerar..." : "Regenerar"}
					</button>
				</div>
				<p class="key-hint">
					Esta chave autentica pedidos de leitura das tuas estatísticas e eventos de calendário em segundo plano com cache de memória de alta performance (&lt;1ms).
				</p>
			</div>

			<!-- Navigation Tabs -->
			<div class="tab-buttons">
				<button
					class:active={activeTab === "homepage"}
					onclick={() => (activeTab = "homepage")}
				>
					Homepage / Dashy
				</button>
				<button
					class:active={activeTab === "ha_sensor"}
					onclick={() => (activeTab = "ha_sensor")}
				>
					Home Assistant (Sensor REST)
				</button>
				<button
					class:active={activeTab === "ha_svg"}
					onclick={() => (activeTab = "ha_svg")}
				>
					Lovelace Card (SVG)
				</button>
				<button
					class:active={activeTab === "ical"}
					onclick={() => (activeTab = "ical")}
				>
					Calendário (iCal Feed)
				</button>
				<button
					class:active={activeTab === "webhook"}
					onclick={() => (activeTab = "webhook")}
				>
					Webhooks em Tempo Real ⚡
				</button>
			</div>

			<!-- Tab Contents -->
			<div class="tab-content">
				{#if activeTab === "homepage"}
					<div class="integration-guide">
						<div class="guide-header">
							<h5>Configuração no <code>services.yaml</code> do Homepage</h5>
							<button
								class="btn-copy-code"
								onclick={() => copyToClipboard(homepageSnippet, true)}
							>
								<Icon i="document" wh={12} />
								{copiedSnippet ? "Copiado!" : "Copiar YAML"}
							</button>
						</div>
						<pre><code>{homepageSnippet}</code></pre>
						<p class="guide-note">
							Endpoint REST: <code>{origin}/api/v1/integrations/homepage?apiKey=...</code>
						</p>
					</div>
				{:else if activeTab === "ha_sensor"}
					<div class="integration-guide">
						<div class="guide-header">
							<h5>Configuração no <code>configuration.yaml</code> do Home Assistant</h5>
							<button
								class="btn-copy-code"
								onclick={() => copyToClipboard(haSensorSnippet, true)}
							>
								<Icon i="document" wh={12} />
								{copiedSnippet ? "Copiado!" : "Copiar YAML"}
							</button>
						</div>
						<pre><code>{haSensorSnippet}</code></pre>
						<p class="guide-note">
							Gera a entidade <code>sensor.watcharr</code> com atributos detalhados para automações e cartões.
						</p>
					</div>
				{:else if activeTab === "ha_svg"}
					<div class="integration-guide">
						<div class="guide-header">
							<h5>Cartão Picture Lovelace para Tablets e Painéis</h5>
							<button
								class="btn-copy-code"
								onclick={() => copyToClipboard(haSvgSnippet, true)}
							>
								<Icon i="document" wh={12} />
								{copiedSnippet ? "Copiado!" : "Copiar YAML"}
							</button>
						</div>
						<pre><code>{haSvgSnippet}</code></pre>
						<p class="guide-note">
							Devolve uma imagem vetorial dinâmica SVG com o próximo episódio, poster e métricas em modo escuro.
						</p>
					</div>
				{:else if activeTab === "ical"}
					<div class="integration-guide">
						<div class="guide-header">
							<h5>Feed iCalendar / Webcal (.ics)</h5>
							<button
								class="btn-copy-code"
								onclick={() => copyToClipboard(icalUrl, true)}
							>
								<Icon i="document" wh={12} />
								{copiedSnippet ? "Copiado!" : "Copiar URL"}
							</button>
						</div>
						<pre><code>{icalUrl}</code></pre>
						<p class="guide-note">
							Compatível com a integração oficial <b>iCalendar</b> do Home Assistant, Google Calendar, Apple Calendar e Thunderbird. Lista os lançamentos da tua biblioteca para os próximos 30 dias.
						</p>
					</div>
				{:else if activeTab === "webhook"}
					<div class="integration-guide">
						<div class="webhook-settings-box">
							<div class="webhook-input-group">
								<label for="webhook-url">URL de Destino (Home Assistant, Discord, Ntfy)</label>
								<div class="webhook-input-row">
									<input
										id="webhook-url"
										type="url"
										placeholder="https://.../api/webhook/watcharr_event ou Discord Webhook"
										bind:value={webhookUrl}
										onblur={saveWebhookSettings}
									/>
									<button
										class="btn-test-wh"
										onclick={testWebhook}
										disabled={isTestingWebhook || !webhookUrl}
									>
										<Icon i="notifications" wh={13} />
										{isTestingWebhook ? "A testar..." : "Testar Webhook"}
									</button>
								</div>
							</div>

							<div class="webhook-controls-row">
								<label class="wh-toggle-label">
									<input
										type="checkbox"
										checked={webhookEnabled}
										onchange={(e) => toggleWebhookEnabled((e.target as HTMLInputElement).checked)}
									/>
									<span>Ativar Disparos em Tempo Real</span>
								</label>

								<div class="wh-filter-wrap">
									<span>Eventos:</span>
									<select bind:value={webhookFilter} onchange={saveWebhookSettings}>
										<option value="all">Todos os Eventos (Finished, Watching, Rating)</option>
										<option value="finished">Apenas Concluídos (Finished)</option>
										<option value="watching">Apenas Em Andamento (Watching)</option>
										<option value="rating">Apenas Avaliações (Rating)</option>
									</select>
								</div>
							</div>
						</div>

						<div class="guide-header" style="margin-top: 14px;">
							<h5>Automação Instantânea no <code>automations.yaml</code> do Home Assistant</h5>
							<button
								class="btn-copy-code"
								onclick={() => copyToClipboard(haAutomationSnippet, true)}
							>
								<Icon i="document" wh={12} />
								{copiedSnippet ? "Copiado!" : "Copiar YAML"}
							</button>
						</div>
						<pre><code>{haAutomationSnippet}</code></pre>
						<p class="guide-note">
							Dispara no exato segundo em que marcas um filme ou episódio, permitindo ligar/desligar luzes e enviar notificações sem esperar pelo scan REST.
						</p>
					</div>
				{/if}
			</div>
		</div>
	</Modal>
{/if}

<style lang="scss">
	.modal-body {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.key-box {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 14px 16px;
		background: rgba(22, 27, 38, 0.75);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 12px;

		.key-label {
			display: flex;
			align-items: center;
			gap: 6px;
			font-size: 11.5px;
			font-weight: 700;
			text-transform: uppercase;
			letter-spacing: 0.5px;
			color: #94a3b8;
		}

		.key-row {
			display: flex;
			align-items: center;
			gap: 8px;

			.key-display {
				flex: 1;
				display: flex;
				align-items: center;
				gap: 8px;
				padding: 8px 12px;
				background: rgba(0, 0, 0, 0.4);
				border: 1px solid rgba(255, 255, 255, 0.1);
				border-radius: 8px;
				overflow: hidden;

				code {
					font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
					font-size: 13px;
					color: #a5b4fc;
					text-overflow: ellipsis;
					overflow: hidden;
					white-space: nowrap;
				}

				span {
					font-size: 12px;
					color: #64748b;
				}
			}

			button {
				display: inline-flex;
				align-items: center;
				gap: 5px;
				padding: 8px 14px;
				font-size: 12px;
				font-weight: 600;
				border-radius: 8px;
				cursor: pointer;
				white-space: nowrap;
				transition: all 0.15s ease;

				&.btn-copy {
					background: #4f46e5;
					border: 1px solid #6366f1;
					color: #fff;

					&:hover {
						background: #4338ca;
					}
				}

				&.btn-regen {
					background: rgba(255, 255, 255, 0.06);
					border: 1px solid rgba(255, 255, 255, 0.12);
					color: #cbd5e1;

					&:hover {
						background: rgba(255, 255, 255, 0.12);
						color: #fff;
					}
				}
			}
		}

		.key-hint {
			margin: 0;
			font-size: 11.5px;
			color: #64748b;
			line-height: 1.4;
		}
	}

	.tab-buttons {
		display: flex;
		align-items: center;
		gap: 6px;
		flex-wrap: wrap;
		border-bottom: 1px solid rgba(255, 255, 255, 0.08);
		padding-bottom: 8px;

		button {
			padding: 6px 12px;
			font-size: 12.5px;
			font-weight: 600;
			border-radius: 8px;
			background: transparent;
			border: 1px solid transparent;
			color: #94a3b8;
			cursor: pointer;
			transition: all 0.15s ease;

			&:hover {
				color: #fff;
				background: rgba(255, 255, 255, 0.05);
			}

			&.active {
				background: rgba(99, 102, 241, 0.15);
				border-color: rgba(99, 102, 241, 0.35);
				color: #c7d2fe;
			}
		}
	}

	.webhook-settings-box {
		display: flex;
		flex-direction: column;
		gap: 12px;
		padding: 14px 16px;
		background: #090d16;
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 10px;

		.webhook-input-group {
			display: flex;
			flex-direction: column;
			gap: 6px;

			label {
				font-size: 11.5px;
				font-weight: 700;
				color: #94a3b8;
				text-transform: uppercase;
				letter-spacing: 0.5px;
			}

			.webhook-input-row {
				display: flex;
				gap: 8px;

				input {
					flex: 1;
					padding: 8px 12px;
					border-radius: 8px;
					background: rgba(255, 255, 255, 0.05);
					border: 1px solid rgba(255, 255, 255, 0.12);
					color: #fff;
					font-size: 13px;

					&:focus {
						outline: none;
						border-color: #6366f1;
					}
				}

				.btn-test-wh {
					display: inline-flex;
					align-items: center;
					gap: 6px;
					padding: 8px 14px;
					background: #6366f1;
					color: white;
					font-weight: 600;
					font-size: 12px;
					border-radius: 8px;
					border: none;
					cursor: pointer;
					white-space: nowrap;

					&:hover:not(:disabled) {
						background: #4f46e5;
					}

					&:disabled {
						opacity: 0.5;
						cursor: not-allowed;
					}
				}
			}
		}

		.webhook-controls-row {
			display: flex;
			align-items: center;
			justify-content: space-between;
			flex-wrap: wrap;
			gap: 12px;
			padding-top: 8px;
			border-top: 1px solid rgba(255, 255, 255, 0.06);

			.wh-toggle-label {
				display: inline-flex;
				align-items: center;
				gap: 8px;
				font-size: 12.5px;
				font-weight: 600;
				color: #e2e8f0;
				cursor: pointer;

				input {
					cursor: pointer;
					accent-color: #6366f1;
				}
			}

			.wh-filter-wrap {
				display: inline-flex;
				align-items: center;
				gap: 8px;
				font-size: 12px;
				color: #94a3b8;

				select {
					padding: 4px 8px;
					border-radius: 6px;
					background: rgba(255, 255, 255, 0.08);
					border: 1px solid rgba(255, 255, 255, 0.14);
					color: #fff;
					font-size: 12px;

					&:focus {
						outline: none;
						border-color: #6366f1;
					}
				}
			}
		}
	}

	.tab-content {
		.integration-guide {
			display: flex;
			flex-direction: column;
			gap: 10px;

			.guide-header {
				display: flex;
				align-items: center;
				justify-content: space-between;

				h5 {
					margin: 0;
					font-size: 13px;
					font-weight: 600;
					color: #e2e8f0;

					code {
						color: #38bdf8;
					}
				}

				.btn-copy-code {
					display: inline-flex;
					align-items: center;
					gap: 5px;
					padding: 4px 10px;
					font-size: 11.5px;
					font-weight: 600;
					border-radius: 6px;
					background: rgba(255, 255, 255, 0.07);
					border: 1px solid rgba(255, 255, 255, 0.12);
					color: #cbd5e1;
					cursor: pointer;

					&:hover {
						background: rgba(255, 255, 255, 0.14);
						color: #fff;
					}
				}
			}

			pre {
				margin: 0;
				padding: 14px;
				background: #090d16;
				border: 1px solid rgba(255, 255, 255, 0.08);
				border-radius: 10px;
				overflow-x: auto;
				font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
				font-size: 12px;
				line-height: 1.5;
				color: #e2e8f0;
			}

			.guide-note {
				margin: 0;
				font-size: 12px;
				color: #94a3b8;
				line-height: 1.45;

				code {
					color: #a5b4fc;
					font-size: 11.5px;
				}

				b {
					color: #fff;
				}
			}
		}
	}
</style>

