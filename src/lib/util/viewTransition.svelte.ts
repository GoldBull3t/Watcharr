/**
 * View Transition State Manager for Watcharr
 * Orchestrates shared element poster morphing and page transitions via the View Transitions API.
 */

class ViewTransitionManager {
	// Active transition key: "${type}-${id}" (e.g. "movie-550", "tv-1399", "game-1020")
	activeKey = $state<string | null>(null);

	// Navigation direction
	direction = $state<"forward" | "back">("forward");

	/**
	 * Sets the active media target for shared element transition.
	 */
	setActive(type: string | undefined, id: number | string | undefined) {
		if (!type || !id) {
			this.activeKey = null;
			return;
		}
		this.activeKey = `${type}-${id}`;
	}

	/**
	 * Checks if a specific media item is the active transition target.
	 */
	isActive(type: string | undefined, id: number | string | undefined): boolean {
		if (!this.activeKey || !type || !id) return false;
		return this.activeKey === `${type}-${id}`;
	}

	/**
	 * Clears the active transition state.
	 */
	clear() {
		this.activeKey = null;
	}
}

export const viewTransition = new ViewTransitionManager();
