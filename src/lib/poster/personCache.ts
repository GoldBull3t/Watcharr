import { req } from "@/lib/util/api";
import type { Media, PersonCreditsResponse } from "@/types";

const creditsCache = new Map<number, Promise<Media[]>>();

/**
 * Fetches and caches person credits with user's watched status populated.
 */
export function getPersonCreditsWithWatched(
	personId: number,
): Promise<Media[]> {
	if (creditsCache.has(personId)) {
		return creditsCache.get(personId)!;
	}
	const promise = req
		.get<PersonCreditsResponse>(`/content/person/${personId}/credits`)
		.then((res) => {
			if (!res?.credits) return [];
			// Deduplicate by TMDB ID
			return res.credits.filter(
				(v, i, a) => a.findIndex((t) => t.ids?.tmdb === v.ids?.tmdb) === i,
			);
		})
		.catch((err) => {
			console.warn("getPersonCreditsWithWatched: Failed to fetch:", err);
			return [];
		});
	creditsCache.set(personId, promise);
	return promise;
}
