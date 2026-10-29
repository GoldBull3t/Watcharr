import type { SupportedMedia, Watched, WatchedStatus } from "@/types";

export type PosterExtraDetails = {
	rating: number | undefined;
	status: WatchedStatus | undefined;
	dateAdded?: string;
	dateModified?: string;
	/**
	 * Only for shows.
	 */
	lastWatched?: string;
	hoursPlayed?: number;
	platform?: string;
};

export function buildExtraDetails(
	t: SupportedMedia | undefined,
	w: Watched,
): PosterExtraDetails {
	const obj = {
		rating: w.rating,
		status: w.status,
		dateAdded: w.createdAt,
		dateModified: w.updatedAt,
		lastWatched: w.watchingSeason,
		hoursPlayed: w.hoursPlayed,
		platform: w.platform,
	} as PosterExtraDetails;
	return obj;
}
