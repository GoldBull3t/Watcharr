<script lang="ts">
	import { resolve } from "$app/paths";
	import Error from "@/lib/Error.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import RequestMovie from "@/lib/request/RequestMovie.svelte";
	import RequestShow from "@/lib/request/RequestShow.svelte";
	import { baseURL, req } from "@/lib/util/api";
	import { toRelativeDate } from "@/lib/util/helpers";
	import { notify } from "@/lib/util/notify";
	import Icon from "@/lib/Icon.svelte";
	import { type ArrRequestResponse, type Media } from "@/types";

	let allRequests: ArrRequestResponse[] | undefined = $state();
	let showBeingApproved: Media | undefined = $state();
	let movieBeingApproved: Media | undefined = $state();
	let beingApprovedOriginalRequest: ArrRequestResponse | undefined = $state();

	async function getRequests() {
		try {
			allRequests = await req.get<ArrRequestResponse[]>(`/arr/request/`);
			if (allRequests && allRequests.length > 0) {
				allRequests = allRequests.sort((a, b) => {
					if (b.status === "PENDING") return 1;
					if (a.status === "PENDING") return -1;
					return Date.parse(b.updatedAt) - Date.parse(a.updatedAt);
				});
			}
		} catch (err) {
			console.error("Failed to get requests!", err);
			throw err;
		}
	}

	async function deny(r: ArrRequestResponse) {
		try {
			await req.post(`/arr/request/deny/${r.id}`);
			getRequests();
		} catch (err) {
			console.error("Failed to deny request!", err);
			notify({ type: "error", text: "Failed when denying request!" });
		}
	}

	async function approve(r: ArrRequestResponse) {
		console.debug("Approving request:", r);
		if (r.content.type === "tv") {
			showBeingApproved = await req.get<Media>(
				`/content/tv/${r.content.tmdbId}`,
			);
		} else if (r.content.type === "movie") {
			movieBeingApproved = await req.get<Media>(
				`/content/movie/${r.content.tmdbId}`,
			);
		} else {
			notify({
				type: "error",
				text: "Unknown content type, can't continue approval!",
			});
			return;
		}
		beingApprovedOriginalRequest = r;
	}
</script>

<svelte:head>
	<title>Media Requests - Watcharr</title>
</svelte:head>

