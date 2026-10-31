<!--
  /import/process is for processing the
  selected files data. Here it will be
  displayed and imported.
 -->

<script lang="ts">
	import { goto } from "$app/navigation";
	import DropDown from "@/lib/DropDown.svelte";
	import Error from "@/lib/Error.svelte";
	import Icon from "@/lib/Icon.svelte";
	import Modal from "@/lib/Modal.svelte";
	import Poster from "@/lib/poster/Poster.svelte";
	import PosterList from "@/lib/poster/PosterList.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import SpinnerTiny from "@/lib/SpinnerTiny.svelte";
	import { sleep } from "@/lib/util/helpers";
	import { notify } from "@/lib/util/notify";
	import { store } from "@/store.svelte";
	import {
		ImportResponseType,
		getContentTypeFromMedia,
		type ImportResponse,
		type ImportedList,
		type Media,
		type WatchedStatus,
	} from "@/types";
	import { req } from "@/lib/util/api";
	import { onDestroy } from "svelte";
	import papa from "papaparse";
	import Status from "@/lib/Status.svelte";
	import { resolve } from "$app/paths";

	interface ImportedListItemMultiProblem {
		original: ImportedList;
		results: Media[];
		callback: (err: Error | string | undefined) => void;
	}

	let rList: ImportedList[] = $state([]);
	let isImporting = $state(false);
	let importText = $state("");
	let cancelled = $state(false);
	let importTableEl: HTMLTableElement | undefined = $state();

	let successCount = $derived(
		rList.filter(
			(l) =>
				l.state === ImportResponseType.IMPORT_SUCCESS ||
				l.state === ImportResponseType.IMPORT_EXISTS,
		).length,
	);
	let failCount = $derived(
		rList.filter(
			(l) =>
				l.state === ImportResponseType.IMPORT_FAILED ||
				l.state === ImportResponseType.IMPORT_NOTFOUND,
		).length,
	);
	let pendingCount = $derived(rList.filter((l) => !l.state).length);

	onDestroy(() => {
		cancelled = true;
	});

	// This isn't reactive to avoid bugs (hopefully this isn't a bug)
	let dropDownSupportedTypes = (() => {
		let t = ["movie", "tv"];
		if (store.serverFeatures?.games) {
			t.push("game");
		}
		return t;
	})();

	// Set when current item being imported gets an IMPORT_MULTI
	// response, which then shows the modal for user to pick correct item.
	let importMultiItem: ImportedListItemMultiProblem | undefined = $state();
	// Set when user clicks 'Change Statuses' button for the change all
	// statuses modal.
	let changeAllStatusesModalCb:
		((newStatus?: WatchedStatus) => void) | undefined = $state();

	async function getList() {
		const list = store.importedList;
		if (!list) {
			console.log("import/process, no list, returning to /import");
			goto(resolve("/import"));
			return;
		}
		console.log("getList", list);
		if (list?.type === "text-list") {
			importText = "Text List";
			// Regex to match a year in between brackets,
			// which we assume is the release year of content.
			const yearRegex = new RegExp(/\([0-9]{4}\)/);
			// Matches supported rating between square brackets.
			// Supports (the sixes can be any number 0-9): [6], [6.6], and [10].
			const ratingRegex = new RegExp(/\[([0-9](?:\.[0-9])?|10)\]/);
			const s = list.data.split("\n");
			for (let i = 0; i < s.length; i++) {
				if (!s[i]) {
					console.warn("getList: ignoring entry:", s[i]);
					continue;
				}
				const l: ImportedList = {};
				l.name = s[i];
				// Try extracting a year.
				const year = l.name.match(yearRegex);
				if (year && year.length > 0) {
					l.year = Number(year[0].replaceAll(/\(|\)/g, ""));
					l.name = l.name.replace(yearRegex, "");
				}
				// Try extracting a rating.
				const rating = l.name.match(ratingRegex);
				if (rating && rating.length > 0) {
					// console.log("found rating", rating);
					l.rating = Number(rating[1]);
					if (typeof rating.index === "number") {
						l.name = l.name.slice(0, rating.index);
					}
				}
				l.name = l.name.trim();
				rList.push(l);
			}
		} else if (list?.type === "tmdb") {
			importText = "TMDB";
			const s = papa.parse(list.data.trim(), { header: true });
			console.debug("parsed csv", s);
			for (let i = 0; i < s.data.length; i++) {
				try {
					// eslint-disable-next-line @typescript-eslint/no-explicit-any
					const el = s.data[i] as any;
					if (el) {
						// Skip if no name or tmdb id
						if (!el.Name && !el["TMDb ID"]) {
							console.warn("Skipping item with no name or tmdb id", el);
							continue;
						}
						const l: ImportedList = { name: el.Name };
						const year = el["Release Date"]
							? new Date(el["Release Date"])
							: undefined;
						if (year) {
							l.year = year.getFullYear();
						}
						if (el.Type === "movie" || el.Type === "tv") {
							l.type = el.Type;
						}
						if (el["TMDb ID"]) {
							l.tmdbId = Number(el["TMDb ID"]);
						}
						if (el["Your Rating"]) {
							l.rating = Math.floor(Number(el["Your Rating"]));
						}
						if (el["Date Rated"]) {
							l.ratingCustomDate = new Date(el["Date Rated"]);
						}
						rList.push(l);
					}
				} catch (err) {
					console.error("Failed to process an item!", err);
					notify({
						type: "error",
						text: "Failed to process an item!",
					});
				}
			}
		} else if (list?.type === "imdb") {
			// There are different types of imdb exports (ratings & watchlist/list),
			// there are common keys between these types that we use below so should
			// be okay with importing either.
			importText = "IMDb";
			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			const s = papa.parse<any>(list.data.trim(), { header: true });
			console.debug("parsed csv", s);
			let anySkipped = false;
			// Sort so that episodes comes last, so they are imported last.
			s.data?.sort((a, b) => {
				const aType = (a["Title Type"] || a["Tipo de título"] || a["Type"] || "")
					?.toLowerCase()
					.trim();
				if (aType === "tv episode" || aType === "tvepisode" || aType === "episódio de tv" || aType === "episodio de tv") {
					return 1;
				}
				const bType = (b["Title Type"] || b["Tipo de título"] || b["Type"] || "")
					?.toLowerCase()
					.trim();
				if (bType === "tv episode" || bType === "tvepisode" || bType === "episódio de tv" || bType === "episodio de tv") {
					return -1;
				}
				return 0;
			});
			for (let i = 0; i < s.data.length; i++) {
				try {
					// eslint-disable-next-line @typescript-eslint/no-explicit-any
					const el = s.data[i] as any;
					if (el) {
						const imdbId = el["Const"] || el["const"];
						const titleVal = el["Title"] || el["Título"] || el["Titulo"] || el["Original Title"];
						const type = (
							el["Title Type"] ||
							el["Tipo de título"] ||
							el["Typ"] ||
							el["Type"] ||
							""
						)
							?.toLowerCase()
							.trim();
						// Skip if no name or tmdb id
						if (!titleVal && !imdbId) {
							console.warn("Skipping item with no title or imdb id", el);
							anySkipped = true;
							continue;
						}
						if (!type) {
							console.warn("Skipping item with no type", el);
							anySkipped = true;
							continue;
						}
						const l: ImportedList = { name: titleVal };
						const releaseDateRaw = el["Release Date"] || el["Data de lançamento"];
						const year = releaseDateRaw
							? new Date(releaseDateRaw)
							: el["Year"]
								? new Date(`${el["Year"]}-01-01`)
								: undefined;
						if (year && !isNaN(year.getTime())) {
							l.year = year.getFullYear();
						}

						switch (type) {
							case "movie":
							case "filme":
							case "película":
							case "film":
							case "video":
							case "vídeo":
							case "tv movie":
							case "tvmovie":
							case "filme para tv":
							case "short":
							case "tvshort":
							case "curta-metragem":
								l.type = "movie";
								break;
							case "tv series":
							case "tvseries":
							case "série de tv":
							case "serie de tv":
							case "tv mini series":
							case "tvminiseries":
							case "minissérie de tv":
							case "tv special":
							case "tvspecial":
							case "tv short":
								l.type = "tv";
								break;
							case "tv episode":
							case "tvepisode":
							case "episódio de tv":
							case "episodio de tv":
								l.type = "tv_episode";
								break;
							default:
								console.warn(
									"Skipping item with invalid type",
									`(${type})`,
									el,
								);
								anySkipped = true;
								continue;
						}

						if (imdbId) {
							l.imdbId = imdbId;
						}
						const yourRating = el["Your Rating"] || el["Sua classificação"];
						if (yourRating) {
							l.rating = Math.floor(Number(yourRating));
						}
						const dateRatedRaw = el["Date Rated"] || el["Data da classificação"] || el["Created"];
						if (dateRatedRaw) {
							const parsedRatedDate = new Date(dateRatedRaw);
							if (!isNaN(parsedRatedDate.getTime())) {
								l.ratingCustomDate = parsedRatedDate;
								l.datesWatched = [parsedRatedDate];
							}
						}
						rList.push(l);
					}
				} catch (err) {
					console.error("Failed to process an item!", err);
					notify({
						type: "error",
						text: "Failed to process an item!",
					});
				}
			}
			notify({
				text: "Any tv episodes will be added to the bottom of the table, don't change the Type on these!",
				time: 20000,
			});
			if (anySkipped) {
				notify({
					type: "error",
					text: "Some items with invalid data may have been skipped (check source data for missing ids/titles or look in console for more details).",
				});
			}
		} else if (list?.type === "movary") {
			importText = "Movary";
			try {
				const s = JSON.parse(list.data);
				// Builds imported list in previous step for ease.
				rList = s;
			} catch (err) {
				console.error("Movary import processing failed!", err);
				notify({
					type: "error",
					text: "Processing failed!. Please report this issue if it persists.",
				});
			}
		} else if (list?.type === "watcharr") {
			importText = "Watcharr";
			try {
				const s = JSON.parse(list.data);
				// Builds imported list in previous step for ease.
				rList = s;
			} catch (err) {
				console.error("Watcharr import processing failed!", err);
				notify({
					type: "error",
					text: "Processing failed!. Please report this issue if it persists.",
				});
			}
		} else if (list?.type === "myanimelist") {
			importText = "MyAnimeList";
			try {
				const parser = new DOMParser();
				const doc = parser.parseFromString(list.data.trim(), "application/xml");
				const errorNode = doc.querySelector("parsererror");
				if (errorNode) {
					console.error("MyAnimeList parse error:", errorNode);
					notify({
						type: "error",
						text: "An error occurred while parsing your MyAnimeList export!",
					});
					return;
				}
				console.log(doc.documentElement.querySelectorAll("anime"));
				const animeNodes = doc.documentElement.querySelectorAll("anime");
				if (animeNodes?.length <= 0) {
					console.error("MyAnimeList: Found no anime nodes:", animeNodes);
					notify({
						type: "error",
						text: "We found no Anime entries in your export file!",
					});
					return;
				}
				for (let i = 0; i < animeNodes.length; i++) {
					const animeNode = animeNodes[i];
					const titleNode = animeNode.querySelector("series_title");
					console.debug("Processing anime:", titleNode?.textContent);
					if (!titleNode?.textContent) {
						console.error("No title found for an anime!", animeNode, titleNode);
						notify({
							type: "error",
							text: "An anime failed to import, a title was not found! Check console for more details.",
						});
						continue;
					}
					const l: ImportedList = { name: titleNode.textContent };
					const scoreNode = animeNode.querySelector("my_score");
					if (scoreNode?.textContent) {
						l.rating = Number(scoreNode.textContent);
					}
					const statusNode = animeNode.querySelector("my_status");
					if (statusNode?.textContent) {
						let malStatus = statusNode.textContent?.toLowerCase();
						if (malStatus === "on-hold") {
							l.status = "HOLD";
						} else if (malStatus === "dropped") {
							l.status = "DROPPED";
						} else if (malStatus === "plan to watch") {
							l.status = "PLANNED";
						} else if (malStatus === "watching") {
							l.status = "WATCHING";
						} else if (malStatus === "completed") {
							l.status = "FINISHED";
						} else {
							console.warn(
								"Anime has no status or an unrecognized status:",
								malStatus,
								"anime_title:",
								titleNode.textContent,
							);
						}
					}
					const typeNode = animeNode.querySelector("series_type");
					if (typeNode?.textContent) {
						const malSeriesType = typeNode.textContent?.toLowerCase();
						if (malSeriesType === "tv" || malSeriesType === "movie") {
							l.type = malSeriesType;
						} else {
							console.warn(
								"Anime has no type or an unrecognized type:",
								malSeriesType,
								"anime_title:",
								titleNode.textContent,
							);
						}
					}
					try {
						const startDateNode = animeNode.querySelector("my_start_date");
						const finishDateNode = animeNode.querySelector("my_finish_date");
						if (
							startDateNode?.textContent &&
							startDateNode?.textContent != "0000-00-00"
						) {
							// For start date, we can simply add the activity manually.
							l.activity = [
								// We don't need all the data when importing activity.
								// customDate must be a date object.
								{
									type: "STATUS_CHANGED",
									data: "WATCHING",
									customDate: new Date(startDateNode.textContent),
								},
								// eslint-disable-next-line @typescript-eslint/no-explicit-any
							] as any[];
						}
						if (
							finishDateNode?.textContent &&
							finishDateNode?.textContent != "0000-00-00"
						) {
							l.datesWatched = [new Date(finishDateNode.textContent)];
						}
					} catch (err) {
						console.error(
							"Processing start/finish times for anime failed!",
							err,
						);
						notify({
							type: "error",
							text: "Failed to process start/finish times for an anime! Check console for more details.",
						});
					}
					rList.push(l);
				}
			} catch (err) {
				console.error("MyAnimeList import failed!", err);
				notify({
					type: "error",
					text: "Failed to process import data!",
				});
			}
		} else if (list?.type === "ryot") {
			importText = "Ryot";
			try {
				const s = JSON.parse(list.data);
				// Builds imported list in previous step for ease.
				rList = s;
			} catch (err) {
				console.error("Ryot import processing failed!", err);
				notify({
					type: "error",
					text: "Processing failed!. Please report this issue if it persists.",
				});
			}
		} else if (list?.type === "todomovies") {
			importText = "TodoMovies";
			try {
				const s = JSON.parse(list.data);
				// Builds imported list in previous step for ease.
				rList = s;
			} catch (err) {
				console.error("TodoMovies import processing failed!", err);
				notify({
					type: "error",
					text: "Processing failed!. Please report this issue if it persists.",
				});
			}
		} else if (list?.type === "yamtrack" || list?.type === "letterboxd") {
			importText = list.type === "yamtrack" ? "Yamtrack" : "Letterboxd";
			try {
				const s = JSON.parse(list.data);
				// Builds imported list in previous step for ease.
				rList = s;
			} catch (err) {
				console.error(`${importText} import processing failed!`, err);
				notify({
					type: "error",
					text: "Processing failed! Please report this issue if it persists.",
				});
			}
		}
	}

	function addRow(
		ev: FocusEvent & { currentTarget: EventTarget & HTMLInputElement },
	) {
		if (!ev.currentTarget.value) {
			return;
		}
		const lo = { name: ev.currentTarget.value } as ImportedList;
		const yearEl = document.getElementById("addYear") as HTMLInputElement;
		if (yearEl?.value) {
			lo.year = Number(yearEl.value);
		}
		rList.push(lo);
		rList = rList;
		ev.currentTarget.value = "";
		yearEl.value = "";
	}

	function removeRow(l: ImportedList) {
		rList = rList.filter((r) => r.name !== l.name);
		rList = rList;
	}

	/**
	 * inputs that have a `data-validateme` property, we will validate with the
	 * browser.
	 */
	function validatableInputsAreValid(): boolean {
		if (!importTableEl) {
			// Somehow the table isn't defined, i guess allow continuing so we
			// dont block.
			console.warn("validatableInputsAreValid: There is no table element.");
			return true;
		}
		const inputs = importTableEl.querySelectorAll<HTMLInputElement>(
			"tbody input[data-validateme]",
		);
		if (inputs.length <= 0) {
			console.warn("validatableInputsAreValid: No inputs found.");
			return true;
		}
		for (const input of inputs) {
			if (!input.reportValidity()) {
				// Stop the for loop after we hit one input that is invalid.
				// No need to continue after we know at least one is invalid.
				console.error("validatableInputsAreValid: Invalid input found.", input);
				return false;
			}
		}
		return true;
	}

	let modalQueue: (() => void)[] = [];
	let isModalOpen = false;

	function acquireModalLock(): Promise<void> {
		return new Promise((resolve) => {
			if (!isModalOpen) {
				isModalOpen = true;
				resolve();
			} else {
				modalQueue.push(resolve);
			}
		});
	}

	function releaseModalLock() {
		if (modalQueue.length > 0) {
			const next = modalQueue.shift();
			next?.();
		} else {
			isModalOpen = false;
		}
	}

	async function startImport() {
		if (!validatableInputsAreValid()) {
			console.warn(
				"startImport: Some fields are not valid, not starting import!",
			);
			notify({
				type: "error",
				text: "An error was found in one of the inputs, please fix it and try starting the import again.",
			});
			return;
		}
		console.log("startImport: Starting with 4 concurrent workers.", rList);
		isImporting = true;
		window.scrollTo(0, 0);

		const concurrency = 4;
		let currentIndex = 0;

		async function worker() {
			while (currentIndex < rList.length) {
				if (cancelled) return;
				const index = currentIndex++;
				const li = rList[index];
				if (!li) break;
				try {
					console.log("Importing", li);
					await doImport(li);
				} catch (err) {
					li.state = ImportResponseType.IMPORT_FAILED;
					console.error("Failed to import item:", li, "reason:", err);
				}
				// Small 15ms tick to yield to browser paint loop
				await sleep(15);
			}
		}

		const workers = Array.from({ length: concurrency }, () => worker());
		await Promise.all(workers);

		store.importedList = undefined;
		if (
			rList.some(
				(i) =>
					i.state == ImportResponseType.IMPORT_FAILED ||
					i.state == ImportResponseType.IMPORT_NOTFOUND,
			)
		) {
			// Some items failed.. go to some-failed
			store.parsedImportedList = rList;
			goto(resolve("/import/some-failed"));
		} else {
			notify({
				type: "success",
				text: "All content successfully imported!",
				time: 15000,
			});
			goto(resolve("/"));
		}
	}

	async function doImport(item: ImportedList) {
		if (!item.name?.trim()) {
			item.state = ImportResponseType.IMPORT_NOTFOUND;
			rList = rList;
			return;
		}
		const resp = await req.post<ImportResponse>("/import", item);
		if (resp.type === ImportResponseType.IMPORT_MULTI) {
			console.log("Import found multiple responses for content", resp);
			let results = resp.results;
			if (!results || results.length <= 0) {
				item.state = ImportResponseType.IMPORT_NOTFOUND;
				rList = rList;
				return;
			}
			if (item.year) {
				results.sort((a, b) => {
					try {
						const ar = a.releaseDate;
						const ay = ar ? new Date(Date.parse(ar)).getFullYear() : undefined;
						const br = b.releaseDate;
						const by = br ? new Date(Date.parse(br)).getFullYear() : undefined;
						if (ay == item.year) return -1;
						else if (by == item.year) return 1;
					} catch (err) {
						console.error("doImport: results sort failed", err);
					}
					return 0;
				});
			}
			await acquireModalLock();
			return new Promise((res, rej) => {
				importMultiItem = {
					original: item,
					results: results,
					callback: (err) => {
						releaseModalLock();
						if (err) {
							item.state = ImportResponseType.IMPORT_NOTFOUND;
							rList = rList;
							rej(err);
						} else {
							res(0);
						}
					},
				};
			});
		} else if (resp.type === ImportResponseType.IMPORT_SUCCESS) {
			item.state = ImportResponseType.IMPORT_SUCCESS;
			const w = resp.watchedEntry;
			if (w) {
				const release = w.media?.releaseDate;
				if (release) item.year = new Date(Date.parse(release)).getFullYear();
				const t = w.media ? getContentTypeFromMedia(w.media) : undefined;
				if (t) item.type = t;
			}
			rList = rList;
		} else {
			item.state = resp.type;
			rList = rList;
		}
	}

	/**
	 * Helper to allow user to quickly update
	 * all statuses to a new one.
	 */
	function changeAllStatuses() {
		changeAllStatusesModalCb = (newStatus) => {
			try {
				console.debug("changeAllStatusesModalCb: newStatus:", newStatus);
				if (!newStatus) {
					// User cancelled flow
					changeAllStatusesModalCb = undefined;
					return;
				}
				if (!rList) {
					console.error("changeAllStatusesModalCb: No list to modify!");
					changeAllStatusesModalCb = undefined;
					return;
				}
				for (let i = 0; i < rList.length; i++) {
					const r = rList[i];
					r.status = newStatus;
				}
				rList = rList;
			} catch (err) {
				console.error("changeAllStatusesModalCb: Failed!", err);
				notify({
					type: "error",
					text: "Failed when updating all statuses. Please try again!",
				});
			}
			changeAllStatusesModalCb = undefined;
		};
	}

	// Not sure why this happens:
	// https://github.com/sveltejs/svelte/discussions/14692#discussioncomment-11569475
	const alist = getList();
