<script lang="ts">
	import Modal from "@/lib/Modal.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import Icon from "@/lib/Icon.svelte";
	import { req } from "@/lib/util/api";
	import { notify } from "@/lib/util/notify";
	import type { SteamStatusResponse, SteamSyncResponse } from "@/types";
	import { onMount } from "svelte";

	interface Props {
		onClose: () => void;
	}

	let { onClose }: Props = $props();

	let loading = $state(true);
	let actionLoading = $state(false);
	let status = $state<SteamStatusResponse>({
		connected: false,
		totalGames: 0,
		hasCustomKey: false,
	});

	let steamIdInput = $state("");
	let apiKeyInput = $state("");
	let syncResult = $state<SteamSyncResponse | undefined>(undefined);

	async function fetchStatus() {
		loading = true;
		try {
			const res = await req.get<SteamStatusResponse>("/game/steam/status");
			status = res;
			if (res.steamId) {
				steamIdInput = res.steamId;
			}
		} catch (err) {
			console.error("fetchStatus failed", err);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		fetchStatus();
	});

	async function handleConnect(e?: Event) {
		if (e) e.preventDefault();
		if (!steamIdInput.trim()) {
			notify({ text: "Insere o teu SteamID ou link do perfil Steam.", type: "error" });
			return;
		}

		actionLoading = true;
		syncResult = undefined;
		try {
			const res = await req.post<SteamStatusResponse>("/game/steam/connect", {
				steamId: steamIdInput.trim(),
				apiKey: apiKeyInput.trim() ? apiKeyInput.trim() : undefined,
			});
			status = res;
			notify({
				text: `Conta Steam conectada com sucesso (${res.personaName || res.steamId})!`,
				type: "success",
			});
		} catch (err: any) {
			console.error("handleConnect failed", err);
			notify({
				text: err?.error || "Falha ao conectar conta Steam.",
				type: "error",
			});
		} finally {
			actionLoading = false;
		}
	}

	async function handleSync() {
		actionLoading = true;
		syncResult = undefined;
		try {
			const res = await req.post<SteamSyncResponse>("/game/steam/sync");
			syncResult = res;
			notify({
				text: res.message || `Sincronizados ${res.syncedCount} jogos!`,
				type: res.syncedCount > 0 ? "success" : "info",
			});
		} catch (err: any) {
			console.error("handleSync failed", err);
			notify({
				text: err?.error || "Falha na sincronização com a Steam.",
				type: "error",
			});
		} finally {
			actionLoading = false;
		}
	}

	async function handleDisconnect() {
		actionLoading = true;
		syncResult = undefined;
		try {
			await req.delete("/game/steam/disconnect");
			status = {
				connected: false,
				totalGames: 0,
				hasCustomKey: false,
			};
			steamIdInput = "";
			apiKeyInput = "";
			notify({ text: "Conta Steam desconectada.", type: "info" });
		} catch (err) {
			notify({ text: "Falha ao desconectar.", type: "error" });
		} finally {
			actionLoading = false;
		}
	}
</script>

