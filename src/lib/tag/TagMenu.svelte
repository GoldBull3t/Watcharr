<script lang="ts">
	import { store } from "@/store.svelte";
	import Icon from "../Icon.svelte";
	import CreateTagModal from "./CreateTagModal.svelte";
	import type { Tag as TagT } from "@/types";
	import Tag from "./Tag.svelte";
	import DeleteTagModal from "./DeleteTagModal.svelte";
	import Menu, { type MenuConfig } from "../Menu.svelte";

	interface Props {
		titleText?: string | undefined;
		onTagClick?: (tag: TagT, remove: boolean) => void | undefined;
		selectedTags?: TagT[] | undefined;
		/**
		 * When `showManageBtn` is true, a manage icon will appear at top
		 * of menu for the user to click.
		 */
		showManageBtn?: boolean;
		menuConfig?: MenuConfig;
	}

	const defaultMenuConfig: MenuConfig = {
		width: "230px",
		right: "47px",
		arrowLeft: "78px",
	};

	let {
		titleText = undefined,
		onTagClick = undefined!,
		selectedTags = undefined,
		showManageBtn = false,
		menuConfig = {},
	}: Props = $props();

	let allTags = $derived(store.tags);

	let tagModalOpen = $state(false);
	let inManageMode = $state(false);
	let tagToEdit: TagT | undefined = $state(undefined);
	let tagToDelete: TagT | undefined = $state(undefined);
	let filterQuery = $state("");

	let filteredTags = $derived(
		allTags
			? allTags.filter((t) =>
					filterQuery
						? t.name.toLowerCase().includes(filterQuery.toLowerCase())
						: true,
				)
			: [],
	);

	function deleteTag(t: TagT) {
		tagToDelete = t;
	}

	function editTag(t: TagT) {
		tagToEdit = t;
	}
</script>

