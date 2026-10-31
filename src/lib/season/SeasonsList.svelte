<script lang="ts">
	import type {
		MediaSeason,
		TMDBSeasonDetails,
		Watched,
		WatchedEpisode,
		WatchedSeason,
		WatchedStatus,
	} from "@/types";
	import { req } from "../util/api";
	import Spinner from "@/lib/Spinner.svelte";
	import Error from "@/lib/Error.svelte";
	import SeasonsListEpisode from "./SeasonsListEpisode.svelte";
	import PosterStatus from "@/lib/poster/PosterStatus.svelte";
	import { notify } from "@/lib/util/notify";
	import PosterRating from "@/lib/poster/PosterRating.svelte";
	import Icon from "@/lib/Icon.svelte";
	import { watchedStatuses } from "@/lib/util/helpers";
	import { removeWatchedSeason, updateWatchedSeason } from "./api";
	import { onMount } from "svelte";

	interface Props {
		tvId: number;
		seasons: MediaSeason[];
		watchedItem: Watched | undefined;
		lastViewedSeason: number | undefined;
		lastViewedSeasonChanged: (wid: number, lvs: number) => void;
		onStatusChanged?: () => void;
	}

	let {
		tvId,
		seasons,
		watchedItem,
		lastViewedSeason,
		lastViewedSeasonChanged,
		onStatusChanged,
	}: Props = $props();

	let activeSeason = $state<number>(1);
	let userSelectedSeason = $state(false);
	let seasonDetailsReq: Promise<TMDBSeasonDetails> | undefined = $state();

	$effect(() => {
		if (!userSelectedSeason && typeof lastViewedSeason === "number") {
			activeSeason = lastViewedSeason;
			seasonDetailsReq = sdr(activeSeason);
		}
	});

	onMount(() => {
		console.debug("SeasonsList: Mounted.");
		if (!seasonDetailsReq) {
			seasonDetailsReq = sdr(activeSeason);
		}
	});

	async function sdr(seasonNum: number) {
		const wid = watchedItem?.id;
		console.debug("SeasonList: sdr: Called.", tvId, seasonNum, wid);
		const resp = await req.getWhole<TMDBSeasonDetails>(
			`/content/tv/${tvId}/season/${seasonNum}`,
			{
				params: {
					watchedId: wid,
				},
			},
		);
		try {
			if (wid) {
				// If we sent a watched id, expect a 'watcharr-lastviewedseason-saved' header in the response.
				const hVal = resp.headers.get("watcharr-lastviewedseason-saved");
				if (hVal) {
					lastViewedSeasonChanged(wid, seasonNum);
				} else {
					console.error(
						"SeasonList: sdr: No header in response indicating that the lastviewedseason was saved.",
					);
					notify({
						type: "error",
						text: "Failed when saving last viewed season",
					});
				}
			}
		} catch (err) {
			console.error(
				"SeasonList: sdr: Failed to process lastviewedseason-saved header.",
				err,
			);
		}
		return resp.body;
	}

	async function handleStatusClick(
		type: WatchedStatus | "DELETE",
		seasonNumber: number,
		seasonDetails?: TMDBSeasonDetails,
	) {
		if (!watchedItem) {
			console.error("handleStatusClick: No watched item.");
			return;
		}
		if (type === "DELETE") {
			const ws = watchedItem.watchedSeasons?.find(
				(s) => s.seasonNumber === seasonNumber,
			);
			if (!ws) {
				notify({
					text: "Failed to find watched season id. Please try refreshing.",
					type: "error",
				});
				return;
			}
			removeWatchedSeason(watchedItem, ws.id);
			return;
		}

		// When marking season as FINISHED, proactively mark all episodes of this season as FINISHED
		if (type === "FINISHED" && seasonDetails?.episodes?.length) {
			if (!watchedItem.watchedEpisodes) {
				watchedItem.watchedEpisodes = [];
			}
			for (const ep of seasonDetails.episodes) {
				const existing = watchedItem.watchedEpisodes.find(
					(we) =>
						we.seasonNumber === seasonNumber &&
						we.episodeNumber === ep.episode_number,
				);
				if (existing) {
					existing.status = "FINISHED";
				} else {
					watchedItem.watchedEpisodes.push({
						id: -ep.id,
						seasonNumber,
						episodeNumber: ep.episode_number,
						status: "FINISHED",
						rating: 0,
					} as WatchedEpisode);
				}
			}
			watchedItem.watchedEpisodes = [...watchedItem.watchedEpisodes];
		}

		await updateWatchedSeason(watchedItem, seasonNumber, { status: type });
		onStatusChanged?.();
	}

	function handleStarClick(rating: number, seasonNumber: number) {
		if (!watchedItem) {
			console.error("handleStarClick: No watchedItem!");
			return;
		}
		updateWatchedSeason(watchedItem, seasonNumber, { rating });
	}

	function checkSeasonStatus(
		watchedSeasons: WatchedSeason[] | undefined,
		currentSeason: MediaSeason,
	): WatchedStatus | undefined {
		if (watchedSeasons) {
			const watchedSeason = watchedSeasons.find(
				(ws) => ws.seasonNumber === currentSeason.number,
			);
			console.debug(
				"checkSeasonStatus:",
				currentSeason.number,
				watchedSeason?.status,
			);
			return watchedSeason?.status;
		}
		return undefined;
	}

	function getSeasonWatchedCount(seasonNumber: number): number {
		if (!watchedItem?.watchedEpisodes) return 0;
		return watchedItem.watchedEpisodes.filter(
			(we) => we.seasonNumber === seasonNumber && we.status === "FINISHED",
		).length;
	}
