<script lang="ts">
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import { page } from "$app/state";
	import Icon from "@/lib/Icon.svelte";
	import { type AuthResponse, type AvailableAuthProviders } from "@/types";
	import { noAuthReq } from "@/lib/util/api";
	import { onMount } from "svelte";
	import { notify, unNotify } from "@/lib/util/notify";
	import { ReqerError } from "@/lib/util/fetch";

	let error: string | undefined = $state();
	let login = $state(true);
	let availableProviders: string[] = $state([]);
	let apHeader = $state(false);
	let apPlex = $state(false);
	let apOidc = $state(false);
	let oidcProviderName = $state("");
	let signupEnabled = $state(true);
	let useEmby = $state(false);
	let noAuto = $state(false);

	onMount(() => {
		if (localStorage.getItem("token")) {
			goto(resolve("/"));
		}

		if (!error && page.url.searchParams.get("again")) {
			error = "Please Login Again";
		}
		if (page.url.searchParams.get("noAuto") == "1") {
			console.info(
				"login: Found noAuto param.. auto logins should be disabled now.",
			);
			noAuto = true;
		}

		processAvailableLoginMethods();
		handleOidcCallback();
	});

	/**
	 * Get and process available login methods.
	 */
	async function processAvailableLoginMethods() {
		let r: AvailableAuthProviders;
		try {
			r = await noAuthReq.get<AvailableAuthProviders>("/auth/available");
		} catch {
			notify({
				type: "error",
				time: 10000,
				text: "Failed to get available login methods",
			});
			return;
		}
		if (!r) {
			console.log("AvailableAuth: no response data");
		}
		if (r.isInSetup) {
			console.log(
				"AvailableAuth: Server is in setup.. navigating to web setup page.",
			);
			goto(resolve("/setup"));
		}
		availableProviders = r.available;
		apHeader = availableProviders?.includes("header");
		apPlex = availableProviders?.includes("plex");
		apOidc = availableProviders?.includes("oidc");
		oidcProviderName = r.oidcProviderName ?? "";
		signupEnabled = r.signupEnabled;
		useEmby = r.useEmby;
		if (r.headerAuthAutoLogin && !noAuto) {
			console.log(
				"AvailableAuth: handling headerAuthAutoLogin.. calling proxyLogin automatically now.",
			);
			proxyLogin(true);
		}
	}

	function handleLogin(ev: SubmitEvent) {
		ev.preventDefault();
		const fd = new FormData(ev.target! as HTMLFormElement);
		const user = fd.get("username");
		const pass = fd.get("password");

		if (!user || !pass) {
			error = "Username and Password fields are required";
			return;
		}

		let customAuthEP = "";
		if ((ev.submitter as HTMLButtonElement)?.name === "jellyfin") {
			customAuthEP = "jellyfin";
		}

		const nid = notify({ text: "Logging in", type: "loading" });
		noAuthReq
			.post<AuthResponse>(`/auth${login ? `/${customAuthEP}` : "/register"}`, {
				username: user,
				password: pass,
			})
			.then((resp) => {
				if (resp?.token) {
					console.log("Received token... logging in.");
					localStorage.setItem("token", resp.token);
					if (useEmby) {
						localStorage.setItem("useEmby", "1");
					} else {
						localStorage.removeItem("useEmby");
					}
					goto(resolve("/"));
					notify({ id: nid, text: `Welcome ${user}!`, type: "success" });
				}
			})
			.catch((err) => {
				error = ReqerError.getMsg(err, "Login failed");
				unNotify(nid);
			});
	}

	async function plexLogin() {
		try {
			const { preparePlexAuth, doPlexLogin, plexPinPoll } =
				await import("@/lib/util/plex");
			const p = preparePlexAuth();
			const pin = await doPlexLogin(p);
			plexPinPoll(pin, p, (err, token) => {
				if (err) {
					error = "Plex Auth Failed";
					console.error("Plex auth failed!", err);
					return;
				}
				const nid = notify({ text: "Logging in", type: "loading" });
				noAuthReq
					.post<AuthResponse>("/auth/plex", {
						token,
						clientIdentifier: p.clientId,
					})
					.then((resp) => {
						if (resp?.token) {
							console.log("Received token... logging in.");
							localStorage.setItem("token", resp.token);
							goto(resolve("/"));
							notify({ id: nid, text: `Welcome!`, type: "success" });
						}
					})
					.catch((err) => {
						console.error("plexLogin: Fail", err);
						error = ReqerError.getMsg(err, "Login failed");
						notify({ id: nid, text: `Failed!`, type: "error" });
					});
			});
		} catch (err) {
			console.error("plexLogin: failed!", err);
			error = "Plex login failed";
		}
	}

	const oidcRedirectUri = () => `/login`;

	/** Start the OAuth2 / OIDC login by redirecting to the provider (#461). */
	async function oidcLogin() {
		try {
			const r = await noAuthReq.get<{ authUrl: string }>("/auth/oidc/login", {
				params: { redirectUri: oidcRedirectUri() },
			});
			if (!r?.authUrl) throw new Error("missing authUrl");
			window.location.assign(r.authUrl);
		} catch (err) {
			console.error("oidcLogin: failed!", err);
			error = ReqerError.getMsg(err, "Failed to start SSO login");
		}
	}

	/** Complete the OIDC login when the provider redirects back with ?code=...&state=.... */
	function handleOidcCallback() {
		const code = page.url.searchParams.get("code");
		const state = page.url.searchParams.get("state");
		const oauthErr = page.url.searchParams.get("error_description") ?? page.url.searchParams.get("error");
		if (oauthErr && !code) {
			error = `SSO login failed: ${oauthErr}`;
			history.replaceState(history.state, "", location.pathname);
			return;
		}
		if (!code || !state) return;
		// Remove the one-time code from the address bar right away.
		history.replaceState(history.state, "", location.pathname);
		const nid = notify({ text: "Logging in", type: "loading" });
		noAuthReq
			.post<AuthResponse>("/auth/oidc/callback", {
				code,
				state,
				redirectUri: oidcRedirectUri(),
			})
			.then((resp) => {
				if (resp?.token) {
					localStorage.setItem("token", resp.token);
					goto(resolve("/"));
					notify({ id: nid, text: `Welcome!`, type: "success" });
				}
			})
			.catch((err) => {
				console.error("handleOidcCallback: Fail", err);
				error = ReqerError.getMsg(err, "SSO login failed");
				unNotify(nid);
			});
	}

	function proxyLogin(auto = false) {
		const nid = notify({ text: "Logging in", type: "loading" });
		noAuthReq
			.post<AuthResponse>(`/auth/proxy`)
			.then((resp) => {
				if (resp?.token) {
					console.log("Received token... logging in.");
					localStorage.setItem("token", resp.token);
					goto(resolve("/"));
					notify({ id: nid, text: `Welcome!`, type: "success" });
				}
			})
			.catch((err) => {
				error = ReqerError.getMsg(err, "Login failed");
				if (auto) {
					notify({
						id: nid,
						text: `Automatic SSO Login Failed!`,
						type: "error",
					});
				} else {
					unNotify(nid);
				}
			});
	}
