import { clearAllStores } from "@/store.svelte";

/**
 * Helper to clear local data in client for logout
 * process. Logic is here so it's not duplciated,
 * this should help avoid forgetting to copy any
 * new logic to other places, which could break stuff.
 * Since this is reusable, technically unrelated logic
 * should not be included here (eg: redirecting to /login).
 */
export function clearWatcharrData() {
	localStorage.removeItem("token");
	clearAllStores();
	if (typeof window !== "undefined" && "caches" in window) {
		caches
			.keys()
			.then((keys) => {
				for (const key of keys) {
					if (key.includes("api-cache") || key.includes("local-images-cache")) {
						caches.delete(key);
					}
				}
			})
			.catch(() => {
				// ignore cache clear errors
			});
	}
}
