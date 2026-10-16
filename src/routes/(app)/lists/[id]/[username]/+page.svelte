<script lang="ts">
	import Icon from "@/lib/Icon.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import UserAvatar from "@/lib/img/UserAvatar.svelte";
	import { followUser, req, unfollowUser } from "@/lib/util/api.js";
	import { clearActiveFilters, store } from "@/store.svelte.js";
	import type { Media, PaginationResponse, PublicUser } from "@/types.js";
	import { onDestroy, untrack } from "svelte";
	import paginatedLoader, {
		PaginatedLoaderRunFnAction,
	} from "@/lib/util/paginatedLoader.svelte.js";
	import infScroll from "@/lib/util/infScroll.js";
	import { page } from "$app/state";
	import PosterList from "@/lib/poster/PosterList.svelte";
	import Poster from "@/lib/poster/Poster.svelte";
	import Error from "@/lib/Error.svelte";
	import { afterNavigate } from "$app/navigation";

	let meta = $derived.by(() => {
		return {
			id: page.params.id,
			username: page.params.username,
		};
	});

	let followBtnDisabled = $state(false);
	let user: PublicUser | undefined = $state();

	let isFollowing = $derived(
		!!store.follows?.find((f) => f.followedUser.id === Number(meta.id)),
	);

	const scroll = infScroll({ callback: onScrollToBottom });
	const dataLoader = paginatedLoader<Media, undefined>(load);

	let nextLoadParams: {
		page: number;
		[x: string]: unknown;
	} = $derived({
		page: dataLoader.state.page + 1,
		...store.sortAndFiltersForQueryParams,
	});

	async function load(signal: AbortSignal) {
		console.debug("load: loadParams:", nextLoadParams);
		if (nextLoadParams.page === dataLoader.state.page) {
			console.warn("load: Already on this page, not loading it again!");
			return;
		}
		if (!meta.id || !meta.username) {
			console.warn("load: Missing id or username!");
			return;
		}
		const r = await req.get<PaginationResponse<Media, undefined>>(
			`/watched/${meta.id}/${meta.username}`,
			{
				params: nextLoadParams,
				signal,
			},
		);
		scroll.dataLoaded();
		return r;
	}

	async function onScrollToBottom() {
		// If an error is being shown, no more infinite scroll.
		if (dataLoader.state.reqLoadError) {
			return;
		}
		console.debug("onScrollToBottom");
		dataLoader.runFn();
	}

	// NOTE: This effect also handles initial load of data.
	$effect(() => {
		// When our sort/filter query params change,
		// load our list again.
		// Since it exists at load, this performs our
		// initial load of data too.
		if (store.sortAndFiltersForQueryParams) {
			untrack(() => {
				// We don't want to trigger another re-run of this
				// effect when state inside these funcs changes.
				dataLoader.reset();
				dataLoader.runFn();
			});
		}
	});

	async function getPublicUser() {
		return await req.get<PublicUser>(
			`/user/public/${meta.id}/${meta.username}`,
		);
	}

	async function follow() {
		followBtnDisabled = true;
		console.log(isFollowing);
		try {
			if (isFollowing) {
				await unfollowUser(Number(meta.id));
			} else {
				await followUser(Number(meta.id));
			}
		} finally {
			followBtnDisabled = false;
		}
	}

	$effect(() => {
		user = undefined;
		if (meta?.id && meta?.username) {
			getPublicUser()
				.then((u) => {
					user = u;
				})
				.catch((err) => {
					console.error("getPublicUser failed!", err);
				});
		}
	});

	afterNavigate((e) => {
		if (!e.from?.route?.id?.toLowerCase()?.includes("/lists")) {
			// Ensure afterNavigate can only runFn when we are coming
			// from another list page.
			// OnMount of a list page we don't want to have this also run
			// because that breaks our loader and our effect will handle loading
			// in that case already.
			return;
		}
		console.log("afterNavigate.");
		dataLoader.abortReq("navigated away");
		dataLoader.runFn(PaginatedLoaderRunFnAction.Reset);
	});

	onDestroy(() => {
		console.debug("PAGE DESTROYED");
		scroll.destroy();
		dataLoader.abortReq("page destroyed");
	});
</script>

<svelte:head>
	<title>{meta.username}'s Watched List - Watcharr</title>
</svelte:head>

