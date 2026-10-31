<script lang="ts">
	import { store } from "@/store.svelte";
	import Icon from "../Icon.svelte";
	import CreateTagModal from "./CreateTagModal.svelte";
	import type { Tag as TagT } from "@/types";
	import Tag from "./Tag.svelte";
	import DeleteTagModal from "./DeleteTagModal.svelte";
	import Menu, { type MenuConfig } from "../Menu.svelte";
	import { onMount } from "svelte";

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
		width: "240px",
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

	const TAG_ORDER_STORAGE_KEY = "watcharr_tag_order";
	const TAG_SORT_MODE_KEY = "watcharr_tag_sort_mode";

	let allTags = $derived(store.tags);
	let customOrderIds = $state<number[]>([]);
	let sortMode = $state<"custom" | "alpha">("custom");
	let draggedTagId = $state<number | null>(null);
	let dragOverTagId = $state<number | null>(null);

	let tagModalOpen = $state(false);
	let inManageMode = $state(false);
	let tagToEdit: TagT | undefined = $state(undefined);
	let tagToDelete: TagT | undefined = $state(undefined);
	let filterQuery = $state("");

	onMount(() => {
		try {
			const savedOrder = localStorage.getItem(TAG_ORDER_STORAGE_KEY);
			if (savedOrder) {
				const parsed = JSON.parse(savedOrder);
				if (Array.isArray(parsed)) {
					customOrderIds = parsed.filter((id) => typeof id === "number");
				}
			}
			const savedMode = localStorage.getItem(TAG_SORT_MODE_KEY);
			if (savedMode === "alpha" || savedMode === "custom") {
				sortMode = savedMode;
			}
		} catch (err) {
			console.debug("TagMenu: Failed to load saved tag order", err);
		}
	});

	let orderedTags = $derived.by(() => {
		if (!allTags) return [];
		const list = [...allTags];
		if (sortMode === "alpha") {
			return list.sort((a, b) => a.name.localeCompare(b.name));
		}
		if (customOrderIds.length > 0) {
			return list.sort((a, b) => {
				const idxA = customOrderIds.indexOf(a.id);
				const idxB = customOrderIds.indexOf(b.id);
				if (idxA === -1 && idxB === -1) return 0;
				if (idxA === -1) return 1;
				if (idxB === -1) return -1;
				return idxA - idxB;
			});
		}
		return list;
	});

	let filteredTags = $derived(
		orderedTags.filter((t) =>
			filterQuery
				? t.name.toLowerCase().includes(filterQuery.toLowerCase())
				: true,
		),
	);

	function toggleSortMode() {
		sortMode = sortMode === "custom" ? "alpha" : "custom";
		try {
			localStorage.setItem(TAG_SORT_MODE_KEY, sortMode);
		} catch {
			// Ignore storage errors
		}
	}

	function handleDragStart(ev: DragEvent, tagId: number) {
		draggedTagId = tagId;
		if (ev.dataTransfer) {
			ev.dataTransfer.effectAllowed = "move";
			ev.dataTransfer.setData("text/plain", String(tagId));
		}
	}

	function handleDragOver(ev: DragEvent, targetTagId: number) {
		ev.preventDefault();
		if (draggedTagId === null || draggedTagId === targetTagId) return;
		dragOverTagId = targetTagId;
	}

	function handleDrop(ev: DragEvent, targetTagId: number) {
		ev.preventDefault();
		if (draggedTagId === null || draggedTagId === targetTagId || !allTags) {
			draggedTagId = null;
			dragOverTagId = null;
			return;
		}
		const currentIds = orderedTags.map((t) => t.id);
		const fromIdx = currentIds.indexOf(draggedTagId);
		const toIdx = currentIds.indexOf(targetTagId);
		if (fromIdx !== -1 && toIdx !== -1) {
			currentIds.splice(fromIdx, 1);
			currentIds.splice(toIdx, 0, draggedTagId);
			customOrderIds = currentIds;
			sortMode = "custom";
			try {
				localStorage.setItem(TAG_ORDER_STORAGE_KEY, JSON.stringify(currentIds));
				localStorage.setItem(TAG_SORT_MODE_KEY, "custom");
			} catch {
				// Ignore storage errors
			}
		}
		draggedTagId = null;
		dragOverTagId = null;
	}

	function handleDragEnd() {
		draggedTagId = null;
		dragOverTagId = null;
	}

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
		{#if allTags && allTags.length > 1}
			<button
				class="plain sort-btn {sortMode === 'alpha' ? 'manage-on' : ''}"
				onclick={toggleSortMode}
				title={sortMode === "alpha"
					? "Sorted A–Z (click for custom drag order)"
					: "Custom order — drag tags to reorder (click for A–Z)"}
			>
				<Icon i="sort" wh={15} />
			</button>
		{/if}
		{#if showManageBtn}
			<button
				class={["plain", inManageMode ? "manage-on" : ""].join(" ")}
				onclick={() => (inManageMode = !inManageMode)}
				title={inManageMode
					? "Exit manage mode"
					: "Manage tags (edit / delete / reorder)"}
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
				<span>Drag to reorder, or edit/delete your tags:</span>
			</div>
		{/if}
		<div class="list" role="list">
			{#each filteredTags as t (t.id)}
				{@const isSelected = selectedTags
					? selectedTags.find((tag) => tag.id === t.id)
						? true
						: false
					: false}
				<div
					class="tag-row {inManageMode ? 'manage-row' : ''} {dragOverTagId === t.id
						? 'drag-over'
						: ''}"
					role="listitem"
					draggable={!filterQuery}
					ondragstart={(e) => handleDragStart(e, t.id)}
					ondragover={(e) => handleDragOver(e, t.id)}
					ondrop={(e) => handleDrop(e, t.id)}
					ondragend={handleDragEnd}
				>
					<div class="tag-with-indicator">
						{#if !filterQuery && allTags.length > 1}
							<span class="drag-grip" title="Drag to reorder">⋮⋮</span>
						{/if}
						<Tag
							tag={t}
							size="sm"
							onClick={() => {
								if (!inManageMode && typeof onTagClick === "function") {
									onTagClick(t, isSelected);
								}
							}}
						/>
						{#if t.isPublic}
							<span class="public-tag-indicator" title="Lista Pública">🌐</span>
						{/if}
					</div>

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
		gap: 4px;

		button.plain {
			display: flex;
			align-items: center;
			justify-content: center;
			width: 26px;
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
			padding: 3px 4px;
			border-radius: 6px;
			border: 1px solid transparent;
			transition: border-color 120ms ease, background 120ms ease;

			&.manage-row {
				background: rgba(255, 255, 255, 0.02);
			}

			&.drag-over {
				border-color: rgba(99, 102, 241, 0.6);
				background: rgba(99, 102, 241, 0.12);
			}

			.tag-with-indicator {
				display: inline-flex;
				align-items: center;
				gap: 5px;

				.drag-grip {
					font-size: 10px;
					letter-spacing: -1px;
					color: rgba(255, 255, 255, 0.3);
					cursor: grab;
					user-select: none;
					line-height: 1;

					&:hover {
						color: rgba(255, 255, 255, 0.7);
					}
				}

				.public-tag-indicator {
					font-size: 11px;
					opacity: 0.85;
					cursor: default;
				}
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
