<script lang="ts">
	import { store, clearSecondaryFilters } from "@/store.svelte";
	import Menu, { type MenuConfig } from "../Menu.svelte";

	interface Props {
		menuConfig?: MenuConfig;
	}

	let { menuConfig }: Props = $props();

	function setTypeFilter(t: "all" | "movie" | "tv" | "game") {
		if (t === "all") {
			store.activeFilters = {
				...store.activeFilters,
				type: [],
			};
		} else {
			store.activeFilters = {
				...store.activeFilters,
				type: [t],
			};
		}
	}

	function setYearFilter(y: number | undefined) {
		store.activeFilters = {
			...store.activeFilters,
			year: y,
		};
	}

	function setRatingFilter(r: number | undefined) {
		store.activeFilters = {
			...store.activeFilters,
			minRating: r,
		};
	}

	function setGenreFilter(g: string | undefined) {
		store.activeFilters = {
			...store.activeFilters,
			genre: g,
		};
	}

		function setProductionStatusFilter(ps: string | undefined) {
		store.activeFilters = {
			...store.activeFilters,
			productionStatus: ps,
		};
	}

	function toggleTagFilter(tagId: number) {
		const currentTags = store.activeFilters?.tags || [];
		const exists = currentTags.includes(tagId);
		const newTags = exists
			? currentTags.filter((id) => id !== tagId)
			: [...currentTags, tagId];
		store.activeFilters = {
			...store.activeFilters,
			tags: newTags,
		};
	}

	function toggleTagMode() {
		const newMode = store.activeFilters?.tagMode === "or" ? "and" : "or";
		store.activeFilters = {
			...store.activeFilters,
			tagMode: newMode,
		};
	}

	const years = [
		2026, 2025, 2024, 2023, 2022, 2021, 2020, 2019, 2018, 2017, 2016, 2015,
		2010, 2005, 2000, 1990, 1980,
	];

	const genres = [
		"Action",
		"Adventure",
		"Animation",
		"Comedy",
		"Crime",
		"Documentary",
		"Drama",
		"Family",
		"Fantasy",
		"Horror",
		"Mystery",
		"Romance",
		"Science Fiction",
		"Thriller",
		"War",
		"Western",
	];

	const gameGenres = ["Role-playing (RPG)", "Strategy", "Shooter"];
</script>

<Menu
	conf={menuConfig || {
		width: "270px",
		right: "0px",
		top: "calc(100% + 6px)",
	}}
