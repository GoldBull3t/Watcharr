<script lang="ts">
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import type { AuthResponse, AvailableAuthProviders } from "@/types";
	import { onMount } from "svelte";
	import { notify, unNotify } from "@/lib/util/notify";
	import { noAuthReq } from "@/lib/util/api";
	import { ReqerError } from "@/lib/util/fetch";
	import Icon from "@/lib/Icon.svelte";

	let error: string | undefined = $state();

	onMount(() => {
		if (localStorage.getItem("token")) {
			goto(resolve("/"));
		}

		noAuthReq.get<AvailableAuthProviders>("/auth/available").then((r) => {
			if (r) {
				if (!r.isInSetup) {
					console.log("Server not in setup.. navigating to login page.");
					goto(resolve("/login"));
				}
			}
		});
	});

	function handleLogin(ev: SubmitEvent) {
		ev.preventDefault();
		const fd = new FormData(ev.target! as HTMLFormElement);
		const user = fd.get("username");
		const pass = fd.get("password");

		if (!user || !pass) {
			error = "Username and Password fields are required";
			return;
		}

		const nid = notify({ text: "Setting Up Admin User", type: "loading" });
		noAuthReq
			.post<AuthResponse>("/setup/create_admin", {
				username: user,
				password: pass,
			})
			.then((resp) => {
				if (resp.token) {
					console.log("Received token... logging in.");
					localStorage.setItem("token", resp.token);
					goto(resolve("/"));
					notify({ id: nid, text: `Welcome ${user}!`, type: "success" });
				}
			})
			.catch((err) => {
				error = ReqerError.getMsg(err, "Setting up user failed");
				unNotify(nid);
			});
	}
</script>

<svelte:head>
	<title>Setup Admin - Watcharr</title>
</svelte:head>

<div class="setup-wrapper">
	<div class="setup-card">
		<div class="setup-header">
			<div class="icon-bubble">
				<Icon i="lock-closed" wh={26} />
			</div>
			<h2>Setup Administrator</h2>
			<p class="subtitle">
				Welcome to Watcharr! Configure your primary administrator credentials to
				secure and manage your server.
			</p>
		</div>

		{#if error}
			<div class="error-banner">
				<Icon i="close-circle" wh={18} />
				<span>{error}</span>
			</div>
		{/if}

		<form onsubmit={handleLogin} class="setup-form">
			<div class="form-group">
				<label for="username">Admin Username</label>
				<input
					id="username"
					type="text"
					name="username"
					placeholder="Choose an admin username"
					autocomplete="username"
					required
				/>
			</div>

			<div class="form-group">
				<label for="password">Admin Password</label>
				<input
					id="password"
					type="password"
					name="password"
					placeholder="Create a strong password"
					autocomplete="new-password"
					required
				/>
			</div>

			<div class="submit-actions">
				<button type="submit" class="primary-btn"> Complete Setup </button>
			</div>
		</form>
	</div>
</div>

<style lang="scss">
	.setup-wrapper {
		display: flex;
		justify-content: center;
		align-items: center;
		width: 100%;
		padding: 20px;
	}

	.setup-card {
		display: flex;
		flex-direction: column;
		width: 100%;
		max-width: 440px;
		padding: 36px 32px;
		border-radius: 20px;
		background: rgba(255, 255, 255, 0.035);
		border: 1px solid rgba(255, 255, 255, 0.09);
		box-shadow: 0 20px 50px -10px rgba(0, 0, 0, 0.5);
		backdrop-filter: blur(16px);
		-webkit-backdrop-filter: blur(16px);
	}

	.setup-header {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
		margin-bottom: 24px;
		text-align: center;

		.icon-bubble {
			display: flex;
			align-items: center;
			justify-content: center;
			width: 52px;
			height: 52px;
			border-radius: 50%;
			background: rgba(99, 102, 241, 0.15);
			border: 1px solid rgba(99, 102, 241, 0.3);
			color: #818cf8;
			margin-bottom: 6px;
		}

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

	.setup-form {
		display: flex;
		flex-direction: column;
		gap: 18px;
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

		.submit-actions {
			margin-top: 8px;

			.primary-btn {
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
</style>