</script>

{#await alist}
	<Spinner />
{:then}
	<div class="content">
		<div class="inner">
			{#if rList}
				<div class="process-header">
					<div class="header-titles">
						<h2>Importing {importText ? `From ${importText}` : ""}</h2>
						<p class="subtitle">
							{#if !isImporting}
								Review your imported list and adjust metadata before starting
								import.
							{:else}
								Import in progress... Failed imports can be corrected once
								completed.
							{/if}
						</p>
					</div>
					<div class="header-stats">
						<span class="stat-pill total">{rList.length} Items</span>
						{#if isImporting || successCount > 0 || failCount > 0}
							<span class="stat-pill success">{successCount} Succeeded</span>
							{#if failCount > 0}
								<span class="stat-pill fail">{failCount} Failed</span>
							{/if}
							{#if pendingCount > 0 && isImporting}
								<span class="stat-pill pending">{pendingCount} Remaining</span>
							{/if}
						{/if}
					</div>
				</div>
				<div class="table-wrap">
					<table
						bind:this={importTableEl}
						class={isImporting ? "is-importing" : ""}
					>
						<thead>
							<tr>
								{#if isImporting}
									<th class="loading-col"></th>
								{/if}
								<th>Name</th>
								<th>Year</th>
								<th>Type</th>
								<th>Status</th>
								<th>Rating</th>
								{#if !isImporting}
									<th></th>
								{/if}
							</tr>
						</thead>
						<tbody>
							{#each rList as l, i (l.tmdbId ? `t-${l.tmdbId}` : l.igdbId ? `g-${l.igdbId}` : `${i}-${l.name}`)}
								<tr>
									{#if isImporting}
										<td class="icon-cell">
											<div>
												{#if !l.state}
													<SpinnerTiny />
												{:else if l.state === ImportResponseType.IMPORT_SUCCESS}
													<Icon i="check" wh={22} />
												{:else if l.state === ImportResponseType.IMPORT_NOTFOUND}
													<Icon i="close" wh={22} />
												{:else if l.state === ImportResponseType.IMPORT_FAILED}
													<Icon i="close" wh={22} />
												{:else if l.state === ImportResponseType.IMPORT_EXISTS}
													<Icon i="check" wh={22} />
												{/if}
											</div>
										</td>
									{/if}
									<td class="name">
										<input
											class="plain"
											bind:value={l.name}
											disabled={isImporting}
										/>
									</td>
									<td class="year">
										<input
											class="plain"
											bind:value={l.year}
											placeholder="YYYY"
											type="number"
											disabled={isImporting}
										/>
									</td>
									<td class="type">
										<DropDown
											options={dropDownSupportedTypes}
											bind:active={l.type}
											placeholder="Type"
											blendIn={true}
											disabled={isImporting}
										/>
									</td>
									<td class="status">
										<DropDown
											options={[
												"FINISHED",
												"PLANNED",
												"WATCHING",
												"HOLD",
												"DROPPED",
											]}
											bind:active={l.status}
											placeholder="Status"
											blendIn={true}
											disabled={isImporting}
										/>
									</td>
									<td class="rating">
										<input
											class="plain"
											data-validateme
											bind:value={l.rating}
											placeholder="0"
											type="number"
											min="0"
											max="10"
											step="0.1"
											disabled={isImporting}
										/>
									</td>
									{#if !isImporting}
										<td>
											<button
												class="plain delete"
												onclick={() => {
													removeRow(l);
												}}
											>
												<Icon i="close" wh="25" />
											</button>
										</td>
									{/if}
								</tr>
							{/each}
							{#if !isImporting}
								<tr>
									<td
										><input
											class="plain"
											placeholder="Name"
											onblur={addRow}
										/></td
									>
									<td class="year">
										<input
											class="plain"
											id="addYear"
											placeholder="YYYY"
											type="number"
										/>
									</td>
									<td class="type"></td>
									<td class="status"></td>
									<td class="rating"></td>
									<td></td>
								</tr>
							{/if}
						</tbody>
					</table>
				</div>
				<div class="btns">
					<button onclick={() => goto(resolve("/import"))}>
						<Icon i="arrow" />Back
					</button>
					<button onclick={() => changeAllStatuses()} disabled={isImporting}>
						Change All Statuses
					</button>
					<button onclick={startImport} disabled={isImporting}>
						Start Importing
					</button>
				</div>
				{#if typeof changeAllStatusesModalCb === "function"}
					<Modal
						title="Select a New Status"
						desc="Override the status for all content in the table"
						maxWidth="600px"
						onClose={() =>
							typeof changeAllStatusesModalCb === "function"
								? changeAllStatusesModalCb()
								: undefined}
					>
						<Status status={undefined} onChange={changeAllStatusesModalCb} />
					</Modal>
				{/if}
			{:else}
				<h2>No list</h2>
			{/if}
		</div>
	</div>

	<!-- Multiple results found modal -->
	{#if importMultiItem}
		<Modal
			title="Multiple Results Found"
			desc="Select the correct item for {importMultiItem.original
				.name} {importMultiItem.original.year &&
				'(' + importMultiItem.original.year + ')'}"
			onClose={() => {
				importMultiItem?.callback("closed results modal");
				importMultiItem = undefined;
			}}
		>
			<PosterList type="vertical">
				{#each importMultiItem.results as r (r.ids)}
					<Poster
						media={r}
						small={true}
						disableInteraction={true}
						hideButtons={true}
						onClick={async () => {
							const item = rList.find(
								(i) => i.name === importMultiItem?.original.name,
							);
							console.log(
								"MultipleResultsFound: Poster clicked. Item in rList:",
								item,
							);
							if (item) {
								// We found the item in our import list, update it
								// to match the selected choice and do the import with it.
								item.type = getContentTypeFromMedia(r);
								if (item.type === "game") {
									item.igdbId = r.ids.igdb;
								} else if (item.type === "movie" || item.type === "tv") {
									item.tmdbId = r.ids.tmdb;
								} else {
									item.state = ImportResponseType.IMPORT_FAILED;
									notify({
										type: "error",
										text: "Can't import selected result because it has an unsupported type associated with it!",
										time: 10000,
									});
									return;
								}
								try {
									await doImport(item);
									importMultiItem?.callback(undefined);
								} catch (err) {
									importMultiItem?.callback(String(err));
								}
								importMultiItem = undefined;
							} else {
								// TODO: show error notif and update state with error icon
							}
							console.log("multi: Poster clicked", r);
						}}
					/>
				{/each}
			</PosterList>
		</Modal>
	{/if}
{:catch err}
	<Error error={err} pretty="Failed to process list!" />
{/await}

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
			max-width: 1200px;
		}

		.process-header {
			display: flex;
			flex-wrap: wrap;
			align-items: flex-start;
			justify-content: space-between;
			gap: 16px;
			margin-bottom: 24px;
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

			.header-stats {
				display: flex;
				align-items: center;
				gap: 8px;
				flex-wrap: wrap;

				.stat-pill {
					font-size: 12px;
					font-weight: 600;
					padding: 4px 10px;
					border-radius: 12px;

					&.total {
						background: rgba(255, 255, 255, 0.06);
						color: rgba(255, 255, 255, 0.85);
					}

					&.success {
						background: rgba(54, 186, 152, 0.15);
						color: #36ba98;
						border: 1px solid rgba(54, 186, 152, 0.25);
					}

					&.fail {
						background: rgba(255, 68, 68, 0.15);
						color: #ff6b6b;
						border: 1px solid rgba(255, 68, 68, 0.25);
					}

					&.pending {
						background: rgba(255, 188, 10, 0.15);
						color: #ffbc0a;
						border: 1px solid rgba(255, 188, 10, 0.25);
					}
				}
			}
		}
	}

	.table-wrap {
		overflow: auto;
		background: rgba(255, 255, 255, 0.02);
		border: 1px solid rgba(255, 255, 255, 0.06);
		border-radius: 12px;
		padding: 4px 8px;
	}

	table {
		width: 100%;
		border-collapse: collapse;

		td {
			&.name {
				width: 100%;
			}

			&.year {
				width: 70px;
				min-width: 67px;
			}

			&.type {
				width: 120px;
			}

			&.rating {
				width: 82px;
			}
		}
	}

	table.is-importing {
		th {
			padding-left: 3px;
		}

		td {
			padding-left: 3px;

			input {
				padding: 7px 0;

				&:focus {
					padding: 7px 5px;
					padding-left: 3px;
				}
			}
		}
	}

	.btns {
		display: flex;
		flex-flow: row;
		margin-top: 20px;
		gap: 10px;

		button {
			display: inline-flex;
			align-items: center;
			gap: 6px;
			padding: 8px 16px;
			border-radius: 8px;
			font-size: 13px;
			font-weight: 500;
			cursor: pointer;

			&:last-of-type {
				margin-left: auto;
			}
		}
	}
</style>
