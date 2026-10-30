<script lang="ts">
	import { onMount } from "svelte";
	import Icon from "@/lib/Icon.svelte";
	import SpinnerTiny from "@/lib/SpinnerTiny.svelte";
	import {
		getPushStatus,
		subscribeToPush,
		unsubscribeFromPush,
		sendTestNotification,
		type PushStatus,
	} from "./pushNotifications";

	let status: PushStatus = $state({
		supported: false,
		permission: "unsupported",
		isSubscribed: false,
	});

	let loading = $state(false);
	let testing = $state(false);

	async function refreshStatus() {
		status = await getPushStatus();
	}

	onMount(() => {
		refreshStatus();
	});

	async function handleToggle() {
		loading = true;
		try {
			if (status.isSubscribed) {
				await unsubscribeFromPush();
			} else {
				await subscribeToPush();
			}
			await refreshStatus();
		} finally {
			loading = false;
		}
	}

	async function handleTest() {
		testing = true;
		try {
			await sendTestNotification();
		} finally {
			testing = false;
		}
	}
</script>

<div class="push-settings-container">
	<div class="header">
		<div class="info">
			<div class="title-row">
				<Icon i="notifications" wh={18} />
				<h4 class="card-title">Notificações Web Push & PWA</h4>
				{#if !status.supported}
					<span class="badge unsupported">Não suportado</span>
				{:else if status.isSubscribed}
					<span class="badge active">Ativo neste dispositivo</span>
				{:else}
					<span class="badge inactive">Desativado</span>
				{/if}
			</div>
			<p class="card-desc">
				Recebe notificações push no ecrã de bloqueio e ambiente de trabalho sobre estreias de episódios, filmes e atualizações de pedidos.
			</p>
		</div>

		<div class="actions">
			{#if status.supported}
				<button
					class="btn {status.isSubscribed ? 'btn-secondary' : 'btn-primary'}"
					disabled={loading}
					onclick={handleToggle}
				>
					{#if loading}
						<SpinnerTiny />
					{:else}
						<Icon i={status.isSubscribed ? "close" : "notifications"} wh={16} />
						<span>{status.isSubscribed ? "Desativar" : "Ativar Notificações"}</span>
					{/if}
				</button>

				{#if status.isSubscribed}
					<button class="btn btn-outline" disabled={testing} onclick={handleTest}>
						{#if testing}
							<SpinnerTiny />
						{:else}
							<Icon i="notifications" wh={16} />
							<span>Enviar Teste</span>
						{/if}
					</button>
				{/if}
			{/if}
		</div>
	</div>
</div>

<style lang="scss">
	.push-settings-container {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.header {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
	}

	.info {
		display: flex;
		flex-direction: column;
		gap: 4px;
		flex: 1;
		min-width: 260px;
	}

	.title-row {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.card-title {
		font-size: 1.05rem;
		font-weight: 600;
		color: $text-color;
		margin: 0;
	}

	.card-desc {
		font-size: 0.88rem;
		color: $text-muted;
		margin: 0;
		line-height: 1.4;
	}

	.badge {
		font-size: 0.72rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		padding: 2px 8px;
		border-radius: 12px;

		&.active {
			background: rgba(34, 197, 94, 0.18);
			color: #4ade80;
			border: 1px solid rgba(74, 222, 128, 0.3);
		}

		&.inactive {
			background: rgba(234, 179, 8, 0.15);
			color: #facc15;
			border: 1px solid rgba(250, 204, 21, 0.3);
		}

		&.unsupported {
			background: rgba(239, 68, 68, 0.15);
			color: #f87171;
			border: 1px solid rgba(248, 113, 113, 0.3);
		}
	}

	.actions {
		display: flex;
		align-items: center;
		gap: 10px;
		flex-wrap: wrap;
	}

	.btn {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		padding: 8px 16px;
		font-size: 0.88rem;
		font-weight: 600;
		border-radius: 8px;
		border: none;
		cursor: pointer;
		transition: all 0.2s ease;

		&:disabled {
			opacity: 0.6;
			cursor: not-allowed;
		}

		&.btn-primary {
			background: $accent-color;
			color: #fff;

			&:hover:not(:disabled) {
				filter: brightness(1.15);
			}
		}

		&.btn-secondary {
			background: rgba(255, 255, 255, 0.1);
			color: $text-color;

			&:hover:not(:disabled) {
				background: rgba(255, 255, 255, 0.16);
			}
		}

		&.btn-outline {
			background: transparent;
			color: $text-color;
			border: 1px solid rgba(255, 255, 255, 0.2);

			&:hover:not(:disabled) {
				background: rgba(255, 255, 255, 0.08);
			}
		}
	}
</style>

