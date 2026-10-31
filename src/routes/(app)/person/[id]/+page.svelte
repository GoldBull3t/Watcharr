<script lang="ts">
	import Error from "@/lib/Error.svelte";
	import Poster from "@/lib/poster/Poster.svelte";
	import PosterList from "@/lib/poster/PosterList.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import DropDown from "@/lib/DropDown.svelte";
	import type {
		Media,
		PersonCreditsResponse,
		PersonDetailsResponse,
	} from "@/types";
	import { req } from "@/lib/util/api.js";
	import Checkbox from "@/lib/Checkbox.svelte";
	import Icon from "@/lib/Icon.svelte";
	import PageBackdrop from "@/lib/generic/PageBackdrop.svelte";
	import PosterImage from "@/lib/content/PosterImage.svelte";
	import ExpandableText from "@/lib/content/ExpandableText.svelte";
	import MediaDetailSkeleton from "@/lib/content/MediaDetailSkeleton.svelte";

	let { data } = $props();

	let person: PersonDetailsResponse | undefined = $state();
	let pageError: unknown | undefined = $state();
	let sortOption = $state("Vote count");
	let credits: PersonCreditsResponse | undefined = $state();
	let onMyListFilter = $state(false);

	async function loadPerson(signal?: AbortSignal) {
		try {
			person = undefined;
			pageError = undefined;
			if (!data.personId) return;
			const p = await req.get<PersonDetailsResponse>(
				`/content/person/${data.personId}`,
				{ signal },
			);
			if (signal?.aborted) return;
			person = p;
			const c = await req.get<PersonCreditsResponse>(
				`/content/person/${data.personId}/credits`,
				{ signal },
			);
			if (signal?.aborted) return;
			c.credits = c.credits?.filter(
				(v, i, a) => a.findIndex((t) => t.ids.tmdb === v.ids.tmdb) === i,
			);
			credits = c;
		} catch (err: unknown) {
			if (signal?.aborted) return;
			person = undefined;
			pageError = err;
		}
	}

	$effect(() => {
		const controller = new AbortController();
		if (data.personId) {
			loadPerson(controller.signal);
		}
		return () => controller.abort();
	});

	let sortedCredits = $derived.by(() => {
		if (!credits?.credits || credits.credits.length === 0) return [];
		const list = [...credits.credits];
		switch (sortOption) {
			case "Vote count":
				return list.sort((a, b) => (b.ratingCount ?? 0) - (a.ratingCount ?? 0));
			case "Newest":
				return list.sort((a, b) => newestOldestSort(a, b, 0));
			case "Oldest":
				return list.sort((a, b) => newestOldestSort(a, b, 1));
			default:
				return list;
		}
	});

	function newestOldestSort(
		a: Media,
		b: Media,
		/**
		 * 0 = Newest,
		 * 1 = Oldest
		 */
		n: 0 | 1,
	) {
		// Assume missing release date means future release (TBD)
		if (!a.releaseDate && !b.releaseDate) {
			// Both releases have no date, return as equals
			// here to avoid an infinite loop.
			return 0;
		}
		if (!a.releaseDate) return n === 0 ? -1 : 1;
		if (!b.releaseDate) return n === 0 ? 1 : -1;

		const dateA = new Date(a.releaseDate).valueOf();
		const dateB = new Date(b.releaseDate).valueOf();

		if (n === 0) {
			return dateB - dateA;
		} else {
			return dateA - dateB;
		}
	}
</script>

<svelte:head>
	<title>{person?.name ? `${person.name} - ` : ""}Person</title>
</svelte:head>

