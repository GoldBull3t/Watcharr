<script lang="ts">
	import { onMount } from "svelte";
	import Modal from "$lib/Modal.svelte";
	import Icon from "$lib/Icon.svelte";
	import { notify } from "$lib/util/notify";
	import { req } from "$lib/util/api";

	interface Props {
		open: boolean;
		onClose: () => void;
	}

	interface ApiKeyResponse {
		apiKey: string;
	}

	let { open = $bindable(false), onClose }: Props = $props();

	let apiKey = $state<string>("");
	let loading = $state<boolean>(true);
	let copied = $state<boolean>(false);
	let hostUrl = $state<string>("");

	onMount(async () => {
		hostUrl = window.location.origin;
		await fetchApiKey();
	});

	async function fetchApiKey() {
		loading = true;
		try {
			const res = await req.get<ApiKeyResponse>("/user/api-key");
			if (res && res.apiKey) {
				apiKey = res.apiKey;
			}
		} catch (err) {
			console.error("Falha ao obter API Key:", err);
			notify({ text: "Não foi possível carregar a API Key do utilizador.", type: "error" });
		} finally {
			loading = false;
		}
	}

	async function regenerateApiKey() {
		if (!confirm("Tens a certeza que queres gerar uma nova chave? Todos os addons e integrações anteriores deixarão de funcionar.")) {
			return;
		}
		loading = true;
		try {
			const res = await req.post<ApiKeyResponse>("/user/api-key/regenerate");
			if (res && res.apiKey) {
				apiKey = res.apiKey;
				notify({ text: "Nova chave de API gerada com sucesso!", type: "success" });
			}
		} catch (err) {
			notify({ text: "Erro ao gerar nova chave de API.", type: "error" });
		} finally {
			loading = false;
		}
	}

	let addonManifestUrl = $derived(
		apiKey ? `${hostUrl}/api/addon/${apiKey}/manifest.json` : ""
	);

	let stremioUrl = $derived(
		addonManifestUrl ? addonManifestUrl.replace(/^https?:\/\//, "stremio://") : ""
	);

	function copyUrl() {
		if (!addonManifestUrl) return;
		navigator.clipboard.writeText(addonManifestUrl);
		copied = true;
		notify({ text: "URL do Addon copiado para a área de transferência!", type: "success" });
		setTimeout(() => {
			copied = false;
		}, 3000);
	}
</script>

{#if open}
	<Modal title="Addon Watcharr para Android TV (NuvioTV & Stremio)" desc="Sincroniza a tua biblioteca e progresso de episódios diretamente na Android TV" maxWidth="620px">
		<div class="addon-modal-content">
			<div class="banner">
				<div class="badge">NATIVO & SEM SERVIÇOS PAGOS</div>
				<p class="desc">
					O Watcharr atua como um Addon oficial compatível com <strong>NuvioTV</strong> e <strong>Stremio</strong>.
					As tuas séries a ver, filmes planeados e episódios sincronizam em tempo real entre a tua TV e o teu servidor.
				</p>
			</div>

			<div class="catalogs-preview">
				<span class="catalogs-title">Catálogos Sincronizados:</span>
				<div class="tags-row">
					<span class="tag">🎬 Séries A Ver</span>
					<span class="tag">🍿 Filmes A Ver</span>
					<span class="tag">⏳ Séries Para Ver</span>
					<span class="tag">📋 Filmes Para Ver</span>
				</div>
			</div>

			<div class="section">
				<label class="label" for="addon-url-input">URL de Instalação do Addon:</label>
				<div class="url-row">
					<input
						id="addon-url-input"
						class="url-input"
						type="text"
						readonly
						value={loading ? "A carregar chave..." : addonManifestUrl}
					/>
					<button class="copy-btn" onclick={copyUrl} disabled={loading || !apiKey} type="button">
						<Icon i={copied ? "check" : "clipboard"} wh={16} />
						{copied ? "Copiado!" : "Copiar"}
					</button>
				</div>
				{#if addonManifestUrl}
					<div class="links-row">
						<a class="sub-link" href={addonManifestUrl} target="_blank" rel="noopener noreferrer">
							<Icon i="globe" wh={13} />
							Testar no Browser
						</a>
						<a class="sub-link" href={stremioUrl}>
							<Icon i="play" wh={13} />
							Instalar no Stremio (1-clique)
						</a>
					</div>
				{/if}
			</div>

			<div class="instructions-card">
				<h4>Como configurar na Android TV (NuvioTV):</h4>
				<ol>
					<li>
						Abre o <strong>NuvioTV</strong> na Android TV e vai a <strong>Definições &rarr; Addons</strong> (ou cola na interface web do Nuvio no telemóvel).
					</li>
					<li>
						Clica em <strong>Adicionar Addon por URL</strong> e cola o link acima.
					</li>
					<li>
						Em <strong>Definições &rarr; Conteúdo e Descoberta &rarr; Reordenar Catálogos</strong>, ativa os catálogos do Watcharr para a Página Inicial.
					</li>
				</ol>
			</div>

			<div class="actions-row">
				<button class="regen-btn" onclick={regenerateApiKey} disabled={loading} type="button">
					<Icon i="arrow-path" wh={14} />
					Regenerar Chave de Acesso
				</button>
				<button class="close-btn" onclick={onClose} type="button">
					Fechar
				</button>
			</div>
		</div>
	</Modal>
{/if}

<style>
	.addon-modal-content {
		display: flex;
		flex-direction: column;
		gap: 16px;
		margin-top: 8px;
	}

	.banner {
		background: rgba(99, 102, 241, 0.08);
		border: 1px solid rgba(99, 102, 241, 0.25);
		border-radius: 8px;
		padding: 14px;
	}

	.badge {
		display: inline-block;
		font-size: 10px;
		font-weight: 700;
		letter-spacing: 0.5px;
		background: #6366f1;
		color: #ffffff;
		padding: 2px 8px;
		border-radius: 4px;
		margin-bottom: 8px;
	}

	.desc {
		font-size: 13px;
		color: var(--text-color);
		margin: 0;
		line-height: 1.4;
	}

	.catalogs-preview {
		display: flex;
		flex-direction: column;
		gap: 6px;
		background: var(--bg-surface-2, rgba(255, 255, 255, 0.02));
		border: 1px solid var(--border-color);
		padding: 10px 12px;
		border-radius: 6px;
	}

	.catalogs-title {
		font-size: 11px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		color: var(--text-muted);
	}

	.tags-row {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}

	.tag {
		font-size: 12px;
		background: rgba(99, 102, 241, 0.15);
		color: var(--text-color);
		border: 1px solid rgba(99, 102, 241, 0.3);
		padding: 3px 8px;
		border-radius: 4px;
		font-weight: 500;
	}

	.section {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.label {
		font-size: 12px;
		font-weight: 600;
		color: var(--text-muted);
	}

	.url-row {
		display: flex;
		gap: 8px;
	}

	.url-input {
		flex: 1;
		background: var(--bg-surface-2, rgba(0, 0, 0, 0.2));
		border: 1px solid var(--border-color);
		color: var(--text-color);
		padding: 8px 12px;
		font-size: 13px;
		font-family: monospace;
		border-radius: 6px;
	}

	.links-row {
		display: flex;
		gap: 16px;
		margin-top: 4px;
	}

	.sub-link {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		font-size: 12px;
		color: var(--primary-color, #6366f1);
		text-decoration: none;
		transition: opacity 0.15s ease;
	}

	.sub-link:hover {
		text-decoration: underline;
		opacity: 0.85;
	}

	.copy-btn {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		background: var(--primary-color, #6366f1);
		color: #fff;
		border: none;
		border-radius: 6px;
		padding: 8px 14px;
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
		transition: opacity 0.15s ease;
	}

	.copy-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.instructions-card {
		background: var(--bg-surface-2, rgba(255, 255, 255, 0.03));
		border: 1px solid var(--border-color);
		border-radius: 8px;
		padding: 14px;
	}

	.instructions-card h4 {
		margin: 0 0 10px 0;
		font-size: 13px;
		font-weight: 600;
	}

	.instructions-card ol {
		margin: 0;
		padding-left: 20px;
		font-size: 12px;
		line-height: 1.6;
		color: var(--text-muted);
	}

	.instructions-card li strong {
		color: var(--text-color);
	}

	.actions-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-top: 8px;
	}

	.regen-btn {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		background: transparent;
		border: 1px solid var(--border-color);
		color: var(--text-muted);
		border-radius: 6px;
		padding: 6px 12px;
		font-size: 12px;
		cursor: pointer;
	}

	.regen-btn:hover {
		color: var(--text-color);
		border-color: var(--text-muted);
	}

	.close-btn {
		background: var(--bg-surface-2, rgba(255, 255, 255, 0.08));
		border: 1px solid var(--border-color);
		color: var(--text-color);
		border-radius: 6px;
		padding: 6px 16px;
		font-size: 13px;
		cursor: pointer;
	}
</style>
