<script lang="ts">
	import Modal from "@/lib/Modal.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import SpinnerTiny from "@/lib/SpinnerTiny.svelte";
	import Icon from "@/lib/Icon.svelte";
	import Checkbox from "@/lib/Checkbox.svelte";
	import { req, baseURL } from "@/lib/util/api";
	import { notify } from "@/lib/util/notify";
	import { toRelativeTime } from "@/lib/util/helpers";
	import type {
		BackupInfo,
		BackupConfig,
		GetBackupsResponse,
	} from "@/types";
	import { onMount } from "svelte";

	interface Props {
		onClose: () => void;
	}

	let { onClose }: Props = $props();

	let loading = $state(true);
	let creating = $state(false);
	let uploading = $state(false);
	let restoring = $state(false);
	let deletingFile: string | null = $state(null);

	let backups: BackupInfo[] = $state([]);
	let config: BackupConfig | null = $state(null);

	// Restore confirmation state
	let restoreTarget: BackupInfo | null = $state(null);
	let fileInput: HTMLInputElement;

	async function loadBackups() {
		loading = true;
		try {
			const res = await req.get<GetBackupsResponse>("/backup/");
			if (res) {
				backups = res.backups || [];
				config = res.config || null;
			}
		} catch (err: any) {
			console.error("Failed to load backups:", err);
			notify({
				type: "error",
				text: "Falha ao carregar lista de backups.",
			});
		} finally {
			loading = false;
		}
	}

	async function triggerManualBackup(format: "auto" | "sqlite" | "json" = "auto") {
		creating = true;
		const nid = notify({
			type: "loading",
			text: "A gerar snapshot seguro e comprimido...",
		});
		try {
			const res = await req.post<BackupInfo>("/backup/", { format });
			notify({
				id: nid,
				type: "success",
				text: `Backup ${res.filename} criado com sucesso (${res.formattedSize})!`,
			});
			await loadBackups();
		} catch (err: any) {
			console.error("Backup generation failed:", err);
			notify({
				id: nid,
				type: "error",
				text: "Erro ao gerar backup da base de dados.",
			});
		} finally {
			creating = false;
		}
	}

	async function updateSettings(updates: {
		enabled?: boolean;
		retention?: number;
		scheduleSeconds?: number;
	}) {
		if (!config) return;
		try {
			const updated = await req.post<BackupConfig>("/backup/config", updates);
			if (updated) {
				config = updated;
				notify({
					type: "success",
					text: "Configuração de backup atualizada.",
				});
			}
		} catch (err) {
			console.error("Failed to update backup config:", err);
			notify({
				type: "error",
				text: "Erro ao atualizar definições de backup.",
			});
		}
	}

	async function confirmRestore() {
		if (!restoreTarget) return;
		restoring = true;
		const targetName = restoreTarget.filename;
		const nid = notify({
			type: "loading",
			text: `A restaurar base de dados a partir de ${targetName}...`,
		});
		try {
			await req.post(`/backup/restore/${targetName}`, {});
			notify({
				id: nid,
				type: "success",
				text: "Base de dados restaurada com sucesso! Snapshot de segurança anterior criado.",
				time: 7000,
			});
			restoreTarget = null;
			await loadBackups();
		} catch (err: any) {
			console.error("Restore failed:", err);
			notify({
				id: nid,
				type: "error",
				text: err?.message || "Falha ao restaurar base de dados.",
				time: 7000,
			});
		} finally {
			restoring = false;
		}
	}

	async function deleteBackup(filename: string) {
		if (!confirm(`Tens a certeza que queres eliminar o backup ${filename}?`)) {
			return;
		}
		deletingFile = filename;
		try {
			await req.delete(`/backup/${filename}`);
			notify({
				type: "success",
				text: "Backup eliminado com sucesso.",
			});
			backups = backups.filter((b) => b.filename !== filename);
		} catch (err) {
			console.error("Delete backup failed:", err);
			notify({
				type: "error",
				text: "Erro ao eliminar backup.",
			});
		} finally {
			deletingFile = null;
		}
	}

	async function handleFileUpload(e: Event) {
		const target = e.target as HTMLInputElement;
		const file = target?.files?.[0];
		if (!file) return;

		uploading = true;
		const nid = notify({
			type: "loading",
			text: `A carregar ${file.name}...`,
		});

		try {
			const formData = new FormData();
			formData.append("file", file);

			const token = localStorage.getItem("token") || "";
			const res = await fetch(`${baseURL}/backup/upload`, {
				method: "POST",
				headers: {
					Authorization: `Bearer ${token}`,
				},
				body: formData,
			});

			if (!res.ok) {
				const errorJson = await res.json().catch(() => ({}));
				throw new Error(errorJson.error || "Falha no upload");
			}

			notify({
				id: nid,
				type: "success",
				text: `Backup ${file.name} carregado com sucesso!`,
			});
			await loadBackups();
		} catch (err: any) {
			console.error("Upload failed:", err);
			notify({
				id: nid,
				type: "error",
				text: err?.message || "Erro ao carregar ficheiro de backup.",
			});
		} finally {
			uploading = false;
			if (fileInput) fileInput.value = "";
		}
	}

	function formatDate(d: string | Date | undefined) {
		if (!d) return "—";
		const date = new Date(d);
		if (isNaN(date.getTime())) return "—";
		return date.toLocaleString();
	}

	onMount(() => {
		loadBackups();
	});
