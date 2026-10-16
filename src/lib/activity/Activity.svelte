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

	function safeJsonParse<T = any>(str: string | undefined): T | undefined {
		if (!str) return undefined;
		try {
			return JSON.parse(str);
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
					if (typeof data === "number") {
						return `Rating Changed to ${data}`;
					}
					if (data && typeof data === "object" && typeof data.rating !== "undefined") {
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
			if (text.toUpperCase() === "CANCELLED" || text.toUpperCase() === "CANCELED") {
				return "Cancelled";
			}
			return text
				.split(" ")
				.map((l) => (l[0] ? l[0].toUpperCase() + l.substring(1).toLowerCase() : l))
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
		const parsed = raw instanceof Date ? raw.getTime() : Date.parse(String(raw));
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
				const key = isNaN(date.getTime()) || ts === 0
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
				<h3>{k}</h3>

				{#each groupedActivities[k] as a (a.id)}
					{@const d = new Date(getCreatedAtVis(a))}
					<li>
						<button class="plain" onclick={() => openEditor(a)}>
							<span title={d.toDateString()}>{toDayTime(d)}</span>
							<span>{getMsg(a)}</span>
						</button>
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
					</li>
				{/each}
			{/each}
		</ul>
	{:else}
		<span>You Have No Activity!</span>
	{/if}
</div>

<style lang="scss">
	.activity {
		width: 100%;

		div.title {
			display: flex;
			flex-flow: row;
			align-items: center;
			gap: 10px;

			.act-count {
				font-family:
					"Shrikhand",
					system-ui,
					-apple-system,
					BlinkMacSystemFont;
				font-size: 13px;
				color: black;
				background-color: $gold;
				padding: 3px 6px;
				border-radius: 50px;
				min-width: 25px;
				min-height: 25px;
				height: min-content;
				display: flex;
				align-items: center;
				justify-content: center;
			}
		}

		ul {
			display: flex;
			flex-flow: column;
			gap: 8px;
			margin-top: 8px;
			margin-left: calc(30px + 8px);
			list-style: none;
			max-height: 250px;
			overflow-y: auto;
			overflow-x: hidden;

			h3 {
				position: sticky;
				top: 0;
				font-size: 16px;
				font-family:
					sans-serif,
					system-ui,
					-apple-system,
					BlinkMacSystemFont;
				padding-bottom: 1px;
				background-color: $bg-color;
			}

			li {
				display: flex;
				flex-flow: row;
				gap: 8px;
				align-items: center;

				button {
					all: unset;
					display: flex;
					flex-flow: row;
					align-items: center;
					flex-shrink: 1;
					gap: 8px;
					width: max-content;
					min-width: 0;
					max-width: 100%;
					cursor: pointer;
				}

				span {
					width: max-content;
					margin-left: 0px;

					&:first-child {
						min-width: max-content;
					}

					&:last-of-type {
						background-color: $accent-color;
						color: $text-color;
						border-radius: 8px;
						padding: 10px 12px;
					}
				}
			}
		}

		span {
			margin-left: calc(30px);
			width: 100%;
			display: flex;
			justify-content: center;
		}
	}

	h2 {
		font-size: 30px;
		font-weight: bold;
		margin-left: 30px;
	}
</style>