</script>

<div class="ctr">
	<ul class="seasons">
		{#each seasons as season (season.number)}
			{@const watchedCount = getSeasonWatchedCount(season.number)}
			{@const pct =
				season.episodeCount > 0
					? Math.min(
							100,
							Math.round((watchedCount / season.episodeCount) * 100),
						)
					: 0}
			<button
				class="plain season-card-btn"
				class:active={activeSeason === season.number}
				onclick={() => {
					console.debug("SeasonsList: Season button pressed", season.number);
					userSelectedSeason = true;
					activeSeason = season.number;
					seasonDetailsReq = sdr(activeSeason);
				}}
			>
				<div class="season-btn-main">
					<div class="season-info-left">
						<h4 class="season-name">{season.name}</h4>
						<span class="season-episodes">
							{#if season.episodeCount > 0}
								{#if watchedItem && watchedCount > 0}
									{watchedCount}/{season.episodeCount} watched
								{:else}
									{season.episodeCount} Episodes
								{/if}
							{:else}
								No Episodes Yet
							{/if}
						</span>
					</div>
					<div class="season-info-right">
						{#if season.releaseDate}
							<span class="season-date">
								{new Date(Date.parse(season.releaseDate)).getFullYear()}
							</span>
						{:else if season.number > 0}
							<span class="season-date">TBD</span>
						{/if}

						{#if watchedItem}
							{@const status = checkSeasonStatus(
								watchedItem.watchedSeasons,
								season,
							)}
							{#if status}
								<div class="plain season-status" title={`Season ${status}`}>
									<Icon i={watchedStatuses[status]} wh={18} />
								</div>
							{/if}
						{/if}
					</div>
				</div>

				{#if watchedItem && season.episodeCount > 0 && watchedCount > 0}
					<div class="season-mini-bar">
						<div class="season-mini-fill" style={`width: ${pct}%`}></div>
					</div>
				{/if}
			</button>
		{/each}
		<div class="last"></div>
	</ul>

	{#if seasonDetailsReq}
		<div class="episodes">
			{#await seasonDetailsReq}
				<Spinner />
			{:then season}
				<div class="episodes-topbar">
					<h3>{season.name}</h3>
					{#if watchedItem}
						{@const ws = watchedItem?.watchedSeasons?.find(
							(s) => s.seasonNumber === season.season_number,
						)}
						{#if ws}
							<div class="rating">
								<PosterRating
									rating={ws?.rating}
									btnTooltip="Season Rating"
									handleStarClick={(r) =>
										handleStarClick(r, season.season_number)}
									minimal={true}
									direction="bot"
								/>
							</div>
						{/if}
						<div class="status">
							<PosterStatus
								status={ws?.status}
								btnTooltip="Season Status"
								handleStatusClick={(t) =>
									handleStatusClick(t, season.season_number, season)}
								minimal={true}
								direction="bot"
							/>
						</div>
					{/if}
				</div>

				{#if season?.episodes?.length > 0}
					<ul>
						{#each season.episodes as ep (ep.id)}
							<SeasonsListEpisode {ep} {tvId} {watchedItem} {onStatusChanged} />
						{/each}
					</ul>
				{:else}
					<h3 class="norm">No episodes in this season yet!</h3>
				{/if}
			{:catch err}
				<Error pretty="Failed to load season details!" error={err} />
			{/await}
		</div>
	{:else}
		<h3 class="norm">Try selecting a season.</h3>
	{/if}
</div>

<style lang="scss">
	.ctr {
		display: flex;
		flex-flow: row;
		gap: 24px;
		width: 100%;
	}

	.episodes {
		width: 100%;
		min-width: 0;

		ul {
			display: flex;
			flex-flow: column;
			list-style: none;
			gap: 16px;
			padding: 0 0 80px 0;
			margin: 0;
		}
	}

	ul.seasons {
		display: flex;
		flex-flow: column;
		list-style: none;
		gap: 10px;
		min-width: fit-content;
		height: 100vh;
		overflow-y: auto;
		position: sticky;
		top: 0px;
		padding: 4px 0 20px;
		margin: 0;
		transition: top 200ms ease-in-out;

		.season-card-btn {
			display: flex;
			flex-flow: column;
			width: 240px;
			background: rgba(22, 27, 38, 0.7);
			border: 1px solid rgba(255, 255, 255, 0.08);
			border-radius: 14px;
			padding: 12px 14px;
			cursor: pointer;
			text-align: left;
			box-shadow: 0 4px 14px rgba(0, 0, 0, 0.25);
			backdrop-filter: blur(12px);
			-webkit-backdrop-filter: blur(12px);
			transition:
				background 150ms ease,
				border-color 150ms ease,
				transform 150ms ease,
				box-shadow 150ms ease;

			&:hover {
				background: rgba(30, 36, 52, 0.85);
				border-color: rgba(255, 255, 255, 0.16);
				transform: translateY(-1px);
			}

			&.active {
				background: rgba(99, 102, 241, 0.14);
				border-color: rgba(99, 102, 241, 0.5);
				box-shadow:
					0 4px 20px -2px rgba(99, 102, 241, 0.3),
					0 0 0 1px rgba(99, 102, 241, 0.15) inset;

				.season-name {
					color: #ffffff;
				}
			}

			.season-btn-main {
				display: flex;
				justify-content: space-between;
				align-items: flex-start;
				width: 100%;
				gap: 8px;
			}

			.season-info-left {
				display: flex;
				flex-direction: column;
				gap: 4px;
				min-width: 0;

				.season-name {
					margin: 0;
					font-size: 15px;
					font-weight: 700;
					color: rgba(255, 255, 255, 0.95);
					white-space: nowrap;
					overflow: hidden;
					text-overflow: ellipsis;
				}

				.season-episodes {
					font-size: 12.5px;
					color: rgba(255, 255, 255, 0.6);
					font-weight: 500;
				}
			}

			.season-info-right {
				display: flex;
				flex-direction: column;
				align-items: flex-end;
				gap: 4px;
				flex-shrink: 0;

				.season-date {
					font-size: 12px;
					font-weight: 600;
					color: rgba(255, 255, 255, 0.45);
				}

				.season-status {
					display: flex;
					align-items: center;
					color: #a5b4fc;
				}
			}

			.season-mini-bar {
				width: 100%;
				height: 4.5px;
				background: rgba(255, 255, 255, 0.08);
				border-radius: 999px;
				overflow: hidden;
				margin-top: 10px;

				.season-mini-fill {
					height: 100%;
					background: $brand-color;
					border-radius: 999px;
					transition: width 220ms ease;
				}
			}
		}

		/* extra scroll space under last item */
		.last {
			padding: 10px;
		}
	}

	.episodes-topbar {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 20px;
		padding: 14px 20px;
		background: rgba(22, 27, 38, 0.6);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 14px;
		backdrop-filter: blur(12px);
		-webkit-backdrop-filter: blur(12px);
		position: relative;
		z-index: 25;

		&:hover,
		&:focus-within,
		&:has(:global(.shown)),
		&:has(:global(.open)) {
			z-index: 55 !important;
		}

		h3 {
			margin: 0;
			font-size: 18px;
			font-weight: 700;
			color: #ffffff;
		}

		div {
			transition: width 100ms ease;

			&:first-of-type {
				margin-left: auto;
			}

			&.rating {
				height: 40px;
				min-height: 40px;
				position: relative;
				overflow: visible;

				&:hover,
				&:focus-within,
				&:has(:global(.shown)),
				&:has(:global(.open)) {
					z-index: 65 !important;
				}
			}

			&.status {
				width: 45px;
				min-height: 40px;
				height: 40px;
				overflow: visible;
				position: relative;

				&:hover,
				&:focus-within,
				&:has(:global(.shown)),
				&:has(:global(.open)) {
					z-index: 65 !important;
				}
			}
		}
	}

	@media screen and (min-width: 960px) {
		:global(body.nav-shown) ul.seasons {
			top: $nav-height;
			height: calc(100vh - $nav-height);
		}
	}

	@media screen and (max-width: 960px) {
		.ctr {
			flex-flow: column;
		}

		:global(body.nav-shown) ul.seasons {
			top: $nav-height;
		}

		ul.seasons {
			flex-flow: row;
			flex-wrap: nowrap;
			position: unset;
			height: unset;
			justify-content: unset;
			overflow-x: auto;
			min-width: unset;
			position: sticky;
			top: 0px;
			padding: 10px 0;
			z-index: 5;
			@include nav-blur;

			.season-card-btn {
				width: 180px;
				flex-shrink: 0;

				&.active {
					position: unset;
				}
			}
		}
	}
</style>
