<script lang="ts">
	import { goto } from "$app/navigation";
	import { resolve } from "$app/paths";
	import Icon from "@/lib/Icon.svelte";
	import { clearActiveFilters, store } from "@/store.svelte";
</script>

<div class="empty-list" role="region" aria-label="Empty library">
	<div class="empty-icon-wrap">
		<Icon i={store.hasActiveFilters ? "filter-circle" : "reel"} wh={48} />
	</div>
	<h2 class="norm">
		{#if store.hasActiveFilters}
			No Matching Media
		{:else}
			Your Library is Empty
		{/if}
	</h2>
	<p class="empty-desc">
		{#if store.hasActiveFilters}
			No titles match your active filters or search terms. Try clearing them to
			see all your media.
		{:else}
			Start building your personal collection by discovering movies, TV shows,
			and games.
		{/if}
	</p>
	<div class="empty-actions">
		{#if store.hasActiveFilters}
			<button
				type="button"
				class="btn primary"
				onclick={() => clearActiveFilters()}
			>
				Clear All Filters
			</button>
		{:else}
			<button
				type="button"
				class="btn primary"
				onclick={() => goto(resolve("/discover"))}
			>
				<Icon i="compass" wh={16} />
				Discover Media
			</button>
			<button
				type="button"
				class="btn secondary"
				onclick={() => goto(resolve("/import"))}
			>
				<Icon i="document" wh={16} />
				Import List
			</button>
		{/if}
	</div>
</div>

<style lang="scss">
	.empty-list {
		display: flex;
		flex-flow: column;
		align-items: center;
		justify-content: center;
		padding: 56px 28px;
		margin: 36px auto;
		max-width: 480px;
		width: 100%;
		text-align: center;
		background: rgba(16, 21, 30, 0.7);
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
			width: 80px;
			height: 80px;
			border-radius: 50%;
			background: radial-gradient(
				circle,
				rgba(245, 158, 11, 0.16) 0%,
				rgba(245, 158, 11, 0.03) 70%
			);
			border: 1px solid rgba(245, 158, 11, 0.25);
			color: #f59e0b;
			margin-bottom: 18px;
			box-shadow: 0 0 24px rgba(245, 158, 11, 0.12);
		}

		h2 {
			font-size: 20px;
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
				width: max-content;
				padding: 8px 18px;
				font-size: 13px;
			}
		}
	}
</style>
