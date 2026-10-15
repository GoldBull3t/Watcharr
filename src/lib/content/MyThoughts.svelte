<script lang="ts">
	import Modal from "../Modal.svelte";
	import Icon from "../Icon.svelte";
	import { notify } from "../util/notify";

	interface Props {
		contentTitle?: string;
		thoughts: string;
		onChange: (newThoughts: string) => Promise<boolean>;
	}

	let { contentTitle = "this", thoughts, onChange }: Props = $props();

	let modalOpen = $state(false);
	let textarea: HTMLTextAreaElement | undefined = $state();
	let thoughtsToDisplay = $derived(
		thoughts ? thoughts : `Set thoughts on ${contentTitle}`,
	);

	function resizeTextarea() {
		if (!textarea) {
			return;
		}
		textarea.style.height = "";
		textarea.style.height = textarea.scrollHeight + "px";
	}

	$effect(() => {
		if (textarea) resizeTextarea();
	});
</script>

<button
	class={`plain thoughts${thoughtsToDisplay?.length > 100 ? " long" : ""}${thoughts ? "" : " placeholdered"}`}
	onclick={() => {
		modalOpen = !modalOpen;
		if (modalOpen) {
			resizeTextarea();
		}
	}}
>
	<i><Icon i="pencil" wh={24} /></i>
	<p>{thoughtsToDisplay}</p>
</button>

{#if modalOpen}
	<Modal
		title="Your Thoughts"
		desc="View or modify your thoughts on {contentTitle}"
		onClose={async () => {
			if (!textarea) {
				notify({
					text: "Failed to find the text box! Please copy your changes to avoid losing them and try again!",
				});
				return;
			}
			// If thoughts weren't changed or changes saved successfully.
			if (thoughts === textarea.value || (await onChange(textarea.value))) {
				modalOpen = false;
			}
		}}
	>
		<textarea
			name="Thoughts"
			rows="3"
			placeholder={`My thoughts on ${contentTitle}`}
			value={thoughts}
			bind:this={textarea}
			oninput={resizeTextarea}></textarea>
	</Modal>
{/if}

<style lang="scss">
	button.thoughts {
		position: relative;
		width: 100%;
		text-align: start;
		padding: 10px 14px;
		background: rgba(255, 255, 255, 0.04);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 10px;
		max-height: 100px;
		color: $text-color;
		cursor: pointer;
		transition: all 160ms ease;

		&:hover {
			background: rgba(255, 255, 255, 0.07);
			border-color: rgba(255, 255, 255, 0.2);
		}

		&.placeholdered {
			color: $text-color-accent;
			font-style: italic;
		}

		i {
			display: flex;
			position: absolute;
			bottom: 8px;
			right: 8px;
			color: $text-color-accent;
			opacity: 0.7;
			transition: all 150ms ease;
		}

		&:hover i {
			color: $brand-color;
			opacity: 1;
		}

		p {
			margin: 0;
			max-height: 90px;
			overflow: hidden;
			font-size: 13.5px;
			line-height: 1.45;
		}

		&.long p {
			mask: linear-gradient(
				to bottom,
				rgba(0, 0, 0, 1) 0,
				rgba(0, 0, 0, 1) 50%,
				rgba(0, 0, 0, 0) 95%
			);
		}
	}

	textarea {
		width: 100%;
		min-height: 120px;
		padding: 12px 14px;
		border-radius: 10px;
		border: 1px solid rgba(255, 255, 255, 0.12);
		background-color: var(--card-bg, #161b26);
		color: $text-color;
		font-family: inherit;
		font-size: 14px;
		line-height: 1.5;
		resize: vertical;
		outline: none;
		box-sizing: border-box;

		&:focus {
			border-color: $brand-color;
			box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.2);
		}
	}
</style>
