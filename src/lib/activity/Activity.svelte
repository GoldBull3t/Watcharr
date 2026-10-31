<script lang="ts">
	import { type Activity, ActivityCreatedBy } from "@/types";
	import {
		getOrdinalSuffix,
		months,
		seasonAndEpToReadable,
	} from "../util/helpers";
	import ActivityEditor from "./ActivityEditor.svelte";
	import ActivityIcon from "./ActivityIcon.svelte";

	interface Props {
		activity: Activity[] | undefined;
		onRemoved: (activity: Activity) => void;
	}

	let { activity = undefined, onRemoved }: Props = $props();

	let clickedActivity: Activity | undefined = $state();
	let groupedActivities: { [index: string]: Activity[] } = $derived(
		getGroupedActivity(activity),
	);
	let count = $derived(activity ? activity.length : 0);

	interface ActivityPayload {
		status?: string;
		rating?: number;
		season?: number;
		episode?: number;
	}

	function safeJsonParse(str: string | undefined): ActivityPayload | undefined {
		if (!str) return undefined;
		try {
			const parsed: unknown = JSON.parse(str);
			if (typeof parsed === "number") {
				return { rating: parsed };
			}
			if (parsed && typeof parsed === "object") {
				return parsed as ActivityPayload;
			}
			return undefined;
		} catch {
			return undefined;
		}
	}

	function getMsg(a: Activity) {
		switch (a?.type) {
			case "ADDED_WATCHED":
				if (a.data) {
					const data = safeJsonParse(a.data);
					return `Added to Watched List${data?.status ? ` as ${toFullTitleCase(data.status)}` : ""}${
						data?.rating ? ` with ${data.rating} Stars` : ""
					}`;
				}
				return "Added to Watched List";
			case "REMOVED_WATCHED":
				return "Removed from Watched List";
			case "RATING_CHANGED":
				if (a.data) {
					return `Rating Changed to ${a.data}`;
				}
				return "Rating Changed";
			case "STATUS_CHANGED":
				if (a.data) {
					let status = a.data;
					if (a.data.startsWith("{")) {
						const data = safeJsonParse(a.data);
						if (data?.status) status = data.status;
					}
					return `Status Changed to ${toFullTitleCase(status)}`;
				}
				return "Status Changed";
			case "THOUGHTS_CHANGED":
				return "Thoughts Changed";
			case "THOUGHTS_REMOVED":
				return "Thoughts Removed";
			case "IMPORTED_RATING":
				if (a.data) {
					const data = safeJsonParse(a.data);
					if (data && typeof data.rating !== "undefined") {
						return `Rating Changed to ${data.rating}`;
					}
					if (!isNaN(Number(a.data))) {
						return `Rating Changed to ${a.data}`;
					}
					return "Added to Watchlist with No Rating";
				}
				return "Imported Rating";
			case "IMPORTED_ADDED_WATCHED":
				return "Imported Watch Date";
			case "SEASON_ADDED":
				if (a.data) {
					const data = safeJsonParse(a.data);
					if (data?.season) {
						return `Season ${data.season} Added as ${toFullTitleCase(data.status)}`;
					}
				}
				return "Season Added";
			case "SEASON_RATING_CHANGED":
				if (a.data) {
					const data = safeJsonParse(a.data);
					if (data?.season) {
						return `Changed Season ${data.season} Rating to ${data.rating}`;
					}
				}
				return "Season Rating Changed";
			case "SEASON_STATUS_CHANGED":
				if (a.data) {
					const data = safeJsonParse(a.data);
					if (data?.season) {
						return `Changed Season ${data.season} Status to ${toFullTitleCase(data.status)}`;
					}
				}
				return "Season Status Changed";
			case "SEASON_REMOVED":
				if (a.data) {
					const data = safeJsonParse(a.data);
					if (data?.season) {
						return `Season ${data.season} Status Removed`;
					}
				}
				return "Season Removed";
			case "EPISODE_ADDED":
				if (a.data) {
					const data = safeJsonParse(a.data);
					if (data) {
						return `${seasonAndEpToReadable(data.season, data.episode)} Added ${data.status ? `as ${toFullTitleCase(data.status)}` : data.rating ? `with Rating ${data.rating}` : ""}`;
					}
				}
				return "Episode Added";
			case "EPISODE_RATING_CHANGED":
				if (a.data) {
					const data = safeJsonParse(a.data);
					if (data) {
						return `${seasonAndEpToReadable(data.season, data.episode)} Rating Changed to ${data.rating}`;
					}
				}
				return "Episode Rating Changed";
			case "EPISODE_STATUS_CHANGED":
				if (a.data) {
					const data = safeJsonParse(a.data);
					if (data) {
						return `${seasonAndEpToReadable(data.season, data.episode)} Status Changed to ${toFullTitleCase(data.status)}`;
					}
				}
				return "Episode Status Changed";
			case "EPISODE_REMOVED":
				if (a.data) {
					const data = safeJsonParse(a.data);
					if (data) {
						return `${seasonAndEpToReadable(data.season, data.episode)} Removed`;
					}
				}
				return "Episode Removed";
			default:
				return a?.type || "Activity";
		}
	}

	function toFullTitleCase(text: string | undefined) {
		if (text) {
			if (text.toUpperCase() === "DROPPED") {
				return "Stopped Watching";
			}
			if (
				text.toUpperCase() === "CANCELLED" ||
				text.toUpperCase() === "CANCELED"
			) {
				return "Cancelled";
			}
			return text
				.split(" ")
				.map((l) =>
					l[0] ? l[0].toUpperCase() + l.substring(1).toLowerCase() : l,
				)
				.join(" ");
		}
		return "Unknown";
	}

	function toDayTime(d: Date | string | undefined | null) {
		if (!d) return "Unknown";
		const date = d instanceof Date ? d : new Date(d);
		if (isNaN(date.getTime())) return "Unknown";
		return `${date.getDate()}${getOrdinalSuffix(date.getDate())} at ${date.toLocaleTimeString()}`;
	}

	/**
	 * Get what will be the visible date the activity was 'created'.
	 * @returns timestamp in milliseconds, or 0 if invalid.
	 */
	function getCreatedAtVis(a: Activity): number {
		const raw: unknown = a.customDate ?? a.createdAt;
		if (!raw) return 0;
		const parsed =
			raw instanceof Date ? raw.getTime() : Date.parse(String(raw));
		return isNaN(parsed) ? 0 : parsed;
	}

	function getGroupedActivity(activities?: Activity[]) {
		activities = activities?.filter((a) => a.type);
		const a = activities?.sort(
			(a, b) => getCreatedAtVis(b) - getCreatedAtVis(a),
		);
		let grouped: { [index: string]: Activity[] } = {};
		if (a) {
			for (let i = 0; i < a.length; i++) {
				const activity = a[i];
				const ts = getCreatedAtVis(activity);
				const date = new Date(ts);
				const key =
					isNaN(date.getTime()) || ts === 0
						? "Unknown Date"
						: `${months[date.getMonth()] ?? "Unknown"} ${date.getFullYear()}`;
				if (grouped[key]) {
					grouped[key].push(activity);
				} else {
					grouped[key] = [activity];
				}
			}
		}
		return grouped;
	}

	function openEditor(a: Activity) {
		clickedActivity = a;
		return;
	}
