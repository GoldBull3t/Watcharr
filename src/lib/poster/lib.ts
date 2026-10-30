import type { Media, SupportedMedia, Watched, WatchedStatus } from "@/types";

export interface SeasonProgressStats {
	totalSeasons: number;
	finishedSeasons: number[];
	watchingSeasons: number[];
	plannedSeasons: number[];
	nextUpSeason?: number;
	nextUpStatus?: "FINISHED" | "WATCHING" | "PLANNED" | "UNWATCHED";
	percentage: number;
	displayText: string;
	seasons: {
		number: number;
		status: "FINISHED" | "WATCHING" | "PLANNED" | "UNWATCHED";
	}[];
}

export function calculateSeasonProgress(
	numberOfSeasons?: number,
	watchedSeasons?: { seasonNumber: number; status: WatchedStatus }[],
): SeasonProgressStats | null {
	const total =
		numberOfSeasons && numberOfSeasons > 0
			? numberOfSeasons
			: watchedSeasons && watchedSeasons.length > 0
				? Math.max(...watchedSeasons.map((s) => s.seasonNumber), 1)
				: 0;

	if (!total || total <= 0) return null;

	const statusMap = new Map<number, WatchedStatus>();
	if (watchedSeasons) {
		for (const ws of watchedSeasons) {
			if (ws.seasonNumber > 0) {
				statusMap.set(ws.seasonNumber, ws.status);
			}
		}
	}

	const finished: number[] = [];
	const watching: number[] = [];
	const planned: number[] = [];
	const seasonsList: SeasonProgressStats["seasons"] = [];

	for (let i = 1; i <= total; i++) {
		const st = statusMap.get(i);
		if (st === "FINISHED") {
			finished.push(i);
			seasonsList.push({ number: i, status: "FINISHED" });
		} else if (st === "WATCHING") {
			watching.push(i);
			seasonsList.push({ number: i, status: "WATCHING" });
		} else if (st === "PLANNED") {
			planned.push(i);
			seasonsList.push({ number: i, status: "PLANNED" });
		} else {
			seasonsList.push({ number: i, status: "UNWATCHED" });
		}
	}

	if (finished.length === 0 && watching.length === 0 && planned.length === 0) {
		return null;
	}

	let nextUpSeason: number | undefined;
	let nextUpStatus: SeasonProgressStats["nextUpStatus"];

	if (watching.length > 0) {
		nextUpSeason = watching[0];
		nextUpStatus = "WATCHING";
	} else if (planned.length > 0) {
		nextUpSeason = planned[0];
		nextUpStatus = "PLANNED";
	} else {
		const maxFinished = finished.length > 0 ? Math.max(...finished) : 0;
		if (maxFinished < total) {
			nextUpSeason = maxFinished + 1;
			nextUpStatus = "UNWATCHED";
		}
	}

	const percentage = Math.min(100, Math.round((finished.length / total) * 100));

	let displayText = `${finished.length}/${total} Temp.`;
	if (nextUpSeason) {
		if (nextUpStatus === "PLANNED") {
			displayText += ` (T${nextUpSeason} Planeada)`;
		} else if (nextUpStatus === "WATCHING") {
			displayText += ` (A ver T${nextUpSeason})`;
		}
	} else if (finished.length === total) {
		displayText = `Completa (${total}/${total} Temp.)`;
	}

	return {
		totalSeasons: total,
		finishedSeasons: finished,
		watchingSeasons: watching,
		plannedSeasons: planned,
		nextUpSeason,
		nextUpStatus,
		percentage,
		displayText,
		seasons: seasonsList,
	};
}

export type PosterExtraDetails = {
	rating: number | undefined;
	status: WatchedStatus | undefined;
	dateAdded?: string;
	dateModified?: string;
	/**
	 * Only for shows.
	 */
	lastWatched?: string;
	seasonProgress?: SeasonProgressStats | null;
	hoursPlayed?: number;
	platform?: string;
};

export function buildExtraDetails(
	t: SupportedMedia | undefined,
	w: Watched,
	media?: Media,
): PosterExtraDetails {
	const isTv = t === "tv" || media?.type === "tv";
	const seasonProgress = isTv
		? calculateSeasonProgress(media?.numberOfSeasons, w.watchedSeasons)
		: null;

	const obj = {
		rating: w.rating,
		status: w.status,
		dateAdded: w.createdAt,
		dateModified: w.updatedAt,
		lastWatched: w.watchingSeason,
		seasonProgress,
		hoursPlayed: w.hoursPlayed,
		platform: w.platform,
	} as PosterExtraDetails;
	return obj;
}
