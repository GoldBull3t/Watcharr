<script lang="ts">
	import Checkbox from "@/lib/Checkbox.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import { notify } from "@/lib/util/notify";
	import type {
		Content,
		RadarrSettings,
		ServerConfig,
		SonarrSettings,
	} from "@/types";
	import SonarrModal from "./modals/SonarrModal.svelte";
	import SettingsList from "@/lib/settings/SettingsList.svelte";
	import Setting from "@/lib/settings/Setting.svelte";
	import SettingButton from "@/lib/settings/SettingButton.svelte";
	import RadarrModal from "./modals/RadarrModal.svelte";
	import { getServerFeatures, req } from "@/lib/util/api";
	import Stats from "@/lib/stats/Stats.svelte";
	import Error from "@/lib/Error.svelte";
	import Stat from "@/lib/stats/Stat.svelte";
	import TwitchModal from "./modals/TwitchModal.svelte";
	import RegionDropDown from "@/lib/RegionDropDown.svelte";
	import LanguageDropDown from "@/lib/LanguageDropDown.svelte";
	import TaskScheduleModal from "./modals/TaskScheduleModal.svelte";
	import TrustedHeaderAuthModal from "./modals/TrustedHeaderAuthModal.svelte";
	import { resolve } from "$app/paths";

	let serverConfig: ServerConfig | undefined = $state();
	let jellyfinOrEmby = $derived(serverConfig?.USE_EMBY ? "Emby" : "Jellyfin");
	let sonarrModalOpen = $state(false);
	let sonarrServerEditing: SonarrSettings | undefined = $state();
	let sonarrModalEditing = $state(false);
	let radarrModalOpen = $state(false);
	let radarrServerEditing: RadarrSettings | undefined = $state();
	let radarrModalEditing = $state(false);
	let twitchModalOpen = $state(false);
	let taskScheduleModalOpen = $state(false);
	let headerSSOModalOpen = $state(false);
	// Disabled vars for disabling inputs until api request completes
	let signupDisabled = $state(false);
	let debugDisabled = $state(false);
	let jfDisabled = $state(false);
	let jfWebhookEnabledDisabled = $state(false);
	let tmdbkDisabled = $state(false);
	let omdbkDisabled = $state(false);
	let mdblistkDisabled = $state(false);
	let plexHostDisabled = $state(false);
	let countryDisabled = $state(false);
	let languageDisabled = $state(false);
	let useEmbyDisabled = $state(false);
	let overseerrHostDisabled = $state(false);
	let overseerrKeyDisabled = $state(false);
	let overseerrTesting = $state(false);

	async function testOverseerrConnection() {
		if (!serverConfig?.OVERSEERR_HOST || !serverConfig?.OVERSEERR_KEY) {
			notify({ type: "error", text: "Host and API Key must be set to test!" });
			return;
		}
		overseerrTesting = true;
		try {
			const res = await req.post<{ version?: string; commitTag?: string }>("/seerr/test", {
				host: serverConfig.OVERSEERR_HOST,
				key: serverConfig.OVERSEERR_KEY,
			});
			notify({
				type: "success",
				text: `Connected to Overseerr/Jellyseerr! Version: ${res.version || res.commitTag || "OK"}`,
			});
		} catch (err) {
			console.error("Overseerr test failed:", err);
			notify({
				type: "error",
				text: "Failed to connect to Overseerr/Jellyseerr. Please check Host and Key.",
			});
		} finally {
			overseerrTesting = false;
		}
	}

	let jellyfinWebhookEndpoint: string | undefined = $derived.by(() => {
		if (
			!serverConfig ||
			!serverConfig.JELLYFIN_WEBHOOK_ENABLED ||
			!serverConfig.JELLYFIN_WEBHOOK_KEY
		) {
			return "Error: Please try reloading the page OR re-toggling the webhook setting to get the address!";
		}
		const route = serverConfig.USE_EMBY ? "emby" : "jellyfin";
		return `${location.protocol}//${location.host}/api/${route}/webhook/${serverConfig.JELLYFIN_WEBHOOK_KEY}`;
	});

	async function getServerConfig() {
		serverConfig = await req.get<ServerConfig>(`/server/config`);
	}

	export function updateServerConfig<K extends keyof ServerConfig>(
		name: K,
		value: ServerConfig[K],
		done?: (respData?: object) => void,
	) {
		if (!serverConfig) {
			console.error("updateServerConfig: No server config to update!");
			notify({ type: "error", text: "No server config to update!" });
			return;
		}
		console.log("Updating server setting", name, "to", value);
		const originalValue = serverConfig[name];
		const nid = notify({ type: "loading", text: "Updating" });
		let ep = "/server/config";
		if (name === "PLEX_HOST") {
			ep = "/server/config/plex_host";
		} else if (name === "JELLYFIN_WEBHOOK_ENABLED") {
			ep = "/server/config/JELLYFIN_WEBHOOK_ENABLED";
		}
		req
			.postWhole<object>(ep, { key: name, value: value })
			.then((r) => {
				if (r.status === 200) {
					serverConfig![name] = value;
					notify({ id: nid, type: "success", text: "Updated" });
					if (typeof done !== "undefined") done(r?.body);
				}
			})
			.catch((err) => {
				console.error("Failed to update user setting", err);
				notify({ id: nid, type: "error", text: "Couldn't Update" });
				serverConfig![name] = originalValue;
				if (typeof done !== "undefined") done();
			});
	}

	interface ServerStats {
		users: number;
		privateUsers: number;
		watchedMovies: number;
		watchedShows: number;
		watchedSeasons: number;
		mostWatchedMovie: Content;
		mostWatchedShow: Content;
		activities: number;
	}

	async function getServerStats() {
		return await req.get<ServerStats>("/server/stats");
	}