<div>
	{#if pageError}
		<Error
			pretty="Failed to load person!"
			error={pageError}
			onRetry={() => loadPerson()}
		/>
	{:else if !person}
		<MediaDetailSkeleton type="person" />
	{:else if Object.keys(person).length > 0}
		{#if credits?.credits && credits.credits.length > 0 && credits.credits[0].extBackdropPath}
			<PageBackdrop
				src={"https://www.themoviedb.org/t/p/w1920_and_h800_multi_faces" +
					credits.credits[0].extBackdropPath}
			/>
		{/if}
		<div class="content">
			<div class="details-wrap">
				<div class="details-container">
					{#if person.extPosterPath}
						<PosterImage
							src={"https://image.tmdb.org/t/p/w500" + person.extPosterPath}
						/>
					{/if}

					<div class="details">
						<span class="title-container">
							<a
								href={person.homepage}
								rel="noreferrer noopener external"
								target="_blank"
							>
								{person.name}
							</a>
							<span></span>
						</span>

						<ExpandableText title="Biography" text={person.biography} />

						<div class="detail-info">
							{#if person.knownForDepartment}
								<div>
									<span>Department</span>
									<span>{person.knownForDepartment}</span>
								</div>
							{/if}
							{#if person.placeOfBirth}
								<div>
									<span>Born In</span>
									<span>{person.placeOfBirth}</span>
								</div>
							{/if}
							{#if person.birthday}
								<div>
									<span>Born On</span>
									<span
										>{new Date(
											Date.parse(person.birthday),
										).toLocaleDateString()}</span
									>
								</div>
							{/if}
							{#if person.deathday}
								<div>
									<span>Died On</span>
									<span>
										{new Date(Date.parse(person.deathday)).toLocaleDateString()}
									</span>
								</div>
							{/if}
							{#if person.age}
								<div>
									<span>Age</span>
									<span>{person.age} Years</span>
								</div>
							{/if}
						</div>
					</div>
				</div>
			</div>
		</div>
		{#if credits}
			{#if credits?.credits && credits?.credits?.length > 0}
				<div class="filters">
					<div class="listFilter">
						<span>On my list</span>
						<Checkbox name="On my list" bind:value={onMyListFilter} />
					</div>
					<DropDown
						bind:active={sortOption}
						placeholder="Vote count"
						options={["Vote count", "Newest", "Oldest"]}
						isDropDownItem={false}
						showActiveElementsInOptions={true}
					/>
				</div>
				<div class="page">
					<PosterList>
						{#each sortedCredits as c (c.ids.tmdb)}
							<Poster
								media={c}
								bind:watched={c.watched}
								fluidSize
								hideIfNotOnList={onMyListFilter}
							/>
						{/each}
					</PosterList>
				</div>
			{:else}
				<div class="no-credits-message">
					<Icon i="star" wh={80} />
					<h2 class="norm">We found no credits!</h2>
					<h4 class="norm">It seems that this person has no credits.</h4>
				</div>
			{/if}
		{:else}
			<Spinner />
		{/if}
	{:else}
		<Error error="Person not found" pretty="Person not found" />
	{/if}
</div>

<style lang="scss">
	@use "../../../../lib/content/page.scss";

	.filters {
		display: flex;
		flex-flow: row wrap;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		margin: 12px auto 20px;
		padding: 12px 20px;
		border-radius: 10px;
		background: rgba(15, 19, 28, 0.72);
		border: 1px solid rgba(255, 255, 255, 0.08);
		width: calc(100% - 64px);
		max-width: 1296px;

		@media screen and (max-width: 680px) {
			width: calc(100% - 32px);
		}

		.listFilter {
			display: flex;
			align-items: center;
			gap: 10px;
			font-size: 14px;
			font-weight: 500;
			color: $text-color;
		}
	}

	.content {
		position: relative;
		color: white;
		margin-bottom: 8px;

		.details-container .details {
			.title-container {
				a {
					color: #ffffff;
					text-decoration: none;
					font-size: 28px;
					font-weight: 700;
					letter-spacing: -0.02em;
					line-height: 1.2;

					&:hover {
						text-decoration: underline;
					}
				}

				span {
					font-size: 18px;
					color: rgba(255, 255, 255, 0.7);
				}
			}

			.detail-info {
				display: flex;
				flex-wrap: wrap;
				gap: 10px;
				margin-top: 12px;

				div {
					display: flex;
					flex-flow: column;
					gap: 2px;
					padding: 8px 14px;
					border-radius: 8px;
					background: rgba(255, 255, 255, 0.04);
					border: 1px solid rgba(255, 255, 255, 0.08);

					span:first-child {
						font-size: 10.5px;
						font-weight: 700;
						text-transform: uppercase;
						letter-spacing: 0.05em;
						color: rgba(255, 255, 255, 0.48);
					}

					span:last-child {
						font-size: 13.5px;
						font-weight: 600;
						color: #ffffff;
					}
				}
			}
		}
	}

	.page {
		display: flex;
		flex-flow: column;
		align-items: center;
		width: 100%;
		max-width: 1360px;
		margin-inline: auto;
		gap: 28px;
		padding: 10px 32px 60px;

		@media screen and (max-width: 680px) {
			padding: 10px 16px 40px;
		}
	}

	.no-credits-message {
		display: flex;
		flex-flow: column;
		gap: 5px;
		align-items: center;
		margin-top: 20px;

		h2 {
			margin-top: 10px;
		}

		h4 {
			font-weight: normal;
		}
	}
</style>
