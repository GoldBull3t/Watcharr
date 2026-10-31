<script lang="ts">
	import Modal from "@/lib/Modal.svelte";
	import Icon from "@/lib/Icon.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import SpinnerTiny from "@/lib/SpinnerTiny.svelte";
	import { baseURL, req } from "@/lib/util/api";
	import { notify } from "@/lib/util/notify";
	import type { BackupInfo, BackupConfig, GetBackupsResponse } from "@/types";
	import { onMount } from "svelte";

	interface Props {
		open?: boolean;
		onClose: () => void;
	}

	let { open = $bindable(false), onClose }: Props = $props();

	let loading = $state(true);
	let backups: BackupInfo[] = $state([]);
	let config: BackupConfig | undefined = $state();

	let isCreating = $state(false);
	let isRestoring = $state(false);
	let deletingFilename: string | null = $state(null);
	let selectedFormat: "auto" | "sqlite" | "json" = $state("auto");

	let confirmRestoreModalOpen = $state(false);
	let backupToRestore: BackupInfo | null = $state(null);
	let fileInputEl: HTMLInputElement | undefined = $state();
	let isUploading = $state(false);

	async function loadBackups() {
		loading = true;
		try {
			const res = await req.get<GetBackupsResponse>("/backup");
			if (res) {
				backups = res.backups || [];
				config = res.config;
			}
		} catch (err: any) {
			console.error("loadBackups failed:", err);
			notify({
				text: err?.message || "Failed to load backups. Admin access required.",
				type: "error",
			});
		} finally {
			loading = false;
		}
	}

	async function createBackup() {
		if (isCreating) return;
		isCreating = true;
		const nid = notify({
			text: "Creating database snapshot...",
			type: "loading",
		});

		try {
			await req.post("/backup", {
				format: selectedFormat,
			});
			notify({
				id: nid,
				text: "Backup created successfully!",
				type: "success",
			});
			await loadBackups();
		} catch (err: any) {
			console.error("createBackup failed:", err);
			notify({
				id: nid,
				text: err?.message || "Failed to create backup",
				type: "error",
			});
		} finally {
			isCreating = false;
		}
	}

	async function deleteBackup(filename: string) {
		if (
			!confirm(`Are you sure you want to permanently delete "${filename}"?`)
		) {
			return;
		}
		deletingFilename = filename;
		try {
			await req.delete(`/backup/${encodeURIComponent(filename)}`);
			notify({
				text: "Backup deleted successfully.",
				type: "success",
			});
			backups = backups.filter((b) => b.filename !== filename);
		} catch (err: any) {
			console.error("deleteBackup failed:", err);
			notify({
				text: err?.message || "Failed to delete backup",
				type: "error",
			});
		} finally {
			deletingFilename = null;
		}
	}

	async function downloadBackup(b: BackupInfo) {
		try {
			const token = localStorage.getItem("token") || "";
			const res = await fetch(
				`${baseURL}/backup/download/${encodeURIComponent(b.filename)}`,
				{
					headers: {
						Authorization: `Bearer ${token}`,
					},
				},
			);
			if (!res.ok) {
				throw new Error(`Download failed with ${res.status}`);
			}
			const blob = await res.blob();
			const url = URL.createObjectURL(blob);
			const a = document.createElement("a");
			a.href = url;
			a.download = b.filename;
			a.click();
			URL.revokeObjectURL(url);
		} catch (err: any) {
			console.error("downloadBackup failed:", err);
			notify({
				text: err?.message || "Failed to download backup archive",
				type: "error",
			});
		}
	}

	function promptRestore(b: BackupInfo) {
		backupToRestore = b;
		confirmRestoreModalOpen = true;
	}

	async function executeRestore() {
		if (!backupToRestore || isRestoring) return;
		isRestoring = true;
		const filename = backupToRestore.filename;
		confirmRestoreModalOpen = false;

		const nid = notify({
			text: `Restoring database from ${filename}...`,
			type: "loading",
		});

		try {
			await req.post(`/backup/restore/${encodeURIComponent(filename)}`, {});
			notify({
				id: nid,
				text: "Database restored successfully! Refreshing view...",
				type: "success",
				time: 4000,
			});
			setTimeout(() => {
				window.location.reload();
			}, 1500);
		} catch (err: any) {
			console.error("executeRestore failed:", err);
			notify({
				id: nid,
				text: err?.message || "Failed to restore database from backup.",
				type: "error",
			});
		} finally {
			isRestoring = false;
			backupToRestore = null;
		}
	}

	async function handleFileUpload(e: Event) {
		const target = e.target as HTMLInputElement;
		if (!target?.files || target.files.length === 0) return;
		const file = target.files[0];
		if (
			!file.name.endsWith(".db.gz") &&
			!file.name.endsWith(".json.gz") &&
			!file.name.endsWith(".db") &&
			!file.name.endsWith(".json")
		) {
			notify({
				text: "Unsupported archive format. Use .db.gz or .json.gz",
				type: "error",
			});
			return;
		}

		isUploading = true;
		const nid = notify({
			text: `Uploading backup "${file.name}"...`,
			type: "loading",
		});

		try {
			const token = localStorage.getItem("token") || "";
			const formData = new FormData();
			formData.append("backup", file);

			const res = await fetch(`${baseURL}/backup/upload`, {
				method: "POST",
				headers: {
					Authorization: `Bearer ${token}`,
				},
				body: formData,
			});

			if (!res.ok) {
				const errBody = await res.json().catch(() => ({}));
				throw new Error(errBody.error || "Upload failed");
			}

			notify({
				id: nid,
				text: "Backup uploaded and verified successfully!",
				type: "success",
			});
			await loadBackups();
		} catch (err: any) {
			console.error("handleFileUpload failed:", err);
			notify({
				id: nid,
				text: err?.message || "Failed to upload backup.",
				type: "error",
			});
		} finally {
			isUploading = false;
			if (fileInputEl) fileInputEl.value = "";
		}
	}

	function formatDate(d: string): string {
		try {
			const dt = new Date(d);
			if (isNaN(dt.getTime())) return d;
			return dt.toLocaleString(undefined, {
				dateStyle: "medium",
				timeStyle: "short",
			});
		} catch {
			return d;
		}
	}

	onMount(() => {
		loadBackups();
	});
