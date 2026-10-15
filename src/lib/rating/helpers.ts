import { store } from "@/store.svelte";
import { RatingStep, RatingSystem } from "@/types";

/**
 * Used for scaling users 'actual' rating we store in db
 * into one we can show that takes into account their
 * settings for how they want stars displayed.
 * Only used for star ratings, not thumbs.
 */
export function toShowableRating(r?: number) {
	if (!r) {
		return 0;
	}
	if (
		!store.userSettings ||
		(!store.userSettings.ratingSystem && !store.userSettings.ratingStep)
	) {
		return Math.round(r);
	}
	if (store.userSettings.ratingSystem === RatingSystem.OutOf100) {
		return r * 10;
	}
	if (store.userSettings.ratingSystem === RatingSystem.OutOf5) {
		if (store.userSettings.ratingStep === RatingStep.Point5) {
			return Math.ceil((r / 2) * 2) / 2;
		}
		if (store.userSettings.ratingStep === RatingStep.Point1) {
			return Math.round((r / 2) * 10) / 10;
		}
		return Math.round(r / 2);
	}
	if (store.userSettings.ratingSystem === RatingSystem.OutOf10) {
		if (store.userSettings.ratingStep === RatingStep.Point5) {
			return Math.ceil(r * 2) / 2;
		}
		if (store.userSettings.ratingStep === RatingStep.Point1) {
			return r;
		}
		return Math.round(r);
	}
	return Math.round(r);
}

export function toWhichThumb(r?: number) {
	if (!r) {
		return;
	}
	const rr = Math.round(r);
	if (rr > 0 && rr <= 4) {
		return -1;
	} else if (r >= 4 && r <= 7) {
		return 0;
	} else if (r >= 8) {
		return 1;
	}
}

/**
 * Formats a public / popular media rating (e.g. from TMDB or IGDB)
 * to a 10-point scale with 1 decimal place.
 * Returns undefined if the rating is not set or <= 0.
 */
export function formatPublicRating(rating?: number): number | undefined {
	if (typeof rating !== "number" || rating <= 0) {
		return undefined;
	}
	// TMDB ratings are often stored as rating * 10 (e.g. 74 for 7.4), while some sources
	// might provide a 0-10 float. If > 10, scale down from 100 to 10.
	const val = rating > 10 ? rating / 10 : rating;
	const rounded = Math.round(val * 10) / 10;
	return rounded > 0 ? rounded : undefined;
}

/**
 * Formats vote count to human-readable string (e.g. 1520 -> "1.5k", 1200000 -> "1.2m").
 * Returns undefined if count is not provided or <= 0.
 */
export function formatVoteCount(count?: number): string | undefined {
	if (typeof count !== "number" || count <= 0) {
		return undefined;
	}
	if (count >= 1000000) {
		return `${(count / 1000000).toFixed(1)}m`;
	}
	if (count >= 1000) {
		return `${(count / 1000).toFixed(1)}k`;
	}
	return String(count);
}
