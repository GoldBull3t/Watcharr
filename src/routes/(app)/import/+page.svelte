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
	} from "@/types";
	import Icon from "@/lib/Icon.svelte";
	import { resolve } from "$app/paths";

	let isLoading = $state(false);

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
			{/if}
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
	}
</style>
