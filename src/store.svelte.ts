import type {
	Filters,
	Follow,
	ImportedList,
	PrivateUser,
	ServerFeatures,
	Tag,
	Theme,
	UserSettings,
	WLDetailedViewOption,
} from "./types";
import type { Notification } from "./lib/util/notify";
import { browser } from "$app/environment";
import { toggleTheme } from "./lib/util/theme";

type ActiveSort = [string, string | undefined] | [];

export const defaultSort: ActiveSort = ["DATEADDED", "DOWN"];

interface Store {
	userInfo: PrivateUser | undefined;
	userSettings: UserSettings | undefined;
	notifications: Notification[];
	activeSort: ActiveSort;
	activeFilters: Filters;
	sortAndFiltersForQueryParams: object;
	appTheme: Theme;
	importedList:
		| {
				data: string;
				type:
					| "text-list"
					| "tmdb"
					| "movary"
					| "watcharr"
					| "myanimelist"
					| "ryot"
					| "todomovies"
					| "imdb"
					| "yamtrack"
					| "letterboxd";
		  }
		| undefined;
	parsedImportedList: ImportedList[] | undefined;
	searchQuery: string;
	serverFeatures: ServerFeatures | undefined;
	follows: Follow[];
	wlDetailedView: WLDetailedViewOption[];
	tags: Tag[];
}

/**
 * This is our actual (private) store.
 */
const _store: Store = $state({
	notifications: [],
	activeSort: defaultSort,
	activeFilters: { type: [], status: [] },
	appTheme: "system",
	sortAndFiltersForQueryParams: {},
	importedList: undefined,
	parsedImportedList: undefined,
	searchQuery: "",
	userInfo: undefined,
	userSettings: undefined,
	serverFeatures: undefined,
	follows: [],
	wlDetailedView: [],
	tags: [],
});

function computeSortAndFiltersForQueryParams(
	sort: ActiveSort,
	filters: Filters,
): Record<string, string> {
	try {
		const qp: Record<string, string> = {};
		if (sort?.length === 2) {
			qp.sort = sort[0];
			qp.sortDir = sort[1] === "UP" ? "asc" : "desc";
		}
		if (filters) {
			const t = filters?.type?.join(",");
			if (t) {
				qp["type"] = t;
			}
			const s = filters?.status?.join(",");
			if (s) {
				qp["status"] = s;
			}
			if (filters?.year) {
				qp["year"] = String(filters.year);
			}
			if (filters?.minRating) {
				qp["minRating"] = String(filters.minRating);
			}
			if (filters?.maxRating) {
				qp["maxRating"] = String(filters.maxRating);
			}
			if (filters?.query) {
				qp["query"] = filters.query;
			}
			if (filters?.genre) {
				qp["genre"] = filters.genre;
			}
						if (filters?.productionStatus) {
				qp["productionStatus"] = filters.productionStatus;
			}
			if (filters?.tags && filters.tags.length > 0) {
				qp["tags"] = filters.tags.join(",");
				if (filters.tagMode) {
					qp["tagMode"] = filters.tagMode;
				}
			}
		}
		return qp;
	} catch (err) {
		console.error("computeSortAndFiltersForQueryParams: Failed!", err);
		return {};
	}
}

/**
 * Expose store to app through getters/setters
 * to control what can and can't be accessed.
 * With setters we can easily and more reliably
 * save certain properties to localStorage when
 * they are updated.
 */
