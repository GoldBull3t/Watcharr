<script lang="ts">
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import Icon from "@/lib/Icon.svelte";
	import { clearActiveFilters, store } from "@/store.svelte";
	import { t } from "@/lib/i18n";
	import { req } from "@/lib/util/api";
	import type { Media, PaginationResponse } from "@/types";
	import Poster from "@/lib/poster/Poster.svelte";
	import { onMount } from "svelte";

	let trendingItems: Media[] = $state([]);
	let trendingLoading = $state(true);

	onMount(async () => {
		if (store.hasActiveFilters) {
			trendingLoading = false;
			return;
		}
		try {
			const res = await req.get<PaginationResponse<Media, undefined>>(
				"/discover",
				{
					params: { page: 1, type: "all", filter: "trending" },
				},
			);
			if (res?.results) {
				trendingItems = res.results.slice(0, 6);
			}
		} catch (e) {
			console.warn("Failed to load trending suggestions for empty state", e);
		} finally {
			trendingLoading = false;
		}
	});
</script>

<div class="empty-library-container" role="region" aria-label="Empty library">
	<div class="empty-list">
		<div class="empty-icon-wrap">
			<Icon i={store.hasActiveFilters ? "filter-circle" : "reel"} wh={48} />
		</div>
		<h2 class="norm">
			{#if store.hasActiveFilters}
				{t("empty.no_match_title")}
			{:else}
				{t("empty.library_title")}
			{/if}
		</h2>
		<p class="empty-desc">
			{#if store.hasActiveFilters}
				{t("empty.no_match_desc")}
			{:else}
				{t("empty.library_desc")}
			{/if}
		</p>
		<div class="empty-actions">
			{#if store.hasActiveFilters}
				<button
					type="button"
					class="btn primary"
					onclick={() => clearActiveFilters()}
				>
					{t("empty.clear_filters")}
				</button>
			{:else}
				<button
					type="button"
					class="btn primary"
					onclick={() => goto(resolve("/discover"))}
				>
					<Icon i="compass" wh={16} />
					{t("empty.discover_media")}
				</button>
				<button
					type="button"
					class="btn secondary"
					onclick={() => goto(resolve("/import"))}
				>
					<Icon i="document" wh={16} />
					{t("empty.import_list")}
				</button>
			{/if}
		</div>
	</div>

	{#if !store.hasActiveFilters && trendingItems.length > 0}
		<div class="trending-section">
			<div class="trending-header">
				<Icon i="sparkles" wh={18} />
				<h3>{t("empty.trending_title")}</h3>
			</div>
			<div class="trending-row">
				{#each trendingItems as item, i (item.ids.tmdb ?? item.ids.igdb ?? i)}
					<Poster media={item} fluidSize={true} small={true} />
				{/each}
			</div>
		</div>
	{/if}
</div>

<style lang="scss">
	.empty-library-container {
		display: flex;
		flex-flow: column;
		align-items: center;
		justify-content: center;
		width: 100%;
		max-width: 1100px;
		margin: 32px auto 60px auto;
		padding: 0 16px;
		grid-column: 1 / -1;
	}

	.empty-list {
		display: flex;
		flex-flow: column;
		align-items: center;
		justify-content: center;
		padding: 48px 32px;
		max-width: 480px;
		width: 100%;
		text-align: center;
		background: rgba(16, 21, 30, 0.75);
		border: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
		border-radius: 18px;
		box-shadow: 0 16px 40px -8px rgba(0, 0, 0, 0.45);
		backdrop-filter: blur(16px);
		-webkit-backdrop-filter: blur(16px);
		animation: emptyFadeIn 220ms cubic-bezier(0.16, 1, 0.3, 1) forwards;

		@keyframes emptyFadeIn {
			from {
				opacity: 0;
				transform: translateY(8px);
			}
			to {
				opacity: 1;
				transform: translateY(0);
			}
		}

		.empty-icon-wrap {
			display: flex;
			align-items: center;
			justify-content: center;
			width: 68px;
			height: 68px;
			border-radius: 50%;
			background: rgba(255, 255, 255, 0.04);
			border: 1px solid rgba(255, 255, 255, 0.08);
			color: $brand-color;
			margin-bottom: 18px;
		}

		h2 {
			font-size: 21px;
			font-weight: 700;
			letter-spacing: -0.02em;
			color: $text-color;
			margin-bottom: 8px;
		}

		.empty-desc {
			font-size: 13.5px;
			color: $text-color-accent;
			line-height: 1.55;
			margin-bottom: 24px;
			max-width: 380px;
		}

		.empty-actions {
			display: flex;
			gap: 12px;
			flex-wrap: wrap;
			justify-content: center;

			button {
				display: inline-flex;
				align-items: center;
				gap: 8px;
				width: max-content;
				padding: 9px 20px;
				font-size: 13px;
				font-weight: 600;
				border-radius: 10px;
			}
		}
	}

	.trending-section {
		width: 100%;
		margin-top: 48px;
		animation: emptyFadeIn 350ms cubic-bezier(0.16, 1, 0.3, 1) forwards;

		.trending-header {
			display: flex;
			align-items: center;
			gap: 8px;
			margin-bottom: 16px;
			color: #f59e0b;

			h3 {
				font-size: 15px;
				font-weight: 700;
				text-transform: uppercase;
				letter-spacing: 0.05em;
				color: $text-color;
				margin: 0;
			}
		}

		.trending-row {
			display: grid;
			grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
			gap: 14px;
			width: 100%;
		}
	}
</style>
