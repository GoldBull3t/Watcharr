<script lang="ts">
	import { resolve } from "$app/paths";
	import { SvelteDate } from "svelte/reactivity";
	import Icon from "@/lib/Icon.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import Error from "@/lib/Error.svelte";
	import Modal from "@/lib/Modal.svelte";
	import { req } from "@/lib/util/api";
	import { store } from "@/store.svelte";
	import type { CalendarEvent, CalendarFilter, CalendarType } from "@/types";
	import { toRelativeDate } from "@/lib/util/helpers";

	type CalendarViewMode = "month" | "week" | "agenda";

	let viewMode: CalendarViewMode = $state("month");
	let scopeFilter: CalendarFilter = $state("library");
	let typeFilter: CalendarType = $state("all");

	// Current reference date (defaults to today)
	const currentDate = new SvelteDate();

	let events: CalendarEvent[] = $state([]);
	let loading: boolean = $state(true);
	let error: unknown | undefined = $state();

	// State for day releases modal when day has multiple releases
	let activeDayModalDate: string | undefined = $state();
	let activeDayModalEvents: CalendarEvent[] = $state([]);

	function openDayModal(dateStr: string, eventsList: CalendarEvent[]) {
		activeDayModalDate = dateStr;
		activeDayModalEvents = eventsList;
	}

	function closeDayModal() {
		activeDayModalDate = undefined;
		activeDayModalEvents = [];
	}

	// Helper to format local date YYYY-MM-DD
	function formatLocalDate(d: Date | SvelteDate): string {
		const year = d.getFullYear();
		const month = String(d.getMonth() + 1).padStart(2, "0");
		const day = String(d.getDate()).padStart(2, "0");
		return `${year}-${month}-${day}`;
	}

	// Compute date range for current view and date
	const dateRange = $derived.by(() => {
		const y = currentDate.getFullYear();
		const m = currentDate.getMonth();

		if (viewMode === "month") {
			// Month start and end
			const firstDayOfMonth = new SvelteDate(y, m, 1);
			const lastDayOfMonth = new SvelteDate(y, m + 1, 0);

			// Pad with days from previous month to start on Monday
			const dayOfWeek = (firstDayOfMonth.getDay() + 6) % 7;
			const startDate = new SvelteDate(firstDayOfMonth);
			startDate.setDate(startDate.getDate() - dayOfWeek);

			// Pad with days from next month to complete weeks
			const endDate = new SvelteDate(lastDayOfMonth);
			const trailingDays =
				(7 - ((lastDayOfMonth.getDay() + 6) % 7) - 1 + 7) % 7;
			endDate.setDate(endDate.getDate() + trailingDays);

			return {
				start: formatLocalDate(startDate),
				end: formatLocalDate(endDate),
				title: firstDayOfMonth.toLocaleString(undefined, {
					month: "long",
					year: "numeric",
				}),
			};
		} else if (viewMode === "week") {
			// Monday to Sunday of the current week
			const dayOfWeek = (currentDate.getDay() + 6) % 7;
			const startDate = new SvelteDate(currentDate);
			startDate.setDate(startDate.getDate() - dayOfWeek);

			const endDate = new SvelteDate(startDate);
			endDate.setDate(endDate.getDate() + 6);

			const titleStart = startDate.toLocaleString(undefined, {
				month: "short",
				day: "numeric",
			});
			const titleEnd = endDate.toLocaleString(undefined, {
				month: "short",
				day: "numeric",
				year: "numeric",
			});

			return {
				start: formatLocalDate(startDate),
				end: formatLocalDate(endDate),
				title: `${titleStart} – ${titleEnd}`,
			};
		} else {
			// Agenda: 30 days ahead from currentDate
			const startDate = new SvelteDate(currentDate);
			const endDate = new SvelteDate(currentDate);
			endDate.setDate(endDate.getDate() + 30);

			return {
				start: formatLocalDate(startDate),
				end: formatLocalDate(endDate),
				title: `${startDate.toLocaleString(undefined, { month: "short", day: "numeric" })} + 30 days`,
			};
		}
	});

	// Fetch calendar events whenever parameters change
	$effect(() => {
		const { start, end } = dateRange;
		const region = store.userSettings?.country || "";

		loading = true;
		error = undefined;

		req
			.get<CalendarEvent[]>("/calendar", {
				params: {
					start,
					end,
					filter: scopeFilter,
					type: typeFilter,
					region,
				},
			})
			.then((res) => {
				events = Array.isArray(res) ? res : [];
			})
			.catch((err) => {
				console.error("Calendar: Failed to load events:", err);
				error = err;
				events = [];
			})
			.finally(() => {
				loading = false;
			});
	});

	// Period navigation
	function prevPeriod() {
		if (viewMode === "month") {
			currentDate.setMonth(currentDate.getMonth() - 1);
		} else if (viewMode === "week") {
			currentDate.setDate(currentDate.getDate() - 7);
		} else {
			currentDate.setDate(currentDate.getDate() - 30);
		}
	}

	function nextPeriod() {
		if (viewMode === "month") {
			currentDate.setMonth(currentDate.getMonth() + 1);
		} else if (viewMode === "week") {
			currentDate.setDate(currentDate.getDate() + 7);
		} else {
			currentDate.setDate(currentDate.getDate() + 30);
		}
	}

	function jumpToday() {
		currentDate.setTime(Date.now());
	}

	// Month grid days computation
	const monthDays = $derived.by(() => {
		if (viewMode !== "month") return [];
		const y = currentDate.getFullYear();
		const m = currentDate.getMonth();

		const firstDayOfMonth = new SvelteDate(y, m, 1);
		const dayOfWeek = (firstDayOfMonth.getDay() + 6) % 7;

		const start = new SvelteDate(firstDayOfMonth);
		start.setDate(start.getDate() - dayOfWeek);

		const days: {
			date: SvelteDate;
			dateStr: string;
			isCurrentMonth: boolean;
			isToday: boolean;
		}[] = [];
		const todayStr = formatLocalDate(new SvelteDate());

		// 6 weeks * 7 days = 42 cells grid
		for (let i = 0; i < 42; i++) {
			const d = new SvelteDate(start);
			d.setDate(d.getDate() + i);
			const dateStr = formatLocalDate(d);

			days.push({
				date: d,
				dateStr,
				isCurrentMonth: d.getMonth() === m,
				isToday: dateStr === todayStr,
			});
		}

		return days;
	});

	// Week days computation
	const weekDays = $derived.by(() => {
		if (viewMode !== "week") return [];
		const dayOfWeek = (currentDate.getDay() + 6) % 7;
		const start = new SvelteDate(currentDate);
		start.setDate(start.getDate() - dayOfWeek);

		const todayStr = formatLocalDate(new SvelteDate());
		const days: { date: SvelteDate; dateStr: string; isToday: boolean }[] = [];

		for (let i = 0; i < 7; i++) {
			const d = new SvelteDate(start);
			d.setDate(d.getDate() + i);
			const dateStr = formatLocalDate(d);
			days.push({
				date: d,
				dateStr,
				isToday: dateStr === todayStr,
			});
		}

		return days;
	});

	// Group events by date string
	const eventsByDate = $derived.by(() => {
		const map: Record<string, CalendarEvent[]> = {};
		for (const ev of events) {
			if (!ev.releaseDate) continue;
			if (!map[ev.releaseDate]) {
				map[ev.releaseDate] = [];
			}
			map[ev.releaseDate].push(ev);
		}
		return map;
	});

	// Unique dates for Agenda view
	const agendaDates = $derived.by(() => {
		if (viewMode !== "agenda") return [];
		const dates = Object.keys(eventsByDate).sort();
		return dates.map((dateStr) => ({
			dateStr,
			date: new SvelteDate(dateStr + "T00:00:00"),
			events: eventsByDate[dateStr] || [],
		}));
	});

	function getMediaHref(
		ev: CalendarEvent,
	): `${`/${"movie" | "tv"}/${string}`}` {
		return ev.mediaType === "movie"
			? `/movie/${ev.mediaId}`
			: `/tv/${ev.mediaId}`;
	}

	function getPosterSrc(ev: CalendarEvent): string | undefined {
		return ev.posterPath
			? `https://image.tmdb.org/t/p/w200${ev.posterPath}`
			: undefined;
	}
