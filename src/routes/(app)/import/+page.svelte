<!-- 
  /import is for getting the user to select
  the file they want to import and reading
  it. The data is set in a store for
  /import/process to process.
 -->

<script lang="ts">
	import { goto } from "$app/navigation";
	import DropFileButton from "@/lib/DropFileButton.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import { notify } from "@/lib/util/notify";
	import { store } from "@/store.svelte";
	import { onMount } from "svelte";
	import papa from "papaparse";
	import type {
		ImportedList,
		MovaryHistory,
		MovaryRatings,
		MovaryWatchlist,
		WatchedStatus,
		TodoMoviesExport,
		TagAddRequest,
		TodoMoviesCustomList,
		TodoMoviesMovie,
		WatchedEpisode,
		YamtrackRow,
		LetterboxdRow,
	} from "@/types";
	import Icon from "@/lib/Icon.svelte";
	import { resolve } from "$app/paths";
	import { baseURL } from "@/lib/util/api";

	let isLoading = $state(false);
	let isExporting = $state(false);

	function processFiles(
		files: FileList | null | undefined,
		type: "text-list" | "tmdb" | "imdb",
	) {
		try {
			console.log("processFiles", files);
			if (!files || files?.length <= 0) {
				console.error("processFiles", "No files to process!");
				notify({
					type: "error",
					text: "File not found in dropped items. Please try again or refresh.",
					time: 6000,
				});
				return;
			}
			isLoading = true;
			if (files.length > 1) {
				notify({
					type: "error",
					text: "Only one file at a time is supported. Continuing with the first.",
					time: 6000,
				});
			}
			// Currently only support for importing one file at a time
			const file = files[0];
			if (type === "text-list" && file.type !== "text/plain") {
				notify({
					type: "error",
					text: "Text list export must be a .txt file!",
				});
				isLoading = false;
				return;
			}
			if ((type === "tmdb" || type === "imdb") && file.type !== "text/csv") {
				notify({
					type: "error",
					text: `${type} export must be a .csv file!`,
				});
				isLoading = false;
				return;
			}
			const r = new FileReader();
			r.addEventListener(
				"load",
				() => {
					if (r.result) {
						store.importedList = {
							data: r.result.toString(),
							type,
						};
						goto(resolve("/import/process"));
					}
				},
				false,
			);
			r.readAsText(file);
		} catch (err) {
			isLoading = false;
			notify({ type: "error", text: "Failed to read file!" });
			console.error("import: Failed to read file!", err);
		}
	}

	async function readFile(fr: FileReader, file: File): Promise<string> {
		return new Promise((resolve, reject) => {
			const res = () => {
				fr.removeEventListener("load", res);
				fr.removeEventListener("error", rej);
				if (fr.result) {
					resolve(fr.result.toString());
				} else {
					reject("no result");
				}
			};
			const rej = () => {
				fr.removeEventListener("load", res);
				fr.removeEventListener("error", rej);
				reject();
			};
			fr.addEventListener("load", res);
			fr.addEventListener("error", rej);
			fr.readAsText(file);
		});
	}

	/**
	 * Process movary import files.
	 *
	 * Movary exports 3 different files:
	 *
	 *  - watchlist.csv = Planned movies.
	 *  - history.csv   = Watched movies, movie can be watched multiple times.
	 *  - ratings.csv   = Ratings for movies. One rating per movie.
	 *
	 * Export types explained better here:
	 * https://github.com/sbondCo/Watcharr/issues/332#issuecomment-1920662244
	 */
	async function processFilesMovary(files?: FileList | null) {
		try {
			console.log("processFilesMovary", files);
			if (!files || files?.length <= 0) {
				console.error("processFilesMovary", "No files to process!");
				notify({
					type: "error",
					text: "File not found in dropped items. Please try again or refresh.",
					time: 6000,
				});
				return;
			}
			if (files.length !== 3) {
				notify({
					type: "error",
					text: "You must select or drop 3 files: history.csv, ratings.csv and watchlist.csv.",
					time: 6000,
				});
				return;
			}
			isLoading = true;
			// Read file data into strings
			let history: string | undefined;
			let ratings: string | undefined;
			let watchlist: string | undefined;
			const r = new FileReader();
			for (let i = 0; i < files.length; i++) {
				const f = files[i];
				if (f.name === "history.csv") {
					history = await readFile(r, f);
				} else if (f.name === "ratings.csv") {
					ratings = await readFile(r, f);
				} else if (f.name === "watchlist.csv") {
					watchlist = await readFile(r, f);
				}
			}
			if (!history || !ratings || !watchlist) {
				notify({
					type: "error",
					text: "Failed to read history, ratings or watchlist. Ensure you have attached 3 files: history.csv, ratings.csv and watchlist.csv.",
					time: 6000,
				});
				isLoading = false;
				return;
			}
			console.log("loaded all files");
			// Convert csv strings into json
			const historyJson = papa.parse<MovaryHistory>(history.trim(), {
				header: true,
			});
			const ratingsJson = papa.parse<MovaryRatings>(ratings.trim(), {
				header: true,
			});
			const watchlistJson = papa.parse<MovaryWatchlist>(watchlist.trim(), {
				header: true,
			});
			// Build toImport array
			const toImport: ImportedList[] = [];
			// Add all history movies (watched). There can be multiple entries for each movie.
			for (let i = 0; i < historyJson.data.length; i++) {
				const h = historyJson.data[i];
				// Skip if no tmdb id.
				if (!h.tmdbId) {
					continue;
				}
				// Skip if already added. The first time it is added we get all info needed from other entries.
				if (toImport.filter((ti) => ti.tmdbId == Number(h.tmdbId)).length > 0) {
					continue;
				}
				const ratingsEntry = ratingsJson.data.find((r) => r.tmdbId == h.tmdbId);
				const t: ImportedList = {
					name: h.title,
					tmdbId: Number(h.tmdbId),
					status: "FINISHED",
					type: "movie", // movary only supports movies
					datesWatched: [],
					thoughts: "",
				};
				// Movie can be watched more than once, get all entries to store all watch dates.
				const allEntries = historyJson.data.filter(
					(he) => he.tmdbId === h.tmdbId,
				);
				for (let i = 0; i < allEntries.length; i++) {
					const e = allEntries[i];
					if (e.watchedAt) {
						t.datesWatched?.push(new Date(e.watchedAt));
					}
					if (e.comment) {
						t.thoughts += e.comment + "\n";
					}
				}
				if (h.year) {
					t.year = Number(h.year);
				}
				if (ratingsEntry && ratingsEntry?.userRating) {
					t.rating = Number(ratingsEntry.userRating);
				}
				toImport.push(t);
			}
			// Add all watchlist movies (planned).
			for (let i = 0; i < watchlistJson.data.length; i++) {
				const wl = watchlistJson.data[i];
				const existing = toImport.find((ti) => ti.tmdbId == Number(wl.tmdbId));
				// If already exists in toImport, simply update status to PLANNED.
				// The movie must have been completed in past, but added back to
				// the users movary watch list as they are planning to watch it again.
				if (existing) {
					existing.status = "PLANNED";
					continue;
				}
				toImport.push({
					name: wl.title,
					tmdbId: Number(wl.tmdbId),
					status: "PLANNED",
					type: "movie", // movary only supports movies
				});
			}
			console.log("toImport:", toImport);
			store.importedList = {
				data: JSON.stringify(toImport),
				type: "movary",
			};
			goto(resolve("/import/process"));
		} catch (err) {
			isLoading = false;
			notify({ type: "error", text: "Failed to read files!" });
			console.error("import: Failed to read files!", err);
		}
	}

	async function processWatcharrFile(files?: FileList | null) {
		try {
			console.log("processWatcharrFile", files);
			if (!files || files?.length <= 0) {
				console.error("processWatcharrFile", "No files to process!");
				notify({
					type: "error",
					text: "File not found in dropped items. Please try again or refresh.",
					time: 6000,
				});
				return;
			}
			isLoading = true;
			if (files.length > 1) {
				notify({
					type: "error",
					text: "Only one file at a time is supported. Continuing with the first.",
					time: 6000,
				});
			}
			// Currently only support for importing one file at a time
			const file = files[0];
			if (file.type !== "application/json") {
				notify({
					type: "error",
					text: "Must be a Watcharr JSON export file",
				});
				isLoading = false;
				return;
			}
			// Build toImport array
			const toImport: ImportedList[] = [];
			const fileText = await readFile(new FileReader(), file);
			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			const jsonData = JSON.parse(fileText) as any[];
			let invalidStructureErrorOccurred = false;
			let processedGame = false;
			for (const v of jsonData) {
				const t: ImportedList = {
					rating: v.rating,
					status: v.status,
					thoughts: v.thoughts,
					activity: v.activity,

					watchedEpisodes: v.watchedEpisodes,
					watchedSeasons: v.watchedSeasons,
				};
				if (v.content) {
					t.tmdbId = v.content.tmdbId;
					t.name = v.content.title;
					t.year = new Date(v.content.release_date)?.getFullYear();
					t.type = v.content.type;
				} else if (v.game) {
					t.igdbId = v.game.igdbId;
					t.name = v.game.name;
					t.year = new Date(v.game.releaseDate)?.getFullYear();
					t.type = "game";
					processedGame = true;
				} else {
					console.error("processWatcharrFile: Went over invalid item.", v);
					if (!invalidStructureErrorOccurred) {
						notify({
							type: "error",
							text: "Item(s) in export has an invalid structure! Look in console for more details.",
						});
						// So we don't spam it...
						invalidStructureErrorOccurred = true;
					}
					continue;
				}
				toImport.push(t);
			}
			if (processedGame && !store.serverFeatures?.games) {
				// We know the server hasnt told us it supports games yet
				// for this session, so alert the user of this before they start
				// importing.
				notify({
					type: "error",
					text:
						"It looks like this server doesn't support 'Games'! " +
						"The server should be configured properly before " +
						"proceeding otherwise none of your games will be imported!",
					time: Infinity,
				});
				// We still allow the user to continue (BECAUSE EIN CUSTOMER ISH ALWAYS REIGHT!!!!
				// or maybe they don't care about games, or maybe they configured
				// it in another tab and know for a fact that it should work.)
			}
			console.log("toImport:", toImport);
			store.importedList = {
				data: JSON.stringify(toImport),
				type: "watcharr",
			};
			goto(resolve("/import/process"));
		} catch (err) {
			isLoading = false;
			notify({ type: "error", text: "Failed to read file!" });
			console.error("import: Failed to read file!", err);
		}
	}

	function processFilesMyAnimeList(files?: FileList | null) {
		try {
			console.log("processFilesMyAnimeList", files);
			if (!files || files?.length <= 0) {
				console.error("processFilesMyAnimeList", "No files to process!");
				notify({
					type: "error",
					text: "File not found in dropped items. Please try again or refresh.",
					time: 6000,
				});
				return;
			}
			isLoading = true;
			if (files.length > 1) {
				notify({
					type: "error",
					text: "Only one file at a time is supported. Continuing with the first.",
					time: 6000,
				});
			}
			// Currently only support for importing one file at a time
			const file = files[0];
			if (file.type !== "text/xml") {
				notify({
					type: "error",
					text: "Your MyAnimeList export should be a xml file.",
				});
				isLoading = false;
				return;
			}
			const r = new FileReader();
			r.addEventListener(
				"load",
				() => {
					if (r.result) {
						store.importedList = {
							data: r.result.toString(),
							type: "myanimelist",
						};
						goto(resolve("/import/process"));
					}
				},
				false,
			);
			r.readAsText(file);
		} catch (err) {
			isLoading = false;
			notify({ type: "error", text: "Failed to read file!" });
			console.error("import: Failed to read file!", err);
		}
	}

	async function processRyotFile(files?: FileList | null) {
		try {
			console.log("processRyotFile", files);
			if (!files || files?.length <= 0) {
				console.error("processRyotFile", "No files to process!");
				notify({
					type: "error",
					text: "File not found in dropped items. Please try again or refresh.",
					time: 6000,
				});
				return;
			}
			isLoading = true;
			if (files.length > 1) {
				notify({
					type: "error",
					text: "Only one file at a time is supported. Continuing with the first.",
					time: 6000,
				});
			}

			// Currently only support for importing one file at a time
			const file = files[0];
			if (file.type !== "application/json") {
				notify({
					type: "error",
					text: "Must be a Ryot JSON export file",
				});
				isLoading = false;
				return;
			}

			// Build toImport array
			const toImport: ImportedList[] = [];
			const fileText = await readFile(new FileReader(), file);
			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			const jsonData = JSON.parse(fileText)["metadata"] as any[];
			for (const v of jsonData) {
				if (
					!v.source_id ||
					!v.identifier ||
					!(v.lot == "show" || v.lot == "movie")
				) {
					notify({
						type: "error",
						text: "Item in export either has no title, TMDB identifier or is not a movie/tv show! Look in console for more details.",
					});
					console.error(
						"Can't add export item to import table! It has title, TMDB identifier or is not a movie/tv show! Item:",
						v,
					);
					continue;
				}

				// Define the main general status of the movie/show
				// In Ryot, it can be marked as multiple of the following, so choose the most relevant
				const statusRanks: [string, WatchedStatus][] = [
					["", "DROPPED"],
					["Watchlist", "PLANNED"],
					["Monitoring", "PLANNED"],
					["In Progress", "WATCHING"],
					["Completed", "FINISHED"],
				];
				let rank = 0;
				for (const s of v.collections) {
					rank = Math.max(
						rank,
						statusRanks.findIndex((pair) => pair[0] == s),
					);
				}

				/**
				 * Ryot ratings are scored out of 100,
				 * scale this down to fit with watcharrs
				 * ratings that are out of 10.
				 */
				const validifyRating = (ryotRating: number) => {
					const r = Number(ryotRating);
					if (isNaN(r)) {
						return 0;
					}
					return Math.floor(r / 10);
				};

				const t: ImportedList = {
					tmdbId: Number(v.identifier),
					name: v.source_id,
					type: v.lot === "show" ? "tv" : v.lot,
					status: statusRanks[rank][1],

					// In Ryot, shows can have one review for each episode - Not supported in Watcharr
					// Will ignore the episodes' reviews
					thoughts:
						v.lot === "movie" && v.reviews?.length
							? v.reviews[0].review?.text
							: "",

					rating: v.reviews?.length
						? validifyRating(Number(v.reviews[0].rating))
						: undefined,

					datesWatched:
						v.lot === "movie" && v.seen_history?.length
							? // eslint-disable-next-line @typescript-eslint/no-explicit-any
								v.seen_history.map((seen: any) => new Date(seen.ended_on))
							: [],

					// Episode ratings are on a separate field: "reviews"
					watchedEpisodes:
						v.lot === "show"
							? // eslint-disable-next-line @typescript-eslint/no-explicit-any
								v.seen_history?.map((episode: any) => ({
									status: episode.progress === "100" ? "FINISHED" : "WATCHING",

									// Linear :( search the reviews for a match
									rating:
										validifyRating(
											Number(
												(
													v.reviews?.find(
														// eslint-disable-next-line @typescript-eslint/no-explicit-any
														(review: any) =>
															review.show_season_number ===
																episode.show_season_number &&
															review.show_episode_number ===
																episode.show_episode_number,
													) || {}
												)?.rating,
											),
										) || null,

									seasonNumber: episode.show_season_number,
									episodeNumber: episode.show_episode_number,
									createdAt: episode.ended_on
										? new Date(episode.ended_on)
										: undefined,
								}))
							: undefined,
				};
				toImport.push(t);
			}
			console.log("toImport:", toImport);
			store.importedList = {
				data: JSON.stringify(toImport),
				type: "ryot",
			};
			goto(resolve("/import/process"));
		} catch (err) {
			isLoading = false;
			notify({ type: "error", text: "Failed to read file!" });
			console.error("import: Failed to read file!", err);
		}
	}

	async function processTodoMoviesFile(files?: FileList | null) {
		try {
			console.log("processFilesTodoMovies", files);
			if (!files || files?.length <= 0) {
				console.error("processFilesTodoMovies", "No files to process!");
				notify({
					type: "error",
					text: "File not found in dropped items. Please try again or refresh.",
					time: 6000,
				});
				return;
			}
			isLoading = true;
			if (files.length > 1) {
				notify({
					type: "error",
					text: "Only one file at a time is supported. Continuing with the first.",
					time: 6000,
				});
			}

			// Currently only support for importing one file at a time
			const file = files[0];
			if (file.type !== "" && !file.name.endsWith(".todomovieslist")) {
				notify({
					type: "error",
					text: "Must be a TodoMovies backup file (.todomovieslist)",
				});
				isLoading = false;
				return;
			}

			// Read file data into strings
			let exportTodoMoviesStr: string | undefined;
			const r = new FileReader();
			exportTodoMoviesStr = await readFile(r, file);
			if (!exportTodoMoviesStr) {
				notify({
					type: "error",
					text: "Failed to read export file. Ensure you have attached the correct file.",
					time: 6000,
				});
				isLoading = false;
				return;
			}
			console.log("Loaded file");

			const exportTodoMovies: TodoMoviesExport =
				JSON.parse(exportTodoMoviesStr);

			console.log("exportTodoMovies:", exportTodoMovies);

			const movieList: TodoMoviesMovie[] = exportTodoMovies.Movie;
			const customLists: TodoMoviesCustomList[] = exportTodoMovies.MovieList;

			// Build toImport array
			const toImport: ImportedList[] = [];

			// Convert the timestamp to seconds (NSDate uses seconds since 2001-01-01 00:00:00 UTC)
			const referenceDate = new Date(2001, 0, 1);

			// Add all history movies. There is only one entry for every movie.
			// Movies which have already been watched but which have been added back in planning are also included, as this could also have ratings and comments.
			for (let i = 0; i < movieList.length; i++) {
				const h = movieList[i];
				// Skip if no tmdb id.
				if (!h.Attrs.tmdbID) {
					continue;
				}
				// Skip if already added. The first time it is added we get all info needed from other entries.
				if (
					toImport.filter((ti) => ti.tmdbId == Number(h.Attrs.tmdbID)).length >
					0
				) {
					continue;
				}

				const nsInsertionDate = h.Attrs.insertionDate.Value;
				const date = new Date(referenceDate.getTime() + nsInsertionDate * 1000);
				const tagsIds = h.Rels.lists.Items;
				const tags = tagsIds.map((tagId) => {
					const tag = customLists.find((list) => list.ObjectID == tagId);
					return {
						name: "TodoMovies list: " + tag?.Attrs.name,
						color: "#000000",
						bgColor: tag?.Attrs.colorInHex,
					} as TagAddRequest;
				});
				const t: ImportedList = {
					name: h.Attrs.title,
					tmdbId: Number(h.Attrs.tmdbID),
					status: h.Attrs.isWatched == 1 ? "FINISHED" : "PLANNED",
					type: "movie", // TodoMovies only supports movies
					datesWatched: [date], // use activities instead
					thoughts: "", // no comments in TodoMovies
					rating: h.Attrs.myScore,
					tags: tags,
				};
				toImport.push(t);
			}

			console.log("toImport:", toImport);
			store.importedList = {
				data: JSON.stringify(toImport),
				type: "todomovies",
			};

			goto(resolve("/import/process"));
		} catch (err) {
			isLoading = false;
			notify({ type: "error", text: "Failed to read files!" });
			console.error("import: Failed to read files!", err);
		}
	}

	async function processYamtrackFile(files?: FileList | null) {
		try {
			console.log("processYamtrackFile", files);
			if (!files || files?.length <= 0) {
				console.error("processYamtrackFile", "No files to process!");
				notify({
					type: "error",
					text: "File not found in dropped items. Please try again or refresh.",
					time: 6000,
				});
				return;
			}
			isLoading = true;
			if (files.length > 1) {
				notify({
					type: "error",
					text: "Only one file at a time is supported. Continuing with the first.",
					time: 6000,
				});
			}

			const file = files[0];
			if (
				file.type !== "text/csv" &&
				!file.name.toLowerCase().endsWith(".csv")
			) {
				notify({
					type: "error",
					text: "Yamtrack export must be a .csv file!",
				});
				isLoading = false;
				return;
			}

			const csvText = await readFile(new FileReader(), file);
			if (!csvText || !csvText.trim()) {
				notify({
					type: "error",
					text: "The selected file is empty!",
				});
				isLoading = false;
				return;
			}

			const parsed = papa.parse<YamtrackRow>(csvText.trim(), {
				header: true,
				skipEmptyLines: true,
			});

			if (!parsed.data || parsed.data.length === 0) {
				notify({
					type: "error",
					text: "Could not find any items in the Yamtrack CSV file!",
				});
				isLoading = false;
				return;
			}

			const toImport: ImportedList[] = [];

			for (let i = 0; i < parsed.data.length; i++) {
				const row = parsed.data[i];
				if (!row || (!row.title && !row.media_id)) {
					continue;
				}

				const rawMediaId = String(row.media_id || "").trim();
				const source = String(row.source || "tmdb")
					.toLowerCase()
					.trim();
				const rawType = String(row.media_type || "")
					.toLowerCase()
					.trim();
				const type: "movie" | "tv" | "game" =
					rawType === "movie" || rawType === "film"
						? "movie"
						: rawType === "game" || rawType === "videogame"
							? "game"
							: "tv";

				const tmdbId =
					source === "tmdb" && rawMediaId && !isNaN(Number(rawMediaId))
						? Number(rawMediaId)
						: undefined;
				const imdbId = source === "imdb" && rawMediaId ? rawMediaId : undefined;
				const igdbId =
					source === "igdb" && rawMediaId && !isNaN(Number(rawMediaId))
						? Number(rawMediaId)
						: undefined;

				// Status mapping
				const rawStatus = String(row.status || "")
					.toLowerCase()
					.trim();
				let status: WatchedStatus = "FINISHED";
				if (
					rawStatus.includes("complete") ||
					rawStatus.includes("finish") ||
					rawStatus.includes("watched")
				) {
					status = "FINISHED";
				} else if (
					rawStatus.includes("watching") ||
					rawStatus.includes("progress") ||
					rawStatus.includes("current")
				) {
					status = "WATCHING";
				} else if (
					rawStatus.includes("plan") ||
					rawStatus.includes("want") ||
					rawStatus.includes("queue") ||
					rawStatus.includes("list")
				) {
					status = "PLANNED";
				} else if (rawStatus.includes("hold") || rawStatus.includes("pause")) {
					status = "HOLD";
				} else if (
					rawStatus.includes("drop") ||
					rawStatus.includes("cancel") ||
					rawStatus.includes("abandon")
				) {
					status = "DROPPED";
				}

				// Rating mapping (Yamtrack can score 1-10 or 1-100)
				let rating: number | undefined = undefined;
				if (
					typeof row.score !== "undefined" &&
					row.score !== null &&
					String(row.score).trim() !== ""
				) {
					const num = Number(row.score);
					if (!isNaN(num) && num > 0) {
						rating =
							num > 10 ? Math.min(10, Math.round(num / 10)) : Math.min(10, num);
					}
				}

				// Dates
				const datesWatched: Date[] = [];
				let ratingCustomDate: Date | undefined = undefined;
				if (row.end_date) {
					const d = new Date(row.end_date);
					if (!isNaN(d.getTime())) {
						datesWatched.push(d);
						ratingCustomDate = d;
					}
				} else if (row.start_date) {
					const d = new Date(row.start_date);
					if (!isNaN(d.getTime())) {
						ratingCustomDate = d;
					}
				} else if (row.created_at) {
					const d = new Date(row.created_at);
					if (!isNaN(d.getTime())) {
						ratingCustomDate = d;
					}
				}

				// Check if item already exists in toImport (e.g. from multiple episode/watch rows)
				const existing = toImport.find((item) => {
					if (tmdbId && item.tmdbId === tmdbId && item.type === type)
						return true;
					if (imdbId && item.imdbId === imdbId) return true;
					if (igdbId && item.igdbId === igdbId) return true;
					if (
						!tmdbId &&
						!imdbId &&
						!igdbId &&
						item.name === row.title &&
						item.type === type
					)
						return true;
					return false;
				});

				const seasonNum = row.season_number
					? Number(row.season_number)
					: undefined;
				const epNum = row.episode_number
					? Number(row.episode_number)
					: undefined;

				if (existing) {
					// Merge additional watch date
					if (datesWatched.length > 0) {
						if (!existing.datesWatched) existing.datesWatched = [];
						existing.datesWatched.push(...datesWatched);
					}
					// Update rating if not set yet
					if (!existing.rating && rating) {
						existing.rating = rating;
					}
					// Append notes if any
					if (
						row.notes?.trim() &&
						!existing.thoughts?.includes(row.notes.trim())
					) {
						existing.thoughts = existing.thoughts
							? `${existing.thoughts}\n${row.notes.trim()}`
							: row.notes.trim();
					}
					// Episode tracking
					if (seasonNum && epNum) {
						if (!existing.watchedEpisodes) existing.watchedEpisodes = [];
						if (
							!existing.watchedEpisodes.some(
								(e) =>
									e.seasonNumber === seasonNum && e.episodeNumber === epNum,
							)
						) {
							existing.watchedEpisodes.push({
								seasonNumber: seasonNum,
								episodeNumber: epNum,
								status: "FINISHED",
							} as WatchedEpisode);
						}
					}
					continue;
				}

				// New item
				const t: ImportedList = {
					name: row.title?.trim() || `Media ${rawMediaId}`,
					type,
					status,
					thoughts: row.notes?.trim() || "",
				};

				if (tmdbId) t.tmdbId = tmdbId;
				if (imdbId) t.imdbId = imdbId;
				if (igdbId) t.igdbId = igdbId;
				if (rating) t.rating = rating;
				if (datesWatched.length > 0) t.datesWatched = datesWatched;
				if (ratingCustomDate) t.ratingCustomDate = ratingCustomDate;

				if (seasonNum && epNum) {
					t.watchedEpisodes = [
						{
							seasonNumber: seasonNum,
							episodeNumber: epNum,
							status: "FINISHED",
						} as WatchedEpisode,
					];
				}

				toImport.push(t);
			}

			if (toImport.length === 0) {
				notify({
					type: "error",
					text: "No valid media items found in the Yamtrack CSV file!",
				});
				isLoading = false;
				return;
			}

			console.log("Yamtrack toImport:", toImport);
			store.importedList = {
				data: JSON.stringify(toImport),
				type: "yamtrack",
			};

			goto(resolve("/import/process"));
		} catch (err) {
			isLoading = false;
			notify({ type: "error", text: "Failed to read Yamtrack file!" });
			console.error("import: Failed to read Yamtrack file!", err);
		}
	}

	async function processLetterboxdFile(files?: FileList | null) {
		try {
			console.log("processLetterboxdFile", files);
			if (!files || files.length <= 0) {
				notify({
					type: "error",
					text: "No files dropped. Please select one or more Letterboxd CSV files.",
					time: 5000,
				});
				return;
			}
			isLoading = true;

			// Map by `${title.toLowerCase()}_${year}`
			const itemsMap = new Map<string, ImportedList>();

			for (let f = 0; f < files.length; f++) {
				const file = files[f];
				const fileName = file.name.toLowerCase();
				if (!fileName.endsWith(".csv") && file.type !== "text/csv") {
					continue;
				}

				const isWatchlist = fileName.includes("watchlist");

				const text = await readFile(new FileReader(), file);
				if (!text || !text.trim()) continue;

				const parsed = papa.parse<LetterboxdRow>(text.trim(), {
					header: true,
					skipEmptyLines: true,
				});

				if (!parsed.data || parsed.data.length === 0) continue;

				for (let i = 0; i < parsed.data.length; i++) {
					const row = parsed.data[i];
					if (!row || !row.Name) continue;

					const title = String(row.Name).trim();
					let year: number | undefined = undefined;
					if (row.Year) {
						const y = Number(row.Year);
						if (!isNaN(y) && y > 1880 && y < 2100) {
							year = y;
						}
					}

					const key = `${title.toLowerCase()}_${year || ""}`;

					// Default status: watchlist.csv -> PLANNED, others -> FINISHED
					let status: WatchedStatus = isWatchlist ? "PLANNED" : "FINISHED";

					// Rating: Letterboxd is 0.5 to 5.0 -> multiply by 2 for Watcharr's 1-10 scale
					let rating: number | undefined = undefined;
					if (
						typeof row.Rating !== "undefined" &&
						row.Rating !== null &&
						String(row.Rating).trim() !== ""
					) {
						const r = Number(row.Rating);
						if (!isNaN(r) && r > 0) {
							rating = Math.min(10, Math.round(r * 2 * 10) / 10);
						}
					}

					// Dates
					const datesWatched: Date[] = [];
					let ratingCustomDate: Date | undefined = undefined;
					const rawDate = row["Watched Date"] || row.Date;
					if (rawDate) {
						const d = new Date(rawDate);
						if (!isNaN(d.getTime())) {
							datesWatched.push(d);
							ratingCustomDate = d;
						}
					}

					// Tags
					const tags: TagAddRequest[] = [];
					if (row.Tags && typeof row.Tags === "string" && row.Tags.trim()) {
						const tagNames = row.Tags.split(",")
							.map((s) => s.trim())
							.filter(Boolean);
						for (const tn of tagNames) {
							tags.push({ name: tn, color: "#ffffff", bgColor: "#e11d48" });
						}
					}

					const existing = itemsMap.get(key);
					if (existing) {
						// Merge
						if (status === "FINISHED") {
							existing.status = "FINISHED";
						}
						if (!existing.rating && rating) {
							existing.rating = rating;
						}
						if (datesWatched.length > 0) {
							if (!existing.datesWatched) existing.datesWatched = [];
							existing.datesWatched.push(...datesWatched);
						}
						if (ratingCustomDate && !existing.ratingCustomDate) {
							existing.ratingCustomDate = ratingCustomDate;
						}
						if (tags.length > 0) {
							if (!existing.tags) existing.tags = [];
							for (const t of tags) {
								if (
									!existing.tags.some(
										(et) => et.name.toLowerCase() === t.name.toLowerCase(),
									)
								) {
									existing.tags.push(t);
								}
							}
						}
					} else {
						const item: ImportedList = {
							name: title,
							year,
							type: "movie",
							status,
							rating,
							datesWatched: datesWatched.length > 0 ? datesWatched : undefined,
							ratingCustomDate,
							tags: tags.length > 0 ? tags : undefined,
						};
						itemsMap.set(key, item);
					}
				}
			}

			const toImport = Array.from(itemsMap.values());
			if (toImport.length === 0) {
				notify({
					type: "error",
					text: "Could not find any valid movies in the selected Letterboxd file(s)!",
					time: 5000,
				});
				isLoading = false;
				return;
			}

			console.log("Letterboxd toImport count:", toImport.length);
			store.importedList = {
				data: JSON.stringify(toImport),
				type: "letterboxd",
			};

			goto(resolve("/import/process"));
		} catch (err) {
			isLoading = false;
			notify({ type: "error", text: "Failed to process Letterboxd file(s)!" });
			console.error("import: Failed to process Letterboxd:", err);
		}
	}

	async function downloadExport(format: "json" | "csv") {
		try {
			isExporting = true;
			const nid = notify({
				type: "loading",
				text: `Generating ${format.toUpperCase()} export...`,
			});

			const token = localStorage.getItem("token") || "";
			const res = await fetch(`${baseURL}/watched/export?format=${format}`, {
				headers: {
					Authorization: token,
				},
			});

			if (!res.ok) {
				throw new Error(`Export failed with status: ${res.status}`);
			}

			const blob = await res.blob();
			const url = window.URL.createObjectURL(blob);
			const a = document.createElement("a");
			a.href = url;
			const dateStr = new Date().toISOString().split("T")[0];
			a.download = `watcharr_export_${dateStr}.${format}`;
			document.body.appendChild(a);
			a.click();
			window.URL.revokeObjectURL(url);
			a.remove();

			notify({
				id: nid,
				type: "success",
				text: `Successfully exported library as .${format}!`,
				time: 4000,
			});
		} catch (err) {
			console.error("downloadExport: Failed!", err);
			notify({
				type: "error",
				text: "Failed to export your library. Please try again.",
				time: 6000,
			});
		} finally {
			isExporting = false;
		}
	}

	onMount(() => {
		if (!localStorage.getItem("token")) {
			goto(resolve("/login"));
		}
	});
