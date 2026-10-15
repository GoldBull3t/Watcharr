import { baseURL } from "@/lib/util/api";
import { MediaTypeE, type Media } from "@/types";

export function getMediaLink(
	w: Media,
): "/" | `/movie/${string}` | `/tv/${string}` | `/game/${string}` {
	switch (w.type) {
		case MediaTypeE.tmdbMovie:
			return `/movie/${w.ids.tmdb}`;
		case MediaTypeE.tmdbShow:
			return `/tv/${w.ids.tmdb}`;
		case MediaTypeE.igdbGame:
			return `/game/${w.ids.igdb}`;
		default:
			return "/";
	}
}

export function getPosterImage(w: Media): string | undefined {
	if (w.poster?.path) {
		return `${baseURL}/${w.poster.path}`;
	}
	if (w.extPosterPath) {
		if (w.type === MediaTypeE.tmdbMovie || w.type === MediaTypeE.tmdbShow) {
			return w.watched
				? `${baseURL}/img${w.extPosterPath}`
				: `https://image.tmdb.org/t/p/w300${w.extPosterPath}`;
		} else if (w.type === MediaTypeE.igdbGame) {
			return `https://images.igdb.com/igdb/image/upload/t_cover_big/${w.extPosterPath}.jpg`;
		}
	}
	return undefined;
}

export function getReleaseYear(w: Media): string {
	const rawDate =
		(
			w as unknown as {
				release_date?: string;
				first_air_date?: string;
				date?: string;
			}
		).release_date ||
		(w as unknown as { first_air_date?: string }).first_air_date ||
		(w as unknown as { date?: string }).date;
	if (rawDate) {
		const d = new Date(rawDate);
		if (!isNaN(d.getFullYear())) {
			return String(d.getFullYear());
		}
	}
	return "";
}