</script>

<svelte:head>
	<title>{login ? "Sign In" : "Create Account"} - Watcharr</title>
</svelte:head>

<div class="auth-wrapper">
	<div class="auth-card">
		{#if signupEnabled}
			<nav class="auth-tabs" aria-label="Authentication modes">
				<button
					type="button"
					class={`tab-pill ${login ? "active" : ""}`}
					onclick={() => {
						login = true;
						error = undefined;
					}}
				>
					Sign In
				</button>
				<button
					type="button"
					class={`tab-pill ${!login ? "active" : ""}`}
					onclick={() => {
						login = false;
						error = undefined;
					}}
				>
					Create Account
				</button>
			</nav>
		{/if}

		<div class="auth-header">
			<h2>{login ? "Welcome Back" : "Create Account"}</h2>
			<p class="subtitle">
				{login
					? "Enter your credentials to access your media collection."
					: "Set up your credentials to start tracking movies, series, and games."}
			</p>
		</div>

		{#if error}
			<div class="error-banner">
				<Icon i="close-circle" wh={18} />
				<span>{error}</span>
			</div>
		{/if}

		<form onsubmit={handleLogin} class="auth-form">
			<div class="form-group">
				<label for="username">Username</label>
				<input
					id="username"
					type="text"
					name="username"
					placeholder="Enter your username"
					autocomplete="username"
					required
				/>
			</div>

			<div class="form-group">
				<label for="password">Password</label>
				<input
					id="password"
					type="password"
					name="password"
					placeholder="Enter your password"
					autocomplete={login ? "current-password" : "new-password"}
					required
				/>
			</div>

			<div class="login-actions">
				<button type="submit" class="primary-submit-btn">
					{login ? "Sign In" : "Create Account"}
				</button>
			</div>

			{#if login && availableProviders?.length > 0}
				<div class="provider-divider">
					<span>or continue with</span>
				</div>

				<div class="external-providers">
					{#if availableProviders.find((ap) => ap === "jellyfin")}
						<button type="submit" name="jellyfin" class="provider-btn jellyfin">
							{#if useEmby}
								<Icon i="emby" wh={18} />
								<span>Emby</span>
							{:else}
								<Icon i="jellyfin" wh={18} />
								<span>Jellyfin</span>
							{/if}
						</button>
					{/if}

					{#if apPlex}
						<button type="button" onclick={plexLogin} class="provider-btn plex">
							<Icon i="plex" wh={18} />
							<span>Plex</span>
						</button>
					{/if}

					{#if apOidc}
						<button
							type="button"
							onclick={oidcLogin}
							class="provider-btn sso"
						>
							<Icon i="lock-closed" wh={18} />
							<span>Continue with {oidcProviderName || "SSO"}</span>
						</button>
					{/if}

					{#if apHeader}
						<button
							type="button"
							onclick={() => proxyLogin()}
							class="provider-btn sso"
						>
							<Icon i="lock-closed" wh={18} />
							<span>Single Sign-On</span>
						</button>
					{/if}
				</div>
			{/if}
		</form>
	</div>
</div>

<style lang="scss">
	.auth-wrapper {
		display: flex;
		justify-content: center;
		align-items: center;
		width: 100%;
		padding: 20px;
	}

	.auth-card {
		display: flex;
		flex-direction: column;
		width: 100%;
		max-width: 420px;
		padding: 32px 30px;
		border-radius: 20px;
		background: rgba(255, 255, 255, 0.035);
		border: 1px solid rgba(255, 255, 255, 0.09);
		box-shadow: 0 20px 50px -10px rgba(0, 0, 0, 0.5);
		backdrop-filter: blur(16px);
		-webkit-backdrop-filter: blur(16px);
	}

	.auth-tabs {
		display: inline-flex;
		padding: 4px;
		background: rgba(0, 0, 0, 0.25);
		border: 1px solid rgba(255, 255, 255, 0.06);
		border-radius: 30px;
		margin-bottom: 24px;

		.tab-pill {
			flex: 1;
			display: flex;
			align-items: center;
			justify-content: center;
			height: 36px;
			border-radius: 20px;
			font-size: 13.5px;
			font-weight: 500;
			color: $text-muted;
			background: transparent;
			border: none;
			cursor: pointer;
			transition: all 150ms ease;

			&.active {
				background: $accent-color;
				color: #fff;
				font-weight: 600;
				box-shadow: 0 2px 8px rgba(0, 0, 0, 0.25);
			}

			&:hover:not(.active) {
				color: #fff;
			}
		}
	}

	.auth-header {
		display: flex;
		flex-direction: column;
		gap: 6px;
		margin-bottom: 22px;
		text-align: center;

		h2 {
			font-size: 24px;
			font-weight: 700;
			letter-spacing: -0.02em;
			color: #fff;
			margin: 0;
		}

		.subtitle {
			font-size: 13.5px;
			color: $text-muted;
			margin: 0;
			line-height: 1.45;
		}
	}

	.error-banner {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px 16px;
		margin-bottom: 20px;
		border-radius: 12px;
		background: rgba(239, 68, 68, 0.12);
		border: 1px solid rgba(239, 68, 68, 0.25);
		color: #fca5a5;
		font-size: 13.5px;
		font-weight: 500;
	}

	.auth-form {
		display: flex;
		flex-direction: column;
		gap: 16px;
		width: 100%;

		.form-group {
			display: flex;
			flex-direction: column;
			gap: 6px;

			label {
				font-size: 13px;
				font-weight: 600;
				color: rgba(255, 255, 255, 0.85);
			}

			input {
				width: 100%;
				height: 44px;
				padding: 0 14px;
				font-size: 14px;
				border-radius: 10px;
				background: rgba(0, 0, 0, 0.25);
				border: 1px solid rgba(255, 255, 255, 0.12);
				color: #fff;
				transition: all 150ms ease;

				&:focus {
					outline: none;
					border-color: $accent-color;
					box-shadow: 0 0 0 3px rgba(229, 160, 0, 0.15);
				}

				&::placeholder {
					color: rgba(255, 255, 255, 0.35);
				}
			}
		}

		.login-actions {
			margin-top: 6px;

			.primary-submit-btn {
				width: 100%;
				height: 44px;
				border-radius: 10px;
				font-size: 15px;
				font-weight: 600;
				background: $accent-color;
				color: #fff;
				border: none;
				cursor: pointer;
				box-shadow: 0 4px 14px rgba(0, 0, 0, 0.25);
				transition: all 150ms ease;

				&:hover {
					background: $accent-color-hover;
					transform: translateY(-1px);
				}
			}
		}
	}

	.provider-divider {
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 8px 0;
		position: relative;

		&::before {
			content: "";
			position: absolute;
			left: 0;
			right: 0;
			height: 1px;
			background: rgba(255, 255, 255, 0.08);
		}

		span {
			position: relative;
			background: #141720;
			padding: 0 12px;
			font-size: 12px;
			color: rgba(255, 255, 255, 0.45);
			text-transform: uppercase;
			letter-spacing: 0.05em;
		}
	}

	.external-providers {
		display: flex;
		flex-direction: column;
		gap: 8px;

		.provider-btn {
			display: flex;
			align-items: center;
			justify-content: center;
			gap: 10px;
			width: 100%;
			height: 40px;
			border-radius: 10px;
			font-size: 13.5px;
			font-weight: 500;
			background: rgba(255, 255, 255, 0.04);
			border: 1px solid rgba(255, 255, 255, 0.08);
			color: #fff;
			cursor: pointer;
			transition: all 150ms ease;

			&:hover {
				background: rgba(255, 255, 255, 0.09);
				border-color: rgba(255, 255, 255, 0.16);
				transform: translateY(-1px);
			}
		}
	}
</style>