</script>

<Modal
	title="Disaster Recovery & Backups"
	desc="Backups automáticos agendados, proteção com compressão WAL e recuperação sem perda de dados."
	{onClose}
>
	<div class="backup-container">
		{#if loading}
			<div class="loader-wrap">
				<Spinner />
				<p>A verificar estado do sistema de backups...</p>
			</div>
		{:else}
			<!-- System Overview Header Metrics -->
			<div class="metrics-grid">
				<div class="metric-card">
					<div class="metric-icon">
						<Icon i="archive" wh={22} />
					</div>
					<div class="metric-info">
						<span class="metric-val">{config?.totalBackups ?? backups.length}</span>
						<span class="metric-label">Backups Guardados</span>
						<span class="metric-sub">{config?.formattedTotalSize ?? "0 B"} total</span>
					</div>
				</div>

				<div class="metric-card">
					<div class="metric-icon">
						<Icon i="database" wh={22} />
					</div>
					<div class="metric-info">
						<span class="metric-val">{config?.driver?.toUpperCase() || "SQLITE"}</span>
						<span class="metric-label">Motor de Dados</span>
						<span class="metric-sub">
							{config?.enabled ? "Auto-Backup Ativo" : "Backups Desativados"}
						</span>
					</div>
				</div>

				<div class="metric-card">
					<div class="metric-icon">
						<Icon i="clock" wh={22} />
					</div>
					<div class="metric-info">
						<span class="metric-val">
							{config?.lastBackup ? toRelativeTime((new Date(config.lastBackup).getTime() - Date.now()) / 1000) : "Nunca"}
						</span>
						<span class="metric-label">Último Snapshot</span>
						<span class="metric-sub">
							{config?.lastBackup ? formatDate(config.lastBackup) : "Nenhum registo"}
						</span>
					</div>
				</div>
			</div>

			<!-- Configuration & Policy Controls -->
			<div class="config-bar">
				<div class="config-item toggle">
					<label class="toggle-label" for="backup-enabled-toggle">
						<Checkbox
							id="backup-enabled-toggle"
							value={config?.enabled ?? true}
							toggled={(val) => updateSettings({ enabled: val })}
						/>
						<span>Backup Automático</span>
					</label>
				</div>

				<div class="config-item">
					<label for="retention-select">Retenção:</label>
					<select
						id="retention-select"
						value={config?.retention ?? 7}
						onchange={(e) => updateSettings({ retention: parseInt(e.currentTarget.value) })}
					>
						<option value="3">Últimos 3 backups</option>
						<option value="7">Últimos 7 backups (Recomendado)</option>
						<option value="14">Últimos 14 backups</option>
						<option value="30">Últimos 30 backups</option>
					</select>
				</div>

				<div class="config-item">
					<label for="schedule-select">Frequência:</label>
					<select
						id="schedule-select"
						value={config?.scheduleSeconds ?? 86400}
						onchange={(e) => updateSettings({ scheduleSeconds: parseInt(e.currentTarget.value) })}
					>
						<option value="43200">A cada 12 horas</option>
						<option value="86400">Diário (a cada 24 horas)</option>
						<option value="604800">Semanal (a cada 7 dias)</option>
					</select>
				</div>
			</div>

			<!-- Actions Bar -->
			<div class="actions-row">
				<button
					class="btn primary-btn"
					disabled={creating || restoring}
					onclick={() => triggerManualBackup("auto")}
				>
					{#if creating}
						<SpinnerTiny />
						<span>A criar snapshot...</span>
					{:else}
						<Icon i="archive" wh={18} />
						<span>Criar Backup Agora</span>
					{/if}
				</button>

				<input
					type="file"
					accept=".gz,.db,.json"
					bind:this={fileInput}
					onchange={handleFileUpload}
					style="display: none;"
				/>

				<button
					class="btn secondary-btn"
					disabled={uploading || restoring}
					onclick={() => fileInput?.click()}
				>
					{#if uploading}
						<SpinnerTiny />
						<span>A carregar...</span>
					{:else}
						<Icon i="upload" wh={18} />
						<span>Carregar Ficheiro</span>
					{/if}
				</button>
			</div>

			<!-- Backups List -->
			<div class="table-container">
				{#if backups.length === 0}
					<div class="empty-state">
						<Icon i="archive" wh={36} />
						<p>Nenhum backup disponível de momento.</p>
						<span>Clica em "Criar Backup Agora" para criar o teu primeiro ponto de restauro.</span>
					</div>
				{:else}
					<table class="backups-table">
						<thead>
							<tr>
								<th>Ficheiro</th>
								<th>Tipo</th>
								<th>Data</th>
								<th>Tamanho</th>
								<th class="actions-col">Ações</th>
							</tr>
						</thead>
						<tbody>
							{#each backups as b (b.filename)}
								{@const isPreRestore = b.filename.includes("pre_restore")}
								<tr>
									<td class="filename-cell">
										<div class="file-icon-wrap" class:prerestore={isPreRestore}>
											<Icon i={b.type === "sqlite" ? "database" : "document"} wh={18} />
										</div>
										<span class="file-name" title={b.filename}>{b.filename}</span>
									</td>
									<td>
										{#if isPreRestore}
											<span class="badge safety">Segurança Prévia</span>
										{:else if b.isAuto}
											<span class="badge auto">Automático</span>
										{:else}
											<span class="badge manual">Manual</span>
										{/if}
									</td>
									<td class="date-cell">{formatDate(b.createdAt)}</td>
									<td class="size-cell">{b.formattedSize}</td>
									<td class="actions-cell">
										<!-- Download button -->
										<a
											class="action-icon-btn download"
											href={`${baseURL}/backup/download/${b.filename}`}
											download={b.filename}
											title="Descarregar backup"
										>
											<Icon i="download" wh={16} />
										</a>

										<!-- Restore button -->
										<button
											class="action-icon-btn restore"
											title="Restaurar a partir deste backup"
											disabled={restoring}
											onclick={() => (restoreTarget = b)}
										>
											<Icon i="undo" wh={16} />
										</button>

										<!-- Delete button -->
										<button
											class="action-icon-btn delete"
											title="Eliminar backup"
											disabled={deletingFile === b.filename || restoring}
											onclick={() => deleteBackup(b.filename)}
										>
											{#if deletingFile === b.filename}
												<SpinnerTiny />
											{:else}
												<Icon i="trash" wh={16} />
											{/if}
										</button>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				{/if}
			</div>
		{/if}
	</div>

	<!-- Restore Confirmation Dialog -->
	{#if restoreTarget}
		<div class="restore-dialog-backdrop">
			<div class="restore-dialog">
				<div class="restore-header">
					<Icon i="undo" wh={28} />
					<h4>Restaurar Base de Dados?</h4>
				</div>
				<p class="restore-desc">
					Tens a certeza que queres restaurar o Watcharr para o estado do ficheiro:
					<strong class="highlight-file">{restoreTarget.filename}</strong>?
				</p>

				<div class="safety-banner">
					<Icon i="check" wh={18} />
					<span>
						<strong>Zero Perda de Dados:</strong> Um snapshot de segurança automático do estado atual será criado antes do restauro.
					</span>
				</div>

				<div class="restore-actions">
					<button
						class="btn cancel-btn"
						disabled={restoring}
						onclick={() => (restoreTarget = null)}
					>
						Cancelar
					</button>
					<button
						class="btn confirm-btn"
						disabled={restoring}
						onclick={confirmRestore}
					>
						{#if restoring}
							<SpinnerTiny />
							<span>A restaurar...</span>
						{:else}
							<span>Confirmar Restauro</span>
						{/if}
					</button>
				</div>
			</div>
		</div>
	{/if}
</Modal>

<style lang="scss">
	.backup-container {
		display: flex;
		flex-direction: column;
		gap: 20px;
		width: 100%;
		min-width: 650px;
		max-width: 820px;
	}

	.loader-wrap {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 40px;
		gap: 12px;
		color: $text-color;
	}

	.metrics-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 12px;

		@media (max-width: 700px) {
			grid-template-columns: 1fr;
		}
	}

	.metric-card {
		display: flex;
		align-items: center;
		gap: 14px;
		background: rgba(255, 255, 255, 0.04);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 12px;
		padding: 14px 16px;

		.metric-icon {
			display: flex;
			align-items: center;
			justify-content: center;
			width: 42px;
			height: 42px;
			border-radius: 10px;
			background: rgba($accent-color, 0.15);
			color: $accent-color;
			flex-shrink: 0;
		}

		.metric-info {
			display: flex;
			flex-direction: column;
			gap: 2px;
			overflow: hidden;

			.metric-val {
				font-size: 17px;
				font-weight: 700;
				color: #fff;
				letter-spacing: -0.01em;
			}

			.metric-label {
				font-size: 12px;
				color: rgba(255, 255, 255, 0.7);
			}

			.metric-sub {
				font-size: 11px;
				color: rgba(255, 255, 255, 0.45);
			}
		}
	}

	.config-bar {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 16px;
		background: rgba(0, 0, 0, 0.2);
		border: 1px solid rgba(255, 255, 255, 0.06);
		border-radius: 10px;
		padding: 12px 16px;

		.config-item {
			display: flex;
			align-items: center;
			gap: 8px;
			font-size: 13px;
			color: $text-color;

			label {
				font-weight: 500;
				color: rgba(255, 255, 255, 0.8);
			}

			select {
				background: rgba(255, 255, 255, 0.06);
				border: 1px solid rgba(255, 255, 255, 0.12);
				color: #fff;
				border-radius: 8px;
				padding: 6px 10px;
				font-size: 12px;
				cursor: pointer;

				&:focus {
					outline: none;
					border-color: $accent-color;
				}
			}

			&.toggle {
				.toggle-label {
					display: flex;
					align-items: center;
					gap: 8px;
					cursor: pointer;
					user-select: none;
				}
			}
		}
	}

	.actions-row {
		display: flex;
		align-items: center;
		gap: 12px;

		.btn {
			display: inline-flex;
			align-items: center;
			gap: 8px;
			padding: 9px 16px;
			font-size: 13px;
			font-weight: 600;
			border-radius: 9px;
			cursor: pointer;
			transition: all 180ms ease;
			border: none;

			&.primary-btn {
				background: $accent-color;
				color: #fff;

				&:hover:not(:disabled) {
					filter: brightness(1.1);
					transform: translateY(-1px);
				}
			}

			&.secondary-btn {
				background: rgba(255, 255, 255, 0.08);
				color: $text-color;
				border: 1px solid rgba(255, 255, 255, 0.1);

				&:hover:not(:disabled) {
					background: rgba(255, 255, 255, 0.14);
				}
			}

			&:disabled {
				opacity: 0.6;
				cursor: not-allowed;
			}
		}
	}

	.table-container {
		background: rgba(0, 0, 0, 0.25);
		border: 1px solid rgba(255, 255, 255, 0.07);
		border-radius: 12px;
		overflow: hidden;
		max-height: 380px;
		overflow-y: auto;
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 40px 20px;
		gap: 8px;
		color: rgba(255, 255, 255, 0.4);
		text-align: center;

		p {
			font-weight: 600;
			color: rgba(255, 255, 255, 0.7);
			margin: 0;
		}

		span {
			font-size: 12px;
		}
	}

	.backups-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 13px;

		thead {
			background: rgba(255, 255, 255, 0.03);
			border-bottom: 1px solid rgba(255, 255, 255, 0.08);

			th {
				text-align: left;
				padding: 10px 14px;
				font-size: 11px;
				font-weight: 600;
				color: rgba(255, 255, 255, 0.5);
				text-transform: uppercase;
				letter-spacing: 0.04em;

				&.actions-col {
					text-align: right;
					padding-right: 18px;
				}
			}
		}

		tbody {
			tr {
				border-bottom: 1px solid rgba(255, 255, 255, 0.04);
				transition: background 120ms ease;

				&:hover {
					background: rgba(255, 255, 255, 0.03);
				}

				&:last-child {
					border-bottom: none;
				}

				td {
					padding: 11px 14px;
					color: rgba(255, 255, 255, 0.85);

					&.filename-cell {
						display: flex;
						align-items: center;
						gap: 10px;
						max-width: 280px;

						.file-icon-wrap {
							display: flex;
							align-items: center;
							justify-content: center;
							width: 28px;
							height: 28px;
							border-radius: 6px;
							background: rgba(255, 255, 255, 0.06);
							color: $accent-color;
							flex-shrink: 0;

							&.prerestore {
								background: rgba(255, 179, 0, 0.15);
								color: #ffb300;
							}
						}

						.file-name {
							white-space: nowrap;
							overflow: hidden;
							text-overflow: ellipsis;
							font-family: monospace;
							font-size: 12px;
						}
					}

					&.date-cell,
					&.size-cell {
						white-space: nowrap;
						font-size: 12px;
						color: rgba(255, 255, 255, 0.65);
					}

					&.actions-cell {
						text-align: right;
						white-space: nowrap;
					}
				}
			}
		}
	}

	.badge {
		display: inline-block;
		padding: 3px 8px;
		border-radius: 6px;
		font-size: 11px;
		font-weight: 600;

		&.auto {
			background: rgba(0, 180, 255, 0.15);
			color: #40c4ff;
			border: 1px solid rgba(0, 180, 255, 0.25);
		}

		&.manual {
			background: rgba(76, 175, 80, 0.15);
			color: #81c784;
			border: 1px solid rgba(76, 175, 80, 0.25);
		}

		&.safety {
			background: rgba(255, 152, 0, 0.15);
			color: #ffb74d;
			border: 1px solid rgba(255, 152, 0, 0.25);
		}
	}

	.action-icon-btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 30px;
		height: 30px;
		border-radius: 7px;
		background: rgba(255, 255, 255, 0.05);
		color: $text-color;
		border: 1px solid rgba(255, 255, 255, 0.08);
		cursor: pointer;
		margin-left: 6px;
		text-decoration: none;
		transition: all 140ms ease;

		&:hover {
			background: rgba(255, 255, 255, 0.12);
			transform: scale(1.05);
		}

		&.download:hover {
			color: #40c4ff;
			border-color: rgba(64, 196, 255, 0.4);
		}

		&.restore:hover {
			color: #81c784;
			border-color: rgba(129, 199, 132, 0.4);
		}

		&.delete:hover {
			color: #e57373;
			border-color: rgba(229, 115, 115, 0.4);
		}

		&:disabled {
			opacity: 0.4;
			cursor: not-allowed;
		}
	}

	.restore-dialog-backdrop {
		position: fixed;
		top: 0;
		left: 0;
		width: 100vw;
		height: 100vh;
		background: rgba(0, 0, 0, 0.7);
		backdrop-filter: blur(4px);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 10000;
	}

	.restore-dialog {
		background: $bg-color;
		border: 1px solid rgba(255, 255, 255, 0.15);
		border-radius: 14px;
		padding: 24px;
		max-width: 480px;
		width: 90%;
		display: flex;
		flex-direction: column;
		gap: 16px;
		box-shadow: 0 16px 40px rgba(0, 0, 0, 0.5);

		.restore-header {
			display: flex;
			align-items: center;
			gap: 12px;
			color: #ffb74d;

			h4 {
				margin: 0;
				font-size: 18px;
				font-weight: 700;
				color: #fff;
			}
		}

		.restore-desc {
			font-size: 14px;
			color: $text-color;
			line-height: 1.5;
			margin: 0;

			.highlight-file {
				font-family: monospace;
				color: $accent-color;
				word-break: break-all;
			}
		}

		.safety-banner {
			display: flex;
			align-items: flex-start;
			gap: 10px;
			background: rgba(76, 175, 80, 0.1);
			border: 1px solid rgba(76, 175, 80, 0.25);
			border-radius: 8px;
			padding: 10px 12px;
			font-size: 12px;
			color: #c8e6c9;
			line-height: 1.4;

			:global(svg) {
				color: #81c784;
				flex-shrink: 0;
				margin-top: 2px;
			}
		}

		.restore-actions {
			display: flex;
			justify-content: flex-end;
			gap: 12px;
			margin-top: 8px;

			.cancel-btn {
				background: rgba(255, 255, 255, 0.08);
				color: $text-color;
				border: 1px solid rgba(255, 255, 255, 0.12);
				padding: 8px 16px;
				border-radius: 8px;
				cursor: pointer;
				font-size: 13px;

				&:hover:not(:disabled) {
					background: rgba(255, 255, 255, 0.14);
				}
			}

			.confirm-btn {
				background: #e57373;
				color: #fff;
				border: none;
				padding: 8px 18px;
				border-radius: 8px;
				cursor: pointer;
				font-size: 13px;
				font-weight: 600;

				&:hover:not(:disabled) {
					filter: brightness(1.15);
				}
			}
		}
	}
</style>