<div class="content">
	<div class="inner">
		<div class="user-profile-banner">
			<div class="avatar-wrap">
				<UserAvatar img={user?.avatar} />
			</div>
			<div class="user-meta">
				<div class="meta-top">
					<h2 title={user?.username} class="username">
						{meta.username}
					</h2>
					<button
						class={`follow-btn ${isFollowing ? "following" : ""}`}
						disabled={followBtnDisabled}
						onclick={follow}
					>
						<Icon i={isFollowing ? "person-minus" : "person-add"} wh={16} />
						<span>{isFollowing ? "Following" : "Follow"}</span>
					</button>
				</div>
				{#if user?.bio}
					<p class="bio" title={user.bio}>{user.bio}</p>
				{/if}
			</div>
			{#if dataLoader.state.data && dataLoader.state.data.length > 0}
				<div class="list-stats">
					<span class="count-pill">{dataLoader.state.data.length} Titles</span>
				</div>
			{/if}
		</div>
	</div>
</div>

<PosterList>
	{#if dataLoader.state.data?.length > 0}
		{#each dataLoader.state.data as w, i (`${i}-${w.type}`)}
			{#if w}
				<Poster
					watched={dataLoader.state.data[i].watched}
					media={w}
					fluidSize={true}
					disableInteraction={true}
				/>
			{/if}
		{/each}
	{:else if !dataLoader.state.reqLoading && !dataLoader.state.reqLoadError}
		<div class="empty-list">
			<div class="empty-icon-wrap">
				<Icon i={store.hasActiveFilters ? "filter-circle" : "reel"} wh={48} />
			</div>
			<h2 class="norm">This list is quiet</h2>
			<p class="empty-desc">
				{#if store.hasActiveFilters}
					No items match your active filters.
				{:else}
					This user hasn't added any media to their list yet.
				{/if}
			</p>
			{#if store.hasActiveFilters}
				<button class="btn" onclick={() => clearActiveFilters()}
					>Clear Filters</button
				>
			{/if}
		</div>
	{/if}
</PosterList>

{#if dataLoader.state.reqLoading}
	<div class="loader-wrap">
		<Spinner />
	</div>
{/if}

{#if dataLoader.state.reqLoadError}
	<div class="error-wrap">
		<Error
			pretty="Failed to load results!"
			error={dataLoader.state.reqLoadError}
			onRetry={() => {
				dataLoader.state.reqLoadError = undefined;
				dataLoader.runFn();
			}}
		/>
	</div>
{/if}

<style lang="scss">
	.content {
		display: flex;
		width: 100%;
		justify-content: center;
		padding: 24px 20px 16px 20px;

		.inner {
			width: 100%;
			max-width: 1200px;
		}
	}

	.user-profile-banner {
		display: flex;
		align-items: center;
		gap: 24px;
		padding: 24px 28px;
		border-radius: 20px;
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid rgba(255, 255, 255, 0.08);
		backdrop-filter: blur(12px);

		.avatar-wrap {
			flex-shrink: 0;
		}

		.user-meta {
			display: flex;
			flex-direction: column;
			gap: 6px;
			flex-grow: 1;

			.meta-top {
				display: flex;
				align-items: center;
				gap: 16px;

				.username {
					font-size: 24px;
					font-weight: 700;
					letter-spacing: -0.02em;
					color: #fff;
					margin: 0;
				}

				.follow-btn {
					display: inline-flex;
					align-items: center;
					gap: 6px;
					height: 32px;
					padding: 0 14px;
					border-radius: 16px;
					font-size: 13px;
					font-weight: 600;
					background: $accent-color;
					color: #fff;
					border: none;
					cursor: pointer;
					transition: all 150ms ease;

					&:hover {
						background: $accent-color-hover;
						transform: translateY(-1px);
					}

					&.following {
						background: rgba(255, 255, 255, 0.08);
						border: 1px solid rgba(255, 255, 255, 0.15);
						color: $text-color;

						&:hover {
							background: rgba(239, 68, 68, 0.15);
							border-color: rgba(239, 68, 68, 0.3);
							color: #fca5a5;
						}
					}
				}
			}

			.bio {
				font-size: 13.5px;
				color: $text-muted;
				margin: 0;
				max-width: 600px;
				line-height: 1.45;
				display: -webkit-box;
				-webkit-line-clamp: 2;
				line-clamp: 2;
				-webkit-box-orient: vertical;
				overflow: hidden;
			}
		}

		.list-stats {
			margin-left: auto;

			.count-pill {
				font-size: 12px;
				font-weight: 600;
				padding: 6px 14px;
				border-radius: 20px;
				background: rgba(255, 255, 255, 0.06);
				border: 1px solid rgba(255, 255, 255, 0.1);
				color: $text-muted;
			}
		}

		@media screen and (max-width: 640px) {
			flex-direction: column;
			align-items: center;
			text-align: center;

			.user-meta .meta-top {
				flex-direction: column;
				gap: 10px;
			}

			.list-stats {
				margin-left: 0;
			}
		}
	}

	.empty-list {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 60px 20px;
		text-align: center;
		grid-column: 1 / -1;

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

		h2 {
			font-size: 20px;
			font-weight: 600;
			margin: 0 0 6px 0;
			color: $text-color;
		}

		.empty-desc {
			font-size: 14px;
			color: $text-muted;
			max-width: 360px;
			margin: 0 0 16px 0;
		}
	}

	.loader-wrap,
	.error-wrap {
		display: flex;
		justify-content: center;
		margin: 40px 0 60px 0;
	}
</style>