>
	<div class="filter-popover-content">
		<div class="popover-header">
			<span class="header-title">Filters</span>
			{#if store.hasSecondaryFilters}
				<button
					type="button"
					class="plain reset-all-btn"
					onclick={() => clearSecondaryFilters()}
				>
					Reset all
				</button>
			{/if}
		</div>

		<!-- Media Type Segmented Pills -->
		<div class="filter-group">
			<label class="group-label" for="filter-type-all">Media Type</label>
			<div class="type-pill-group">
				<button
					id="filter-type-all"
					type="button"
					class="type-pill"
					class:active={!store.activeFilters?.type ||
						store.activeFilters.type.length === 0}
					onclick={() => setTypeFilter("all")}
				>
					All
				</button>
				<button
					type="button"
					class="type-pill"
					class:active={store.activeFilters?.type?.includes("movie")}
					onclick={() => setTypeFilter("movie")}
				>
					Movies
				</button>
				<button
					type="button"
					class="type-pill"
					class:active={store.activeFilters?.type?.includes("tv")}
					onclick={() => setTypeFilter("tv")}
				>
					Shows
				</button>
				{#if store.serverFeatures?.games}
					<button
						type="button"
						class="type-pill"
						class:active={store.activeFilters?.type?.includes("game")}
						onclick={() => setTypeFilter("game")}
					>
						Games
					</button>
				{/if}
			</div>
		</div>

		<!-- Release Year -->
		<div class="filter-group">
			<label class="group-label" for="filter-year">Release Year</label>
			<select
				id="filter-year"
				class="filter-select"
				value={store.activeFilters?.year
					? String(store.activeFilters.year)
					: "all"}
				onchange={(e) => {
					const val = (e.target as HTMLSelectElement).value;
					setYearFilter(val === "all" ? undefined : parseInt(val));
				}}
				aria-label="Filter by release year"
			>
				<option value="all">All Years</option>
				{#each years as y}
					<option value={String(y)}>{y}</option>
				{/each}
			</select>
		</div>

		<!-- Minimum Rating -->
		<div class="filter-group">
			<label class="group-label" for="filter-rating">Minimum Rating</label>
			<select
				id="filter-rating"
				class="filter-select"
				value={store.activeFilters?.minRating
					? String(store.activeFilters.minRating)
					: "all"}
				onchange={(e) => {
					const val = (e.target as HTMLSelectElement).value;
					setRatingFilter(val === "all" ? undefined : parseFloat(val));
				}}
				aria-label="Filter by minimum rating"
			>
				<option value="all">All Ratings</option>
				<option value="9">★ 9+ Outstanding</option>
				<option value="8">★ 8+ Great</option>
				<option value="7">★ 7+ Good</option>
				<option value="6">★ 6+ Decent</option>
			</select>
		</div>

		<!-- Genre -->
		<div class="filter-group">
			<label class="group-label" for="filter-genre">Genre</label>
			<select
				id="filter-genre"
				class="filter-select"
				value={store.activeFilters?.genre || "all"}
				onchange={(e) => {
					const val = (e.target as HTMLSelectElement).value;
					setGenreFilter(val === "all" ? undefined : val);
				}}
				aria-label="Filter by genre"
			>
				<option value="all">All Genres</option>
				{#each genres as g}
					<option value={g}>{g}</option>
				{/each}
				{#if store.serverFeatures?.games}
					{#each gameGenres as gg}
						<option value={gg}>{gg}</option>
					{/each}
				{/if}
			</select>
		</div>

		<!-- Production Status (Canceladas pela produtora, etc.) -->
		<div class="filter-group">
			<label class="group-label" for="filter-production-status"
				>Production Status</label
			>
			<select
				id="filter-production-status"
				class="filter-select"
				value={store.activeFilters?.productionStatus || "all"}
				onchange={(e) => {
					const val = (e.target as HTMLSelectElement).value;
					setProductionStatusFilter(val === "all" ? undefined : val);
				}}
				aria-label="Filter by production status"
			>
				<option value="all">All Production Statuses</option>
				<option value="canceled">❌ Cancelada pelos Produtores</option>
				<option value="returning">📺 Em Exibição (Novas Temporadas)</option>
				<option value="ended">✅ Terminada (Com Final)</option>
						</select>
		</div>

		<!-- Tags Filter -->
		{#if store.tags && store.tags.length > 0}
			<div class="filter-group">
				<div class="tag-header-row">
					<span class="group-label">Tags / Listas</span>
					{#if (store.activeFilters?.tags || []).length > 1}
						<button
							type="button"
							class="tag-mode-btn"
							onclick={toggleTagMode}
							title="Alternar entre AND (todas) e OR (qualquer)"
						>
							{store.activeFilters?.tagMode === "or" ? "Modo: OU (Qualquer)" : "Modo: E (Todas)"}
						</button>
					{/if}
				</div>
				<div class="tags-pill-wrap small-scrollbar">
					{#each store.tags as t}
						{@const isSelected = (store.activeFilters?.tags || []).includes(t.id)}
						<button
							type="button"
							class="tag-chip"
							class:active={isSelected}
							onclick={() => toggleTagFilter(t.id)}
							title={isSelected ? `Remover filtro da tag ${t.name}` : `Filtrar por tag ${t.name}`}
						>
							<span class="tag-color-dot" style={t.color ? `background-color: ${t.color}` : ""}></span>
							<span class="tag-name">{t.name}</span>
							{#if isSelected}
								<span class="tag-check">&#10004;</span>
							{/if}
						</button>
					{/each}
				</div>
			</div>
		{/if}
	</div>
</Menu>

<style lang="scss">
	.filter-popover-content {
		display: flex;
		flex-flow: column;
		gap: 12px;
		padding: 4px;

		.popover-header {
			display: flex;
			align-items: center;
			justify-content: space-between;
			padding-bottom: 6px;
			border-bottom: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));

			.header-title {
				font-size: 13px;
				font-weight: 700;
				color: $text-color;
				letter-spacing: 0.02em;
			}

			.reset-all-btn {
				font-size: 11.5px;
				font-weight: 600;
				color: #eab308;
				cursor: pointer;
				padding: 2px 4px;
				border-radius: 4px;
				transition: opacity 140ms ease;

				&:hover {
					text-decoration: underline;
					opacity: 0.85;
				}
			}
		}

		.filter-group {
			display: flex;
			flex-flow: column;
			gap: 5px;

			.group-label {
				font-size: 11px;
				font-weight: 600;
				text-transform: uppercase;
				letter-spacing: 0.05em;
				color: $text-color-accent;
			}

			.type-pill-group {
				display: grid;
				grid-template-columns: repeat(3, 1fr);
				gap: 4px;
				background-color: rgba(0, 0, 0, 0.25);
				padding: 3px;
				border-radius: 8px;
				border: 1px solid var(--card-border, rgba(255, 255, 255, 0.06));

				&:has(button:nth-child(4)) {
					grid-template-columns: repeat(4, 1fr);
				}

				.type-pill {
					padding: 5px 4px;
					font-size: 11.5px;
					font-weight: 500;
					text-align: center;
					border-radius: 6px;
					border: none;
					background: transparent;
					color: $text-color-accent;
					cursor: pointer;
					transition: all 130ms ease;

					&:hover {
						color: $text-color;
						background-color: rgba(255, 255, 255, 0.05);
					}

					&.active {
						color: #ffffff;
						background-color: $accent-color;
						font-weight: 600;
						box-shadow: 0 1px 4px rgba(0, 0, 0, 0.25);
					}
				}
			}

			.filter-select {
				appearance: none;
				-webkit-appearance: none;
				-moz-appearance: none;
				width: 100%;
				padding: 7px 28px 7px 10px;
				border-radius: 7px;
				border: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
				background-color: rgba(18, 22, 30, 0.95);
				background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='rgba(255,255,255,0.45)' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
				background-repeat: no-repeat;
				background-position: right 9px center;
				color: $text-color;
				color-scheme: dark;
				font-size: 12px;
				font-weight: 500;
				outline: none;
				cursor: pointer;
				transition:
					border-color 140ms ease,
					box-shadow 140ms ease;

				option {
					background-color: #12161f;
					color: #f1f5f9;
				}

				&:hover {
					border-color: rgba(255, 255, 255, 0.2);
				}

				&:focus {
					border-color: #eab308;
					box-shadow: 0 0 0 2px rgba(234, 179, 8, 0.2);
				}
			}
			.tag-header-row {
			display: flex;
			align-items: center;
			justify-content: space-between;
			gap: 6px;

			.tag-mode-btn {
				font-size: 10px;
				font-weight: 700;
				color: #eab308;
				background: rgba(234, 179, 8, 0.12);
				border: 1px solid rgba(234, 179, 8, 0.25);
				border-radius: 4px;
				padding: 1px 6px;
				cursor: pointer;
				transition: all 120ms ease;

				&:hover {
					background: rgba(234, 179, 8, 0.22);
				}
			}
		}

		.tags-pill-wrap {
			display: flex;
			flex-wrap: wrap;
			gap: 5px;
			max-height: 120px;
			overflow-y: auto;
			padding-right: 2px;

			.tag-chip {
				display: inline-flex;
				align-items: center;
				gap: 5px;
				padding: 3px 8px;
				border-radius: 6px;
				border: 1px solid rgba(255, 255, 255, 0.1);
				background: rgba(255, 255, 255, 0.05);
				color: #e2e8f0;
				font-size: 11px;
				font-weight: 500;
				cursor: pointer;
				transition: all 120ms ease;

				&:hover {
					border-color: rgba(255, 255, 255, 0.25);
					background: rgba(255, 255, 255, 0.1);
				}

				&.active {
					border-color: #eab308;
					background: rgba(234, 179, 8, 0.16);
					color: #fef08a;
					font-weight: 600;
				}

				.tag-color-dot {
					width: 7px;
					height: 7px;
					border-radius: 50%;
					background-color: #94a3b8;
					flex-shrink: 0;
				}

				.tag-name {
					max-width: 130px;
					white-space: nowrap;
					overflow: hidden;
					text-overflow: ellipsis;
				}

				.tag-check {
					font-size: 9px;
					color: #eab308;
				}
			}
		}
	}
}
</style>
