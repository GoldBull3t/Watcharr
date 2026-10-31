<script lang="ts">
	import Modal from "@/lib/Modal.svelte";
	import Icon from "@/lib/Icon.svelte";
	import SpinnerTiny from "@/lib/SpinnerTiny.svelte";
	import Checkbox from "@/lib/Checkbox.svelte";
	import { store } from "@/store.svelte";
	import { req } from "@/lib/util/api";
	import { notify } from "@/lib/util/notify";
	import type { PrivateUser, UserSettings } from "@/types";

	interface Props {
		open?: boolean;
		onClose: () => void;
	}

	let { open = $bindable(false), onClose }: Props = $props();

	let webhookUrl = $state(store.userSettings?.activityWebhookUrl ?? "");
	let isEnabled = $state(store.userSettings?.activityWebhookEnabled ?? false);
	let filterType: "all" | "finished" | "watching" | "rating" = $state(
		store.userSettings?.activityWebhookFilter ?? "all",
	);

	let isSaving = $state(false);

	$effect(() => {
		if (store.userSettings) {
			webhookUrl = store.userSettings.activityWebhookUrl ?? "";
			isEnabled = store.userSettings.activityWebhookEnabled ?? false;
			filterType = store.userSettings.activityWebhookFilter ?? "all";
		}
	});

	let detectedPlatform = $derived.by(() => {
		const u = webhookUrl.toLowerCase().trim();
		if (!u) return null;
		if (u.includes("discord.com") || u.includes("discordapp.com"))
			return "discord";
		if (u.includes("hooks.slack.com")) return "slack";
		if (u.includes("ntfy.sh") || u.includes("/ntfy")) return "ntfy";
		if (u.includes(":8123") || u.includes("/api/webhook/"))
			return "homeassistant";
		return "generic";
	});

	async function saveSettings() {
		isSaving = true;
		try {
			const res = await req.put<UserSettings>("/user/settings", {
				activityWebhookUrl: webhookUrl.trim(),
				activityWebhookEnabled: isEnabled,
				activityWebhookFilter: filterType,
			});
			if (res) {
				store.userSettings = res;
			}
			notify({
				text: "Real-time activity webhook saved successfully!",
				type: "success",
			});
			onClose();
		} catch (err: any) {
			console.error("Failed to save activity webhook settings:", err);
			notify({
				text: err?.message || "Failed to save webhook settings",
				type: "error",
			});
		} finally {
			isSaving = false;
		}
	}
</script>

<Modal title="Real-Time Outbound Activity Webhooks" maxWidth="620px" {onClose}>
	<div class="webhook-modal-body">
		<p class="modal-desc">
			Broadcast events (Watched, Status Changes, Ratings) in real time to your
			Homelab, Home Assistant automation, Discord channel, Slack, or Ntfy topic.
		</p>

		<!-- Enable Toggle -->
		<div class="setting-row">
			<label class="checkbox-label" for="activity_webhook_enabled">
				<Checkbox name="activity_webhook_enabled" bind:value={isEnabled} />
				<div class="label-text">
					<strong>Enable Activity Webhook</strong>
					<span class="desc"
						>Automatically send an outbound JSON payload whenever you update
						your media list.</span
					>
				</div>
			</label>
		</div>

		<!-- Webhook URL Input -->
		<div class="input-block">
			<div class="label-row">
				<label for="act_webhook_url">Webhook URL</label>
				<div class="platform-indicators">
					{#if detectedPlatform === "discord"}
						<span class="badge discord"
							><Icon i="discord" wh={12} /> Discord</span
						>
					{:else if detectedPlatform === "homeassistant"}
						<span class="badge ha">Home Assistant</span>
					{:else if detectedPlatform === "slack"}
						<span class="badge slack">Slack</span>
					{:else if detectedPlatform === "ntfy"}
						<span class="badge ntfy">Ntfy</span>
					{:else if detectedPlatform === "generic"}
						<span class="badge generic">JSON Webhook</span>
					{/if}
				</div>
			</div>
			<input
				id="act_webhook_url"
				type="url"
				placeholder="https://discord.com/api/webhooks/... or http://homeassistant:8123/api/webhook/..."
				bind:value={webhookUrl}
			/>
			<span class="input-hint">
				Works with Discord webhook URLs, Home Assistant webhooks, Slack, or any
				HTTP endpoint.
			</span>
		</div>

		<!-- Filter Type -->
		<div class="input-block">
			<label for="act_filter">Event Filter</label>
			<select id="act_filter" bind:value={filterType}>
				<option value="all"
					>All Events (Status changes, Ratings, Watched)</option
				>
				<option value="finished"
					>Only Finished (When an item or episode is marked completed)</option
				>
				<option value="watching">Only Watching (When actively watching)</option>
				<option value="rating">Only Ratings (When rating is updated)</option>
			</select>
		</div>

		<!-- Action Footer -->
		<div class="modal-footer">
			<button class="btn-action secondary" onclick={onClose}> Cancel </button>
			<button
				class="btn-action primary"
				disabled={isSaving}
				onclick={saveSettings}
			>
				{#if isSaving}
					<SpinnerTiny /> Saving...
				{:else}
					Save Settings
				{/if}
			</button>
		</div>
	</div>
</Modal>

<style lang="scss">
	.webhook-modal-body {
		display: flex;
		flex-direction: column;
		gap: 20px;
	}

	.modal-desc {
		font-size: 13px;
		color: var(--text-color-accent, #94a3b8);
		margin: 0;
		line-height: 1.5;
	}

	.setting-row {
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 8px;
		padding: 12px;
	}

	.input-block {
		display: flex;
		flex-direction: column;
		gap: 6px;

		label {
			font-size: 13px;
			font-weight: 600;
			color: var(--text-color, #fff);
		}

		input,
		select {
			background: rgba(0, 0, 0, 0.35);
			border: 1px solid rgba(255, 255, 255, 0.15);
			border-radius: 8px;
			padding: 9px 12px;
			color: var(--text-color, #fff);
			font-size: 13px;
			outline: none;

			&:focus {
				border-color: var(--accent-color, #ffb703);
			}
		}

		.input-hint {
			font-size: 11px;
			color: var(--text-color-accent, #94a3b8);
		}
	}

	.label-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.platform-indicators {
		display: flex;
		gap: 6px;

		.badge {
			font-size: 11px;
			font-weight: 600;
			padding: 2px 7px;
			border-radius: 4px;

			&.discord {
				background: rgba(88, 101, 242, 0.2);
				color: #818cf8;
				border: 1px solid rgba(88, 101, 242, 0.3);
			}

			&.ha {
				background: rgba(14, 165, 233, 0.2);
				color: #38bdf8;
				border: 1px solid rgba(14, 165, 233, 0.3);
			}

			&.slack {
				background: rgba(234, 179, 8, 0.2);
				color: #facc15;
				border: 1px solid rgba(234, 179, 8, 0.3);
			}

			&.ntfy {
				background: rgba(16, 185, 129, 0.2);
				color: #34d399;
				border: 1px solid rgba(16, 185, 129, 0.3);
			}

			&.generic {
				background: rgba(255, 255, 255, 0.1);
				color: #fff;
			}
		}
	}

	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: 10px;
		margin-top: 10px;
	}

	.btn-action {
		padding: 8px 16px;
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

		&:disabled {
			opacity: 0.6;
			cursor: not-allowed;
		}
	}
</style>