</script>

{#if clickedActivity}
	<ActivityEditor
		activity={clickedActivity}
		activityMessage={getMsg(clickedActivity)}
		onClose={() => (clickedActivity = undefined)}
		onUpdated={(activityId, updatedActivity) => {
			if (!activity) {
				console.error(
					"ActivityEditor->onUpdated: 'activity' doesn't exist somehow..",
				);
				return;
			}
			const ai = activity.findIndex((a) => a.id === activityId);
			if (ai !== -1) {
				activity[ai] = updatedActivity;
			}
		}}
		onRemoved={(a) => {
			if (!activity) {
				console.error(
					"ActivityEditor->onRemoved: 'activity' doesn't exist somehow..",
				);
				return;
			}
			onRemoved(a);
		}}
	/>
{/if}

<div class="activity">
	<div class="title">
		<h2>Activity</h2>
		<div class="act-count">{count}</div>
	</div>
	{#if groupedActivities && Object.keys(groupedActivities).length > 0}
		<ul>
			{#each Object.keys(groupedActivities) as k (k)}
				<li class="month-group-header">
					<h3>{k}</h3>
				</li>

				{#each groupedActivities[k] as a (a.id)}
					{@const d = new Date(getCreatedAtVis(a))}
					<li class="activity-row">
						<button class="plain row-btn" onclick={() => openEditor(a)}>
							<span class="act-msg">{getMsg(a)}</span>
							<span class="act-date" title={d.toDateString()}
								>{toDayTime(d)}</span
							>
						</button>
						<div class="act-icons">
							{#if a.createdBy == ActivityCreatedBy.Watcharr}
								<ActivityIcon
									icon="sparkles"
									tooltipText={a.reason
										? `(Automated) ${a.reason}`
										: "Completed by an automation."}
								/>
							{:else if a.createdBy == ActivityCreatedBy.GenericImport}
								<ActivityIcon icon="calendar" tooltipText="Imported" />
							{:else if a.createdBy == ActivityCreatedBy.JellyfinImport}
								<ActivityIcon icon="jellyfin" tooltipText="Jellyfin Import" />
							{:else if a.createdBy == ActivityCreatedBy.JellyfinWebhook}
								<ActivityIcon icon="jellyfin" tooltipText="Jellyfin Webhook" />
							{:else if a.createdBy == ActivityCreatedBy.PlexImport}
								<ActivityIcon icon="plex" tooltipText="Plex Import" />
							{/if}
							{#if a.countAsPlay}
								<ActivityIcon icon="play" tooltipText="Counts as a Play." />
							{/if}
						</div>
					</li>
				{/each}
			{/each}
		</ul>
	{:else}
		<span class="empty-msg">No activity recorded yet.</span>
	{/if}
</div>

<style lang="scss">
	.activity {
		width: 100%;
		background: rgba(15, 19, 28, 0.72);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 10px;
		padding: 20px 24px;

		@media screen and (max-width: 600px) {
			padding: 16px;
		}

		div.title {
			display: flex;
			flex-flow: row;
			align-items: center;
			gap: 10px;
			margin-bottom: 14px;
			padding-bottom: 12px;
			border-bottom: 1px solid rgba(255, 255, 255, 0.06);

			h2 {
				margin: 0;
				font-size: 18px;
				font-weight: 700;
				letter-spacing: -0.01em;
				color: #ffffff;
			}

			.act-count {
				font-size: 11.5px;
				font-weight: 700;
				color: $brand-color;
				background: rgba(245, 158, 11, 0.12);
				border: 1px solid rgba(245, 158, 11, 0.28);
				padding: 2px 8px;
				border-radius: 6px;
				display: inline-flex;
				align-items: center;
				justify-content: center;
			}
		}

		ul {
			display: flex;
			flex-flow: column;
			gap: 6px;
			margin: 0;
			padding: 0 4px 0 0;
			list-style: none;
			max-height: 300px;
			overflow-y: auto;
			overflow-x: hidden;
			scrollbar-gutter: stable;
			overscroll-behavior: contain;

			li.month-group-header {
				margin-top: 8px;

				&:first-child {
					margin-top: 0;
				}

				h3 {
					margin: 0;
					font-size: 11px;
					font-weight: 700;
					text-transform: uppercase;
					letter-spacing: 0.06em;
					color: rgba(255, 255, 255, 0.45);
					padding: 4px 0;
				}
			}

			li.activity-row {
				display: flex;
				flex-flow: row;
				gap: 10px;
				align-items: center;
				justify-content: space-between;
				padding: 8px 12px;
				border-radius: 6px;
				background: rgba(255, 255, 255, 0.025);
				border: 1px solid rgba(255, 255, 255, 0.05);
				transition:
					background 140ms ease,
					border-color 140ms ease;

				&:hover {
					background: rgba(255, 255, 255, 0.05);
					border-color: rgba(255, 255, 255, 0.12);
				}

				button.row-btn {
					all: unset;
					display: flex;
					flex-flow: row;
					align-items: center;
					justify-content: space-between;
					flex: 1;
					gap: 12px;
					min-width: 0;
					cursor: pointer;

					@media screen and (max-width: 560px) {
						flex-direction: column;
						align-items: flex-start;
						gap: 4px;
					}
				}

				.act-msg {
					font-size: 13.5px;
					font-weight: 500;
					color: #f1f5f9;
				}

				.act-date {
					font-size: 12px;
					color: rgba(255, 255, 255, 0.48);
					white-space: nowrap;
				}

				.act-icons {
					display: inline-flex;
					align-items: center;
					gap: 6px;
					flex-shrink: 0;
				}
			}
		}

		.empty-msg {
			display: block;
			font-size: 13.5px;
			color: rgba(255, 255, 255, 0.45);
			padding: 8px 0;
		}
	}
</style>
