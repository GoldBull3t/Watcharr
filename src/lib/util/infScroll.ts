import { page } from "$app/state";

export interface ToolTipOptions {
	/**
	 * Infinite scroll threshold.
	 * Distance from bottom in pixels to trigger prefetch.
	 * Defaults to 800px (or 75% of viewport height).
	 */
	threshold?: number;
	/**
	 * Ran when user scrolls near the end.
	 * This callback should load the new page of data.
	 */
	callback: () => Promise<void>;
}

/**
 * High-performance Infinite Scroll with predictive prefetching.
 *
 * Features:
 * - RAF (requestAnimationFrame) throttling to eliminate scroll event jank.
 * - Dynamic prefetch distance (triggers when ~75% through current viewport, before reaching absolute bottom).
 * - Concurrency guard preventing multiple duplicate fetch triggers.
 * - Accurate cross-browser scrollHeight calculations.
 */
export default function infScroll(opts: ToolTipOptions) {
	const { threshold = 800, callback } = opts;

	const startPagePathLower = page.url?.pathname?.toLowerCase();

	let isRunning = false;
	let rafId: number | null = null;

	const getEffectiveThreshold = () => {
		if (typeof window === "undefined") return threshold;
		return Math.max(threshold, Math.round(window.innerHeight * 0.75));
	};

	const isNearBottom = () => {
		if (typeof window === "undefined" || typeof document === "undefined") return false;
		const scrollPos = window.innerHeight + Math.round(window.scrollY);
		const scrollTarget = Math.max(
			document.documentElement.scrollHeight,
			document.body.offsetHeight,
		);
		return scrollPos + getEffectiveThreshold() >= scrollTarget;
	};

	const run = async () => {
		if (isRunning) return;
		if (isNearBottom()) {
			isRunning = true;
			try {
				await callback();
			} finally {
				isRunning = false;
			}
		}
	};

	const onScrollOrResize = () => {
		if (rafId !== null) return;
		rafId = window.requestAnimationFrame(() => {
			rafId = null;
			run();
		});
	};

	const addEvents = () => {
		window.addEventListener("scroll", onScrollOrResize, { passive: true });
		window.addEventListener("resize", onScrollOrResize, { passive: true });
	};

	const removeEvents = () => {
		if (rafId !== null) {
			window.cancelAnimationFrame(rafId);
			rafId = null;
		}
		window.removeEventListener("scroll", onScrollOrResize);
		window.removeEventListener("resize", onScrollOrResize);
	};

	/**
	 * Run after data is loaded to ensure that if the newly rendered
	 * cards still do not fill the screen (or user has high DPI),
	 * the next page is fetched without waiting for scroll.
	 */
	const dataLoaded = () => {
		setTimeout(() => {
			if (startPagePathLower === page.url?.pathname?.toLowerCase()) {
				run();
			}
		}, 180);
	};

	addEvents();

	return {
		isAtBottom: isNearBottom,
		run,
		dataLoaded,
		destroy: () => {
			removeEvents();
		},
	};
}