export const store = {
	get notifications() {
		return _store.notifications;
	},
	set notifications(v) {
		_store.notifications = v;
	},
	get activeSort() {
		return _store.activeSort;
	},
	set activeSort(v) {
		_store.activeSort = v;
		if (browser) {
			try {
				localStorage.setItem("activeFilter", JSON.stringify(v));
			} catch (e) {
				console.warn("Store: Failed to save activeSort:", e);
			}
		}
		console.debug("Store: Saved activeSort:", v);
	},
	get activeFilters() {
		return _store.activeFilters;
	},
	get hasActiveFilters(): boolean {
		return (
			this.activeFilters &&
			(this.activeFilters.status?.length > 0 ||
				this.activeFilters.type?.length > 0 ||
				Boolean(this.activeFilters.year) ||
				Boolean(this.activeFilters.minRating) ||
				Boolean(this.activeFilters.maxRating) ||
				Boolean(this.activeFilters.query) ||
				Boolean(this.activeFilters.genre) || Boolean(this.activeFilters.tags && this.activeFilters.tags.length > 0))
		);
	},
	get hasSecondaryFilters(): boolean {
		return Boolean(
			this.activeFilters &&
			(this.activeFilters.type?.length > 0 ||
				Boolean(this.activeFilters.year) ||
				Boolean(this.activeFilters.minRating) ||
				Boolean(this.activeFilters.maxRating) ||
				Boolean(this.activeFilters.genre) ||
				Boolean(this.activeFilters.productionStatus) || Boolean(this.activeFilters.tags && this.activeFilters.tags.length > 0)),
		);
	},
	get secondaryFilterCount(): number {
		if (!this.activeFilters) return 0;
		let count = 0;
		if (this.activeFilters.type && this.activeFilters.type.length > 0) count++;
		if (this.activeFilters.year) count++;
		if (this.activeFilters.minRating || this.activeFilters.maxRating) count++;
		if (this.activeFilters.genre) count++;
				if (this.activeFilters.productionStatus) count++;
		if (this.activeFilters.tags && this.activeFilters.tags.length > 0) count += this.activeFilters.tags.length;
		return count;
	},
	set activeFilters(v) {
		_store.activeFilters = {
			type: Array.isArray(v?.type) ? v.type : [],
			status: Array.isArray(v?.status) ? v.status : [],
			year: v?.year,
			minRating: v?.minRating,
			maxRating: v?.maxRating,
			query: v?.query,
						genre: v?.genre,
			productionStatus: v?.productionStatus,
			tags: Array.isArray(v?.tags) ? v.tags : [],
			tagMode: v?.tagMode || "and",
		};
		if (browser) {
			try {
				localStorage.setItem(
					"activeFilterReal",
					JSON.stringify(_store.activeFilters),
				);
			} catch (e) {
				console.warn("Store: Failed to save activeFilters:", e);
			}
		}
		console.debug("Store: Saved activeFilters:", _store.activeFilters);
	},
	/**
	 * Return our `activeSort` and `activeFilters` in an object
	 * that is in the correct format for our get watched page
	 * http requests.
	 */
	get sortAndFiltersForQueryParams() {
		return computeSortAndFiltersForQueryParams(
			_store.activeSort,
			_store.activeFilters,
		);
	},
	get appTheme() {
		return _store.appTheme;
	},
	/**
	 * Only set appTheme through toggleTheme() helper.
	 */
	set appTheme(v) {
		_store.appTheme = v;
		localStorage.setItem("theme", v);
		console.debug("Store: Saved appTheme:", v);
	},
	get importedList() {
		return _store.importedList;
	},
	set importedList(v) {
		_store.importedList = v;
	},
	get parsedImportedList() {
		return _store.parsedImportedList;
	},
	set parsedImportedList(v) {
		_store.parsedImportedList = v;
	},
	get searchQuery() {
		return _store.searchQuery;
	},
	set searchQuery(v) {
		_store.searchQuery = v;
	},
	get userInfo() {
		return _store.userInfo;
	},
	set userInfo(v) {
		_store.userInfo = v;
	},
	get userSettings() {
		return _store.userSettings;
	},
	set userSettings(v) {
		_store.userSettings = v;
	},
	get serverFeatures() {
		return _store.serverFeatures;
	},
	set serverFeatures(v) {
		_store.serverFeatures = v;
	},
	get follows() {
		return _store.follows;
	},
	set follows(v) {
		_store.follows = v;
	},
	get wlDetailedView() {
		return _store.wlDetailedView;
	},
	set wlDetailedView(v) {
		_store.wlDetailedView = v;
		if (browser) {
			try {
				if (v) {
					localStorage.setItem(
						"wlDetailedView",
						JSON.stringify(_store.wlDetailedView),
					);
					console.debug("Store: Saved wlDetailedView:", v);
				} else {
					localStorage.removeItem("wlDetailedView");
					console.debug("Store: Removed wlDetailedView");
				}
			} catch (e) {
				console.warn("Store: Failed to save wlDetailedView:", e);
			}
		}
	},
	get tags() {
		return _store.tags;
	},
	set tags(v) {
		_store.tags = v;
	},
};

