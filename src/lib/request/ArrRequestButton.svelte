<script lang="ts">
	import type {
		ArrDetailsResponse,
		ArrInfoResponse,
		ArrRequestResponse,
		ArrRequestStatus,
		ContentType,
		SeerrContentStatus,
	} from "@/types";
	import { onMount } from "svelte";
	import Icon from "../Icon.svelte";
	import tooltip from "../actions/tooltip";
	import { msToAmountsOfTime } from "../util/helpers";
	import { notify } from "../util/notify";
	import { req } from "../util/api";
	import { ReqerError } from "../util/fetch";
	import store from "@/store.svelte";

	interface Props {
		type: ContentType;
		tmdbId: number;
		openRequestModal: () => void;
	}

	let { type, tmdbId, openRequestModal }: Props = $props();

	let existingRequest: ArrRequestResponse | undefined;
	let info: ArrInfoResponse | undefined = $state();
	let seerrStatus: SeerrContentStatus | undefined = $state();
	let isRequestingSeerr = $state(false);

	// The status, extra added string types are set in this file.
	let status:
		| ArrDetailsResponse
		| "available"
		| "requested"
		| ArrRequestStatus
		| undefined = $state();
	let estimatedCompletionIn: string | undefined = $state();

	function formatDiskSize(bytes?: number): string {
		if (!bytes || bytes <= 0) return "";
		const gb = bytes / (1024 * 1024 * 1024);
		if (gb >= 1) return `${gb.toFixed(1)} GB`;
		const mb = bytes / (1024 * 1024);
		return `${mb.toFixed(0)} MB`;
	}

	let availableLabel = $derived.by(() => {
		if (seerrStatus?.isAvailable) {
			return "Available";
		}
		if (!info) return "Available";
		const parts: string[] = ["Available"];
		if (info.qualityName) {
			parts.push(info.qualityName);
		} else if (info.statistics && info.statistics.totalEpisodeCount > 0) {
			parts.push(`${info.statistics.episodeFileCount}/${info.statistics.totalEpisodeCount} eps`);
		}
		const sizeStr = formatDiskSize(info.sizeOnDisk);
		if (sizeStr) {
			parts.push(`(${sizeStr})`);
		}
		return parts.join(" • ");
	});

	async function getInfo() {
		try {
			if (!existingRequest) {
				console.warn("getInfo called before existingRequest exists.");
				return;
			}
			if (
				existingRequest.status !== "APPROVED" &&
				existingRequest.status !== "AUTO_APPROVED" &&
				existingRequest.status !== "FOUND"
			) {
				console.debug(
					"getInfo: Request is not in an approved state.. not continuing.",
				);
				return;
			}
			const resp = await req.get<ArrInfoResponse>(
				`/arr/${type === "movie" ? "rad" : "son"}/info/${existingRequest.id}`,
			);
			if (resp) {
				info = resp;
				if (info.hasFile) {
					status = "available";
				} else {
					getStatus();
				}
			}
		} catch (err) {
			if (
				ReqerError.withBody(err) &&
				err.response?.status === 404 &&
				err.hasErrorInBody() &&
				err.body.error === "request deleted"
			) {
				return;
			}
			console.error("ArrRequestButton: getInfo failed!", err);
			notify({
				text: `Failed when getting info from ${type === "movie" ? "Radarr" : "Sonarr"}`,
				type: "error",
			});
		}
	}

	async function getStatus(ev: MouseEvent | undefined = undefined) {
		try {
			if (!existingRequest) {
				console.warn("getStatus called before existingRequest exists.");
				return;
			}
			const statusResp = await req.getWhole<ArrDetailsResponse>(
				`/arr/${type === "movie" ? "rad" : "son"}/status/${existingRequest.serverName}/${existingRequest.arrId}`,
			);
			if (statusResp.status === 204) {
				status = "requested";
			} else if (statusResp.body) {
				status = statusResp.body;
				const estMs =
					new Date(status.estimatedCompletionTime).getTime() -
					new Date(Date.now()).getTime();
				const est = msToAmountsOfTime(estMs);
				if (est.days > 0) {
					estimatedCompletionIn = `${est.days} Day${est.days > 1 ? "s" : ""}`;
				} else if (est.hours > 0) {
					estimatedCompletionIn = `${est.hours} Hour${est.hours > 1 ? "s" : ""}`;
				} else if (est.minutes > 0) {
					estimatedCompletionIn = `${est.minutes} Minute${est.minutes > 1 ? "s" : ""}`;
				} else if (est.seconds > 0) {
					estimatedCompletionIn = `${est.seconds} Second${est.seconds > 1 ? "s" : ""}`;
				} else {
					estimatedCompletionIn = undefined;
				}
			}
			if (ev?.target) {
				setTimeout(() => {
					(ev?.target as HTMLButtonElement)?.blur();
				}, 250);
			}
		} catch (err) {
			console.error("ArrRequestButton: getStatus failed!", err);
			notify({
				text: `Failed to fetch content status from ${type === "movie" ? "Radarr" : "Sonarr"}`,
				type: "error",
			});
		}
	}

	async function checkSeerrStatus(): Promise<boolean> {
		if (!store.serverFeatures?.overseerr) return false;
		try {
			const res = await req.get<SeerrContentStatus>(`/seerr/status/${type}/${tmdbId}`);
			if (res && res.configured) {
				seerrStatus = res;
				if (res.isAvailable) {
					status = "available";
					return true;
				} else if (res.isProcessing) {
					status = "requested";
					return true;
				} else if (res.isPending) {
					status = "PENDING";
					return true;
				} else if (res.status === 4) {
					status = "available";
					return true;
				}
			}
		} catch (err) {
			console.error("checkSeerrStatus error:", err);
		}
		return false;
	}

	async function requestViaSeerr() {
		try {
			isRequestingSeerr = true;
			await req.post("/seerr/request", {
				mediaType: type,
				mediaId: tmdbId,
				seasons: "all",
			});
			notify({
				type: "success",
				text: "Request submitted to Jellyseerr/Overseerr!",
			});
			status = "PENDING";
		} catch (err) {
			console.error("requestViaSeerr failed:", err);
			notify({
				type: "error",
				text: "Failed to submit request to Jellyseerr/Overseerr",
			});
		} finally {
			isRequestingSeerr = false;
		}
	}

	function handleRequestClick() {
		if (store.serverFeatures?.overseerr) {
			requestViaSeerr();
		} else {
			openRequestModal();
		}
	}

	async function lookForExisting() {
		try {
			const existingRequestResp = await req.get<ArrRequestResponse>(
				`/arr/${type === "movie" ? "rad" : "son"}/request/${tmdbId}`,
			);
			if (existingRequestResp && existingRequestResp.arrId) {
				existingRequest = existingRequestResp;
				getInfo();
			} else if (existingRequestResp) {
				// If no arrId, use status in request (pending, denied, etc)
				console.log(
					"No arrId in request resp.. using request status for btn status if set.",
				);
				if (existingRequestResp.status) {
					status = existingRequestResp.status;
				}
			}
		} catch (err) {
			console.error("ArrRequestButton: lookForExisting failed!", err);
			notify({
				text: `Failed when looking for an existing request for this content`,
				type: "error",
			});
		}
	}

	// Used by Request modals to set existing request data in this component from their request response.
	export function setExistingRequest(r: ArrRequestResponse) {
		if (existingRequest) {
			console.error(
				"ArrRequestButton: setExistingRequest: Existing request has already been defined! Not continuing.",
			);
			return;
		}
		if (!r) {
			console.error(
				"ArrRequestButton: setExistingRequest: No response passed in.",
				r,
			);
			return;
		}
		console.debug("ArrRequestButton: setExistingRequest: Running..", r);
		existingRequest = r;
		if (existingRequest?.status) {
			status = existingRequest?.status;
		}
		getInfo();
	}

	onMount(async () => {
		const handledBySeerr = await checkSeerrStatus();
		if (!handledBySeerr) {
			const hasArr = (type === "movie" && store.serverFeatures?.radarr) || (type === "tv" && store.serverFeatures?.sonarr);
			if (hasArr) {
				lookForExisting();
			}
		}
	});