</script>

<svelte:head>
	<title>Calendar - Watcharr</title>
</svelte:head>

<div class="calendar-page">
	<!-- Top Bar / Controls -->
	<header class="calendar-toolbar">
		<div class="left-controls">
			<div class="title-with-icon">
				<Icon i="calendar" wh={26} />
				<h2 class="page-title">Release Calendar</h2>
			</div>

			<div class="nav-cluster">
				<button
					class="plain icon-btn"
					onclick={prevPeriod}
					aria-label="Previous period"
				>
					<Icon i="chevron" wh={18} facing="left" />
				</button>
				<button class="today-btn" onclick={jumpToday}> Today </button>
				<button
					class="plain icon-btn"
					onclick={nextPeriod}
					aria-label="Next period"
				>
					<Icon i="chevron" wh={18} facing="right" />
				</button>
				<h3 class="period-title">{dateRange.title}</h3>
			</div>
		</div>

		<div class="right-controls">
			<!-- Filter Scope: Library vs All -->
			<div class="pill-group">
				<button
					class="pill-btn"
					class:active={scopeFilter === "library"}
					onclick={() => (scopeFilter = "library")}
				>
					My Library
				</button>
				<button
					class="pill-btn"
					class:active={scopeFilter === "all"}
					onclick={() => (scopeFilter = "all")}
				>
					All Releases
				</button>
			</div>

			<!-- Media Type: All / Movies / TV -->
			<div class="pill-group">
				<button
					class="pill-btn"
					class:active={typeFilter === "all"}
					onclick={() => (typeFilter = "all")}
				>
					All
				</button>
				<button
					class="pill-btn"
					class:active={typeFilter === "movie"}
					onclick={() => (typeFilter = "movie")}
				>
					Movies
				</button>
				<button
					class="pill-btn"
					class:active={typeFilter === "tv"}
					onclick={() => (typeFilter = "tv")}
				>
					Shows
				</button>
			</div>

			<!-- View Mode: Month / Week / Agenda -->
			<div class="pill-group view-switcher">
				<button
					class="pill-btn"
					class:active={viewMode === "month"}
					onclick={() => (viewMode = "month")}
				>
					Month
				</button>
				<button
					class="pill-btn"
					class:active={viewMode === "week"}
					onclick={() => (viewMode = "week")}
				>
					Week
				</button>
				<button
					class="pill-btn"
					class:active={viewMode === "agenda"}
					onclick={() => (viewMode = "agenda")}
				>
					Agenda
				</button>
			</div>
		</div>
	</header>

	{#if error}
		<Error pretty="Failed to load calendar events" {error} />
	{:else if loading && events.length === 0}
		<div class="loading-state">
			<Spinner />
			<p>Loading release calendar...</p>
		</div>
	{:else}
		<!-- MONTH VIEW -->
		{#if viewMode === "month"}
			<div class="month-container">
				<!-- Days of week headers -->
				<div class="month-header-row">
					<div class="day-name">Mon</div>
					<div class="day-name">Tue</div>
					<div class="day-name">Wed</div>
					<div class="day-name">Thu</div>
					<div class="day-name">Fri</div>
					<div class="day-name">Sat</div>
					<div class="day-name">Sun</div>
				</div>

				<!-- 42 day grid -->
				<div class="month-grid">
					{#each monthDays as day (day.dateStr)}
						{@const dayEvents = eventsByDate[day.dateStr] || []}
						<div
							class="month-cell"
							class:other-month={!day.isCurrentMonth}
							class:is-today={day.isToday}
						>
							<div class="cell-top">
								<span class="day-number">{day.date.getDate()}</span>
								{#if dayEvents.length > 0}
									<button
										class="plain day-count-badge"
										onclick={() => openDayModal(day.dateStr, dayEvents)}
										title="View all {dayEvents.length} releases"
										aria-label="View all {dayEvents.length} releases for this day"
									>
										{dayEvents.length}
									</button>
								{/if}
							</div>

							<div class="cell-events small-scrollbar">
								{#if dayEvents.length <= 3}
									{#each dayEvents as ev (ev.id)}
										<a
											href={resolve(getMediaHref(ev))}
											class="mini-event-chip"
											class:is-show={ev.mediaType === "tv"}
											title={`${ev.title}${ev.episodeNumber ? ` - S${ev.seasonNumber}E${ev.episodeNumber}` : ""}`}
										>
											<span class="chip-dot"></span>
											<span class="chip-title">
												{#if ev.episodeNumber}
													S{ev.seasonNumber}E{ev.episodeNumber} {ev.title}
												{:else}
													{ev.title}
												{/if}
											</span>
										</a>
									{/each}
								{:else}
									{#each dayEvents.slice(0, 2) as ev (ev.id)}
										<a
											href={resolve(getMediaHref(ev))}
											class="mini-event-chip"
											class:is-show={ev.mediaType === "tv"}
											title={`${ev.title}${ev.episodeNumber ? ` - S${ev.seasonNumber}E${ev.episodeNumber}` : ""}`}
										>
											<span class="chip-dot"></span>
											<span class="chip-title">
												{#if ev.episodeNumber}
													S{ev.seasonNumber}E{ev.episodeNumber} {ev.title}
												{:else}
													{ev.title}
												{/if}
											</span>
										</a>
									{/each}
									<button
										class="plain more-events-btn"
										onclick={() => openDayModal(day.dateStr, dayEvents)}
										title="View all {dayEvents.length} releases for this day"
									>
										+{dayEvents.length - 2} more
									</button>
								{/if}
							</div>
						</div>
					{/each}
				</div>
			</div>

			<!-- WEEK VIEW -->
		{:else if viewMode === "week"}
			<div class="week-container">
				<div class="week-grid">
					{#each weekDays as day (day.dateStr)}
						{@const dayEvents = eventsByDate[day.dateStr] || []}
						<div class="week-col" class:is-today={day.isToday}>
							<div class="week-col-header">
								<span class="week-day-name">
									{day.date.toLocaleString(undefined, { weekday: "short" })}
								</span>
								<span class="week-day-number">{day.date.getDate()}</span>
							</div>

							<div class="week-col-events small-scrollbar">
								{#if dayEvents.length === 0}
									<div class="no-events-placeholder">
										<span>No releases</span>
									</div>
								{:else}
									{#each dayEvents as ev (ev.id)}
										{@const poster = getPosterSrc(ev)}
										<a href={resolve(getMediaHref(ev))} class="week-event-card">
											<div class="card-thumb">
												{#if poster}
													<img src={poster} alt={ev.title} loading="lazy" />
												{:else}
													<div class="fallback-icon">
														<Icon
															i={ev.mediaType === "movie" ? "film" : "tv"}
															wh={20}
														/>
													</div>
												{/if}
											</div>
											<div class="card-content">
												<strong class="event-title">{ev.title}</strong>
												{#if ev.episodeNumber}
													<span class="ep-tag">
														S{ev.seasonNumber} · E{ev.episodeNumber}
													</span>
													{#if ev.episodeName}
														<span class="ep-name">{ev.episodeName}</span>
													{/if}
												{:else}
													<span class="movie-tag">Movie Premiere</span>
												{/if}
												{#if ev.inLibrary && ev.status}
													<span class="library-tag">{ev.status}</span>
												{/if}
											</div>
										</a>
									{/each}
								{/if}
							</div>
						</div>
					{/each}
				</div>
			</div>

			<!-- AGENDA VIEW -->
		{:else if viewMode === "agenda"}
			<div class="agenda-container">
				{#if agendaDates.length === 0}
					<div class="empty-state">
						<Icon i="calendar" wh={48} />
						<h3>No scheduled releases</h3>
						<p>No premieres or episode air dates found for this period.</p>
						{#if scopeFilter === "library"}
							<button
								class="pill-btn primary"
								onclick={() => (scopeFilter = "all")}
							>
								View All TMDB Releases
							</button>
						{/if}
					</div>
				{:else}
					{#each agendaDates as group (group.dateStr)}
						<div class="agenda-group">
							<div class="agenda-date-header">
								<span class="agenda-date-main">
									{group.date.toLocaleString(undefined, {
										weekday: "long",
										month: "long",
										day: "numeric",
										year: "numeric",
									})}
								</span>
								<span class="agenda-date-relative">
									{toRelativeDate(group.date)}
								</span>
							</div>

							<div class="agenda-events-list">
								{#each group.events as ev (ev.id)}
									{@const poster = getPosterSrc(ev)}
									<a href={resolve(getMediaHref(ev))} class="agenda-event-card">
										<div class="agenda-thumb">
											{#if poster}
												<img src={poster} alt={ev.title} loading="lazy" />
											{:else}
												<div class="fallback-icon">
													<Icon
														i={ev.mediaType === "movie" ? "film" : "tv"}
														wh={24}
													/>
												</div>
											{/if}
										</div>

										<div class="agenda-details">
											<div class="agenda-top-row">
												<h4 class="agenda-title">{ev.title}</h4>
												<div class="agenda-badges">
													{#if ev.episodeNumber}
														<span class="badge ep">
															S{ev.seasonNumber} · E{ev.episodeNumber}
														</span>
													{:else}
														<span class="badge movie">Movie</span>
													{/if}
													{#if ev.inLibrary && ev.status}
														<span class="badge status">{ev.status}</span>
													{/if}
												</div>
											</div>

											{#if ev.episodeName}
												<span class="agenda-ep-title">{ev.episodeName}</span>
											{/if}

											{#if ev.overview}
												<p class="agenda-overview">{ev.overview}</p>
											{/if}
										</div>
									</a>
								{/each}
							</div>
						</div>
					{/each}
				{/if}
			</div>
		{/if}
	{/if}

	{#if activeDayModalDate && activeDayModalEvents.length > 0}
		<Modal
			title={`Releases for ${new SvelteDate(activeDayModalDate + "T00:00:00").toLocaleDateString(undefined, { weekday: "long", month: "long", day: "numeric", year: "numeric" })}`}
			desc={`${activeDayModalEvents.length} items scheduled`}
			onClose={closeDayModal}
			maxWidth="640px"
		>
			<div class="day-modal-list small-scrollbar">
				{#each activeDayModalEvents as ev (ev.id)}
					{@const poster = getPosterSrc(ev)}
					<a
						href={resolve(getMediaHref(ev))}
						class="day-modal-item"
						onclick={closeDayModal}
					>
						<div class="modal-item-thumb">
							{#if poster}
								<img src={poster} alt={ev.title} loading="lazy" />
							{:else}
								<div class="thumb-fallback">
									<Icon i={ev.mediaType === "tv" ? "tv" : "film"} wh={22} />
								</div>
							{/if}
						</div>
						<div class="modal-item-info">
							<div class="item-title-row">
								<strong class="item-title">{ev.title}</strong>
								<span
									class="badge {ev.mediaType === 'tv'
										? 'badge-tv'
										: 'badge-movie'}"
								>
									{ev.mediaType === "tv" ? "TV Show" : "Movie"}
								</span>
							</div>
							{#if ev.episodeNumber}
								<span class="item-ep-badge">
									Season {ev.seasonNumber}, Episode {ev.episodeNumber}
									{#if ev.episodeName}
										— "{ev.episodeName}"
									{/if}
								</span>
							{/if}
							{#if ev.overview}
								<p class="item-overview">{ev.overview}</p>
							{/if}
						</div>
					</a>
				{/each}
			</div>
		</Modal>
	{/if}
</div>

<style lang="scss">
	.calendar-page {
		max-width: 1400px;
		margin: 0 auto;
		padding: 24px 20px 80px;
		display: flex;
		flex-direction: column;
		gap: 20px;
		min-height: calc(100vh - 120px);
	}

	.calendar-toolbar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		flex-wrap: wrap;
		gap: 16px;
		background: rgba(22, 27, 38, 0.7);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 18px;
		padding: 16px 20px;
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
		backdrop-filter: blur(16px);
		-webkit-backdrop-filter: blur(16px);
	}

	.left-controls {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 20px;

		.title-with-icon {
			display: flex;
			align-items: center;
			gap: 10px;
			color: #ffffff;

			.page-title {
				margin: 0;
				font-size: 20px;
				font-weight: 700;
				letter-spacing: -0.01em;
			}
		}

		.nav-cluster {
			display: flex;
			align-items: center;
			gap: 8px;

			.icon-btn {
				display: flex;
				align-items: center;
				justify-content: center;
				width: 32px;
				height: 32px;
				border-radius: 8px;
				background: rgba(255, 255, 255, 0.06);
				border: 1px solid rgba(255, 255, 255, 0.1);
				color: #ffffff;
				cursor: pointer;
				transition: background 120ms ease;

				&:hover {
					background: rgba(255, 255, 255, 0.14);
				}
			}

			.today-btn {
				padding: 5px 12px;
				border-radius: 8px;
				background: rgba(255, 255, 255, 0.08);
				border: 1px solid rgba(255, 255, 255, 0.12);
				color: #ffffff;
				font-size: 13px;
				font-weight: 600;
				cursor: pointer;
				transition: background 120ms ease;

				&:hover {
					background: rgba(255, 255, 255, 0.16);
				}
			}

			.period-title {
				margin: 0 0 0 8px;
				font-size: 16px;
				font-weight: 700;
				color: #f1f5f9;
			}
		}
	}

	.right-controls {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 12px;
	}

	.pill-group {
		display: inline-flex;
		background: rgba(14, 18, 28, 0.85);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 10px;
		padding: 3px;
		gap: 2px;

		.pill-btn {
			padding: 5px 12px;
			border-radius: 8px;
			background: transparent;
			border: none;
			color: rgba(255, 255, 255, 0.65);
			font-size: 13px;
			font-weight: 600;
			cursor: pointer;
			transition:
				background 150ms ease,
				color 150ms ease;

			&:hover:not(.active) {
				color: #ffffff;
				background: rgba(255, 255, 255, 0.05);
			}

			&.active {
				background: #6366f1;
				color: #ffffff;
				box-shadow: 0 2px 8px rgba(99, 102, 241, 0.4);
			}
		}
	}

	.loading-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 80px 20px;
		gap: 16px;
		color: rgba(255, 255, 255, 0.7);
	}

	/* MONTH VIEW STYLES */
	.month-container {
		display: flex;
		flex-direction: column;
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 18px;
		overflow: hidden;
		background: rgba(18, 22, 32, 0.65);
		backdrop-filter: blur(12px);
		-webkit-backdrop-filter: blur(12px);
		box-shadow: 0 8px 30px rgba(0, 0, 0, 0.4);
	}

	.month-header-row {
		display: grid;
		grid-template-columns: repeat(7, 1fr);
		background: rgba(25, 31, 46, 0.85);
		border-bottom: 1px solid rgba(255, 255, 255, 0.08);

		.day-name {
			padding: 12px 8px;
			text-align: center;
			font-size: 13px;
			font-weight: 700;
			letter-spacing: 0.04em;
			text-transform: uppercase;
			color: rgba(255, 255, 255, 0.6);
		}
	}

	.month-grid {
		display: grid;
		grid-template-columns: repeat(7, 1fr);
		grid-auto-rows: minmax(120px, 1fr);
		gap: 1px;
		background: rgba(255, 255, 255, 0.05);

		.month-cell {
			background: #0f141e;
			padding: 8px;
			display: flex;
			flex-direction: column;
			gap: 6px;
			min-width: 0;
			min-height: 120px;
			height: 100%;
			overflow: hidden;
			position: relative;
			transition: background 120ms ease;

			&:hover {
				background: #141a27;
			}

			&.other-month {
				opacity: 0.45;
				background: #0b0e15;
			}

			&.is-today {
				background: rgba(99, 102, 241, 0.08);
				box-shadow: inset 0 0 0 1.5px rgba(99, 102, 241, 0.6);

				.day-number {
					color: #818cf8;
					font-weight: 800;
				}
			}

			.cell-top {
				display: flex;
				justify-content: space-between;
				align-items: center;
				flex-shrink: 0;

				.day-number {
					font-size: 13px;
					font-weight: 600;
					color: rgba(255, 255, 255, 0.85);
				}

				.day-count-badge {
					font-size: 11px;
					font-weight: 700;
					padding: 1px 6px;
					border-radius: 999px;
					background: rgba(255, 255, 255, 0.1);
					color: rgba(255, 255, 255, 0.7);
					cursor: pointer;
					border: none;
					transition: all 120ms ease;

					&:hover {
						background: rgba(99, 102, 241, 0.4);
						color: #ffffff;
					}
				}
			}

			.cell-events {
				display: flex;
				flex-direction: column;
				gap: 4px;
				overflow-y: auto;
				overflow-x: hidden;
				flex: 1;
				min-height: 0;
				max-height: 90px;
				scrollbar-width: thin;

				.mini-event-chip {
					display: flex;
					align-items: center;
					gap: 6px;
					padding: 3px 6px;
					border-radius: 6px;
					background: rgba(22, 27, 40, 0.9);
					border: 1px solid rgba(255, 255, 255, 0.08);
					text-decoration: none;
					color: #e2e8f0;
					font-size: 11.5px;
					line-height: 1.2;
					flex-shrink: 0;
					transition:
						background 120ms ease,
						border-color 120ms ease;

					&:hover {
						background: rgba(30, 36, 52, 1);
						border-color: rgba(99, 102, 241, 0.4);
					}

					.chip-dot {
						width: 6px;
						height: 6px;
						border-radius: 50%;
						background: #f59e0b;
						flex-shrink: 0;
					}

					&.is-show .chip-dot {
						background: #6366f1;
					}

					.chip-title {
						white-space: nowrap;
						overflow: hidden;
						text-overflow: ellipsis;
					}
				}

				.more-events-btn {
					display: flex;
					align-items: center;
					justify-content: center;
					padding: 2px 6px;
					border-radius: 5px;
					background: rgba(99, 102, 241, 0.15);
					border: 1px dashed rgba(99, 102, 241, 0.4);
					color: #a5b4fc;
					font-size: 11px;
					font-weight: 600;
					cursor: pointer;
					margin-top: 1px;
					flex-shrink: 0;
					transition: all 120ms ease;

					&:hover {
						background: rgba(99, 102, 241, 0.3);
						border-color: #818cf8;
						color: #ffffff;
					}
				}
			}
		}
	}

	.day-modal-list {
		display: flex;
		flex-direction: column;
		gap: 10px;
		max-height: 60vh;
		overflow-y: auto;
		padding: 6px 2px;

		.day-modal-item {
			display: flex;
			gap: 14px;
			padding: 12px;
			background: rgba(22, 27, 40, 0.6);
			border: 1px solid rgba(255, 255, 255, 0.08);
			border-radius: 12px;
			text-decoration: none;
			color: inherit;
			transition: all 140ms ease;

			&:hover {
				background: rgba(28, 35, 52, 0.95);
				border-color: rgba(99, 102, 241, 0.4);
				transform: translateY(-1px);
			}

			.modal-item-thumb {
				width: 52px;
				height: 78px;
				border-radius: 6px;
				overflow: hidden;
				background: #000;
				flex-shrink: 0;

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
					color: rgba(255, 255, 255, 0.4);
				}
			}

			.modal-item-info {
				display: flex;
				flex-direction: column;
				gap: 4px;
				flex: 1;
				min-width: 0;

				.item-title-row {
					display: flex;
					align-items: center;
					justify-content: space-between;
					gap: 8px;

					.item-title {
						font-size: 14px;
						font-weight: 700;
						color: #ffffff;
					}

					.badge {
						font-size: 10.5px;
						font-weight: 600;
						padding: 2px 7px;
						border-radius: 4px;

						&.badge-tv {
							background: rgba(99, 102, 241, 0.2);
							color: #818cf8;
						}
						&.badge-movie {
							background: rgba(245, 158, 11, 0.2);
							color: #fbbf24;
						}
					}
				}

				.item-ep-badge {
					font-size: 12px;
					font-weight: 600;
					color: #60a5fa;
				}

				.item-overview {
					margin: 2px 0 0;
					font-size: 12px;
					line-height: 1.4;
					color: rgba(255, 255, 255, 0.6);
					display: -webkit-box;
					-webkit-line-clamp: 2;
					line-clamp: 2;
					-webkit-box-orient: vertical;
					overflow: hidden;
				}
			}
		}
	}

	/* WEEK VIEW STYLES */
	.week-container {
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 18px;
		overflow: hidden;
		background: rgba(18, 22, 32, 0.65);
		backdrop-filter: blur(12px);
		-webkit-backdrop-filter: blur(12px);
	}

	.week-grid {
		display: grid;
		grid-template-columns: repeat(7, 1fr);
		gap: 1px;
		background: rgba(255, 255, 255, 0.05);

		.week-col {
			background: #0f141e;
			display: flex;
			flex-direction: column;
			min-width: 0;
			min-height: 500px;

			&.is-today {
				background: rgba(99, 102, 241, 0.05);
				.week-col-header {
					background: rgba(99, 102, 241, 0.15);
					border-bottom: 2px solid #6366f1;
					.week-day-number {
						color: #818cf8;
						font-weight: 800;
					}
				}
			}

			.week-col-header {
				padding: 12px 10px;
				display: flex;
				flex-direction: column;
				align-items: center;
				gap: 4px;
				background: rgba(25, 31, 46, 0.85);
				border-bottom: 1px solid rgba(255, 255, 255, 0.08);

				.week-day-name {
					font-size: 12px;
					font-weight: 700;
					text-transform: uppercase;
					color: rgba(255, 255, 255, 0.6);
				}

				.week-day-number {
					font-size: 18px;
					font-weight: 700;
					color: #ffffff;
				}
			}

			.week-col-events {
				padding: 10px;
				display: flex;
				flex-direction: column;
				gap: 10px;
				overflow-y: auto;
				flex: 1;

				.no-events-placeholder {
					display: flex;
					align-items: center;
					justify-content: center;
					height: 100px;
					color: rgba(255, 255, 255, 0.25);
					font-size: 12px;
				}

				.week-event-card {
					display: flex;
					flex-direction: column;
					gap: 8px;
					padding: 10px;
					border-radius: 12px;
					background: rgba(22, 27, 40, 0.8);
					border: 1px solid rgba(255, 255, 255, 0.08);
					text-decoration: none;
					transition:
						transform 140ms ease,
						background 140ms ease,
						border-color 140ms ease;

					&:hover {
						transform: translateY(-2px);
						background: rgba(28, 35, 50, 0.95);
						border-color: rgba(99, 102, 241, 0.35);
					}

					.card-thumb {
						width: 100%;
						aspect-ratio: 16 / 9;
						border-radius: 8px;
						overflow: hidden;
						background: #000000;

						img {
							width: 100%;
							height: 100%;
							object-fit: cover;
						}

						.fallback-icon {
							width: 100%;
							height: 100%;
							display: flex;
							align-items: center;
							justify-content: center;
							color: rgba(255, 255, 255, 0.25);
						}
					}

					.card-content {
						display: flex;
						flex-direction: column;
						gap: 3px;

						.event-title {
							font-size: 13.5px;
							font-weight: 700;
							color: #ffffff;
							line-height: 1.3;
						}

						.ep-tag {
							font-size: 11.5px;
							font-weight: 600;
							color: #818cf8;
						}

						.ep-name {
							font-size: 11.5px;
							color: rgba(255, 255, 255, 0.65);
							white-space: nowrap;
							overflow: hidden;
							text-overflow: ellipsis;
						}

						.movie-tag {
							font-size: 11.5px;
							font-weight: 600;
							color: #fbbf24;
						}

						.library-tag {
							align-self: flex-start;
							margin-top: 4px;
							font-size: 10.5px;
							font-weight: 600;
							text-transform: uppercase;
							letter-spacing: 0.04em;
							padding: 2px 6px;
							border-radius: 4px;
							background: rgba(99, 102, 241, 0.15);
							color: #a5b4fc;
						}
					}
				}
			}
		}
	}

	/* AGENDA VIEW STYLES */
	.agenda-container {
		display: flex;
		flex-direction: column;
		gap: 24px;
	}

	.agenda-group {
		display: flex;
		flex-direction: column;
		gap: 12px;

		.agenda-date-header {
			display: flex;
			align-items: baseline;
			gap: 12px;
			padding-bottom: 6px;
			border-bottom: 1px solid rgba(255, 255, 255, 0.08);

			.agenda-date-main {
				font-size: 16px;
				font-weight: 700;
				color: #ffffff;
			}

			.agenda-date-relative {
				font-size: 13px;
				color: rgba(255, 255, 255, 0.5);
			}
		}

		.agenda-events-list {
			display: grid;
			grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
			gap: 14px;

			@media screen and (max-width: 480px) {
				grid-template-columns: 1fr;
			}
		}

		.agenda-event-card {
			display: flex;
			gap: 14px;
			padding: 14px;
			border-radius: 14px;
			background: rgba(22, 27, 40, 0.7);
			border: 1px solid rgba(255, 255, 255, 0.08);
			text-decoration: none;
			backdrop-filter: blur(12px);
			-webkit-backdrop-filter: blur(12px);
			transition:
				transform 140ms ease,
				background 140ms ease,
				border-color 140ms ease;

			&:hover {
				transform: translateY(-2px);
				background: rgba(28, 35, 52, 0.9);
				border-color: rgba(99, 102, 241, 0.35);
			}

			.agenda-thumb {
				width: 75px;
				height: 110px;
				flex-shrink: 0;
				border-radius: 8px;
				overflow: hidden;
				background: #000000;

				img {
					width: 100%;
					height: 100%;
					object-fit: cover;
				}

				.fallback-icon {
					width: 100%;
					height: 100%;
					display: flex;
					align-items: center;
					justify-content: center;
					color: rgba(255, 255, 255, 0.25);
				}
			}

			.agenda-details {
				flex: 1;
				min-width: 0;
				display: flex;
				flex-direction: column;
				gap: 4px;

				.agenda-top-row {
					display: flex;
					align-items: flex-start;
					justify-content: space-between;
					gap: 8px;

					.agenda-title {
						margin: 0;
						font-size: 16px;
						font-weight: 700;
						color: #ffffff;
						line-height: 1.3;
					}

					.agenda-badges {
						display: flex;
						gap: 6px;
						flex-shrink: 0;

						.badge {
							font-size: 11px;
							font-weight: 700;
							padding: 2px 7px;
							border-radius: 6px;

							&.ep {
								background: rgba(99, 102, 241, 0.15);
								border: 1px solid rgba(99, 102, 241, 0.3);
								color: #a5b4fc;
							}

							&.movie {
								background: rgba(245, 158, 11, 0.15);
								border: 1px solid rgba(245, 158, 11, 0.3);
								color: #fcd34d;
							}

							&.status {
								background: rgba(255, 255, 255, 0.08);
								border: 1px solid rgba(255, 255, 255, 0.12);
								color: rgba(255, 255, 255, 0.7);
								text-transform: uppercase;
							}
						}
					}
				}

				.agenda-ep-title {
					font-size: 13px;
					font-weight: 600;
					color: rgba(255, 255, 255, 0.75);
				}

				.agenda-overview {
					margin: 4px 0 0;
					font-size: 12.5px;
					line-height: 1.4;
					color: rgba(255, 255, 255, 0.6);
					display: -webkit-box;
					-webkit-line-clamp: 2;
					line-clamp: 2;
					-webkit-box-orient: vertical;
					overflow: hidden;
				}
			}
		}
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 80px 20px;
		gap: 12px;
		text-align: center;
		color: rgba(255, 255, 255, 0.5);

		h3 {
			margin: 0;
			font-size: 18px;
			color: #ffffff;
		}

		p {
			margin: 0;
			font-size: 14px;
			max-width: 400px;
		}

		.pill-btn.primary {
			background: #6366f1;
			color: #ffffff;
			margin-top: 12px;
			padding: 8px 18px;
			border-radius: 8px;
			border: none;
			font-weight: 600;
			cursor: pointer;
			transition: background 150ms ease;

			&:hover {
				background: #4f46e5;
			}
		}
	}

	@media screen and (max-width: 900px) {
		.calendar-toolbar {
			flex-direction: column;
			align-items: stretch;

			.left-controls,
			.right-controls {
				justify-content: space-between;
			}
		}

		.week-grid {
			display: flex;
			flex-direction: column;

			.week-col {
				min-height: unset;
			}
		}
	}

	@media screen and (max-width: 600px) {
		.calendar-page {
			padding: 14px 10px 80px;
			gap: 14px;
		}

		.calendar-toolbar {
			padding: 12px 14px;
			border-radius: 14px;
			gap: 12px;
		}

		.month-container {
			overflow-x: auto;
			-webkit-overflow-scrolling: touch;
		}

		.month-header-row,
		.month-grid {
			min-width: 320px;
		}

		.month-header-row .day-name {
			font-size: 11px;
			padding: 8px 2px;
		}

		.month-grid .month-cell {
			min-height: 70px;
			padding: 4px;

			.cell-events {
				max-height: 50px;

				.mini-event-chip {
					font-size: 10px;
					padding: 2px 4px;
				}

				.more-events-btn {
					font-size: 9.5px;
					padding: 1px 3px;
				}
			}
		}
	}
</style>