</script>

<div class="content">
	<div class="inner">
		<div class="page-header">
			<div class="header-titles">
				<h2>Import Your Watchlist</h2>
				<p class="subtitle">
					Seamlessly migrate your movies, TV shows, and watch history from
					external services or backups.
				</p>
			</div>
			<span class="beta-badge">Beta</span>
		</div>
		<div class="big-btns">
			{#if isLoading}
				<div class="spinner-wrap">
					<Spinner />
				</div>
			{:else}
				<DropFileButton
					text="Watcharr Export"
					filesSelected={(f) => processWatcharrFile(f)}
				/>

				<DropFileButton
					text=".txt list"
					filesSelected={(f) => processFiles(f, "text-list")}
				/>

				<DropFileButton
					icon="themoviedb"
					text="TMDb"
					filesSelected={(f) => processFiles(f, "tmdb")}
				/>

				<button
					class="plain trakt-import-btn"
					onclick={() => goto(resolve("/import/trakt"))}
				>
					<Icon i="trakt" wh="100%" />
					<h4 class="norm">Trakt Import</h4>
				</button>

				<DropFileButton
					icon="imdb"
					text="IMDb"
					filesSelected={(f) => processFiles(f, "imdb")}
				/>

				<DropFileButton
					icon="movary"
					text="Movary Exports"
					filesSelected={(f) => processFilesMovary(f)}
					allowSelectMultipleFiles
				/>

				<DropFileButton
					icon="myanimelist"
					text="MyAnimeList Export"
					filesSelected={(f) => processFilesMyAnimeList(f)}
				/>

				<DropFileButton
					icon="ryot"
					text="Ryot Exports"
					filesSelected={(f) => processRyotFile(f)}
				/>

				<DropFileButton
					icon="todomovies"
					text="TodoMovies"
					filesSelected={(f) => processTodoMoviesFile(f)}
				/>

				<DropFileButton
					icon="yamtrack"
					text="Yamtrack CSV"
					filesSelected={(f) => processYamtrackFile(f)}
				/>

				<DropFileButton
					icon="letterboxd"
					text="Letterboxd CSV"
					filesSelected={(f) => processLetterboxdFile(f)}
					allowSelectMultipleFiles
				/>
			{/if}
		</div>

		<!-- Export section -->
		<div class="export-section">
			<div class="export-header">
				<div class="export-titles">
					<div class="title-row">
						<Icon i="download" wh={22} />
						<h3>Export Your Library</h3>
					</div>
					<p class="subtitle">
						Download a complete backup of your watched media, reviews, dates,
						ratings, and tags.
					</p>
				</div>
			</div>
			<div class="export-actions">
				<button
					class="export-btn primary"
					disabled={isExporting}
					onclick={() => downloadExport("json")}
				>
					<Icon i="download" wh={18} />
					<span>Export as JSON</span>
					<span class="btn-subtext">Full backup (Watcharr format)</span>
				</button>
				<button
					class="export-btn secondary"
					disabled={isExporting}
					onclick={() => downloadExport("csv")}
				>
					<Icon i="download" wh={18} />
					<span>Export as CSV</span>
					<span class="btn-subtext">Spreadsheet compatible</span>
				</button>
			</div>
		</div>
	</div>
</div>

<style lang="scss">
	.content {
		display: flex;
		width: 100%;
		justify-content: center;
		padding: 24px 20px 80px 20px;

		.inner {
			display: flex;
			flex-flow: column;
			width: 100%;
			max-width: 900px;
		}

		.page-header {
			display: flex;
			align-items: flex-start;
			justify-content: space-between;
			gap: 16px;
			margin-bottom: 28px;
			padding-bottom: 20px;
			border-bottom: 1px solid rgba(255, 255, 255, 0.06);

			.header-titles {
				display: flex;
				flex-direction: column;
				gap: 4px;

				h2 {
					font-size: 26px;
					font-weight: 700;
					letter-spacing: -0.02em;
					color: $text-color;
					margin: 0;
				}

				.subtitle {
					font-size: 14px;
					color: $text-muted;
					margin: 0;
				}
			}

			.beta-badge {
				font-size: 11px;
				font-weight: 600;
				text-transform: uppercase;
				letter-spacing: 0.5px;
				background: rgba(54, 186, 152, 0.15);
				color: #36ba98;
				padding: 3px 8px;
				border-radius: 12px;
				border: 1px solid rgba(54, 186, 152, 0.3);
			}
		}

		.big-btns {
			display: grid;
			grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
			gap: 16px;
			width: 100%;

			.spinner-wrap {
				grid-column: 1 / -1;
				display: flex;
				justify-content: center;
				padding: 60px 0;
			}
		}

		.trakt-import-btn {
			display: flex;
			flex-flow: column;
			justify-content: center;
			align-items: center;
			gap: 14px;
			height: 160px;
			width: 100%;
			padding: 24px;
			background: rgba(255, 255, 255, 0.03);
			border: 1px solid rgba(255, 255, 255, 0.08);
			border-radius: 14px;
			user-select: none;
			transition: all 180ms ease;
			color: rgba(255, 255, 255, 0.9);
			fill: rgba(255, 255, 255, 0.9);
			cursor: pointer;

			&:hover {
				background: rgba(255, 255, 255, 0.06);
				border-color: rgba(54, 186, 152, 0.4);
				color: #ffffff;
				fill: #ffffff;
				transform: translateY(-2px);
			}

			h4 {
				font-size: 14px;
				font-weight: 600;
				letter-spacing: 0.2px;
				margin: 0;
			}
		}

		.export-section {
			margin-top: 48px;
			padding: 26px;
			background: rgba(255, 255, 255, 0.02);
			border: 1px solid rgba(255, 255, 255, 0.06);
			border-radius: 16px;
			display: flex;
			flex-direction: column;
			gap: 20px;

			.export-header {
				display: flex;
				justify-content: space-between;
				align-items: center;

				.export-titles {
					display: flex;
					flex-direction: column;
					gap: 6px;

					.title-row {
						display: flex;
						align-items: center;
						gap: 10px;
						color: $text-color;
						fill: $text-color;

						h3 {
							margin: 0;
							font-size: 18px;
							font-weight: 600;
							letter-spacing: -0.01em;
						}
					}

					.subtitle {
						margin: 0;
						font-size: 13px;
						color: $text-muted;
					}
				}
			}

			.export-actions {
				display: grid;
				grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
				gap: 14px;

				.export-btn {
					display: flex;
					flex-direction: column;
					align-items: center;
					justify-content: center;
					gap: 6px;
					padding: 16px 20px;
					border-radius: 12px;
					cursor: pointer;
					transition: all 0.2s ease;
					user-select: none;
					text-align: center;
					border: 1px solid transparent;

					&:disabled {
						opacity: 0.5;
						cursor: not-allowed;
					}

					span {
						font-weight: 600;
						font-size: 14px;
					}

					.btn-subtext {
						font-size: 11px;
						font-weight: 400;
						opacity: 0.7;
					}

					&.primary {
						background: rgba(54, 186, 152, 0.12);
						color: #36ba98;
						fill: #36ba98;
						border-color: rgba(54, 186, 152, 0.3);

						&:hover:not(:disabled) {
							background: rgba(54, 186, 152, 0.22);
							border-color: rgba(54, 186, 152, 0.5);
							transform: translateY(-2px);
						}
					}

					&.secondary {
						background: rgba(255, 255, 255, 0.04);
						color: $text-color;
						fill: $text-color;
						border-color: rgba(255, 255, 255, 0.08);

						&:hover:not(:disabled) {
							background: rgba(255, 255, 255, 0.08);
							border-color: rgba(255, 255, 255, 0.18);
							transform: translateY(-2px);
						}
					}
				}
			}
		}
	}
</style>
