<script lang="ts">
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import Checkbox from "@/lib/Checkbox.svelte";
	import Error from "@/lib/Error.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import Setting from "@/lib/settings/Setting.svelte";
	import { req, updateUserSetting, baseURL } from "@/lib/util/api";
	import {
		getOrdinalSuffix,
		monthsShort,
		toRelativeDate,
	} from "@/lib/util/helpers";
	import { store } from "@/store.svelte";
	import {
		UserType,
		type Image,
		type Profile,
		type RecentActivityItem,
	} from "@/types";
	import { notify } from "@/lib/util/notify";
	import UserAvatar from "@/lib/img/UserAvatar.svelte";
	import PwChangeModal from "@/routes/(app)/profile/modals/PwChangeModal.svelte";
	import SyncModal from "./modals/SyncModal.svelte";
	import RegionDropDown from "@/lib/RegionDropDown.svelte";
	import LanguageDropDown from "@/lib/LanguageDropDown.svelte";
	import RatingSetting from "@/lib/rating/RatingSetting.svelte";
	import { toggleTheme } from "@/lib/util/theme";
	import ExportListModal from "./modals/ExportListModal.svelte";
	import { ReqerError } from "@/lib/util/fetch";
	import Icon from "@/lib/Icon.svelte";
	import CalendarWebhookModal from "@/lib/calendar/CalendarWebhookModal.svelte";

	let user = $derived(store.userInfo);
	let settings = $derived(store.userSettings);
	let selectedTheme = $derived(store.appTheme);

	let privateDisabled = $state(false);
	let privateThoughtsDisabled = $state(false);
	let exportModalOpen = $state(false);
	let hideSpoilersDisabled = $state(false);
	let countryDisabled = $state(false);
	let languageDisabled = $state(false);
	let includePreviouslyWatchedDisabled = $state(false);
	let automateShowStatusesDisabled = $state(false);
	let pwChangeModalOpen = $state(false);
	let getProfilePromise = $state(getProfile());
	let jellyfinSyncModalOpen = $state(false);
	let plexSyncModalOpen = $state(false);
	let calendarWebhookModalOpen = $state(false);

	let statusTab: "all" | "movie" | "tv" | "game" = $state("all");
	let activityPeriodTab: "month" | "year" = $state("month");

	async function getProfile() {
		return await req.get<Profile>(`/profile`);
	}

	function formatDate(d: Date | string | undefined | null) {
		if (!d) return "Unknown";
		const date = d instanceof Date ? d : new Date(d);
		if (isNaN(date.getTime())) return "Unknown";
		return `${date.getDate()}${getOrdinalSuffix(date.getDate())} ${
			monthsShort[date.getMonth()]
		} ${date.getFullYear()}`;
	}

	function updateBio(
		ev: FocusEvent & { currentTarget: EventTarget & HTMLTextAreaElement },
	) {
		const newBio = ev?.currentTarget?.value;
		if (typeof newBio !== "string") {
			console.warn("updateBio called without any value", newBio);
			return;
		}
		const nid = notify({ text: "Updating Bio", type: "loading" });
		req
			.post("/user/bio", { newBio: newBio })
			.then(() => {
				if (user) {
					user.bio = newBio;
					notify({ id: nid, text: "Updated Bio", type: "success" });
				}
			})
			.catch((err) => {
				notify({
					id: nid,
					text: ReqerError.getMsg(err, "Failed to update bio"),
					type: "error",
				});
			});
	}

	function avatarDropped(ev: Event) {
		const files = (ev.currentTarget as HTMLInputElement)?.files;
		if (!files || files?.length <= 0) {
			console.error("avatarDropped: no file found");
			return;
		}
		const nid = notify({ text: "Uploading avatar", type: "loading" });
		const fd = new FormData();
		fd.append("avatar", files[0]);
		req
			.post<Image>("/user/avatar", fd)
			.then((r) => {
				if (user) {
					user.avatar = r;
					notify({ id: nid, text: "Avatar Uploaded", type: "success" });
				}
			})
			.catch((err) => {
				console.error("uploading avatar failed", err);
				notify({
					id: nid,
					text: ReqerError.getMsg(err, "Failed to upload avatar"),
					type: "error",
				});
			});
	}

	/**
	 * Takes in number of minutes and converts to readable string.
	 */
	function toFormattedTimeLong(m: number) {
		if (!m || m <= 0) return "0 hours";
		const countInMinutes = [
			["month", 43200],
			["week", 10080],
			["day", 1440],
			["hour", 60],
		];
		let ansString = "";
		let tmp;
		for (const c of countInMinutes) {
			tmp = Math.floor(m / (c[1] as number));
			if (tmp) ansString += `${tmp} ${c[0]}${tmp >= 2 ? "s, " : ", "}`;
			m -= tmp * (c[1] as number);
		}
		if (!ansString) {
			return `${m} mins`;
		}
		return ansString.slice(0, -2);
	}

	function getActivityHref(
		item: RecentActivityItem,
	): `${`/${"movie" | "tv" | "game"}/${string}`}` {
		return item.mediaType === "movie"
			? `/movie/${item.mediaId}`
			: item.mediaType === "tv"
				? `/tv/${item.mediaId}`
				: `/game/${item.mediaId}`;
	}

	function getCurrentDist(
		profile: Profile,
		tab: "all" | "movie" | "tv" | "game",
	): Record<string, number> {
		return tab === "all"
			? profile.statusDistribution || {}
			: profile.statusDistributionType?.[tab] || {};
	}

	function getTotalInTab(dist: Record<string, number>): number {
		return Object.values(dist).reduce((a, b) => a + b, 0);
	}

	function getActivityPoster(item: RecentActivityItem): string | undefined {
		if (!item.posterPath) return undefined;
		if (item.posterPath.startsWith("http")) return item.posterPath;
		if (item.mediaType === "game") return `${baseURL}/${item.posterPath}`;
		return `https://image.tmdb.org/t/p/w200${item.posterPath}`;
	}

	function formatActivityLabel(item: RecentActivityItem): string {
		switch (item.type) {
			case "STATUS_CHANGED": {
				let status = item.data;
				if (status?.startsWith("{")) {
					try {
						const parsed = JSON.parse(status);
						if (parsed.status) status = parsed.status;
					} catch {}
				}
				return `Status changed to ${status || "unknown"}`;
			}
			case "RATING_CHANGED":
				return `Rated ${item.data}/10`;
			case "IMPORTED_RATING": {
				let rating = item.data;
				if (rating?.startsWith("{")) {
					try {
						const parsed = JSON.parse(rating);
						if (parsed.rating) rating = parsed.rating;
					} catch {}
				}
				return `Rated ${rating}/10`;
			}
			case "ADDED_WATCHED":
				return "Added to library";
			case "IMPORTED_ADDED_WATCHED":
				return "Imported to library";
			case "EPISODE_STATUS_CHANGED":
				return `Watched episode`;
			default:
				return item.type?.replace(/_/g, " ").toLowerCase() || "activity";
		}
	}