<div class="content">
	<div class="inner">
		<div class="page-header">
			<div>
				<h2>Media Requests</h2>
				<p class="subtitle">
					Review and process community media requests for Sonarr and Radarr.
				</p>
			</div>
			{#if allRequests && allRequests.length > 0}
				{@const pendingCount = allRequests.filter(
					(r) => r.status === "PENDING",
				).length}
				<div class="header-badges">
					{#if pendingCount > 0}
						<span class="pending-badge">{pendingCount} Pending</span>
					{/if}
					<span class="total-badge">{allRequests.length} Total</span>
				</div>
			{/if}
		</div>

		{#await getRequests()}
			<div class="spinner-wrap">
				<Spinner />
			</div>
		{:then}
			<div class="request-container">
				{#if allRequests && allRequests.length > 0}
					{#each allRequests as r (r.id)}
						{@const mediaLink =
							r.content.type !== "tv_episode"
								? `/${r.content.type}/${r.content.tmdbId}`
								: undefined}
						<div class={`request-card ${r.content.type}`}>
							<div class="card-poster">
								{#if r.content?.poster_path}
									<img
										src={`${baseURL}/img${r.content.poster_path}`}
										alt=""
										loading="lazy"
									/>
								{:else}
									<div class="poster-fallback">
										<Icon i="film" wh={28} />
									</div>
								{/if}
								{#if r.serverName}
									<span
										class={`server-badge ${r.content.type === "tv" ? "badge-sonarr" : "badge-radarr"}`}
									>
										{r.serverName}
									</span>
								{/if}
							</div>

							<div class="card-body">
								<div class="card-main-meta">
									<h3 class="card-title">
										{#if mediaLink}
											<a
												href={resolve(
													mediaLink as `/movie/${string}` | `/tv/${string}`,
												)}
												class="title-link"
											>
												{r.content.title}
											</a>
										{:else}
											<span>{r.content.title}</span>
										{/if}
										{#if r.content.release_date}
											<span class="year"
												>({new Date(
													r.content.release_date,
												).getFullYear()})</span
											>
										{/if}
									</h3>
									{#if r.content.overview}
										<p class="overview">{r.content.overview}</p>
									{/if}
								</div>

								<div class="card-footer">
									<div class="requester-info">
										{#if r.username}
											<span class="by-user">
												Requested by <strong>{r.username}</strong>
											</span>
											<span class="time-dot">&#8226;</span>
											<span class="rel-time"
												>{toRelativeDate(new Date(r.createdAt))}</span
											>
										{/if}
									</div>

									<div class="actions">
										{#if r.status === "PENDING"}
											<button class="btn-decline" onclick={() => deny(r)}>
												<Icon i="close" wh={14} /> Decline
											</button>
											<button class="btn-approve" onclick={() => approve(r)}>
												<Icon i="check" wh={14} /> Approve
											</button>
										{:else}
											<span
												class={`status-pill pill-${r.status.toLowerCase()}`}
											>
												{r.status}
											</span>
										{/if}
									</div>
								</div>
							</div>
						</div>
					{/each}
				{:else}
					<div class="empty-requests">
						<div class="empty-icon-wrap">
							<Icon i="ticket" wh={44} />
						</div>
						<h3>No Media Requests</h3>
						<p>
							All requested titles have been reviewed or no requests have been
							submitted yet.
						</p>
					</div>
				{/if}
			</div>
		{:catch err}
			<Error error={err} pretty="Failed to fetch media requests!" />
		{/await}

		{#if showBeingApproved}
			<RequestShow
				content={showBeingApproved}
				approveMode={true}
				originalRequest={beingApprovedOriginalRequest}
				onClose={() => {
					showBeingApproved = undefined;
					getRequests();
				}}
			/>
		{:else if movieBeingApproved}
			<RequestMovie
				content={movieBeingApproved}
				approveMode={true}
				originalRequest={beingApprovedOriginalRequest}
				onClose={() => {
					movieBeingApproved = undefined;
					getRequests();
				}}
			/>
		{/if}
	</div>
</div>

<style lang="scss">
	.content {
		display: flex;
		width: 100%;
		justify-content: center;
		padding: 24px 20px 80px 20px;

		.inner {
			width: 100%;
			max-width: 860px;
			display: flex;
			flex-direction: column;
			gap: 24px;
		}
	}

	.page-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		flex-wrap: wrap;

		h2 {
			font-size: 26px;
			font-weight: 700;
			letter-spacing: -0.02em;
			color: $text-color;
			margin: 0 0 4px 0;
		}

		.subtitle {
			font-size: 14px;
			color: $text-muted;
			margin: 0;
		}

		.header-badges {
			display: flex;
			gap: 8px;

			.pending-badge {
				font-size: 12px;
				font-weight: 700;
				padding: 4px 12px;
				border-radius: 20px;
				background: rgba(245, 158, 11, 0.15);
				color: #fbbf24;
				border: 1px solid rgba(245, 158, 11, 0.3);
			}

			.total-badge {
				font-size: 12px;
				font-weight: 600;
				padding: 4px 12px;
				border-radius: 20px;
				background: rgba(255, 255, 255, 0.06);
				border: 1px solid rgba(255, 255, 255, 0.1);
				color: $text-muted;
			}
		}
	}

	.spinner-wrap {
		display: flex;
		justify-content: center;
		margin: 40px 0;
	}

	.request-container {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.request-card {
		display: flex;
		gap: 20px;
		padding: 16px;
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 16px;
		transition: all 180ms ease;

		&:hover {
			border-color: rgba(255, 255, 255, 0.14);
			background: rgba(255, 255, 255, 0.045);
		}

		.card-poster {
			position: relative;
			flex-shrink: 0;
			width: 100px;
			height: 150px;
			border-radius: 10px;
			overflow: hidden;
			background: #111;

			img {
				width: 100%;
				height: 100%;
				object-fit: cover;
			}

			.poster-fallback {
				display: flex;
				align-items: center;
				justify-content: center;
				width: 100%;
				height: 100%;
				color: $text-muted;
			}

			.server-badge {
				position: absolute;
				bottom: 4px;
				left: 4px;
				right: 4px;
				text-align: center;
				font-size: 10px;
				font-weight: 700;
				padding: 2px 4px;
				border-radius: 6px;
				text-transform: uppercase;
				letter-spacing: 0.04em;
				overflow: hidden;
				text-overflow: ellipsis;
				white-space: nowrap;

				&.badge-sonarr {
					background: #00b4d8;
					color: #000;
				}

				&.badge-radarr {
					background: #ffb703;
					color: #000;
				}
			}
		}

		.card-body {
			display: flex;
			flex-direction: column;
			justify-content: space-between;
			width: 100%;
			gap: 12px;

			.card-main-meta {
				display: flex;
				flex-direction: column;
				gap: 6px;

				.card-title {
					font-size: 18px;
					font-weight: 700;
					margin: 0;
					line-height: 1.3;

					.title-link {
						color: #fff;
						text-decoration: none;

						&:hover {
							text-decoration: underline;
						}
					}

					.year {
						font-size: 15px;
						font-weight: 400;
						color: $text-muted;
						margin-left: 6px;
					}
				}

				.overview {
					font-size: 13.5px;
					color: rgba(255, 255, 255, 0.7);
					line-height: 1.45;
					margin: 0;
					display: -webkit-box;
					-webkit-line-clamp: 2;
					line-clamp: 2;
					-webkit-box-orient: vertical;
					overflow: hidden;
				}
			}

			.card-footer {
				display: flex;
				align-items: center;
				justify-content: space-between;
				flex-wrap: wrap;
				gap: 12px;
				padding-top: 10px;
				border-top: 1px solid rgba(255, 255, 255, 0.05);

				.requester-info {
					display: flex;
					align-items: center;
					gap: 6px;
					font-size: 12.5px;
					color: $text-muted;

					strong {
						color: $text-color;
					}
				}

				.actions {
					display: flex;
					align-items: center;
					gap: 8px;

					.btn-decline {
						display: inline-flex;
						align-items: center;
						gap: 5px;
						padding: 6px 14px;
						border-radius: 8px;
						font-size: 12.5px;
						font-weight: 500;
						background: rgba(239, 68, 68, 0.1);
						border: 1px solid rgba(239, 68, 68, 0.25);
						color: #fca5a5;
						cursor: pointer;
						transition: all 150ms ease;

						&:hover {
							background: rgba(239, 68, 68, 0.2);
						}
					}

					.btn-approve {
						display: inline-flex;
						align-items: center;
						gap: 5px;
						padding: 6px 14px;
						border-radius: 8px;
						font-size: 12.5px;
						font-weight: 600;
						background: $success;
						border: none;
						color: #fff;
						cursor: pointer;
						transition: all 150ms ease;

						&:hover {
							background: $success-hover;
							transform: translateY(-1px);
						}
					}

					.status-pill {
						font-size: 11px;
						font-weight: 700;
						text-transform: uppercase;
						letter-spacing: 0.04em;
						padding: 4px 10px;
						border-radius: 12px;

						&.pill-approved {
							background: rgba(16, 185, 129, 0.15);
							color: #34d399;
							border: 1px solid rgba(16, 185, 129, 0.3);
						}

						&.pill-denied {
							background: rgba(239, 68, 68, 0.15);
							color: #f87171;
							border: 1px solid rgba(239, 68, 68, 0.3);
						}
					}
				}
			}
		}

		@media screen and (max-width: 580px) {
			flex-direction: column;
			align-items: center;

			.card-poster {
				width: 140px;
				height: 210px;
			}

			.card-body .card-footer {
				flex-direction: column;
				align-items: stretch;

				.actions {
					justify-content: stretch;

					button {
						flex: 1;
						justify-content: center;
					}
				}
			}
		}
	}

	.empty-requests {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 60px 20px;
		text-align: center;

		.empty-icon-wrap {
			display: flex;
			align-items: center;
			justify-content: center;
			width: 80px;
			height: 80px;
			border-radius: 50%;
			background: rgba(255, 255, 255, 0.04);
			border: 1px solid rgba(255, 255, 255, 0.08);
			color: $text-muted;
			margin-bottom: 16px;
		}

		h3 {
			font-size: 20px;
			font-weight: 600;
			margin: 0 0 6px 0;
			color: $text-color;
		}

		p {
			font-size: 14px;
			color: $text-muted;
			max-width: 380px;
			margin: 0;
		}
	}
</style>
