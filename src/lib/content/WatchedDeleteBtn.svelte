<script lang="ts">
	import tooltip from "../actions/tooltip";
	import Icon from "../Icon.svelte";
	import { removeWatched } from "../util/api";
	import { notifyUndo } from "../util/notify";
	import type { Watched } from "@/types";

	interface Props {
		watchedId: number;
		watchedEntry?: Watched;
		mediaName?: string;
		onDelete: () => void;
		onRestore?: (w: Watched) => void;
	}

	let {
		watchedId,
		watchedEntry,
		mediaName,
		onDelete,
		onRestore,
	}: Props = $props();

	function onDeleteClicked() {
		const targetId = watchedId;
		const entry = watchedEntry;
		const name = mediaName ?? "Item";
		onDelete();

		notifyUndo({
			text: `<b>${name}</b> removido da lista`,
			undoLabel: "Desfazer",
			time: 6000,
			onUndo: () => {
				if (onRestore && entry) {
					onRestore(entry);
				} else {
					window.location.reload();
				}
			},
			onDismiss: () => {
				removeWatched(targetId);
			},
		});
	}
</script>

<button
	class="delete-btn"
	onclick={() => onDeleteClicked()}
	use:tooltip={{ text: "Delete", pos: "bot" }}
	aria-label="Delete from list"
>
	<Icon i="trash" wh={19} />
</button>

<style lang="scss">
	.delete-btn {
		&:hover {
			color: $error;
		}
	}
</style>
