<script lang="ts">
	import type { Media, MediaProvider } from "@/types";
	import ProviderIcon from "./ProviderIcon.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import { store } from "@/store.svelte";
	import { req } from "@/lib/util/api";
	import {
		POPULAR_REGIONS,
		getAllRegions,
		getCountryName,
		type RegionInfo,
	} from "./watchProviders";
	import { onMount } from "svelte";

	interface Props {
		providers?: MediaProvider[];
		fullListLink?: string;
		fullListLinkText?: string;
		mediaType?: "movie" | "tv" | "game";
		mediaId?: number;
		mediaTitle?: string;
	}

	let {
		providers = [],
		fullListLink,
		fullListLinkText = "JustWatch",
		mediaType = "movie",
		mediaId,
		mediaTitle,
	}: Props = $props();

	// Active country selection (persisted in localStorage or user settings)
	let activeCountry = $state<string>(
		(typeof localStorage !== "undefined" &&
			localStorage.getItem("watcharr_watch_provider_country")) ||
			store.userSettings?.country ||
			"PT",
	);

	let activeCountryName = $derived(getCountryName(activeCountry));

	let displayedProviders = $state<MediaProvider[]>([]);
	let displayedFullListLink = $state<string | undefined>();
	let loading = $state(false);

	// Country selector dropdown state
	let dropdownOpen = $state(false);
	let searchQuery = $state("");
	let allRegions = $state<RegionInfo[]>(POPULAR_REGIONS);
	let selectorElement = $state<HTMLDivElement | undefined>();

	// In-memory cache for fetched providers: key = `${mediaType}-${mediaId}-${countryCode}`
	const providersCache = new Map<
		string,
		{ providers: MediaProvider[]; fullListLink?: string }
	>();

	// Synchronize when initial props change
	$effect(() => {
		const currProviders = providers || [];
		const key = `${mediaType}-${mediaId}-${activeCountry}`;
		providersCache.set(key, {
			providers: currProviders,
			fullListLink,
		});
		displayedProviders = currProviders;
		displayedFullListLink = fullListLink;
	});

	onMount(() => {
		getAllRegions().then((regs) => {
			if (regs?.length) {
				allRegions = regs;
			}
		});

		function handleDocumentClick(e: MouseEvent) {
			if (
				dropdownOpen &&
				selectorElement &&
				!selectorElement.contains(e.target as Node)
			) {
				dropdownOpen = false;
			}
		}

		function handleKeydown(e: KeyboardEvent) {
			if (dropdownOpen && e.key === "Escape") {
				dropdownOpen = false;
			}
		}

		document.addEventListener("click", handleDocumentClick);
		document.addEventListener("keydown", handleKeydown);

		return () => {
			document.removeEventListener("click", handleDocumentClick);
			document.removeEventListener("keydown", handleKeydown);
		};
	});

	let filteredRegions = $derived.by(() => {
		if (!searchQuery.trim()) return allRegions;
		const q = searchQuery.toLowerCase().trim();
		return allRegions.filter(
			(r) =>
				r.name.toLowerCase().includes(q) || r.code.toLowerCase().includes(q),
		);
	});

	async function selectCountry(code: string) {
		if (code === activeCountry) {
			dropdownOpen = false;
			return;
		}

		activeCountry = code;
		if (typeof localStorage !== "undefined") {
			localStorage.setItem("watcharr_watch_provider_country", code);
		}
		dropdownOpen = false;
		searchQuery = "";

		const key = `${mediaType}-${mediaId}-${code}`;
		if (providersCache.has(key)) {
			const cached = providersCache.get(key)!;
			displayedProviders = cached.providers;
			displayedFullListLink = cached.fullListLink;
			return;
		}

		if (!mediaType || !mediaId || mediaType === "game") return;

		loading = true;
		try {
			const resp = await req.get<Media>(`/content/${mediaType}/${mediaId}`, {
				params: { region: code },
			});
			if (resp) {
				const resProviders = resp.providers || [];
				const resLink = resp.providersFullListLink;
				providersCache.set(key, {
					providers: resProviders,
					fullListLink: resLink,
				});
				if (activeCountry === code) {
					displayedProviders = resProviders;
					displayedFullListLink = resLink;
				}
			}
		} catch (err) {
			console.error("Failed to load providers for region", code, err);
		} finally {
			loading = false;
		}
	}
</script>