</script>

<Modal title="Disaster Recovery & Database Backups" maxWidth="820px" {onClose}>
	<div class="backup-modal-container">
		<!-- Summary Cards -->
		{#if config}
			<div class="summary-grid">
				<div class="summary-card">
					<span class="summary-label">Automated Backups</span>
					<div class="summary-val">
						{#if config.enabled}
							<span class="status-pill active">Active</span>
						{:else}
							<span class="status-pill inactive">Disabled</span>
						{/if}
						<span class="sub-text">Every 24h</span>
					</div>
				</div>

				<div class="summary-card">
					<span class="summary-label">Retention Policy</span>
					<div class="summary-val">
						<span class="metric">{config.retention}</span>
						<span class="sub-text">snapshots kept</span>
					</div>
				</div>

				<div class="summary-card">
					<span class="summary-label">Available Backups</span>
					<div class="summary-val">
						<span class="metric">{backups.length}</span>
						<span class="sub-text">({config.formattedTotalSize || "0 B"})</span>
					</div>
				</div>

				<div class="summary-card">
					<span class="summary-label">Database Engine</span>
					<div class="summary-val">
						<span class="engine-badge"
							>{config.driver?.toUpperCase() || "SQLITE"}</span
						>
						<span class="sub-text">Zero-downtime WAL</span>
					</div>
				</div>
			</div>
		{/if}

		<!-- Action Bar -->
		<div class="action-bar">
			<div class="create-cluster">
				<select class="format-select" bind:value={selectedFormat}>
					<option value="auto"
						>Auto Format ({config?.driver === "sqlite"
							? "SQLite WAL"
							: "JSON"})</option
					>
					<option value="sqlite">SQLite Snapshot (.db.gz)</option>
					<option value="json">Universal JSON (.json.gz)</option>
				</select>

				<button
					class="btn-action primary"
					disabled={isCreating}
					onclick={createBackup}
				>
					{#if isCreating}
						<SpinnerTiny /> Creating...
					{:else}
						<Icon i="database" wh={16} /> Create Backup Now
					{/if}
				</button>
			</div>

			<div class="upload-cluster">
				<input
					type="file"
					accept=".db.gz,.json.gz,.db,.json"
					class="hidden-file-input"
					bind:this={fileInputEl}
					onchange={handleFileUpload}
				/>
				<button
					class="btn-action secondary"
					disabled={isUploading}
					onclick={() => fileInputEl?.click()}
				>
					{#if isUploading}
						<SpinnerTiny /> Uploading...
					{:else}
						<Icon i="upload" wh={16} /> Upload Archive
					{/if}
				</button>
			</div>
		</div>

		<!-- Backups Table -->
		<div class="table-wrap small-scrollbar">
			{#if loading}
				<div class="loading-wrap">
					<Spinner />
				</div>
			{:else if backups.length === 0}
				<div class="empty-backups">
					<Icon i="database" wh={42} />
					<p>No database snapshots available yet.</p>
					<span
						>Click "Create Backup Now" to generate your first zero-downtime
						snapshot.</span
					>
				</div>
			{:else}
				<table class="backups-table">
					<thead>
						<tr>
							<th>Type</th>
							<th>Filename</th>
							<th>Created</th>
							<th>Size</th>
							<th class="text-right">Actions</th>
						</tr>
					</thead>
					<tbody>
						{#each backups as b (b.filename)}
							<tr>
								<td>
									{#if b.isAuto}
										<span class="tag auto">Automated</span>
									{:else}
										<span class="tag manual">Manual</span>
									{/if}
								</td>
								<td class="filename-cell">
									<div class="file-name" title={b.filename}>
										{b.filename}
									</div>
									<span class="file-sub"
										>{b.type === "sqlite"
											? "Binary SQLite Snapshot"
											: "Universal JSON Dump"}</span
									>
								</td>
								<td class="date-cell">{formatDate(b.createdAt)}</td>
								<td class="size-cell">{b.formattedSize}</td>
								<td class="actions-cell">
									<button
										class="icon-btn download"
										title="Download backup"
										onclick={() => downloadBackup(b)}
									>
										<Icon i="download" wh={16} />
									</button>
									<button
										class="icon-btn restore"
										title="Restore database from this snapshot"
										disabled={isRestoring}
										onclick={() => promptRestore(b)}
									>
										<Icon i="refresh" wh={16} />
									</button>
									<button
										class="icon-btn delete"
										title="Delete snapshot"
										disabled={deletingFilename === b.filename}
										onclick={() => deleteBackup(b.filename)}
									>
										{#if deletingFilename === b.filename}
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
	</div>
</Modal>

<!-- Restore Confirmation Modal -->
{#if confirmRestoreModalOpen && backupToRestore}
	<Modal
		title="⚠️ Confirm Database Restoration"
		maxWidth="520px"
		onClose={() => (confirmRestoreModalOpen = false)}
	>
		<div class="confirm-restore-body">
			<div class="warning-callout">
				<div>
					<strong>Restoring snapshot:</strong>
					<p class="restore-target-name">{backupToRestore.filename}</p>
				</div>
			</div>

			<p>
				Watcharr will replace the active database content with the snapshot
				data. A mandatory safety snapshot will be created automatically before
				restoring to ensure zero data loss.
			</p>

			<div class="restore-actions">
				<button
					class="btn-action secondary"
					onclick={() => (confirmRestoreModalOpen = false)}
				>
					Cancel
				</button>
				<button
					class="btn-action danger"
					disabled={isRestoring}
					onclick={executeRestore}
				>
					{#if isRestoring}
						<SpinnerTiny /> Restoring...
					{:else}
						Yes, Restore Database
					{/if}
				</button>
			</div>
		</div>
	</Modal>
{/if}

<style lang="scss">
	.backup-modal-container {
		display: flex;
		flex-direction: column;
		gap: 20px;
	}

	.summary-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(170px, 1fr));
		gap: 12px;
	}

	.summary-card {
		background: rgba(255, 255, 255, 0.04);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 10px;
		padding: 12px 14px;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.summary-label {
		font-size: 11px;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--text-color-accent, #94a3b8);
	}

	.summary-val {
		display: flex;
		align-items: baseline;
		gap: 8px;

		.metric {
			font-size: 18px;
			font-weight: 700;
			color: var(--text-color, #fff);
		}

		.sub-text {
			font-size: 12px;
			color: var(--text-color-accent, #94a3b8);
		}
	}

	.status-pill {
		font-size: 11px;
		font-weight: 600;
		padding: 2px 8px;
		border-radius: 9999px;

		&.active {
			background: rgba(16, 185, 129, 0.15);
			color: #10b981;
			border: 1px solid rgba(16, 185, 129, 0.3);
		}

		&.inactive {
			background: rgba(239, 68, 68, 0.15);
			color: #ef4444;
			border: 1px solid rgba(239, 68, 68, 0.3);
		}
	}

	.engine-badge {
		font-size: 12px;
		font-weight: 700;
		color: var(--accent-color, #ffb703);
	}

	.action-bar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		flex-wrap: wrap;
	}

	.create-cluster {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.format-select {
		background: rgba(0, 0, 0, 0.35);
		border: 1px solid rgba(255, 255, 255, 0.15);
		color: var(--text-color, #fff);
		border-radius: 8px;
		padding: 7px 10px;
		font-size: 13px;
		outline: none;

		&:focus {
			border-color: var(--accent-color, #ffb703);
		}
	}

	.hidden-file-input {
		display: none;
	}

	.btn-action {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 8px 14px;
		font-size: 13px;
		font-weight: 600;
		border-radius: 8px;
		cursor: pointer;
		border: none;
		transition: all 0.15s ease;

		&.primary {
			background: var(--accent-color, #ffb703);
			color: #0f172a;

			&:hover:not(:disabled) {
				filter: brightness(1.1);
				transform: translateY(-1px);
			}
		}

		&.secondary {
			background: rgba(255, 255, 255, 0.08);
			color: var(--text-color, #fff);
			border: 1px solid rgba(255, 255, 255, 0.12);

			&:hover:not(:disabled) {
				background: rgba(255, 255, 255, 0.15);
			}
		}

		&.danger {
			background: #ef4444;
			color: #fff;

			&:hover:not(:disabled) {
				background: #dc2626;
			}
		}

		&:disabled {
			opacity: 0.6;
			cursor: not-allowed;
		}
	}

	.table-wrap {
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 10px;
		background: rgba(0, 0, 0, 0.25);
		max-height: 380px;
		overflow-y: auto;
	}

	.loading-wrap {
		display: flex;
		justify-content: center;
		padding: 40px;
	}

	.empty-backups {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
		padding: 40px 20px;
		text-align: center;
		color: var(--text-color-accent, #94a3b8);

		p {
			font-size: 15px;
			font-weight: 600;
			margin: 0;
			color: var(--text-color, #fff);
		}

		span {
			font-size: 12px;
		}
	}

	.backups-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 13px;

		th {
			text-align: left;
			padding: 10px 14px;
			font-size: 11px;
			font-weight: 600;
			text-transform: uppercase;
			letter-spacing: 0.05em;
			color: var(--text-color-accent, #94a3b8);
			border-bottom: 1px solid rgba(255, 255, 255, 0.08);
			background: rgba(255, 255, 255, 0.02);
		}

		td {
			padding: 10px 14px;
			border-bottom: 1px solid rgba(255, 255, 255, 0.04);
			vertical-align: middle;
		}

		tr:last-child td {
			border-bottom: none;
		}

		tr:hover td {
			background: rgba(255, 255, 255, 0.02);
		}
	}

	.tag {
		display: inline-block;
		font-size: 11px;
		font-weight: 600;
		padding: 2px 7px;
		border-radius: 4px;

		&.auto {
			background: rgba(59, 130, 246, 0.15);
			color: #60a5fa;
		}

		&.manual {
			background: rgba(245, 158, 11, 0.15);
			color: #f59e0b;
		}
	}

	.filename-cell {
		max-width: 240px;

		.file-name {
			font-weight: 600;
			color: var(--text-color, #fff);
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}

		.file-sub {
			display: block;
			font-size: 11px;
			color: var(--text-color-accent, #94a3b8);
		}
	}

	.date-cell,
	.size-cell {
		color: var(--text-color-accent, #94a3b8);
		white-space: nowrap;
	}

	.actions-cell {
		text-align: right;
		white-space: nowrap;
	}

	.icon-btn {
		background: transparent;
		border: none;
		color: var(--text-color-accent, #94a3b8);
		padding: 6px;
		border-radius: 6px;
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		transition: all 0.15s ease;

		&:hover {
			color: var(--text-color, #fff);
			background: rgba(255, 255, 255, 0.1);
		}

		&.restore:hover {
			color: #38bdf8;
			background: rgba(56, 189, 248, 0.15);
		}

		&.delete:hover {
			color: #ef4444;
			background: rgba(239, 68, 68, 0.15);
		}

		&:disabled {
			opacity: 0.5;
			cursor: not-allowed;
		}
	}

	.confirm-restore-body {
		display: flex;
		flex-direction: column;
		gap: 16px;
		font-size: 14px;
	}

	.warning-callout {
		display: flex;
		align-items: center;
		gap: 12px;
		background: rgba(245, 158, 11, 0.1);
		border: 1px solid rgba(245, 158, 11, 0.3);
		padding: 12px 14px;
		border-radius: 8px;
		color: #fbbf24;

		.restore-target-name {
			margin: 2px 0 0 0;
			font-family: monospace;
			font-size: 12px;
			color: #fff;
		}
	}

	.restore-actions {
		display: flex;
		justify-content: flex-end;
		gap: 10px;
		margin-top: 8px;
	}
</style>
