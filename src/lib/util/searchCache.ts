/**
 * High-performance In-Memory & Session Search Cache with LRU eviction.
 *
 * Keeps recent search results cached for 5 minutes, eliminating
 * repetitive network requests to TMDB/IGDB and giving instantaneous
 * responses (0ms) when users backspace or retype queries.
 */

interface CacheEntry<T> {
	data: T;
	timestamp: number;
}

class SearchCache {
	private memoryCache = new Map<string, CacheEntry<unknown>>();
	private readonly maxEntries: number;
	private readonly ttlMs: number;

	constructor(maxEntries = 150, ttlMs = 5 * 60 * 1000) {
		this.maxEntries = maxEntries;
		this.ttlMs = ttlMs;
	}

	private normalizeKey(
		query: string,
		type?: string,
		preferMyList?: boolean,
		page = 1,
	): string {
		const cleanQuery = (query || "").trim().toLowerCase();
		const cleanType = (type || "multi").toLowerCase();
		const cleanList = Boolean(preferMyList);
		return `${cleanQuery}::${cleanType}::${cleanList}::${page}`;
	}

	/**
	 * Retrieve a cached search result if available and still fresh.
	 */
	get<T>(
		query: string,
		type?: string,
		preferMyList?: boolean,
		page = 1,
	): T | undefined {
		if (!query || !query.trim()) return undefined;

		const key = this.normalizeKey(query, type, preferMyList, page);
		const entry = this.memoryCache.get(key);

		if (!entry) {
			// Check sessionStorage fallback if available in browser
			if (typeof sessionStorage !== "undefined") {
				try {
					const raw = sessionStorage.getItem(`w_search_${key}`);
					if (raw) {
						const parsed: CacheEntry<T> = JSON.parse(raw);
						if (Date.now() - parsed.timestamp <= this.ttlMs) {
							this.memoryCache.set(key, parsed);
							return parsed.data;
						} else {
							sessionStorage.removeItem(`w_search_${key}`);
						}
					}
				} catch {}
			}
			return undefined;
		}

		// Check expiration
		if (Date.now() - entry.timestamp > this.ttlMs) {
			this.memoryCache.delete(key);
			return undefined;
		}

		// Refresh LRU position (delete and re-insert)
		this.memoryCache.delete(key);
		this.memoryCache.set(key, entry);

		return entry.data as T;
	}

	/**
	 * Cache a search response with automatic LRU eviction.
	 */
	set<T>(
		query: string,
		data: T,
		type?: string,
		preferMyList?: boolean,
		page = 1,
	): void {
		if (!query || !query.trim() || !data) return;

		const key = this.normalizeKey(query, type, preferMyList, page);

		// LRU eviction if full
		if (this.memoryCache.size >= this.maxEntries) {
			const oldestKey = this.memoryCache.keys().next().value;
			if (oldestKey) {
				this.memoryCache.delete(oldestKey);
			}
		}

		const entry: CacheEntry<T> = {
			data,
			timestamp: Date.now(),
		};

		this.memoryCache.set(key, entry);

		// Store first page in sessionStorage for instant return across navigations
		if (page === 1 && typeof sessionStorage !== "undefined") {
			try {
				sessionStorage.setItem(`w_search_${key}`, JSON.stringify(entry));
			} catch {}
		}
	}

	/**
	 * Check if a query is present in cache without moving LRU order.
	 */
	has(
		query: string,
		type?: string,
		preferMyList?: boolean,
		page = 1,
	): boolean {
		const key = this.normalizeKey(query, type, preferMyList, page);
		const entry = this.memoryCache.get(key);
		if (!entry) return false;
		return Date.now() - entry.timestamp <= this.ttlMs;
	}

	/**
	 * Invalidate all cached searches.
	 */
	clear(): void {
		this.memoryCache.clear();
		if (typeof sessionStorage !== "undefined") {
			try {
				const toRemove: string[] = [];
				for (let i = 0; i < sessionStorage.length; i++) {
					const k = sessionStorage.key(i);
					if (k && k.startsWith("w_search_")) {
						toRemove.push(k);
					}
				}
				toRemove.forEach((k) => sessionStorage.removeItem(k));
			} catch {}
		}
	}
}

export const searchCache = new SearchCache();

