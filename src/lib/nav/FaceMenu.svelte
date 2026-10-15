<script lang="ts">
	import { store } from "@/store.svelte";
	import Menu from "../Menu.svelte";
	import { parseTokenPayload, userHasPermission } from "../util/helpers";
	import { UserPermission, UserType } from "@/types";
	import ProxyUserLogoutModal from "../logout/ProxyUserLogoutModal.svelte";
	import { goto } from "$app/navigation";
	import { clearWatcharrData } from "../logout";
	import { notify } from "../util/notify";
	import AboutModal from "./AboutModal.svelte";
	import { resolve } from "$app/paths";

	let user = $derived(store.userInfo);
	let proxyUserLogoutShown = $state(false);
	let aboutModalOpen = $state(false);

	function logout() {
		if (user?.type === UserType.Proxy) {
			// Proxy users logout flow is different.
			proxyUserLogoutShown = true;
			return;
		}
		clearWatcharrData();
		goto(resolve("/login"));
	}

	function profile() {
		goto(resolve("/profile"));
	}

	function serverSettings() {
		goto(resolve("/server"));
	}

	function userManagement() {
		goto(resolve("/manage_users"));
	}

	function requestManagement() {
		goto(resolve("/arr_requests"));
	}

	function shareWatchedList() {
		const nid = notify({ type: "loading", text: "Getting link" });
		const ud = parseTokenPayload();
		if (ud?.userId && ud?.username) {
			const shareLink = `${window.location.origin}/lists/${ud.userId}/${ud.username}`;
			navigator.clipboard
				.writeText(shareLink)
				.then(() => {
					notify({ id: nid, type: "success", text: "Copied share link" });
				})
				.catch((r) => {
					console.error("Failed to copy list share link", r);
					notify({
						id: nid,
						type: "error",
						text: `Failed to copy share link:<br/><a href="${shareLink}" target="_blank">${shareLink}</a>`,
						time: 20000,
					});
				});
		} else {
			notify({ id: nid, type: "error", text: "Failed to get link" });
		}
	}

	function closeAbout() {
		aboutModalOpen = false;
	}
</script>

<Menu conf={{ width: "180px", right: "0px", top: "calc(100% + 8px)" }}>
	{#if user?.username}
		<h5 title={user.username}>Hi, {user.username}!</h5>
	{/if}
	<button type="button" class="plain" onclick={() => profile()}>Profile & Settings</button>
	{#if !store.userSettings?.private}
		<button type="button" class="plain" onclick={() => shareWatchedList()}>Share List</button>
	{/if}
	{#if user && userHasPermission(user.permissions, UserPermission.PERM_ADMIN)}
		<button type="button" class="plain" onclick={() => serverSettings()}>Server Settings</button>
		<button type="button" class="plain" onclick={() => userManagement()}>Users</button>
		{#if store.serverFeatures?.sonarr || store.serverFeatures?.radarr}
			<!-- At least one (sonarr/radarr) should be enabled for requests menu item to display. -->
			<button type="button" class="plain" onclick={() => requestManagement()}>
				Requests
			</button>
		{/if}
	{/if}
	<button type="button" class="plain logout" onclick={() => logout()}>Logout</button>
	{#if proxyUserLogoutShown}
		<ProxyUserLogoutModal onClose={() => (proxyUserLogoutShown = false)} />
	{/if}
	<div class="menu-footer-wrap">
		<button
			type="button"
			class="menu-footer"
			onclick={() => {
				aboutModalOpen = !aboutModalOpen;
			}}
		>
			about
		</button>
		<span>•</span>
		<a
			class="menu-footer"
			href="https://github.com/sbondCo/Watcharr/releases"
			target="_blank"
		>
			v{__WATCHARR_VERSION__}
		</a>
	</div>
</Menu>

{#if aboutModalOpen}
	<AboutModal onClose={closeAbout} />
{/if}

<style lang="scss">
	button.plain.logout {
		color: #fca5a5;
		margin-top: 4px;
		border-top: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
		border-radius: 0 0 7px 7px;

		&:hover {
			color: #ffffff;
			background-color: rgba(239, 68, 68, 0.85);
		}
	}
</style>