/**
 * Reset everything in `store` back to default values.
 */
export const clearAllStores = () => {
	store.notifications = [];
	store.activeSort = defaultSort;
	store.appTheme = "system";
	store.importedList = undefined;
	store.parsedImportedList = undefined;
	store.searchQuery = "";
	store.userInfo = undefined;
	store.userSettings = undefined;
	store.serverFeatures = undefined;
	store.follows = [];
	store.wlDetailedView = [];
	store.tags = [];
	clearActiveFilters();
};

export const clearActiveFilters = () => {
	store.activeFilters = {
		type: [],
		status: [],
		year: undefined,
		minRating: undefined,
		maxRating: undefined,
		query: undefined,
		genre: undefined,
				productionStatus: undefined,
		tags: [],
		tagMode: "and",
	};
};

export const clearSecondaryFilters = () => {
	store.activeFilters = {
		...store.activeFilters,
		type: [],
		year: undefined,
		minRating: undefined,
		maxRating: undefined,
		genre: undefined,
				productionStatus: undefined,
		tags: [],
		tagMode: "and",
	};
};

if (browser) {
	rehydrateStore();
}

/**
 * Restore state from localStorage and apply values into
 * our `store`.
 * Rehydrates directly into `_store` (the real store)
 * to avoid the setters that would trigger a save right
 * after rehydrate.
 */
function rehydrateStore() {
	console.info("rehydrateStore: Running..");
	// Restore activeSort
	const raf = localStorage.getItem("activeFilter");
	if (raf) {
		try {
			_store.activeSort = JSON.parse(raf);
			console.debug(
				"rehydrateStore: Restored activeSort:",
				$state.snapshot(store.activeSort),
			);
		} catch (err) {
			console.warn(
				"rehydrateStore: Failed to parse activeFilter from localStorage, resetting",
				err,
			);
			localStorage.removeItem("activeFilter");
		}
	}
	// Restore activeFilters
	const filters = localStorage.getItem("activeFilterReal");
	if (filters) {
		try {
			const parsed = JSON.parse(filters);
			_store.activeFilters = {
				type: Array.isArray(parsed?.type) ? parsed.type : [],
				status: Array.isArray(parsed?.status) ? parsed.status : [],
				year: typeof parsed?.year === "number" ? parsed.year : undefined,
				minRating:
					typeof parsed?.minRating === "number" ? parsed.minRating : undefined,
				maxRating:
					typeof parsed?.maxRating === "number" ? parsed.maxRating : undefined,
				query: typeof parsed?.query === "string" ? parsed.query : undefined,
				genre: typeof parsed?.genre === "string" ? parsed.genre : undefined,
				productionStatus:
					typeof parsed?.productionStatus === "string"
						? parsed.productionStatus
						: undefined,
			};
			console.debug(
				"rehydrateStore: Restored activeFilters:",
				$state.snapshot(store.activeFilters),
			);
		} catch (err) {
			console.warn(
				"rehydrateStore: Failed to parse activeFilterReal from localStorage, resetting",
				err,
			);
			localStorage.removeItem("activeFilterReal");
		}
	}
	// Restore appTheme
	const theme = localStorage.getItem("theme") as Theme;
	if (theme) {
		_store.appTheme = theme;
		toggleTheme(theme, false);
		console.debug(
			"rehydrateStore: Restored appTheme:",
			$state.snapshot(store.appTheme),
		);
	} else {
		const defTheme: Theme = "system";
		_store.appTheme = defTheme;
		toggleTheme(defTheme, false);
		console.debug(
			"rehydrateStore: appTheme hydrated from system default (wont save):",
			defTheme,
		);
	}
	// Restore wlDetailedView
	const wlDetailedViewR = localStorage.getItem("wlDetailedView");
	if (wlDetailedViewR) {
		try {
			_store.wlDetailedView = JSON.parse(wlDetailedViewR);
			console.debug(
				"rehydrateStore: Restored wlDetailedView:",
				$state.snapshot(store.wlDetailedView),
			);
		} catch (err) {
			console.warn(
				"rehydrateStore: Failed to parse wlDetailedView from localStorage, resetting",
				err,
			);
			localStorage.removeItem("wlDetailedView");
		}
	}
	console.info("rehydrateStore: Done.");
}

export default store;
