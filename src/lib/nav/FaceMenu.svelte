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
	import { t, i18n } from "@/lib/i18n";

	let user = $derived(store.userInfo);
	let proxyUserLogoutShown = $state(false);
	let aboutModalOpen = $state(false);

	function toggleLanguage() {
		i18n.setLocale(i18n.currentLocale === "pt" ? "en" : "pt");
	}

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
		const nid = notify({ type: "loading", text: t("nav.getting_link") });
		const ud = parseTokenPayload();
		if (ud?.userId && ud?.username) {
			const shareLink = `${window.location.origin}/lists/${ud.userId}/${ud.username}`;
			navigator.clipboard
				.writeText(shareLink)
				.then(() => {
					notify({ id: nid, type: "success", text: t("nav.copied_link") });
				})
				.catch((r) => {
					console.error("Failed to copy list share link", r);
					notify({
						id: nid,
						type: "error",
						text: `${t("nav.failed_link")}:<br/><a href="${shareLink}" target="_blank">${shareLink}</a>`,
						time: 20000,
					});
				});
		} else {
			notify({ id: nid, type: "error", text: t("nav.failed_link") });
		}
	}

	function closeAbout() {
		aboutModalOpen = false;
	}
</script>

<Menu conf={{ width: "190px", right: "0px", top: "calc(100% + 8px)" }}>
	{#if user?.username}
		<h5 title={user.username}>Hi, {user.username}!</h5>
	{/if}
	<button type="button" class="plain" onclick={() => profile()}
		>{t("nav.profile_settings")}</button
	>
	{#if !store.userSettings?.private}
		<button type="button" class="plain" onclick={() => shareWatchedList()}
			>{t("nav.share_list")}</button
		>
	{/if}
	{#if user && userHasPermission(user.permissions, UserPermission.PERM_ADMIN)}
		<button type="button" class="plain" onclick={() => serverSettings()}
			>{t("nav.server_settings")}</button
		>
		<button type="button" class="plain" onclick={() => userManagement()}
			>{t("nav.manage_users")}</button
		>
		{#if store.serverFeatures?.sonarr || store.serverFeatures?.radarr}
			<!-- At least one (sonarr/radarr) should be enabled for requests menu item to display. -->
			<button type="button" class="plain" onclick={() => requestManagement()}>
				{t("nav.requests")}
			</button>
		{/if}
	{/if}
	<button type="button" class="plain lang-toggle" onclick={toggleLanguage}>
		🌐 {i18n.currentLocale === "pt" ? "Português" : "English"}
	</button>
	<button type="button" class="plain logout" onclick={() => logout()}
		>{t("nav.logout")}</button
	>
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
