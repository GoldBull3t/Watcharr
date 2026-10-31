<script lang="ts">
	import { req } from "@/lib/util/api";
	import type { FranchiseMatch } from "@/types";
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import Icon from "@/lib/Icon.svelte";

	interface Props {
		mediaType: "movie" | "tv";
		tmdbId: number;
		collectionId?: number;
		collectionName?: string;
		hasCollectionBlock?: boolean;
	}

	let {
		mediaType,
		tmdbId,
		collectionId,
		collectionName,
		hasCollectionBlock = false,
	}: Props = $props();

	let match = $state<FranchiseMatch | undefined>();

	$effect(() => {
		if (tmdbId) {
			req
				.get<FranchiseMatch>(`/franchise/check/${mediaType}/${tmdbId}`)
				.then((res) => {
					match = res;
				})
				.catch((err) => {
					console.warn("Failed checking franchise for content:", err);
				});
		}
	});
</script>

{#if match && match.inFranchise && match.franchiseId}
	<div class="franchise-universe-banner">
		<div class="banner-left">
			<span class="banner-icon">
				<Icon i="compass" wh={18} />
			</span>
			<div class="banner-texts">
				<span class="banner-kicker">Universo Cinematográfico & Franquia</span>
				<div class="name-row">
					<strong class="universe-name">{match.franchiseName}</strong>
					{#if match.phaseOrEra}
						<span class="phase-pill">{match.phaseOrEra}</span>
					{/if}
				</div>
			</div>
		</div>

		<button
			type="button"
			class="explore-btn"
			onclick={() => goto(resolve(`/franchise/${match?.franchiseId}`))}
		>
			<Icon i="clock" wh={14} />
			<span>Ver Linha Cronológica</span>
			<span class="arrow-sym">→</span>
		</button>
	</div>
{:else if collectionId && !hasCollectionBlock}
	<div class="franchise-universe-banner tmdb-col-banner">
		<div class="banner-left">
			<span class="banner-icon">
				<Icon i="film" wh={18} />
			</span>
			<div class="banner-texts">
				<span class="banner-kicker">Coleção & Franquia TMDB</span>
				<strong class="universe-name">{collectionName || "Ver Franquia"}</strong
				>
			</div>
		</div>

		<button
			type="button"
			class="explore-btn"
			onclick={() => goto(resolve(`/franchise/tmdb-${collectionId}`))}
		>
			<Icon i="clock" wh={14} />
			<span>Ver Rota da Franquia</span>
			<span class="arrow-sym">→</span>
		</button>
	</div>
{/if}

<style lang="scss">
	.franchise-universe-banner {
		display: flex;
		align-items: center;
		justify-content: space-between;
		width: 100%;
		gap: 16px;
		background: rgba(15, 19, 28, 0.86);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 10px;
		padding: 14px 18px;
		margin: 0;
		transition: border-color 150ms ease;

		&:hover {
			border-color: rgba(255, 255, 255, 0.22);
		}

		@media (max-width: 600px) {
			flex-direction: column;
			align-items: flex-start;
			gap: 12px;

			.explore-btn {
				width: 100%;
				justify-content: center;
			}
		}
	}

	.banner-left {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.banner-icon {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 36px;
		height: 36px;
		border-radius: 8px;
		background: rgba(245, 158, 11, 0.12);
		border: 1px solid rgba(245, 158, 11, 0.28);
		color: $brand-color;
		flex-shrink: 0;
	}

	.banner-texts {
		display: flex;
		flex-direction: column;
		gap: 3px;
	}

	.banner-kicker {
		font-size: 0.73rem;
		text-transform: uppercase;
		letter-spacing: 0.6px;
		color: #8b949e;
		font-weight: 600;
	}

	.name-row {
		display: flex;
		align-items: center;
		gap: 10px;
		flex-wrap: wrap;
	}

	.universe-name {
		font-size: 1.05rem;
		color: #f0f6fc;
		font-weight: 700;
	}

	.phase-pill {
		background: rgba(56, 139, 253, 0.2);
		border: 1px solid rgba(56, 139, 253, 0.35);
		color: #58a6ff;
		padding: 2px 8px;
		border-radius: 6px;
		font-size: 0.72rem;
		font-weight: 600;
	}

	.explore-btn {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		background: rgba(31, 111, 235, 0.25);
		border: 1px solid rgba(56, 139, 253, 0.45);
		color: #58a6ff;
		padding: 7px 14px;
		border-radius: 8px;
		font-size: 0.85rem;
		font-weight: 600;
		cursor: pointer;
		white-space: nowrap;
		transition: all 0.2s ease;

		&:hover {
			background: #1f6feb;
			color: #ffffff;
			border-color: #58a6ff;
			transform: translateY(-1px);

			.arrow-sym {
				transform: translateX(3px);
			}
		}
	}

	.arrow-sym {
		transition: transform 0.2s ease;
		display: inline-block;
	}
</style>
