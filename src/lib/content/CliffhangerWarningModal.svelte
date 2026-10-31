<script lang="ts">
	import Modal from "../Modal.svelte";
	import Icon from "../Icon.svelte";

	interface Props {
		showName?: string;
		warning?: string;
		description?: string;
		onConfirm: () => void;
		onCancel: () => void;
	}

	let { showName, warning, description, onConfirm, onCancel }: Props = $props();
</script>

<Modal title="Aviso de Cancelamento" maxWidth="500px" onClose={onCancel}>
	<div class="warning-container">
		<div class="warning-icon-badge">
			<Icon i="close-circle" wh={32} />
		</div>

		<div class="warning-content">
			<h4>{showName || "Esta série"} foi cancelada sem desfecho!</h4>
			{#if warning}
				<p class="warning-highlight">{warning}</p>
			{/if}
			{#if description}
				<p class="warning-desc">{description}</p>
			{/if}
			<p class="warning-note">
				A produção foi descontinuada e termina com perguntas importantes e
				mistérios sem resolução. Desejas adicionar à tua lista mesmo assim?
			</p>
		</div>
	</div>

	<div class="actions">
		<button class="secondary" onclick={onCancel}> Cancelar </button>
		<button class="danger-btn" onclick={onConfirm}>
			<Icon i="add" wh={14} />
			Adicionar Mesmo Assim
		</button>
	</div>
</Modal>

<style lang="scss">
	.warning-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
		gap: 16px;
		padding: 8px 0 16px 0;

		@media screen and (min-width: 500px) {
			flex-direction: row;
			align-items: flex-start;
			text-align: left;
			gap: 20px;
		}

		.warning-icon-badge {
			display: flex;
			align-items: center;
			justify-content: center;
			flex-shrink: 0;
			width: 56px;
			height: 56px;
			border-radius: 50%;
			background: rgba(239, 68, 68, 0.15);
			border: 1px solid rgba(239, 68, 68, 0.35);
			color: #f87171;
		}

		.warning-content {
			display: flex;
			flex-direction: column;
			gap: 8px;

			h4 {
				margin: 0;
				font-size: 17px;
				font-weight: 700;
				color: #fff;
			}

			.warning-highlight {
				margin: 0;
				font-size: 13.5px;
				font-weight: 600;
				color: #fca5a5;
				line-height: 1.4;
			}

			.warning-desc {
				margin: 0;
				font-size: 12.5px;
				color: rgba(255, 255, 255, 0.7);
				line-height: 1.45;
				background: rgba(0, 0, 0, 0.25);
				padding: 8px 12px;
				border-radius: 8px;
				border-left: 3px solid #ef4444;
			}

			.warning-note {
				margin: 4px 0 0 0;
				font-size: 12.5px;
				color: rgba(255, 255, 255, 0.6);
				line-height: 1.4;
			}
		}
	}

	.actions {
		display: flex;
		justify-content: flex-end;
		gap: 12px;
		margin-top: 16px;
		padding-top: 12px;
		border-top: 1px solid rgba(255, 255, 255, 0.08);

		button {
			padding: 8px 16px;
			font-size: 13px;
			font-weight: 600;
			border-radius: 8px;
			cursor: pointer;
			transition: all 0.15s ease;

			&.secondary {
				background: rgba(255, 255, 255, 0.08);
				border: 1px solid rgba(255, 255, 255, 0.12);
				color: #e2e8f0;

				&:hover {
					background: rgba(255, 255, 255, 0.14);
				}
			}

			&.danger-btn {
				display: inline-flex;
				align-items: center;
				gap: 6px;
				background: #dc2626;
				border: 1px solid #b91c1c;
				color: #fff;

				&:hover {
					background: #ef4444;
				}
			}
		}
	}
</style>
