<script lang="ts">
	import Checkbox from "@/lib/Checkbox.svelte";
	import Modal from "@/lib/Modal.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import Setting from "@/lib/settings/Setting.svelte";
	import SettingsList from "@/lib/settings/SettingsList.svelte";
	import { req } from "@/lib/util/api";
	import { notify } from "@/lib/util/notify";
	import type { OIDCSetting, ServerConfigByName } from "@/types";
	import { onMount } from "svelte";

	interface Props {
		onClose: () => void;
	}

	let { onClose }: Props = $props();

	let cfg: OIDCSetting = $state({
		enabled: false,
		providerName: "",
		issuerUrl: "",
		clientId: "",
		clientSecret: "",
		redirectUri: "",
		autoRegister: true,
	});
	let loadingCfg = $state(true);
	let saving = $state(false);

	const suggestedRedirect = $derived(`${location.origin}/login`);

	async function getCfg() {
		try {
			const res = await req.get<ServerConfigByName<OIDCSetting>>(
				"/server/config",
				{ params: { s: "OIDC" } },
			);
			cfg = { ...cfg, ...res.value };
		} catch (err) {
			console.error("OIDCModal: getCfg failed!", err);
			notify({
				type: "error",
				text: "Failed to get configuration from server.",
				time: 6000,
			});
		} finally {
			loadingCfg = false;
		}
	}

	async function save() {
		if (cfg.enabled && (!cfg.issuerUrl.trim() || !cfg.clientId.trim())) {
			notify({
				type: "error",
				text: "Issuer URL and Client ID are required to enable OIDC.",
				time: 6000,
			});
			return;
		}
		saving = true;
		const nid = notify({ type: "loading", text: "Saving.." });
		try {
			await req.post("/server/config", cfg, { params: { s: "OIDC" } });
			notify({ id: nid, type: "success", text: "Changes saved!" });
			onClose();
		} catch (err) {
			console.error("OIDCModal: save failed!", err);
			notify({
				id: nid,
				type: "error",
				text: "Failed to save config. Please try again!",
				time: 6000,
			});
		} finally {
			saving = false;
		}
	}

	onMount(() => {
		getCfg();
	});
</script>

<Modal
	title="OpenID Connect (OAuth2)"
	desc="Single sign-on with any OIDC provider (Authentik, Keycloak, Authelia, Google, ...)."
	{onClose}
>
	{#if loadingCfg}
		<Spinner />
	{:else}
		<SettingsList>
			<Setting
				title="Callback URL"
				desc="Register this redirect URI in your provider's application settings."
			>
				<code>{cfg.redirectUri || suggestedRedirect}</code>
			</Setting>
			<Setting title="Enabled" desc="Show the SSO button on the login page." row>
				<Checkbox
					name="OIDC_ENABLED"
					value={cfg.enabled}
					toggled={(on) => (cfg.enabled = on)}
				/>
			</Setting>
			<Setting
				title="Provider Name"
				desc="Shown on the login button, e.g. Authentik."
			>
				<input
					type="text"
					placeholder="SSO"
					bind:value={cfg.providerName}
					disabled={saving}
				/>
			</Setting>
			<Setting
				title="Issuer URL"
				desc="Base URL of the provider. Endpoints are discovered from /.well-known/openid-configuration."
			>
				<input
					type="text"
					placeholder="https://auth.example.com/application/o/watcharr"
					bind:value={cfg.issuerUrl}
					disabled={saving}
				/>
			</Setting>
			<Setting title="Client ID">
				<input type="text" bind:value={cfg.clientId} disabled={saving} />
			</Setting>
			<Setting title="Client Secret" desc="Leave empty for public clients.">
				<input
					type="password"
					bind:value={cfg.clientSecret}
					disabled={saving}
				/>
			</Setting>
			<Setting
				title="Redirect URI"
				desc="Optional. Defaults to {suggestedRedirect}."
			>
				<input
					type="text"
					placeholder={suggestedRedirect}
					bind:value={cfg.redirectUri}
					disabled={saving}
				/>
			</Setting>
			<Setting
				title="Auto Register"
				desc="Create a Watcharr account automatically on first SSO login. New users get the default user permissions."
				row
			>
				<Checkbox
					name="OIDC_AUTO_REGISTER"
					value={cfg.autoRegister}
					toggled={(on) => (cfg.autoRegister = on)}
				/>
			</Setting>
			<div class="btns">
				<button onclick={() => save()} disabled={saving}>Save</button>
			</div>
		</SettingsList>
	{/if}
</Modal>

<style lang="scss">
	.btns {
		display: flex;
		flex-flow: row;
		justify-content: flex-end;

		button {
			width: max-content;
			padding-left: 15px;
			padding-right: 15px;
		}
	}
</style>

