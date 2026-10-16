import type { TMDBRegions } from "@/types";
import { req } from "@/lib/util/api";

export interface RegionInfo {
	code: string;
	name: string;
	flag: string;
}

export const POPULAR_REGIONS: RegionInfo[] = [
	{ code: "PT", name: "Portugal", flag: "🇵🇹" },
	{ code: "BR", name: "Brasil", flag: "🇧🇷" },
	{ code: "ES", name: "Espanha", flag: "🇪🇸" },
	{ code: "US", name: "Estados Unidos", flag: "🇺🇸" },
	{ code: "GB", name: "Reino Unido", flag: "🇬🇧" },
	{ code: "FR", name: "França", flag: "🇫🇷" },
	{ code: "DE", name: "Alemanha", flag: "🇩🇪" },
	{ code: "IT", name: "Itália", flag: "🇮🇹" },
	{ code: "CA", name: "Canadá", flag: "🇨🇦" },
	{ code: "AU", name: "Austrália", flag: "🇦🇺" },
	{ code: "MX", name: "México", flag: "🇲🇽" },
	{ code: "AR", name: "Argentina", flag: "🇦🇷" },
	{ code: "JP", name: "Japão", flag: "🇯🇵" },
];

export function getCountryFlag(code: string): string {
	if (!code || code.length !== 2) return "🌐";
	const upper = code.toUpperCase();
	const first = 127397 + upper.charCodeAt(0);
	const second = 127397 + upper.charCodeAt(1);
	return String.fromCodePoint(first, second);
}

const countryNamesMap: Record<string, string> = {
	PT: "Portugal",
	BR: "Brasil",
	ES: "Espanha",
	US: "Estados Unidos",
	GB: "Reino Unido",
	FR: "França",
	DE: "Alemanha",
	IT: "Itália",
	CA: "Canadá",
	AU: "Austrália",
	MX: "México",
	AR: "Argentina",
	JP: "Japão",
};

export function getCountryName(code: string): string {
	if (!code) return "Desconhecido";
	const upper = code.toUpperCase();
	return countryNamesMap[upper] || upper;
}

let cachedRegions: RegionInfo[] | null = null;
let pendingRegionsPromise: Promise<RegionInfo[]> | null = null;

export async function getAllRegions(): Promise<RegionInfo[]> {
	if (cachedRegions) return cachedRegions;
	if (pendingRegionsPromise) return pendingRegionsPromise;

	pendingRegionsPromise = (async () => {
		try {
			const res = await req.get<TMDBRegions>("/content/regions");
			if (res?.results?.length) {
				const map = new Map<string, RegionInfo>();
				// Seed popular regions with Portuguese localized names
				for (const pop of POPULAR_REGIONS) {
					map.set(pop.code, pop);
				}
				for (const r of res.results) {
					if (!map.has(r.iso_3166_1)) {
						map.set(r.iso_3166_1, {
							code: r.iso_3166_1,
							name: r.english_name,
							flag: getCountryFlag(r.iso_3166_1),
						});
					}
				}
				cachedRegions = Array.from(map.values()).sort((a, b) =>
					a.name.localeCompare(b.name),
				);
				return cachedRegions;
			}
		} catch (err) {
			console.warn("watchProviders: Failed to load all regions", err);
		}
		cachedRegions = POPULAR_REGIONS;
		return cachedRegions;
	})();

	return pendingRegionsPromise;
}