</script>

<svelte:head>
	<title>Server Settings - Watcharr</title>
</svelte:head>

<div class="content">
	<div class="inner">
		<SettingsList>
			<h2>Server Settings</h2>

			<Stats>
				{#await getServerStats()}
					<Spinner />
				{:then stats}
					<Stat
						name="Users"
						value={stats.users}
						href={resolve("/manage_users")}
						large
					/>
					<Stat name="Private Users" value={stats.privateUsers} large />
					<Stat name="Watched Movies" value={stats.watchedMovies} large />
					<Stat name="Watched Shows" value={stats.watchedShows} large />
					<Stat name="Watched Seasons" value={stats.watchedSeasons} large />
					<Stat name="Activities" value={stats.activities} large />
					{#if stats.mostWatchedMovie?.title}
						<Stat
							name="Most Watched Movie"
							value={stats.mostWatchedMovie.title}
							href={resolve(
								`/movie/${stats.mostWatchedMovie.tmdbId}` as `/movie/${string}`,
							)}
						/>
					{/if}
					{#if stats.mostWatchedShow?.title}
						<Stat
							name="Most Watched Show"
							value={stats.mostWatchedShow.title}
							href={resolve(
								`/tv/${stats.mostWatchedShow.tmdbId}` as `/tv/${string}`,
							)}
						/>
					{/if}
				{:catch err}
					<Error error={err} pretty="Failed to get server stats!" />
				{/await}
			</Stats>

			{#await getServerConfig()}
				<Spinner />
			{:then}
				<h3>General</h3>
				{#if serverConfig}
					<Setting
						title="Default Country"
						desc="Default country for new users. This can be changed per user and won't affect existing users."
					>
						<RegionDropDown
							selectedCountry={serverConfig.DEFAULT_COUNTRY}
							disabled={countryDisabled}
							onChange={(c) => {
								countryDisabled = true;
								updateServerConfig("DEFAULT_COUNTRY", c, () => {
									countryDisabled = false;
								});
							}}
						/>
					</Setting>
					<Setting
						title="Default Language"
						desc="Default language for TMDB metadata and new users. This can be changed per user and won't affect existing users."
					>
						<LanguageDropDown
							selectedLanguage={serverConfig.DEFAULT_LANGUAGE}
							disabled={languageDisabled}
							onChange={(l) => {
								languageDisabled = true;
								updateServerConfig("DEFAULT_LANGUAGE", l, () => {
									languageDisabled = false;
								});
							}}
						/>
					</Setting>
					<Setting
						title="{jellyfinOrEmby} Host"
						desc="Point to your {jellyfinOrEmby} server to enable related features. Don't change server after
        already using another."
					>
						<input
							type="text"
							placeholder="https://{jellyfinOrEmby.toLowerCase()}.example.com"
							bind:value={serverConfig.JELLYFIN_HOST}
							onblur={() => {
								jfDisabled = true;
								updateServerConfig(
									"JELLYFIN_HOST",
									serverConfig!.JELLYFIN_HOST,
									() => {
										jfDisabled = false;
									},
								);
							}}
							disabled={jfDisabled}
						/>
					</Setting>
					<Setting
						title="{jellyfinOrEmby} Webhook"
						desc="Do you want to enable {jellyfinOrEmby} Webhook support?"
						row
						handbook="jellyfin-webhook-setup"
					>
						<Checkbox
							name="JELLYFIN_WEBHOOK_ENABLED"
							disabled={jfWebhookEnabledDisabled}
							value={serverConfig.JELLYFIN_WEBHOOK_ENABLED}
							toggled={(on) => {
								jfWebhookEnabledDisabled = true;
								updateServerConfig("JELLYFIN_WEBHOOK_ENABLED", on, (rData) => {
									jfWebhookEnabledDisabled = false;
									if (
										serverConfig &&
										rData &&
										"JELLYFIN_WEBHOOK_KEY" in rData &&
										typeof rData.JELLYFIN_WEBHOOK_KEY === "string"
									) {
										serverConfig.JELLYFIN_WEBHOOK_KEY =
											rData.JELLYFIN_WEBHOOK_KEY;
									}
								});
							}}
						/>
					</Setting>
					{#if serverConfig.JELLYFIN_WEBHOOK_ENABLED}
						<Setting
							title="{jellyfinOrEmby} Webhook Endpoint"
							desc="This is the address you should use when creating the webhook in {jellyfinOrEmby}. Supports native Emby Premiere webhooks as well as the standard Webhooks plugin. This address contains a secret key, don't share it with others!"
						>
							<code>
								{jellyfinWebhookEndpoint}
							</code>
						</Setting>
					{/if}
					<Setting
						title="Use Emby"
						desc="Enable native Emby mode with dedicated /api/emby/webhook endpoints and branding."
						row
					>
						<Checkbox
							name="USE_EMBY"
							disabled={useEmbyDisabled}
							value={serverConfig.USE_EMBY}
							toggled={(on) => {
								useEmbyDisabled = true;
								updateServerConfig("USE_EMBY", on, () => {
									useEmbyDisabled = false;
								});
							}}
						/>
					</Setting>
					<Setting
						title="Plex Host"
						desc="Point to your Plex server to enable related features. Don't change server after
        already using another."
					>
						<input
							type="text"
							placeholder="https://plex.example.com"
							bind:value={serverConfig.PLEX_HOST}
							onblur={() => {
								plexHostDisabled = true;
								updateServerConfig(
									"PLEX_HOST",
									serverConfig!.PLEX_HOST,
									(rData) => {
										plexHostDisabled = false;
										serverConfig!.PLEX_MACHINE_ID =
											rData &&
											"PLEX_MACHINE_ID" in rData &&
											rData?.PLEX_MACHINE_ID
												? String(rData?.PLEX_MACHINE_ID)
												: undefined;
									},
								);
							}}
							disabled={plexHostDisabled}
						/>
						{#if serverConfig.PLEX_MACHINE_ID}
							<span style="font-size: 10px"
								>Machine Id: {serverConfig.PLEX_MACHINE_ID}</span
							>
						{/if}
					</Setting>
					<Setting title="TMDB Key" desc="Provide your own TMDB API Key">
						<input
							type="password"
							placeholder="TMDB Key"
							bind:value={serverConfig.TMDB_KEY}
							onblur={() => {
								tmdbkDisabled = true;
								updateServerConfig("TMDB_KEY", serverConfig!.TMDB_KEY, () => {
									tmdbkDisabled = false;
								});
							}}
							disabled={tmdbkDisabled}
						/>
					</Setting>
					<Setting
						title="OMDb API Key"
						desc="Optional: Provide an OMDb API Key (omdbapi.com) to display IMDb, Rotten Tomatoes, Metacritic scores, Box Office and Oscars/Awards."
					>
						<input
							type="password"
							placeholder="OMDb API Key"
							bind:value={serverConfig.OMDB_KEY}
							onblur={() => {
								omdbkDisabled = true;
								updateServerConfig("OMDB_KEY", serverConfig!.OMDB_KEY, () => {
									omdbkDisabled = false;
								});
							}}
							disabled={omdbkDisabled}
						/>
					</Setting>
					<Setting
						title="MDBList API Key"
						desc="Optional: Provide an MDBList API Key (mdblist.com) to display Letterboxd, Trakt, Rotten Tomatoes (Critics & Popcornmeter), IMDb and Metacritic scores."
					>
						<input
							type="password"
							placeholder="MDBList API Key"
							bind:value={serverConfig.MDBLIST_KEY}
							onblur={() => {
								mdblistkDisabled = true;
								updateServerConfig(
									"MDBLIST_KEY",
									serverConfig!.MDBLIST_KEY,
									() => {
										mdblistkDisabled = false;
									},
								);
							}}
							disabled={mdblistkDisabled}
						/>
					</Setting>
					<Setting
						title="Signup"
						desc="Allow signing up with Watcharr credentials."
						row
					>
						<Checkbox
							name="SIGNUP_ENABLED"
							disabled={signupDisabled}
							value={serverConfig.SIGNUP_ENABLED}
							toggled={(on) => {
								signupDisabled = true;
								updateServerConfig("SIGNUP_ENABLED", on, () => {
									signupDisabled = false;
								});
							}}
						/>
					</Setting>
					<Setting
						title="Debug"
						desc="Enable debug logging. Don't keep this on all the time, only use this when trying to resolve an issue."
						row
					>
						<Checkbox
							name="DEBUG"
							disabled={debugDisabled}
							value={serverConfig.DEBUG}
							toggled={(on) => {
								debugDisabled = true;
								updateServerConfig("DEBUG", on, () => {
									debugDisabled = false;
								});
							}}
						/>
					</Setting>
					<Setting>
						<SettingButton
							title="Task Schedule"
							desc="View and configure server task schedule."
							icon="arrow"
							onClick={() => {
								taskScheduleModalOpen = true;
							}}
						/>
					</Setting>
					{#if taskScheduleModalOpen}
						<TaskScheduleModal onClose={() => (taskScheduleModalOpen = false)}
						></TaskScheduleModal>
					{/if}
					<Setting>
						<SettingButton
							title="Trusted Header Authentication"
							desc="Configure trusted header single sign-on."
							icon="arrow"
							onClick={() => {
								headerSSOModalOpen = true;
							}}
						/>
					</Setting>
					{#if headerSSOModalOpen}
						<TrustedHeaderAuthModal onClose={() => (headerSSOModalOpen = false)}
						></TrustedHeaderAuthModal>
					{/if}
					<div>
						<h3>Services</h3>
						<h5 class="norm">
							These integrations are not in their final stages. Consider them a
							preview/beta, if you have any issues,
							<a
								style="text-decoration: underline;"
								href="https://github.com/sbondCo/Watcharr/issues/new/choose"
								target="_blank"
							>
								please report them.
							</a>
						</h5>
					</div>

					<Setting title="Twitch">
						<SettingButton
							title="Twitch"
							desc="Twitch application credentials for enabling game support (via IGDB)."
							icon={serverConfig.TWITCH &&
							Object.keys(serverConfig.TWITCH).length > 0
								? "arrow"
								: "add"}
							onClick={() => {
								twitchModalOpen = true;
							}}
						/>
					</Setting>

					<Setting title="Sonarr">
						{#if serverConfig.SONARR && serverConfig.SONARR?.length > 0}
							{#each serverConfig.SONARR as server (server.name)}
								<SettingButton
									title={server.name}
									desc={`Configure server at ${server.host}`}
									onClick={() => {
										sonarrServerEditing = server;
										sonarrModalEditing = true;
										sonarrModalOpen = true;
									}}
								/>
							{/each}
						{/if}
						<SettingButton
							title="Sonarr"
							desc="Add a Sonarr server."
							icon="add"
							onClick={() => {
								let name = "Sonarr";
								if (serverConfig?.SONARR && serverConfig.SONARR.length > 0) {
									// if this still exists ya on yur own
									name = `Sonarr${serverConfig!.SONARR.length + 1}`;
								}
								sonarrServerEditing = { name };
								sonarrModalEditing = false;
								sonarrModalOpen = true;
							}}
						/>
					</Setting>

					<Setting title="Radarr">
						{#if serverConfig.RADARR && serverConfig.RADARR?.length > 0}
							{#each serverConfig.RADARR as server (server.name)}
								<SettingButton
									title={server.name}
									desc={`Configure server at ${server.host}`}
									onClick={() => {
										radarrServerEditing = server;
										radarrModalEditing = true;
										radarrModalOpen = true;
									}}
								/>
							{/each}
						{/if}
						<SettingButton
							title="Radarr"
							desc="Add a Radarr server."
							icon="add"
							onClick={() => {
								let name = "Radarr";
								if (serverConfig?.RADARR && serverConfig.RADARR.length > 0) {
									// if this still exists ya on yur own
									name = `Radarr${serverConfig!.RADARR.length + 1}`;
								}
								radarrServerEditing = { name };
								radarrModalEditing = false;
								radarrModalOpen = true;
							}}
						/>
					</Setting>

					<Setting
						title="Jellyseerr / Overseerr"
						desc="Connect to Jellyseerr or Overseerr to handle media requests and check real-time availability."
					>
						<div style="display: flex; flex-direction: column; gap: 8px; width: 100%;">
							<input
								type="text"
								placeholder="Host (e.g. http://jellyseerr:5055 or https://requests.example.com)"
								bind:value={serverConfig.OVERSEERR_HOST}
								onblur={() => {
									overseerrHostDisabled = true;
									updateServerConfig(
										"OVERSEERR_HOST",
										serverConfig!.OVERSEERR_HOST,
										() => {
											overseerrHostDisabled = false;
											getServerFeatures();
										},
									);
								}}
								disabled={overseerrHostDisabled}
							/>
							<input
								type="password"
								placeholder="API Key (from Settings > General in Overseerr)"
								bind:value={serverConfig.OVERSEERR_KEY}
								onblur={() => {
									overseerrKeyDisabled = true;
									updateServerConfig(
										"OVERSEERR_KEY",
										serverConfig!.OVERSEERR_KEY,
										() => {
											overseerrKeyDisabled = false;
											getServerFeatures();
										},
									);
								}}
								disabled={overseerrKeyDisabled}
							/>
							<div style="display: flex; gap: 10px; align-items: center; margin-top: 4px;">
								<button
									class="btn"
									style="padding: 6px 14px; font-size: 13px; font-weight: 600;"
									disabled={overseerrTesting || !serverConfig.OVERSEERR_HOST || !serverConfig.OVERSEERR_KEY}
									onclick={testOverseerrConnection}
								>
									{overseerrTesting ? "Testing..." : "Test Connection"}
								</button>
								{#if serverConfig.OVERSEERR_HOST && serverConfig.OVERSEERR_KEY}
									<span style="font-size: 12px; color: #10b981; font-weight: 600;">Configured</span>
								{/if}
							</div>
						</div>
					</Setting>

					{#if twitchModalOpen}
						<TwitchModal
							cfg={serverConfig.TWITCH}
							onClose={() => {
								// "temporary" solution to showing added servers
								// and reloading data to revert modified but not saved changes.
								getServerConfig();
								getServerFeatures();
								twitchModalOpen = false;
							}}
						/>
					{/if}

					{#if sonarrModalOpen && sonarrServerEditing}
						<SonarrModal
							servarr={sonarrServerEditing}
							isEditing={sonarrModalEditing}
							onClose={() => {
								// "temporary" solution to showing added servers
								// and reloading data to revert modified but not saved changes.
								getServerConfig();
								getServerFeatures();
								sonarrModalOpen = false;
								sonarrServerEditing = undefined;
							}}
						/>
					{/if}

					{#if radarrModalOpen && radarrServerEditing}
						<RadarrModal
							servarr={radarrServerEditing}
							isEditing={radarrModalEditing}
							onClose={() => {
								// "temporary" solution to showing added servers
								// and reloading data to revert modified but not saved changes.
								getServerConfig();
								getServerFeatures();
								radarrModalOpen = false;
								radarrServerEditing = undefined;
							}}
						/>
					{/if}
				{/if}
			{:catch err}
				<Error error={err} pretty="Failed to load server config!" />
			{/await}
		</SettingsList>
	</div>
</div>

<style lang="scss">
	.content {
		display: flex;
		width: 100%;
		justify-content: center;
		padding: 24px 20px 80px 20px;

		.inner {
			width: 100%;
			max-width: 860px;
			display: flex;
			flex-direction: column;
			gap: 24px;

			h2 {
				font-size: 26px;
				font-weight: 700;
				letter-spacing: -0.02em;
				color: $text-color;
				margin: 0;
			}

			h3 {
				font-size: 18px;
				font-weight: 700;
				color: $text-color;
				margin: 16px 0 6px 0;
			}

			& > div:not(:first-of-type) {
				margin-top: 10px;
			}
		}
	}
</style>
