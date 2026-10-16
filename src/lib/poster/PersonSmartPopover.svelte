<script lang="ts">
	import { resolve } from "$app/paths";
	import Icon from "@/lib/Icon.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import { getPersonCreditsWithWatched } from "./personCache";
	import { MediaTypeE, type Media } from "@/types";
	import { onMount } from "svelte";

	interface Props {
		personId: number;
		personName?: string;
		role?: string;
	}

	let { personId, personName, role }: Props = $props();

	let credits = $state<Media[] | null>(null);
	let loading = $state(true);

	let inLibrary = $derived.by(() => {
		if (!credits) return [];
		return credits.filter((c) => !!c.watched);
	});

	let otherKnown = $derived.by(() => {
		if (!credits) return [];
		return credits.filter((c) => !c.watched).slice(0, 3);
	});

	onMount(() => {
		let isMounted = true;
		getPersonCreditsWithWatched(personId).then((res) => {
			if (isMounted) {
				credits = res;
				loading = false;
			}
		});
		return () => {
			isMounted = false;
		};
	});
</script>

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<div
	class="person-smart-popover"
	role="tooltip"
	onclick={(e) => e.stopPropagation()}
	onkeydown={(e) => e.stopPropagation()}
>
	<div class="popover-header">
		<div class="header-titles">
			<span class="person-title">{personName || "Creator"}</span>
			{#if role}
				<span class="person-role">{role}</span>
			{/if}
		</div>

		{#if !loading}
			{#if inLibrary.length > 0}
				<div class="library-badge" title="Títulos na tua lista">
					<Icon i="star" wh={11} />
					<span>{inLibrary.length} na tua lista</span>
				</div>
			{:else}
				<div class="filmography-badge">Filmografia</div>
			{/if}
		{/if}
	</div>

	<div class="popover-body">
		{#if loading}
			<div class="loading-state">
				<Spinner />
				<span>A procurar na biblioteca...</span>
			</div>
		{:else if inLibrary.length > 0}
			<div class="library-items-list">
				{#each inLibrary.slice(0, 3) as item (item.ids?.tmdb || item.name)}
					{@const mediaType =
						item.type === MediaTypeE.tmdbShow ? "tv" : "movie"}
					<a
						href={resolve(`/${mediaType}/${item.ids?.tmdb}`)}
						class="library-item-row"
					>
						<div class="mini-thumb">
							{#if item.extPosterPath}
								<img
									src={`https://image.tmdb.org/t/p/w92${item.extPosterPath}`}
									alt=""
									loading="lazy"
								/>
							{:else}
								<span class="no-thumb">🎬</span>
							{/if}
						</div>
						<div class="item-text">
							<span class="item-title">{item.name}</span>
							<div class="item-meta">
								{#if item.watched?.rating}
									<span class="item-rating">★ {item.watched.rating}</span>
								{/if}
								{#if item.watched?.status}
									<span
										class="item-status status-{item.watched.status.toLowerCase()}"
									>
										{item.watched.status}
									</span>
								{/if}
							</div>
						</div>
					</a>
				{/each}
				{#if inLibrary.length > 3}
					<div class="more-hint">
						+{inLibrary.length - 3} outros títulos na tua lista
					</div>
				{/if}
			</div>
		{:else}
			<div class="no-library-state">
				<p class="no-library-msg">Nenhum título deste criador na tua lista ainda.</p>
				{#if otherKnown.length > 0}
					<div class="known-section">
						<span class="known-header">Títulos mais conhecidos:</span>
						<div class="known-chips">
							{#each otherKnown as item}
								{@const mediaType =
									item.type === MediaTypeE.tmdbShow ? "tv" : "movie"}
								<a
									href={resolve(`/${mediaType}/${item.ids?.tmdb}`)}
									class="known-chip"
								>
									{item.name}
								</a>
							{/each}
						</div>
					</div>
				{/if}
			</div>
		{/if}
	</div>

	<div class="popover-footer">
		<a href={resolve(`/person/${personId}`)} class="view-profile-link">
			Ver filmografia completa
			<Icon i="arrow" facing="right" wh={11} />
		</a>
	</div>
</div>

<style lang="scss">
	.person-smart-popover {
		position: absolute;
		bottom: calc(100% + 12px);
		left: 50%;
		transform: translateX(-50%);
		width: 280px;
		max-width: 90vw;
		border-radius: 14px;
		background: rgba(14, 18, 28, 0.96);
		border: 1px solid rgba(245, 158, 11, 0.35);
		box-shadow:
			0 16px 40px -4px rgba(0, 0, 0, 0.8),
			0 0 20px rgba(245, 158, 11, 0.12);
		backdrop-filter: blur(24px);
		-webkit-backdrop-filter: blur(24px);
		z-index: 1000;
		display: flex;
		flex-flow: column;
		overflow: hidden;
		animation: popoverFade 180ms cubic-bezier(0.16, 1, 0.3, 1);
		pointer-events: auto;
		text-align: left;

		&::after {
			content: "";
			position: absolute;
			top: 100%;
			left: 50%;
			transform: translateX(-50%);
			border: 7px solid transparent;
			border-top-color: rgba(14, 18, 28, 0.96);
		}
	}

	.popover-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
		padding: 12px 14px 10px 14px;
		border-bottom: 1px solid rgba(255, 255, 255, 0.07);
		background: rgba(255, 255, 255, 0.02);

		.header-titles {
			display: flex;
			flex-flow: column;
			gap: 1px;
			min-width: 0;

			.person-title {
				font-size: 13.5px;
				font-weight: 700;
				color: #ffffff;
				white-space: nowrap;
				overflow: hidden;
				text-overflow: ellipsis;
			}

			.person-role {
				font-size: 11px;
				color: rgba(255, 255, 255, 0.55);
				white-space: nowrap;
				overflow: hidden;
				text-overflow: ellipsis;
			}
		}

		.library-badge {
			display: inline-flex;
			align-items: center;
			gap: 4px;
			padding: 2.5px 7px;
			border-radius: 9999px;
			background: rgba(245, 158, 11, 0.18);
			border: 1px solid rgba(245, 158, 11, 0.45);
			color: #fbbf24;
			font-size: 10.5px;
			font-weight: 700;
			white-space: nowrap;
			flex-shrink: 0;
			box-shadow: 0 0 10px rgba(245, 158, 11, 0.2);

			:global(svg) {
				fill: #f59e0b;
			}
		}

		.filmography-badge {
			font-size: 10.5px;
			padding: 2px 6px;
			border-radius: 4px;
			background: rgba(255, 255, 255, 0.08);
			color: rgba(255, 255, 255, 0.6);
			white-space: nowrap;
			flex-shrink: 0;
		}
	}

	.popover-body {
		padding: 10px 14px;

		.loading-state {
			display: flex;
			align-items: center;
			justify-content: center;
			gap: 8px;
			padding: 14px 0;
			font-size: 11.5px;
			color: rgba(255, 255, 255, 0.6);
		}

		.library-items-list {
			display: flex;
			flex-flow: column;
			gap: 6px;

			.library-item-row {
				display: flex;
				align-items: center;
				gap: 10px;
				padding: 5px 8px;
				border-radius: 8px;
				background: rgba(255, 255, 255, 0.035);
				border: 1px solid rgba(255, 255, 255, 0.05);
				text-decoration: none;
				transition: all 130ms ease;

				&:hover {
					background: rgba(255, 255, 255, 0.08);
					border-color: rgba(245, 158, 11, 0.35);
					transform: translateX(2px);
				}

				.mini-thumb {
					width: 28px;
					height: 38px;
					border-radius: 4px;
					overflow: hidden;
					background: #000000;
					flex-shrink: 0;

					img {
						width: 100%;
						height: 100%;
						object-fit: cover;
					}

					.no-thumb {
						display: flex;
						align-items: center;
						justify-content: center;
						height: 100%;
						font-size: 12px;
					}
				}

				.item-text {
					display: flex;
					flex-flow: column;
					gap: 2px;
					min-width: 0;
					flex: 1;

					.item-title {
						font-size: 12px;
						font-weight: 600;
						color: #ffffff;
						white-space: nowrap;
						overflow: hidden;
						text-overflow: ellipsis;
					}

					.item-meta {
						display: flex;
						align-items: center;
						gap: 6px;
						font-size: 10.5px;

						.item-rating {
							color: #fbbf24;
							font-weight: 700;
						}

						.item-status {
							color: rgba(255, 255, 255, 0.5);
							font-size: 10px;
							text-transform: capitalize;

							&.status-finished {
								color: #34d399;
							}
							&.status-watching {
								color: #60a5fa;
							}
							&.status-planned {
								color: #fbbf24;
							}
						}
					}
				}
			}

			.more-hint {
				text-align: center;
				font-size: 10.5px;
				color: rgba(255, 255, 255, 0.45);
				padding: 3px 0 0 0;
			}
		}

		.no-library-state {
			display: flex;
			flex-flow: column;
			gap: 8px;

			.no-library-msg {
				margin: 0;
				font-size: 11.5px;
				color: rgba(255, 255, 255, 0.55);
				line-height: 1.35;
			}

			.known-section {
				display: flex;
				flex-flow: column;
				gap: 4px;

				.known-header {
					font-size: 10.5px;
					font-weight: 600;
					color: rgba(255, 255, 255, 0.45);
					text-transform: uppercase;
					letter-spacing: 0.04em;
				}

				.known-chips {
					display: flex;
					flex-wrap: wrap;
					gap: 4px;

					.known-chip {
						font-size: 10.5px;
						padding: 2.5px 7px;
						border-radius: 6px;
						background: rgba(255, 255, 255, 0.05);
						border: 1px solid rgba(255, 255, 255, 0.08);
						color: rgba(255, 255, 255, 0.85);
						text-decoration: none;
						transition: all 130ms ease;

						&:hover {
							background: rgba(245, 158, 11, 0.15);
							border-color: rgba(245, 158, 11, 0.35);
							color: #fbbf24;
						}
					}
				}
			}
		}
	}

	.popover-footer {
		padding: 8px 14px 10px 14px;
		border-top: 1px solid rgba(255, 255, 255, 0.06);
		background: rgba(0, 0, 0, 0.15);

		.view-profile-link {
			display: inline-flex;
			align-items: center;
			gap: 4px;
			font-size: 11.5px;
			font-weight: 600;
			color: #fbbf24;
			text-decoration: none;
			transition: all 130ms ease;

			&:hover {
				color: #ffffff;
				transform: translateX(2px);
			}
		}
	}

	@keyframes popoverFade {
		from {
			opacity: 0;
			transform: translate(-50%, 6px) scale(0.96);
		}
		to {
			opacity: 1;
			transform: translate(-50%, 0) scale(1);
		}
	}
</style>