</script>

{#if isRequestingSeerr}
	<button disabled>Requesting...</button>
{:else if typeof status === "object" || status === "requested"}
	<button
		onclick={getStatus}
		use:tooltip={{
			text:
				status === "requested"
					? "Waiting for approval or download to start"
					: status.status === "downloading"
						? estimatedCompletionIn
							? `Done in ${estimatedCompletionIn}`
							: "Estimation Unavailable"
						: status.status === "paused"
							? "Download has been paused"
							: "",
			pos: "bot",
		}}
	>
		<div><Icon i="refresh" /></div>
		<span>
			{#if status === "requested"}
				Requested
			{:else}
				{status.status}
				{status.status === "downloading"
					? status?.progress
						? `(${status?.progress}%)`
						: ""
					: ""}
			{/if}
		</span>
	</button>
{:else if status === "available" || status === "FOUND"}
	<button disabled>{availableLabel}</button>
{:else if status === "PENDING"}
	<button disabled>Pending</button>
{:else if status === "APPROVED"}
	<button disabled>Approved</button>
{:else if status === "AUTO_APPROVED"}
	<button disabled>Auto Approved</button>
{:else if status === "DENIED"}
	<button disabled>Denied</button>
{:else}
	<button onclick={handleRequestClick}>Request</button>
{/if}

<style lang="scss">
	button {
		max-width: fit-content;
		text-transform: capitalize;
		position: relative;

		& > div {
			display: none;
			position: absolute;
			left: 50%;
			top: 50%;
			transform: translate(-50%, -50%);
		}

		&:hover:not(:focus) {
			& > div {
				display: block;
			}

			& > span {
				visibility: hidden;
			}
		}
	}
</style>