</script>

<svelte:head>
	<title>My Profile - Watcharr</title>
</svelte:head>

<div class="content">
	<div class="inner">
		<!-- User Profile Card -->
		<section class="profile-hero-card">
			<div class="avatar-wrap">
				<UserAvatar img={user?.avatar} {avatarDropped} />
			</div>
			<div class="profile-meta">
				<div class="meta-header">
					<h2 class="username" title={user?.username}>
						{user?.username}
					</h2>
					{#if user?.type}
						<span class="user-badge">{user.type}</span>
					{/if}
				</div>
				<textarea
					rows="2"
					class="bio-input"
					placeholder="Add a short bio about yourself..."
					onblur={updateBio}
					value={user?.bio}></textarea>
			</div>
		</section>

		<!-- Statistics Section -->
		<section class="stats-section">
			<div class="stats-header-row">
				<h3 class="section-heading">
					<Icon i="reel" wh={20} />
					Cinematic Activity & Statistics
				</h3>
			</div>

			{#await getProfilePromise}
				<div class="stats-loading">
					<Spinner />
					<p>Calculating collection metrics...</p>
				</div>
			{:then profile}
				<!-- Hero Metrics Grid -->
				<div class="hero-stats-grid">
					<div class="stat-card primary">
						<div class="card-icon">
							<Icon i="film" wh={24} />
						</div>
						<div class="stat-body">
							<span class="stat-value">{profile.moviesWatched}</span>
							<span class="stat-label">Movies Watched</span>
							<span class="stat-sub"
								>{toFormattedTimeLong(profile.moviesWatchedRuntime)}</span
							>
						</div>
					</div>

					<div class="stat-card primary">
						<div class="card-icon">
							<Icon i="tv" wh={24} />
						</div>
						<div class="stat-body">
							<span class="stat-value">{profile.episodesWatched || 0}</span>
							<span class="stat-label">Episodes Watched</span>
							<span class="stat-sub"
								>{toFormattedTimeLong(profile.showsWatchedRuntime)}</span
							>
						</div>
					</div>

					<div class="stat-card">
						<div class="card-icon">
							<Icon i="check" wh={24} />
						</div>
						<div class="stat-body">
							<span class="stat-value">{profile.showsWatched}</span>
							<span class="stat-label">Series Completed</span>
							<span class="stat-sub"
								>{profile.showsWatching || 0} in progress</span
							>
						</div>
					</div>

					<div class="stat-card highlight">
						<div class="card-icon">
							<Icon i="clock" wh={24} />
						</div>
						<div class="stat-body">
							<span class="stat-value"
								>{toFormattedTimeLong(profile.totalRuntime)}</span
							>
							<span class="stat-label">Total Screen Time</span>
							<span class="stat-sub"
								>Member since {formatDate(new Date(profile.joined))}</span
							>
						</div>
					</div>
				</div>

				<!-- Partial History Transparency Notice -->
				{#if profile.hasPartialHistory}
					<div class="transparency-notice">
						<Icon i="compass" wh={18} />
						<span>
							Some older imported titles lack exact episode runtimes or custom
							dates. Screen time reflects only records with confirmed duration.
						</span>
					</div>
				{/if}

				<!-- Status Distribution & Rating Analytics Grid -->
				<div class="dashboard-split-grid">
					<!-- Watch Status Distribution -->
					<div class="dashboard-panel">
						<div class="panel-header">
							<div class="panel-title-wrap">
								<h4 class="panel-title">Collection Status</h4>
								<span class="panel-subtitle"
									>Watch states across your media</span
								>
							</div>

							<!-- Filter Tabs -->
							<div class="mini-pills">
								<button
									class="mini-pill"
									class:active={statusTab === "all"}
									onclick={() => (statusTab = "all")}
								>
									All
								</button>
								<button
									class="mini-pill"
									class:active={statusTab === "movie"}
									onclick={() => (statusTab = "movie")}
								>
									Movies
								</button>
								<button
									class="mini-pill"
									class:active={statusTab === "tv"}
									onclick={() => (statusTab = "tv")}
								>
									Shows
								</button>
								{#if profile.gamesPlayed > 0 || (profile.statusDistributionType?.game && Object.values(profile.statusDistributionType.game).some((v) => v > 0))}
									<button
										class="mini-pill"
										class:active={statusTab === "game"}
										onclick={() => (statusTab = "game")}
									>
										Games
									</button>
								{/if}
							</div>
						</div>

						{#if true}
							{@const currentDist = getCurrentDist(profile, statusTab)}
							{@const totalInTab = getTotalInTab(currentDist)}

							<!-- Segmented Bar -->
							<div class="status-bar-track">
								{#if totalInTab > 0}
									{#if (currentDist.FINISHED || 0) > 0}
										<div
											class="bar-segment finished"
											style={`width: ${((currentDist.FINISHED || 0) / totalInTab) * 100}%`}
											title={`Finished: ${currentDist.FINISHED}`}
										></div>
									{/if}
									{#if (currentDist.WATCHING || 0) > 0}
										<div
											class="bar-segment watching"
											style={`width: ${((currentDist.WATCHING || 0) / totalInTab) * 100}%`}
											title={`Watching: ${currentDist.WATCHING}`}
										></div>
									{/if}
									{#if (currentDist.PLANNED || 0) > 0}
										<div
											class="bar-segment planned"
											style={`width: ${((currentDist.PLANNED || 0) / totalInTab) * 100}%`}
											title={`Planned: ${currentDist.PLANNED}`}
										></div>
									{/if}
									{#if (currentDist.HOLD || 0) > 0}
										<div
											class="bar-segment hold"
											style={`width: ${((currentDist.HOLD || 0) / totalInTab) * 100}%`}
											title={`On Hold: ${currentDist.HOLD}`}
										></div>
									{/if}
									{#if (currentDist.DROPPED || 0) > 0}
										<div
											class="bar-segment dropped"
											style={`width: ${((currentDist.DROPPED || 0) / totalInTab) * 100}%`}
											title={`Dropped: ${currentDist.DROPPED}`}
										></div>
									{/if}
								{:else}
									<div class="empty-bar">No titles in this category</div>
								{/if}
							</div>

							<!-- Status Cards Legend -->
							<div class="status-legend-grid">
								<div class="legend-badge finished">
									<span class="dot"></span>
									<span class="label">Finished</span>
									<strong class="count">{currentDist.FINISHED || 0}</strong>
								</div>
								<div class="legend-badge watching">
									<span class="dot"></span>
									<span class="label">Watching</span>
									<strong class="count">{currentDist.WATCHING || 0}</strong>
								</div>
								<div class="legend-badge planned">
									<span class="dot"></span>
									<span class="label">Planned</span>
									<strong class="count">{currentDist.PLANNED || 0}</strong>
								</div>
								<div class="legend-badge hold">
									<span class="dot"></span>
									<span class="label">On Hold</span>
									<strong class="count">{currentDist.HOLD || 0}</strong>
								</div>
								<div class="legend-badge dropped">
									<span class="dot"></span>
									<span class="label">Dropped</span>
									<strong class="count">{currentDist.DROPPED || 0}</strong>
								</div>
							</div>
						{/if}
					</div>

					<!-- Ratings Analytics -->
					<div class="dashboard-panel">
						<div class="panel-header">
							<div class="panel-title-wrap">
								<h4 class="panel-title">Rating Evolution</h4>
								<span class="panel-subtitle">Personal score distribution</span>
							</div>
							{#if profile.ratingStats?.count > 0}
								<span class="rating-counter-badge">
									{profile.ratingStats.count} rated titles
								</span>
							{/if}
						</div>

						{#if profile.ratingStats?.count > 0}
							<div class="rating-content-split">
								<!-- Score Highlight Card -->
								<div class="rating-score-highlight">
									<div class="score-stars">
										<Icon i="star" wh={28} />
									</div>
									<span class="big-score"
										>{profile.ratingStats.average.toFixed(1)}</span
									>
									<span class="score-scale">Average out of 10</span>
									<div class="score-range-pills">
										<span class="range-pill"
											>Min: {profile.ratingStats.lowest}</span
										>
										<span class="range-pill"
											>Max: {profile.ratingStats.highest}</span
										>
									</div>
								</div>

								<!-- Rating Histogram -->
								<div class="rating-histogram">
									{#each [10, 9, 8, 7, 6, 5, 4, 3, 2, 1] as score (score)}
										{@const count =
											profile.ratingStats.distribution?.[String(score)] || 0}
										{@const pct =
											profile.ratingStats.count > 0
												? (count / profile.ratingStats.count) * 100
												: 0}
										<div class="histogram-row">
											<span class="score-num">{score}</span>
											<div class="histo-bar-wrap">
												<div
													class="histo-bar-fill"
													style={`width: ${pct}%`}
												></div>
											</div>
											<span class="histo-count">{count}</span>
										</div>
									{/each}
								</div>
							</div>
						{:else}
							<div class="empty-analytics-placeholder">
								<Icon i="star" wh={32} />
								<p>No ratings logged yet. Rate titles to see your analytics.</p>
							</div>
						{/if}
					</div>
				</div>

				<!-- Activity Timeline & Genre Breakdown Grid -->
				<div class="dashboard-split-grid">
					<!-- Watch Activity Trends -->
					<div class="dashboard-panel">
						<div class="panel-header">
							<div class="panel-title-wrap">
								<h4 class="panel-title">Watch Activity</h4>
								<span class="panel-subtitle">Playback volume over time</span>
							</div>
							<div class="mini-pills">
								<button
									class="mini-pill"
									class:active={activityPeriodTab === "month"}
									onclick={() => (activityPeriodTab = "month")}
								>
									Monthly
								</button>
								<button
									class="mini-pill"
									class:active={activityPeriodTab === "year"}
									onclick={() => (activityPeriodTab = "year")}
								>
									Yearly
								</button>
							</div>
						</div>

						{#if true}
							{@const periods =
								activityPeriodTab === "month"
									? profile.activityByMonth || []
									: profile.activityByYear || []}
							{@const maxPeriodCount = Math.max(
								...periods.map((p) => p.count),
								1,
							)}

							{#if periods.length > 0}
								<div class="activity-bars-container small-scrollbar">
									<div class="activity-bars-flex">
										{#each periods as p (p.period)}
											{@const heightPct = Math.max(
												(p.count / maxPeriodCount) * 100,
												8,
											)}
											<div
												class="activity-col"
												title={`${p.period}: ${p.count}`}
											>
												<div class="col-track">
													<div
														class="col-fill"
														style={`height: ${heightPct}%`}
													></div>
												</div>
												<span class="col-label">
													{activityPeriodTab === "month"
														? p.period.slice(2)
														: p.period}
												</span>
												<span class="col-val">{p.count}</span>
											</div>
										{/each}
									</div>
								</div>
							{:else}
								<div class="empty-analytics-placeholder">
									<Icon i="calendar" wh={32} />
									<p>
										No dated activity recorded yet. Watch events will chart
										here.
									</p>
								</div>
							{/if}
						{/if}
					</div>

					<!-- Genre Distribution -->
					<div class="dashboard-panel">
						<div class="panel-header">
							<div class="panel-title-wrap">
								<h4 class="panel-title">Top Genres</h4>
								<span class="panel-subtitle"
									>Most watched themes & categories</span
								>
							</div>
						</div>

						{#if profile.genreDistribution && profile.genreDistribution.length > 0}
							{@const maxGenreCount = Math.max(
								...profile.genreDistribution.map((g) => g.count),
								1,
							)}
							<div class="genres-ranked-list small-scrollbar">
								{#each profile.genreDistribution as g (g.name)}
									{@const pct = (g.count / maxGenreCount) * 100}
									<div class="genre-item">
										<div class="genre-info">
											<span class="genre-name">{g.name}</span>
											<span class="genre-count">{g.count} titles</span>
										</div>
										<div class="genre-meter-track">
											<div
												class="genre-meter-fill"
												style={`width: ${pct}%`}
											></div>
										</div>
									</div>
								{/each}
							</div>
						{:else}
							<div class="empty-analytics-placeholder">
								<Icon i="tag" wh={32} />
								<p>
									Genre details will populate automatically as your collection
									updates.
								</p>
							</div>
						{/if}
					</div>
				</div>

				<!-- Recent History Stream -->
				{#if profile.recentActivity && profile.recentActivity.length > 0}
					<div class="dashboard-panel full-width">
						<div class="panel-header">
							<div class="panel-title-wrap">
								<h4 class="panel-title">Recent Activity Stream</h4>
								<span class="panel-subtitle"
									>Latest actions in your library</span
								>
							</div>
						</div>

						<div class="recent-stream-grid">
							{#each profile.recentActivity as item (item.id)}
								{@const poster = getActivityPoster(item)}
								<a href={resolve(getActivityHref(item))} class="stream-item">
									<div class="stream-thumb">
										{#if poster}
											<img src={poster} alt={item.mediaTitle} loading="lazy" />
										{:else}
											<div class="thumb-fallback">
												<Icon
													i={item.mediaType === "movie" ? "film" : "tv"}
													wh={18}
												/>
											</div>
										{/if}
									</div>
									<div class="stream-content">
										<strong class="stream-title">{item.mediaTitle}</strong>
										<span class="stream-action">
											{formatActivityLabel(item)}
										</span>
										<time class="stream-time"
											>{toRelativeDate(new Date(item.date))}</time
										>
									</div>
								</a>
							{/each}
						</div>
					</div>
				{/if}
			{:catch err}
				<Error error={err} pretty="Failed to load lifetime stats!" />
			{/await}
		</section>

		<!-- Settings & Preferences Section -->
		<section class="settings-container">
			<h3 class="section-heading">
				<Icon i="compass" wh={18} />
				Preferences & Settings
			</h3>

			<!-- Theme Selector -->
			<div class="settings-card">
				<h4 class="card-title">Appearance</h4>
				<p class="card-desc">Choose how Watcharr looks on this device.</p>
				<div class="theme-grid">
					<button
						class={`theme-option ${selectedTheme === "system" ? "active" : ""}`}
						onclick={() => toggleTheme("system")}
						type="button"
					>
						<div class="preview-box system-preview">
							<div class="half light-side"></div>
							<div class="half dark-side"></div>
						</div>
						<span class="theme-label">System</span>
					</button>

					<button
						class={`theme-option ${selectedTheme === "light" ? "active" : ""}`}
						onclick={() => toggleTheme("light")}
						type="button"
					>
						<div class="preview-box light-preview"></div>
						<span class="theme-label">Light</span>
					</button>

					<button
						class={`theme-option ${selectedTheme === "dark" ? "active" : ""}`}
						onclick={() => toggleTheme("dark")}
						type="button"
					>
						<div class="preview-box dark-preview"></div>
						<span class="theme-label">Dark</span>
					</button>
				</div>
			</div>

			<!-- Regional & Streaming Providers -->
			<div class="settings-card">
				<h4 class="card-title">Regional & Localization</h4>
				<Setting
					title="Country & Region"
					desc="Select your region to display local streaming availability and watch providers."
				>
					<RegionDropDown
						selectedCountry={settings?.country}
						disabled={countryDisabled}
						onChange={(c) => {
							countryDisabled = true;
							updateUserSetting("country", c, () => {
								countryDisabled = false;
							});
						}}
					/>
				</Setting>
				<Setting
					title="Metadata Language"
					desc="Preferred language for movie and show titles, overviews, and season details from TMDB."
				>
					<LanguageDropDown
						selectedLanguage={settings?.language}
						disabled={languageDisabled}
						onChange={(l) => {
							languageDisabled = true;
							updateUserSetting("language", l, () => {
								languageDisabled = false;
							});
						}}
					/>
				</Setting>
			</div>

			<!-- Privacy Settings -->
			<div class="settings-card">
				<h4 class="card-title">Privacy & Social</h4>
				<Setting
					title="Private Account"
					desc="Hide your profile and watched library from community searches."
					row
				>
					<Checkbox
						name="private"
						disabled={privateDisabled}
						value={settings?.private}
						toggled={(on) => {
							privateDisabled = true;
							updateUserSetting("private", on, () => {
								privateDisabled = false;
							});
						}}
					/>
				</Setting>

				{#if !settings?.private}
					<Setting
						title="Private Thoughts"
						desc="Hide your personal notes and written thoughts from followers."
						row
					>
						<Checkbox
							name="privateThoughts"
							disabled={privateThoughtsDisabled}
							value={settings?.privateThoughts}
							toggled={(on) => {
								privateThoughtsDisabled = true;
								updateUserSetting("privateThoughts", on, () => {
									privateThoughtsDisabled = false;
								});
							}}
						/>
					</Setting>
				{/if}
			</div>

			<!-- Library & Episode Behavior -->
			<div class="settings-card">
				<h4 class="card-title">Playback & Library Behavior</h4>
				<Setting
					title="Hide Spoilers"
					desc="Blur episode titles, descriptions, and thumbnails until marked as watched."
					row
				>
					<Checkbox
						name="hideSpoilers"
						disabled={hideSpoilersDisabled}
						value={settings?.hideSpoilers}
						toggled={(on) => {
							hideSpoilersDisabled = true;
							updateUserSetting("hideSpoilers", on, () => {
								hideSpoilersDisabled = false;
							});
						}}
					/>
				</Setting>

				<Setting
					title="Automate Show Statuses"
					desc="Automatically update show and season watch status when episodes are tracked."
					tag="experimental"
					row
				>
					<Checkbox
						name="automateShowStatusesDisabled"
						disabled={automateShowStatusesDisabled}
						value={settings?.automateShowStatuses}
						toggled={(on) => {
							automateShowStatusesDisabled = true;
							updateUserSetting("automateShowStatuses", on, () => {
								automateShowStatusesDisabled = false;
							});
						}}
					/>
				</Setting>

				<Setting
					title="Include Previously Watched"
					desc="Include previously finished items in the 'Finished' library filter."
					row
				>
					<Checkbox
						name="includePreviouslyWatched"
						disabled={includePreviouslyWatchedDisabled}
						value={settings?.includePreviouslyWatched}
						toggled={(on) => {
							includePreviouslyWatchedDisabled = true;
							updateUserSetting("includePreviouslyWatched", on, () => {
								includePreviouslyWatchedDisabled = false;
							});
						}}
					/>
				</Setting>
			</div>

			<!-- Rating System Setting -->
			<div class="settings-card">
				<h4 class="card-title">Rating Scale</h4>
				<RatingSetting />
			</div>

			<!-- Calendar Webhook Notifications Setting -->
			<div class="settings-card">
				<div class="webhook-settings-header">
					<div>
						<h4 class="card-title">Calendar & Release Webhooks</h4>
						<p class="card-desc">
							Recebe alertas matinais automáticos no Discord, Slack ou Ntfy quando estrearem episódios ou filmes da tua biblioteca.
						</p>
					</div>
					<button
						class="btn-action primary-accent"
						onclick={() => (calendarWebhookModalOpen = true)}
					>
						<Icon i="notifications" wh={16} /> Configurar Webhook
					</button>
				</div>
			</div>

			<!-- Data & Integrations Actions -->
			<div class="settings-card">
				<h4 class="card-title">Data & Integrations</h4>
				<p class="card-desc">
					Manage library imports, exports, and server synchronization.
				</p>
				<div class="actions-toolbar">
					<button class="btn-action" onclick={() => goto(resolve("/import"))}>
						<Icon i="document" wh={16} /> Import Data
					</button>
					<button class="btn-action" onclick={() => (exportModalOpen = true)}>
						Export List
					</button>
					{#if user?.type === UserType.Watcharr}
						<button
							class="btn-action"
							onclick={() => {
								pwChangeModalOpen = true;
							}}
						>
							<Icon i="lock-closed" wh={16} /> Change Password
						</button>
					{/if}
					{#if user?.type === UserType?.Jellyfin}
						<button
							class="btn-action"
							onclick={() => (jellyfinSyncModalOpen = true)}
						>
							Sync With {localStorage.getItem("useEmby") ? "Emby" : "Jellyfin"}
						</button>
					{/if}
					{#if user?.type === UserType?.Plex}
						<button
							class="btn-action"
							onclick={() => (plexSyncModalOpen = true)}
						>
							Sync With Plex
						</button>
					{/if}
				</div>
			</div>

			{#if exportModalOpen}
				<ExportListModal
					onClose={() => {
						exportModalOpen = false;
					}}
				/>
			{/if}
			{#if pwChangeModalOpen}
				<PwChangeModal
					userName={user?.username}
					onClose={() => {
						pwChangeModalOpen = false;
					}}
				></PwChangeModal>
			{/if}
			{#if jellyfinSyncModalOpen}
				<SyncModal onClose={() => (jellyfinSyncModalOpen = false)} />
			{/if}
			{#if plexSyncModalOpen}
				<SyncModal type="plex" onClose={() => (plexSyncModalOpen = false)} />
			{/if}
			<CalendarWebhookModal bind:open={calendarWebhookModalOpen} />
		</section>
	</div>
</div>

<style lang="scss">
	.content {
		display: flex;
		width: 100%;
		justify-content: center;
		padding: 28px 20px 80px 20px;

		.inner {
			width: 100%;
			max-width: 960px;
			display: flex;
			flex-direction: column;
			gap: 32px;
		}
	}

	.section-heading {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 18px;
		font-weight: 700;
		color: $text-color;
		margin: 0;
	}

	.profile-hero-card {
		display: flex;
		align-items: center;
		gap: 24px;
		padding: 28px;
		background: rgba(22, 27, 38, 0.75);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 20px;
		backdrop-filter: blur(12px);

		.avatar-wrap {
			flex-shrink: 0;
		}

		.profile-meta {
			display: flex;
			flex-direction: column;
			gap: 10px;
			width: 100%;

			.meta-header {
				display: flex;
				align-items: center;
				gap: 12px;

				.username {
					font-size: 26px;
					font-weight: 700;
					letter-spacing: -0.02em;
					color: $text-color;
					margin: 0;
				}

				.user-badge {
					font-size: 11px;
					font-weight: 700;
					text-transform: uppercase;
					letter-spacing: 0.06em;
					padding: 3px 10px;
					border-radius: 20px;
					background: rgba(99, 102, 241, 0.15);
					color: #818cf8;
					border: 1px solid rgba(99, 102, 241, 0.3);
				}
			}

			.bio-input {
				width: 100%;
				padding: 10px 14px;
				font-size: 14px;
				background: rgba(0, 0, 0, 0.25);
				border: 1px solid rgba(255, 255, 255, 0.08);
				border-radius: 10px;
				color: $text-color;
				resize: vertical;
				min-height: 48px;
				transition: border-color 150ms ease;

				&:focus {
					outline: none;
					border-color: $accent-color;
				}
			}
		}

		@media screen and (max-width: 600px) {
			flex-direction: column;
			align-items: center;
			text-align: center;

			.meta-header {
				justify-content: center;
			}
		}
	}

	/* Statistics Section */
	.stats-section {
		display: flex;
		flex-direction: column;
		gap: 20px;
	}

	.stats-loading {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 48px;
		gap: 12px;
		color: $text-muted;
	}

	.hero-stats-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 16px;

		@media screen and (max-width: 800px) {
			grid-template-columns: repeat(2, 1fr);
		}

		@media screen and (max-width: 480px) {
			grid-template-columns: 1fr;
		}

		.stat-card {
			display: flex;
			flex-direction: column;
			gap: 12px;
			padding: 20px;
			border-radius: 16px;
			background: rgba(22, 27, 38, 0.75);
			border: 1px solid rgba(255, 255, 255, 0.08);
			backdrop-filter: blur(10px);
			transition:
				transform 150ms ease,
				border-color 150ms ease;

			&:hover {
				transform: translateY(-2px);
				border-color: rgba(255, 255, 255, 0.16);
			}

			.card-icon {
				color: rgba(255, 255, 255, 0.6);
				display: flex;
				align-items: center;
			}

			&.primary .card-icon {
				color: #818cf8;
			}

			&.highlight .card-icon {
				color: #f59e0b;
			}

			.stat-body {
				display: flex;
				flex-direction: column;
				gap: 4px;

				.stat-value {
					font-size: 24px;
					font-weight: 700;
					color: #ffffff;
					letter-spacing: -0.02em;
				}

				.stat-label {
					font-size: 13px;
					font-weight: 600;
					color: $text-muted;
				}

				.stat-sub {
					font-size: 11.5px;
					color: rgba(255, 255, 255, 0.45);
				}
			}
		}
	}

	.transparency-notice {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px 16px;
		border-radius: 12px;
		background: rgba(59, 130, 246, 0.08);
		border: 1px solid rgba(59, 130, 246, 0.2);
		color: #93c5fd;
		font-size: 13px;
		line-height: 1.4;
	}

	.dashboard-split-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 20px;

		@media screen and (max-width: 768px) {
			grid-template-columns: 1fr;
		}
	}

	.dashboard-panel {
		display: flex;
		flex-direction: column;
		gap: 16px;
		padding: 24px;
		border-radius: 18px;
		background: rgba(22, 27, 38, 0.75);
		border: 1px solid rgba(255, 255, 255, 0.08);
		backdrop-filter: blur(10px);

		&.full-width {
			grid-column: 1 / -1;
		}

		.panel-header {
			display: flex;
			align-items: flex-start;
			justify-content: space-between;
			flex-wrap: wrap;
			gap: 10px;

			.panel-title-wrap {
				display: flex;
				flex-direction: column;
				gap: 2px;

				.panel-title {
					margin: 0;
					font-size: 16px;
					font-weight: 700;
					color: #ffffff;
				}

				.panel-subtitle {
					font-size: 12.5px;
					color: $text-muted;
				}
			}

			.rating-counter-badge {
				font-size: 12px;
				font-weight: 600;
				padding: 4px 10px;
				border-radius: 20px;
				background: rgba(255, 255, 255, 0.08);
				color: rgba(255, 255, 255, 0.8);
			}
		}
	}

	/* Mini pills for panel tabs */
	.mini-pills {
		display: inline-flex;
		padding: 2px;
		background: rgba(0, 0, 0, 0.35);
		border: 1px solid rgba(255, 255, 255, 0.06);
		border-radius: 8px;
		gap: 2px;

		.mini-pill {
			padding: 4px 10px;
			font-size: 11.5px;
			font-weight: 600;
			border-radius: 6px;
			background: transparent;
			border: none;
			color: rgba(255, 255, 255, 0.6);
			cursor: pointer;
			transition: all 120ms ease;

			&:hover:not(.active) {
				color: #fff;
			}

			&.active {
				background: #6366f1;
				color: #ffffff;
			}
		}
	}

	/* Status Distribution Bar */
	.status-bar-track {
		display: flex;
		height: 12px;
		width: 100%;
		border-radius: 6px;
		overflow: hidden;
		background: rgba(255, 255, 255, 0.05);

		.bar-segment {
			height: 100%;
			transition: width 300ms ease;

			&.finished {
				background: #22c55e;
			}
			&.watching {
				background: #6366f1;
			}
			&.planned {
				background: #f59e0b;
			}
			&.hold {
				background: #a855f7;
			}
			&.dropped {
				background: #f43f5e;
			}
		}

		.empty-bar {
			width: 100%;
			text-align: center;
			font-size: 10px;
			color: rgba(255, 255, 255, 0.4);
			line-height: 12px;
		}
	}

	.status-legend-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 8px;

		@media screen and (max-width: 500px) {
			grid-template-columns: repeat(2, 1fr);
		}

		.legend-badge {
			display: flex;
			align-items: center;
			gap: 8px;
			padding: 8px 12px;
			border-radius: 8px;
			background: rgba(255, 255, 255, 0.03);
			border: 1px solid rgba(255, 255, 255, 0.05);

			.dot {
				width: 8px;
				height: 8px;
				border-radius: 50%;
			}

			.label {
				font-size: 12px;
				color: rgba(255, 255, 255, 0.7);
			}

			.count {
				margin-left: auto;
				font-size: 13px;
				font-weight: 700;
				color: #ffffff;
			}

			&.finished .dot {
				background: #22c55e;
			}
			&.watching .dot {
				background: #6366f1;
			}
			&.planned .dot {
				background: #f59e0b;
			}
			&.hold .dot {
				background: #a855f7;
			}
			&.dropped .dot {
				background: #f43f5e;
			}
		}
	}

	/* Ratings split layout */
	.rating-content-split {
		display: grid;
		grid-template-columns: 140px 1fr;
		gap: 20px;
		align-items: center;

		@media screen and (max-width: 550px) {
			grid-template-columns: 1fr;
		}

		.rating-score-highlight {
			display: flex;
			flex-direction: column;
			align-items: center;
			text-align: center;
			padding: 16px;
			background: rgba(245, 158, 11, 0.08);
			border: 1px solid rgba(245, 158, 11, 0.2);
			border-radius: 14px;
			gap: 4px;

			.score-stars {
				color: #f59e0b;
			}

			.big-score {
				font-size: 36px;
				font-weight: 800;
				letter-spacing: -0.02em;
				color: #ffffff;
			}

			.score-scale {
				font-size: 11px;
				font-weight: 600;
				color: rgba(255, 255, 255, 0.6);
			}

			.score-range-pills {
				display: flex;
				gap: 6px;
				margin-top: 6px;

				.range-pill {
					font-size: 10px;
					font-weight: 700;
					padding: 2px 6px;
					border-radius: 4px;
					background: rgba(255, 255, 255, 0.08);
					color: rgba(255, 255, 255, 0.7);
				}
			}
		}

		.rating-histogram {
			display: flex;
			flex-direction: column;
			gap: 4px;

			.histogram-row {
				display: flex;
				align-items: center;
				gap: 8px;
				font-size: 11px;

				.score-num {
					width: 16px;
					text-align: right;
					color: rgba(255, 255, 255, 0.6);
					font-weight: 600;
				}

				.histo-bar-wrap {
					flex: 1;
					height: 6px;
					border-radius: 3px;
					background: rgba(255, 255, 255, 0.06);
					overflow: hidden;

					.histo-bar-fill {
						height: 100%;
						background: #f59e0b;
						border-radius: 3px;
						transition: width 300ms ease;
					}
				}

				.histo-count {
					width: 24px;
					color: rgba(255, 255, 255, 0.5);
					text-align: left;
				}
			}
		}
	}

	/* Activity Bars */
	.activity-bars-container {
		overflow-x: auto;
		padding-bottom: 8px;

		.activity-bars-flex {
			display: flex;
			align-items: flex-end;
			gap: 8px;
			min-height: 140px;
			padding-top: 16px;

			.activity-col {
				display: flex;
				flex-direction: column;
				align-items: center;
				gap: 6px;
				min-width: 32px;
				flex: 1;

				.col-track {
					width: 16px;
					height: 100px;
					display: flex;
					align-items: flex-end;
					background: rgba(255, 255, 255, 0.04);
					border-radius: 4px;
					overflow: hidden;

					.col-fill {
						width: 100%;
						background: linear-gradient(180deg, #818cf8 0%, #6366f1 100%);
						border-radius: 4px 4px 0 0;
						transition: height 300ms ease;
					}
				}

				.col-label {
					font-size: 10px;
					font-weight: 600;
					color: rgba(255, 255, 255, 0.5);
				}

				.col-val {
					font-size: 10px;
					color: #ffffff;
					font-weight: 700;
				}
			}
		}
	}

	/* Genres list */
	.genres-ranked-list {
		display: flex;
		flex-direction: column;
		gap: 10px;
		max-height: 180px;
		overflow-y: auto;
		padding-right: 4px;

		.genre-item {
			display: flex;
			flex-direction: column;
			gap: 4px;

			.genre-info {
				display: flex;
				justify-content: space-between;
				font-size: 12px;

				.genre-name {
					font-weight: 600;
					color: #ffffff;
				}

				.genre-count {
					color: rgba(255, 255, 255, 0.5);
				}
			}

			.genre-meter-track {
				height: 5px;
				border-radius: 3px;
				background: rgba(255, 255, 255, 0.06);
				overflow: hidden;

				.genre-meter-fill {
					height: 100%;
					border-radius: 3px;
					background: linear-gradient(90deg, #6366f1, #a855f7);
				}
			}
		}
	}

	/* Recent Activity Stream */
	.recent-stream-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 12px;

		@media screen and (max-width: 800px) {
			grid-template-columns: repeat(2, 1fr);
		}

		@media screen and (max-width: 500px) {
			grid-template-columns: 1fr;
		}

		.stream-item {
			display: flex;
			align-items: center;
			gap: 12px;
			padding: 10px 12px;
			border-radius: 12px;
			background: rgba(255, 255, 255, 0.03);
			border: 1px solid rgba(255, 255, 255, 0.06);
			text-decoration: none;
			color: inherit;
			transition: all 150ms ease;

			&:hover {
				background: rgba(255, 255, 255, 0.06);
				border-color: rgba(255, 255, 255, 0.14);
				transform: translateY(-1px);
			}

			.stream-thumb {
				width: 36px;
				height: 54px;
				border-radius: 6px;
				overflow: hidden;
				flex-shrink: 0;
				background: rgba(0, 0, 0, 0.3);

				img {
					width: 100%;
					height: 100%;
					object-fit: cover;
				}

				.thumb-fallback {
					width: 100%;
					height: 100%;
					display: flex;
					align-items: center;
					justify-content: center;
					color: rgba(255, 255, 255, 0.3);
				}
			}

			.stream-content {
				display: flex;
				flex-direction: column;
				gap: 2px;
				overflow: hidden;

				.stream-title {
					font-size: 13px;
					font-weight: 600;
					color: #ffffff;
					white-space: nowrap;
					overflow: hidden;
					text-overflow: ellipsis;
				}

				.stream-action {
					font-size: 11px;
					color: #a5b4fc;
					font-weight: 500;
				}

				.stream-time {
					font-size: 10.5px;
					color: rgba(255, 255, 255, 0.4);
				}
			}
		}
	}

	.empty-analytics-placeholder {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 36px 16px;
		gap: 8px;
		text-align: center;
		color: rgba(255, 255, 255, 0.4);

		p {
			margin: 0;
			font-size: 13px;
			max-width: 300px;
		}
	}

	/* Settings Section */
	.settings-container {
		display: flex;
		flex-direction: column;
		gap: 20px;

		.settings-card {
			display: flex;
			flex-direction: column;
			gap: 16px;
			padding: 24px;
			border-radius: 16px;
			background: rgba(22, 27, 38, 0.75);
			border: 1px solid rgba(255, 255, 255, 0.08);

			.card-title {
				font-size: 16px;
				font-weight: 600;
				color: $text-color;
				margin: 0;
			}

			.card-desc {
				font-size: 13.5px;
				color: $text-muted;
				margin: -8px 0 4px 0;
			}
		}
	}

	.theme-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 16px;

		.theme-option {
			display: flex;
			flex-direction: column;
			align-items: center;
			gap: 10px;
			padding: 14px;
			background: rgba(255, 255, 255, 0.03);
			border: 2px solid rgba(255, 255, 255, 0.08);
			border-radius: 12px;
			cursor: pointer;
			transition: all 180ms ease;

			.preview-box {
				width: 100%;
				height: 52px;
				border-radius: 8px;
				overflow: hidden;
				border: 1px solid rgba(255, 255, 255, 0.1);

				&.light-preview {
					background: #f8fafc;
				}

				&.dark-preview {
					background: #0f172a;
				}

				&.system-preview {
					display: flex;

					.half {
						width: 50%;
						height: 100%;

						&.light-side {
							background: #f8fafc;
						}

						&.dark-side {
							background: #0f172a;
						}
					}
				}
			}

			.theme-label {
				font-size: 13.5px;
				font-weight: 500;
				color: $text-color;
			}

			&:hover {
				border-color: rgba(255, 255, 255, 0.2);
				transform: translateY(-2px);
			}

			&.active {
				border-color: $brand-color;
				box-shadow: 0 0 16px rgba(99, 102, 241, 0.25);

				.theme-label {
					font-weight: 700;
					color: $brand-color;
				}
			}
		}

		@media screen and (max-width: 500px) {
			grid-template-columns: 1fr;
		}
	}

	.webhook-settings-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 20px;
		flex-wrap: wrap;

		.card-desc {
			margin: 0;
		}

		.btn-action.primary-accent {
			background: rgba(54, 186, 152, 0.12);
			border-color: rgba(54, 186, 152, 0.35);
			color: #36ba98;
			fill: #36ba98;
			font-weight: 600;

			&:hover {
				background: rgba(54, 186, 152, 0.22);
				border-color: rgba(54, 186, 152, 0.55);
				transform: translateY(-1px);
			}
		}
	}

	.actions-toolbar {
		display: flex;
		flex-flow: row wrap;
		gap: 10px;

		.btn-action {
			display: inline-flex;
			align-items: center;
			gap: 8px;
			padding: 9px 18px;
			border-radius: 10px;
			font-size: 13.5px;
			font-weight: 500;
			background: rgba(255, 255, 255, 0.05);
			border: 1px solid rgba(255, 255, 255, 0.1);
			color: $text-color;
			cursor: pointer;
			transition: all 180ms ease;

			&:hover {
				background: rgba(255, 255, 255, 0.1);
				border-color: rgba(255, 255, 255, 0.2);
				transform: translateY(-1px);
			}
		}
	}
</style>