<Menu conf={Object.assign(defaultMenuConfig, menuConfig)}>
	<div class="title">
		<h4 class="norm sm-caps">{titleText ? titleText : "my tags"}</h4>
		{#if showManageBtn}
			<button
				class={["plain", inManageMode ? "manage-on" : ""].join(" ")}
				onclick={() => (inManageMode = !inManageMode)}
				title={inManageMode ? "Exit manage mode" : "Manage tags (edit / delete)"}
			>
				<Icon i={inManageMode ? "close" : "settings"} wh={16} />
			</button>
		{/if}
		<button
			class="plain add-btn"
			onclick={() => (tagModalOpen = !tagModalOpen)}
			title="Create new tag"
		>
			<Icon i="add" wh={20} />
		</button>
	</div>

	{#if allTags && allTags.length > 5}
		<div class="search-wrap">
			<Icon i="search" wh={14} />
			<input
				type="text"
				placeholder="Search tags..."
				bind:value={filterQuery}
				class="menu-search-input"
			/>
			{#if filterQuery}
				<button class="plain clear-btn" onclick={() => (filterQuery = "")}>
					<Icon i="close" wh={12} />
				</button>
			{/if}
		</div>
	{/if}

	{#if allTags && allTags.length > 0}
		{#if inManageMode}
			<div class="manage-hint">
				<span>Edit or delete your collections:</span>
			</div>
		{/if}
		<div class="list">
			{#each filteredTags as t (t.id)}
				{@const isSelected = selectedTags
					? selectedTags.find((tag) => tag.id === t.id)
						? true
						: false
					: false}
				<div class="tag-row {inManageMode ? 'manage-row' : ''}">
					<Tag
						tag={t}
						size="sm"
						onClick={() => {
							if (!inManageMode && typeof onTagClick === "function") {
								onTagClick(t, isSelected);
							}
						}}
					/>

					{#if inManageMode}
						<div class="manage-actions">
							<button
								type="button"
								class="plain manage-act-btn edit"
								onclick={() => editTag(t)}
								title="Edit tag appearance & details"
							>
								<Icon i="pencil" wh={14} />
							</button>
							<button
								type="button"
								class="plain manage-act-btn delete"
								onclick={() => deleteTag(t)}
								title="Delete tag"
							>
								<Icon i="trash" wh={14} />
							</button>
						</div>
					{:else if isSelected}
						<span class="selected-check">
							<Icon i="check" wh={16} />
						</span>
					{/if}
				</div>
			{:else}
				<span class="empty-search">No tags match "{filterQuery}"</span>
			{/each}
		</div>
	{:else}
		<div class="no-tags-state">
			<span>You have no tags yet!</span>
			<button
				type="button"
				class="btn primary create-first-btn"
				onclick={() => (tagModalOpen = true)}
			>
				Create First Tag
			</button>
		</div>
	{/if}
</Menu>

{#if tagModalOpen}
	<CreateTagModal onClose={() => (tagModalOpen = false)} />
{/if}

{#if tagToEdit}
	<CreateTagModal
		existingTag={tagToEdit}
		onClose={() => (tagToEdit = undefined)}
	/>
{/if}

{#if tagToDelete}
	<DeleteTagModal tag={tagToDelete} onClose={() => (tagToDelete = undefined)} />
{/if}

<style lang="scss">
	h4 {
		color: $text-color;

		&:not(:first-of-type) {
			margin-top: 8px;
		}
	}

	.title {
		display: flex;
		flex-flow: row;
		align-items: center;
		margin-bottom: 8px;
		gap: 5px;

		button.plain {
			display: flex;
			align-items: center;
			justify-content: center;
			width: 28px;
			height: 26px;
			padding: 2px 3px;
			border-radius: 8px;
			color: rgba(255, 255, 255, 0.7);
			transition: all 140ms ease;

			&:hover {
				color: #ffffff;
				background-color: rgba(255, 255, 255, 0.08);
			}

			&.manage-on {
				color: #818cf8;
				background-color: rgba(99, 102, 241, 0.15);
			}

			&:first-of-type {
				margin-left: auto;
			}
		}
	}

	.search-wrap {
		display: flex;
		align-items: center;
		gap: 6px;
		background: rgba(255, 255, 255, 0.05);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 6px;
		padding: 4px 8px;
		margin-bottom: 8px;
		color: rgba(255, 255, 255, 0.5);

		.menu-search-input {
			background: transparent;
			border: none;
			outline: none;
			color: #ffffff;
			font-size: 12px;
			width: 100%;

			&::placeholder {
				color: rgba(255, 255, 255, 0.35);
			}
		}

		.clear-btn {
			cursor: pointer;
			color: rgba(255, 255, 255, 0.4);

			&:hover {
				color: #ffffff;
			}
		}
	}

	.manage-hint {
		font-size: 11px;
		color: rgba(255, 255, 255, 0.5);
		margin-bottom: 6px;
		padding: 0 2px;
	}

	.list {
		display: flex;
		flex-flow: column;
		gap: 6px;
		max-height: 280px;
		overflow-y: auto;
		padding-right: 2px;

		.tag-row {
			display: flex;
			align-items: center;
			justify-content: space-between;
			gap: 6px;
			color: $text-color;
			padding: 2px;
			border-radius: 6px;

			&.manage-row {
				background: rgba(255, 255, 255, 0.02);
			}

			.selected-check {
				display: flex;
				align-items: center;
				color: #36ba98;
				margin-left: auto;
			}

			.manage-actions {
				display: flex;
				align-items: center;
				gap: 4px;
				margin-left: auto;

				.manage-act-btn {
					display: flex;
					align-items: center;
					justify-content: center;
					width: 24px;
					height: 24px;
					border-radius: 6px;
					cursor: pointer;
					transition: all 120ms ease;

					&.edit {
						color: rgba(255, 255, 255, 0.6);
						&:hover {
							color: #818cf8;
							background: rgba(99, 102, 241, 0.15);
						}
					}

					&.delete {
						color: rgba(255, 255, 255, 0.6);
						&:hover {
							color: #f87171;
							background: rgba(248, 113, 113, 0.15);
						}
					}
				}
			}
		}

		.empty-search {
			font-size: 12px;
			color: rgba(255, 255, 255, 0.4);
			text-align: center;
			padding: 10px 0;
		}
	}

	.no-tags-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
		padding: 12px 0;
		font-size: 13px;
		color: rgba(255, 255, 255, 0.6);

		.create-first-btn {
			font-size: 12px;
			padding: 5px 12px;
		}
	}
</style>