<div class="streaming-providers-card">
	<div class="card-header">
		<div class="header-left">
			<span class="header-title">
				{#if mediaType === "game"}
					Onde Jogar & Comprar
				{:else}
					Onde Assistir
				{/if}
			</span>

			{#if mediaType === "movie" || mediaType === "tv"}
				<div class="country-selector-wrap" bind:this={selectorElement}>
					<button
						type="button"
						class="country-btn"
						class:active={dropdownOpen}
						onclick={() => (dropdownOpen = !dropdownOpen)}
						title="Alterar país para verificar catálogo de streaming"
					>
						<span class="country-code">{activeCountry}</span>
						<span class="chevron">▾</span>
					</button>

					{#if dropdownOpen}
						<div class="country-dropdown-popover">
							<div class="popover-search">
								<input
									type="text"
									placeholder="Pesquisar país..."
									bind:value={searchQuery}
									onclick={(e) => e.stopPropagation()}
									onkeydown={(e) => e.stopPropagation()}
								/>
							</div>

							<div class="popular-chips">
								{#each POPULAR_REGIONS.slice(0, 8) as r}
									<button
										type="button"
										class="popular-chip"
										class:selected={r.code === activeCountry}
										onclick={() => selectCountry(r.code)}
									>
										<span class="chip-code">{r.code}</span>
									</button>
								{/each}
							</div>

							<div class="regions-list">
								{#each filteredRegions as reg (reg.code)}
									<button
										type="button"
										class="region-row"
										class:selected={reg.code === activeCountry}
										onclick={() => selectCountry(reg.code)}
									>
										<span class="row-name">{reg.name}</span>
										<span class="row-code">{reg.code}</span>
									</button>
								{/each}
								{#if filteredRegions.length === 0}
									<div class="empty-regions">Nenhum país encontrado</div>
								{/if}
							</div>
						</div>
					{/if}
				</div>
			{/if}
		</div>

		{#if displayedFullListLink}
			<a
				href={displayedFullListLink}
				class="justwatch-badge"
				target="_blank"
				rel="noreferrer noopener external"
				title="Ver catálogo e opções de compra/aluguer completas no JustWatch"
			>
				<span class="justwatch-name">{fullListLinkText}</span>
				<span class="justwatch-arrow">↗</span>
			</a>
		{/if}
	</div>

	<div class="card-body">
		{#if loading}
			<div class="providers-loading">
				<Spinner />
				<span class="loading-text">
					A procurar catálogo em {activeCountryName} ({activeCountry})...
				</span>
			</div>
		{:else if displayedProviders && displayedProviders.length > 0}
			<div class="providers-badges-grid">
				{#each displayedProviders as provider (provider.name + (provider.type || ""))}
					<ProviderIcon
						i={provider.name}
						href={provider.link}
						{mediaTitle}
						type={provider.type}
						logoPath={provider.logoPath}
					/>
				{/each}
			</div>
		{:else if mediaType === "movie" || mediaType === "tv"}
			<div class="no-providers-alert">
				<span class="no-providers-text">
					Sem plataformas de streaming registadas em {activeCountryName} ({activeCountry}).
				</span>
				<div class="alternative-suggestions">
					<span class="suggest-label">Verificar em:</span>
					{#each ["US", "GB", "ES", "BR", "FR"].filter((c) => c !== activeCountry) as c}
						<button
							type="button"
							class="suggest-pill"
							onclick={() => selectCountry(c)}
						>
							<span class="s-code">{c}</span>
						</button>
					{/each}
				</div>
			</div>
		{/if}
	</div>
</div>

<style lang="scss">
	.streaming-providers-card {
		display: flex;
		flex-flow: column;
		gap: 10px;
		width: 100%;
		margin-top: 14px;
		padding: 14px 16px;
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid rgba(255, 255, 255, 0.08);

		.card-header {
			display: flex;
			align-items: center;
			justify-content: space-between;
			gap: 12px;
			flex-wrap: wrap;

			.header-left {
				display: flex;
				align-items: center;
				gap: 10px;
				flex-wrap: wrap;

				.header-title {
					font-size: 11.5px;
					font-weight: 700;
					text-transform: uppercase;
					letter-spacing: 0.06em;
					color: rgba(255, 255, 255, 0.55);
				}

				.country-selector-wrap {
					position: relative;

					.country-btn {
						display: inline-flex;
						align-items: center;
						gap: 5px;
						padding: 3px 8px;
						border-radius: 6px;
						background: rgba(255, 255, 255, 0.06);
						border: 1px solid rgba(255, 255, 255, 0.12);
						color: #ffffff;
						font-size: 11.5px;
						font-weight: 700;
						cursor: pointer;
						transition: all 140ms ease;

						.country-code {
							letter-spacing: 0.5px;
						}

						.chevron {
							font-size: 10px;
							opacity: 0.6;
							transition: transform 140ms ease;
						}

						&:hover,
						&.active {
							background: rgba(255, 255, 255, 0.15);
							border-color: rgba(255, 255, 255, 0.25);
							transform: translateY(-1px);

							.chevron {
								opacity: 1;
								transform: translateY(1px);
							}
						}
					}

					.country-dropdown-popover {
						position: absolute;
						top: calc(100% + 8px);
						left: 0;
						width: 270px;
						max-height: 380px;
						padding: 12px;
						border-radius: 14px;
						background: rgba(18, 22, 33, 0.95);
						border: 1px solid rgba(255, 255, 255, 0.14);
						box-shadow: 0 12px 36px rgba(0, 0, 0, 0.55);
						backdrop-filter: blur(24px);
						-webkit-backdrop-filter: blur(24px);
						z-index: 1000;
						display: flex;
						flex-flow: column;
						gap: 10px;

						.popover-search {
							input {
								width: 100%;
								padding: 7px 10px;
								border-radius: 8px;
								background: rgba(255, 255, 255, 0.08);
								border: 1px solid rgba(255, 255, 255, 0.12);
								color: #ffffff;
								font-size: 12.5px;
								outline: none;

								&::placeholder {
									color: rgba(255, 255, 255, 0.4);
								}

								&:focus {
									border-color: #a5b4fc;
									background: rgba(255, 255, 255, 0.12);
								}
							}
						}

						.popular-chips {
							display: flex;
							flex-wrap: wrap;
							gap: 5px;

							.popular-chip {
								display: inline-flex;
								align-items: center;
								gap: 4px;
								padding: 3px 7px;
								border-radius: 6px;
								background: rgba(255, 255, 255, 0.06);
								border: 1px solid rgba(255, 255, 255, 0.08);
								color: #e2e8f0;
								font-size: 11px;
								font-weight: 600;
								cursor: pointer;
								transition: all 120ms ease;

								&:hover {
									background: rgba(255, 255, 255, 0.15);
									border-color: rgba(255, 255, 255, 0.2);
								}

								&.selected {
									background: rgba(99, 102, 241, 0.3);
									border-color: #818cf8;
									color: #ffffff;
								}
							}
						}

						.regions-list {
							display: flex;
							flex-flow: column;
							gap: 2px;
							max-height: 210px;
							overflow-y: auto;
							padding-right: 4px;

							&::-webkit-scrollbar {
								width: 5px;
							}
							&::-webkit-scrollbar-thumb {
								background: rgba(255, 255, 255, 0.2);
								border-radius: 4px;
							}

							.region-row {
								display: flex;
								align-items: center;
								gap: 8px;
								padding: 6px 10px;
								border-radius: 8px;
								background: transparent;
								border: none;
								color: #cbd5e1;
								font-size: 12.5px;
								text-align: left;
								cursor: pointer;
								transition: background 120ms ease;

								.row-name {
									flex: 1;
									white-space: nowrap;
									overflow: hidden;
									text-overflow: ellipsis;
								}

								.row-code {
									font-size: 11px;
									font-weight: 700;
									opacity: 0.5;
								}

								&:hover {
									background: rgba(255, 255, 255, 0.1);
									color: #ffffff;
								}

								&.selected {
									background: rgba(99, 102, 241, 0.25);
									color: #a5b4fc;
									font-weight: 700;
								}
							}

							.empty-regions {
								padding: 16px;
								text-align: center;
								font-size: 12px;
								color: rgba(255, 255, 255, 0.4);
							}
						}
					}
				}
			}

			.justwatch-badge {
				display: inline-flex;
				align-items: center;
				gap: 5px;
				padding: 3px 8px;
				border-radius: 6px;
				background: rgba(243, 175, 41, 0.12);
				border: 1px solid rgba(243, 175, 41, 0.28);
				color: #f6c043;
				text-decoration: none;
				font-size: 11px;
				font-weight: 700;
				letter-spacing: 0.4px;
				transition: all 140ms ease;

				.justwatch-arrow {
					font-size: 10px;
				}

				&:hover {
					background: rgba(243, 175, 41, 0.22);
					border-color: rgba(243, 175, 41, 0.5);
					color: #ffffff;
					transform: translateY(-1px);
				}
			}
		}

		.card-body {
			.providers-loading {
				display: flex;
				align-items: center;
				gap: 10px;
				padding: 10px 0;
				color: rgba(255, 255, 255, 0.7);
				font-size: 13px;
			}

			.providers-badges-grid {
				display: flex;
				align-items: center;
				flex-wrap: wrap;
				gap: 10px;
			}

			.no-providers-alert {
				display: flex;
				flex-flow: column;
				gap: 10px;
				padding: 10px 14px;
				border-radius: 10px;
				background: rgba(255, 255, 255, 0.03);
				border: 1px dashed rgba(255, 255, 255, 0.1);

				.no-providers-text {
					font-size: 12.5px;
					color: rgba(255, 255, 255, 0.6);
				}

				.alternative-suggestions {
					display: flex;
					align-items: center;
					flex-wrap: wrap;
					gap: 6px;

					.suggest-label {
						font-size: 11.5px;
						font-weight: 600;
						color: rgba(255, 255, 255, 0.45);
					}

					.suggest-pill {
						display: inline-flex;
						align-items: center;
						gap: 4px;
						padding: 2px 7px;
						border-radius: 6px;
						background: rgba(255, 255, 255, 0.06);
						border: 1px solid rgba(255, 255, 255, 0.1);
						color: #e2e8f0;
						font-size: 11.5px;
						font-weight: 600;
						cursor: pointer;
						transition: all 120ms ease;

						&:hover {
							background: rgba(255, 255, 255, 0.16);
							border-color: rgba(255, 255, 255, 0.25);
							transform: translateY(-1px);
						}
					}
				}
			}
		}
	}
</style>