<Modal title="Sincronização Steam (Gaming 2.0)" maxWidth="560px" {onClose}>
	{#if loading}
		<div class="loading-wrap">
			<Spinner />
			<span>A carregar estado da integração Steam...</span>
		</div>
	{:else}
		<div class="steam-container">
			{#if status.connected}
				<div class="profile-card">
					<div class="avatar-area">
						{#if status.avatar}
							<img src={status.avatar} alt={status.personaName || "Steam User"} class="steam-avatar" />
						{:else}
							<div class="avatar-fallback">
								<Icon i="gamepad" />
							</div>
						{/if}
					</div>
					<div class="user-meta">
						<h4>{status.personaName || "Steam Account"}</h4>
						<span class="subtext">SteamID: <code>{status.steamId}</code></span>
						<span class="badge-library">
							🎮 {status.totalGames} jogos encontrados na Steam
						</span>
					</div>
				</div>

				<div class="info-note">
					<p>
						O Watcharr compara os títulos dos jogos registados na tua biblioteca com a tua conta Steam,
						atualizando automaticamente as <b>horas jogadas</b> e definindo a plataforma como <b>PC / Steam</b>.
					</p>
				</div>

				{#if syncResult}
					<div class="sync-summary">
						<div class="sync-header">
							<Icon i="check" />
							<span>{syncResult.message}</span>
						</div>
						{#if syncResult.updatedGames && syncResult.updatedGames.length > 0}
							<div class="game-chips">
								{#each syncResult.updatedGames as g}
									<span class="game-chip">{g}</span>
								{/each}
							</div>
						{/if}
					</div>
				{/if}

				<div class="actions">
					<button
						class="btn-sync"
						disabled={actionLoading}
						onclick={handleSync}
					>
						{#if actionLoading}
							<Spinner />
							<span>A sincronizar biblioteca...</span>
						{:else}
							<Icon i="refresh" />
							<span>Sincronizar Horas de Jogo Agora</span>
						{/if}
					</button>

					<button
						class="btn-disconnect"
						disabled={actionLoading}
						onclick={handleDisconnect}
					>
						Desconectar
					</button>
				</div>
			{:else}
				<form onsubmit={handleConnect} class="connect-form">
					<p class="desc">
						Liga a tua conta Steam para sincronizar automaticamente as tuas horas jogadas e estatísticas no Watcharr.
					</p>

					<div class="form-group">
						<label for="steam-id-input">SteamID, URL do Perfil ou Nome Personalizado:</label>
						<input
							id="steam-id-input"
							type="text"
							bind:value={steamIdInput}
							placeholder="Ex: 76561198000000000 ou https://steamcommunity.com/id/nome"
							required
						/>
						<span class="hint">Aceita SteamID64 numérico de 17 dígitos, link direto ou Vanity URL.</span>
					</div>

					<div class="form-group">
						<label for="steam-api-key-input">Steam Web API Key (Opcional):</label>
						<input
							id="steam-api-key-input"
							type="password"
							bind:value={apiKeyInput}
							placeholder="Deixa vazio para usar a chave global do servidor (STEAM_API_KEY)"
						/>
						<span class="hint">
							Podes obter uma chave gratuita instantaneamente em
							<a href="https://steamcommunity.com/dev/apikey" target="_blank" rel="noreferrer">
								steamcommunity.com/dev/apikey
							</a>.
						</span>
					</div>

					<button type="submit" class="btn-connect" disabled={actionLoading}>
						{#if actionLoading}
							<Spinner />
							<span>A verificar perfil Steam...</span>
						{:else}
							<span>Ligar Conta Steam</span>
						{/if}
					</button>
				</form>
			{/if}
		</div>
	{/if}
</Modal>

<style lang="scss">
	.loading-wrap {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
		padding: 30px 10px;
		color: var(--text-color-muted, #a0a0a0);
	}

	.steam-container {
		display: flex;
		flex-direction: column;
		gap: 16px;
		padding: 4px 0;
	}

	.profile-card {
		display: flex;
		align-items: center;
		gap: 16px;
		background: rgba(255, 255, 255, 0.04);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 10px;
		padding: 16px;

		.steam-avatar {
			width: 64px;
			height: 64px;
			border-radius: 50%;
			border: 2px solid var(--accent-color, #7099ff);
			object-fit: cover;
		}

		.avatar-fallback {
			width: 64px;
			height: 64px;
			border-radius: 50%;
			background: rgba(255, 255, 255, 0.08);
			display: flex;
			align-items: center;
			justify-content: center;
			font-size: 28px;
		}

		.user-meta {
			display: flex;
			flex-direction: column;
			gap: 4px;

			h4 {
				margin: 0;
				font-size: 1.15rem;
				font-weight: 600;
			}

			.subtext {
				font-size: 0.82rem;
				color: var(--text-color-muted, #8e8e8e);

				code {
					font-family: monospace;
					background: rgba(0, 0, 0, 0.25);
					padding: 1px 5px;
					border-radius: 4px;
				}
			}

			.badge-library {
				display: inline-block;
				margin-top: 4px;
				font-size: 0.8rem;
				font-weight: 500;
				color: #92b9ff;
			}
		}
	}

	.info-note {
		background: rgba(112, 153, 255, 0.08);
		border-left: 3px solid #7099ff;
		padding: 10px 14px;
		border-radius: 0 8px 8px 0;
		font-size: 0.88rem;
		line-height: 1.4;

		p {
			margin: 0;
		}
	}

	.sync-summary {
		background: rgba(46, 204, 113, 0.08);
		border: 1px solid rgba(46, 204, 113, 0.25);
		border-radius: 8px;
		padding: 12px;
		display: flex;
		flex-direction: column;
		gap: 8px;

		.sync-header {
			display: flex;
			align-items: center;
			gap: 8px;
			font-weight: 500;
			color: #2ecc71;
			font-size: 0.92rem;
		}

		.game-chips {
			display: flex;
			flex-wrap: wrap;
			gap: 6px;

			.game-chip {
				background: rgba(46, 204, 113, 0.15);
				color: #2ecc71;
				font-size: 0.8rem;
				padding: 3px 8px;
				border-radius: 6px;
				font-weight: 500;
			}
		}
	}

	.actions {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		margin-top: 8px;

		.btn-sync {
			flex: 1;
			display: flex;
			align-items: center;
			justify-content: center;
			gap: 8px;
			background: var(--accent-color, #7099ff);
			color: #fff;
			border: none;
			padding: 11px 18px;
			border-radius: 8px;
			font-weight: 600;
			cursor: pointer;
			transition: opacity 0.15s ease;

			&:hover:not(:disabled) {
				opacity: 0.9;
			}

			&:disabled {
				opacity: 0.6;
				cursor: not-allowed;
			}
		}

		.btn-disconnect {
			background: transparent;
			color: #ff6b6b;
			border: 1px solid rgba(255, 107, 107, 0.3);
			padding: 10px 16px;
			border-radius: 8px;
			font-size: 0.88rem;
			cursor: pointer;
			transition: background 0.15s ease;

			&:hover:not(:disabled) {
				background: rgba(255, 107, 107, 0.1);
			}
		}
	}

	.connect-form {
		display: flex;
		flex-direction: column;
		gap: 14px;

		.desc {
			font-size: 0.92rem;
			color: var(--text-color-muted, #9e9e9e);
			margin: 0 0 6px 0;
			line-height: 1.4;
		}

		.form-group {
			display: flex;
			flex-direction: column;
			gap: 5px;

			label {
				font-size: 0.85rem;
				font-weight: 500;
			}

			input {
				background: rgba(0, 0, 0, 0.2);
				border: 1px solid rgba(255, 255, 255, 0.12);
				border-radius: 6px;
				padding: 10px 12px;
				color: inherit;
				font-size: 0.9rem;

				&:focus {
					outline: none;
					border-color: var(--accent-color, #7099ff);
				}
			}

			.hint {
				font-size: 0.76rem;
				color: var(--text-color-muted, #7c7c7c);

				a {
					color: var(--accent-color, #7099ff);
					text-decoration: underline;
				}
			}
		}

		.btn-connect {
			display: flex;
			align-items: center;
			justify-content: center;
			gap: 8px;
			background: var(--accent-color, #7099ff);
			color: #fff;
			border: none;
			padding: 12px 18px;
			border-radius: 8px;
			font-weight: 600;
			cursor: pointer;
			margin-top: 8px;
			transition: opacity 0.15s ease;

			&:hover:not(:disabled) {
				opacity: 0.9;
			}

			&:disabled {
				opacity: 0.6;
				cursor: not-allowed;
			}
		}
	}
</style>

