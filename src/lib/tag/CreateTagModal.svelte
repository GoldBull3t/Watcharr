<script lang="ts">
	import { store } from "@/store.svelte";
	import Modal from "../Modal.svelte";
	import Setting from "../settings/Setting.svelte";
	import SettingsList from "../settings/SettingsList.svelte";
	import ColorSelector from "../ColorSelector.svelte";
	import TagComponent from "./Tag.svelte";
	import { notify } from "../util/notify";
	import { req } from "../util/api";
	import type { Tag, TagAddRequest } from "@/types";
	import { onMount } from "svelte";
	import {
		TAG_COLOR_PRESETS,
		TAG_EMOJI_CATEGORIES,
		splitTagEmoji,
		formatTagName,
		getContrastTextColor,
		isValidHex,
	} from "./tagUtil";

	interface Props {
		onClose: () => void;
		// Passing an existing tag will enable 'Edit Tag' mode.
		existingTag?: Tag | undefined;
	}

	let { onClose, existingTag = undefined }: Props = $props();

	// Select a default preset
	const randomPreset =
		TAG_COLOR_PRESETS[Math.floor(Math.random() * TAG_COLOR_PRESETS.length)];

	let selectedEmoji = $state<string | undefined>(undefined);
	let tagLabel = $state("");
	let tagDescription = $state("");
	let isPublic = $state(false);
	let bgColor = $state(randomPreset.bgColor);
	let textColor = $state(randomPreset.textColor);
	let autoContrast = $state(true);
	let customHexInput = $state(randomPreset.bgColor);

	let activeEmojiTab = $state(0);
	let error = $state("");
	let submitDisabled = $state(false);
	let modalTitle = $state("Create A Tag");
	let modalDesc = $state("Create a custom collection or category tag");
	let submitBtnText = $state("Create Tag");

	// Combined tag name with emoji prefix
	let fullTagName = $derived(formatTagName(selectedEmoji, tagLabel));

	// Reactive preview tag object
	let previewTag = $derived<Tag>({
		id: existingTag ? existingTag.id : 0,
		name: fullTagName || "Tag Preview",
		color: textColor,
		bgColor: bgColor,
		description: tagDescription,
		createdAt: existingTag?.createdAt || new Date().toISOString(),
		updatedAt: existingTag?.updatedAt || new Date().toISOString(),
		deletedAt: existingTag?.deletedAt || "",
	});

	function handleBgColorChange(newHex: string) {
		bgColor = newHex;
		customHexInput = newHex;
		if (autoContrast) {
			textColor = getContrastTextColor(newHex);
		}
	}

	function handleCustomHexBlur() {
		let hex = customHexInput.trim();
		if (!hex.startsWith("#")) {
			hex = `#${hex}`;
		}
		if (isValidHex(hex)) {
			handleBgColorChange(hex);
		} else {
			customHexInput = bgColor;
		}
	}

	function selectEmoji(emoji: string) {
		if (selectedEmoji === emoji) {
			selectedEmoji = undefined;
		} else {
			selectedEmoji = emoji;
		}
	}

	async function addTag() {
		const finalName = fullTagName.trim();
		if (!finalName) {
			error = "Tag must have a name!";
			return;
		}
		const nid = notify({ text: "Creating Tag", type: "loading" });
		try {
			const resp = await req.post<Tag>("/tag", {
				name: finalName,
				color: textColor,
				bgColor,
				description: tagDescription,
				isPublic,
			} as TagAddRequest);
			store.tags.push(resp);
			notify({ id: nid, text: "Tag Created!", type: "success" });
			onClose();
		} catch (err) {
			console.error("addTag: Failed!", err);
			notify({
				id: nid,
				text: "Failed to create tag!",
				type: "error",
				time: 2,
			});
			error = "Failed to create tag!";
		}
	}

	async function updateTag() {
		const finalName = fullTagName.trim();
		if (!finalName) {
			error = "Tag must have a name!";
			return;
		}
		const nid = notify({ text: "Updating Tag", type: "loading" });
		try {
			const resp = await req.put<Tag>(`/tag/${existingTag!.id}`, {
				name: finalName,
				color: textColor,
				bgColor,
				description: tagDescription,
				isPublic,
			} as TagAddRequest);
			existingTag!.name = finalName;
			existingTag!.color = textColor;
			existingTag!.bgColor = bgColor;
			existingTag!.description = tagDescription;
			existingTag!.isPublic = isPublic;
			notify({ id: nid, text: "Tag Updated!", type: "success" });
			onClose();
		} catch (err) {
			console.error("updateTag: Failed!", err);
			notify({
				id: nid,
				text: "Failed to update tag!",
				type: "error",
				time: 2,
			});
			error = "Failed to update tag!";
		}
	}

	async function submitClicked() {
		submitDisabled = true;
		try {
			if (existingTag) {
				await updateTag();
			} else {
				await addTag();
			}
		} catch (err) {
			console.log("CreateTagModal: Submit failed!", err);
		}
		submitDisabled = false;
	}

	onMount(() => {
		if (existingTag) {
			modalTitle = "Edit Tag";
			modalDesc = "Customize tag appearance, emoji and details";
			submitBtnText = "Save Changes";

			const parsed = splitTagEmoji(existingTag.name);
			selectedEmoji = parsed.emoji;
			tagLabel = parsed.label;

			textColor = existingTag.color;
			bgColor = existingTag.bgColor;
			customHexInput = existingTag.bgColor;
			tagDescription = existingTag.description || "";
			isPublic = existingTag.isPublic || false;
			autoContrast = false; // Preserve existing custom text color
		}
	});
</script>

<div class="wrap">
	<Modal title={modalTitle} desc={modalDesc} maxWidth="560px" {onClose} {error}>
		<!-- Live Preview Card -->
		<div class="tag-preview-card">
			<div class="preview-header">
				<span class="preview-title">Live Preview</span>
				<span class="preview-hint">How your tag will look across Watcharr</span>
			</div>
			<div class="preview-stage">
				<TagComponent tag={previewTag} size="md" />
				<TagComponent tag={previewTag} size="sm" />
			</div>
		</div>

		<SettingsList>
			<!-- Name & Emoji -->
			<Setting
				title="Name & Emoji"
				desc="Choose an emoji and label for your collection"
			>
				<div class="name-input-group">
					<div class="emoji-indicator" title="Selected Emoji">
						{selectedEmoji || "🏷️"}
					</div>
					<input
						type="text"
						name="name"
						placeholder="e.g. Cozy Watches, Sci-Fi..."
						bind:value={tagLabel}
						class="tag-text-input"
					/>
				</div>

				<!-- Quick Emoji Selector -->
				<div class="emoji-picker-container">
					<div class="emoji-tabs">
						{#each TAG_EMOJI_CATEGORIES as cat, i}
							<button
								type="button"
								class="plain emoji-tab-btn {activeEmojiTab === i
									? 'active'
									: ''}"
								onclick={() => (activeEmojiTab = i)}
							>
								{cat.category}
							</button>
						{/each}
						{#if selectedEmoji}
							<button
								type="button"
								class="plain clear-emoji-btn"
								onclick={() => (selectedEmoji = undefined)}
								title="Remove emoji"
							>
								Clear
							</button>
						{/if}
					</div>

					<div class="emoji-grid">
						{#each TAG_EMOJI_CATEGORIES[activeEmojiTab].emojis as emoji}
							<button
								type="button"
								class="plain emoji-btn {selectedEmoji === emoji
									? 'selected'
									: ''}"
								onclick={() => selectEmoji(emoji)}
								title={`Use ${emoji}`}
							>
								{emoji}
							</button>
						{/each}
					</div>
				</div>
			</Setting>

			<!-- Color Palette Presets -->
			<Setting
				title="Color Palette"
				desc="Choose a curated tone or customize your hex"
			>
				<div class="swatches-grid">
					{#each TAG_COLOR_PRESETS as preset}
						<button
							type="button"
							class="plain swatch-btn {bgColor.toLowerCase() ===
							preset.bgColor.toLowerCase()
								? 'active'
								: ''}"
							style:background={preset.bgColor}
							title={preset.name}
							onclick={() => handleBgColorChange(preset.bgColor)}
						>
							{#if bgColor.toLowerCase() === preset.bgColor.toLowerCase()}
								<svg
									xmlns="http://www.w3.org/2000/svg"
									width="14"
									height="14"
									viewBox="0 0 24 24"
									fill="none"
									stroke={preset.textColor}
									stroke-width="3"
									stroke-linecap="round"
									stroke-linejoin="round"
								>
									<polyline points="20 6 9 17 4 12"></polyline>
								</svg>
							{/if}
						</button>
					{/each}
				</div>

				<!-- Custom HEX input row -->
				<div class="hex-input-row">
					<div class="hex-field">
						<span class="hex-hash">HEX</span>
						<input
							type="text"
							maxlength="7"
							bind:value={customHexInput}
							onblur={handleCustomHexBlur}
							onkeydown={(e) => e.key === "Enter" && handleCustomHexBlur()}
							placeholder="#6366f1"
							class="hex-text-input"
						/>
					</div>
					<div class="color-picker-wrap" title="Custom color picker">
						<ColorSelector
							value={bgColor}
							style="width: 36px; height: 36px; border-radius: 8px;"
						/>
					</div>
					<label class="auto-contrast-toggle">
						<input
							type="checkbox"
							bind:checked={autoContrast}
							onchange={() => {
								if (autoContrast) {
									textColor = getContrastTextColor(bgColor);
								}
							}}
						/>
						<span>Auto Contrast</span>
					</label>
				</div>
			</Setting>

			{#if !autoContrast}
				<Setting
					title="Custom Text Color"
					desc="Fine-tune text readability"
					row
				>
					<ColorSelector bind:value={textColor} style="max-width: 120px;" />
				</Setting>
			{/if}

			<!-- Description -->
			<Setting
				title="Description"
				desc="Optional note or theme for this collection."
			>
				<input
					type="text"
					name="description"
					placeholder="e.g. Must-watch horror for spooky season..."
					bind:value={tagDescription}
				/>
			</Setting>

			<!-- Public Sharing Setting -->
			<Setting
				title="Lista Pública & Partilhável"
				desc="Permite a qualquer pessoa aceder e ver esta lista temática através de um link direto."
				row
			>
				<input
					type="checkbox"
					name="isPublic"
					bind:checked={isPublic}
					style="width: 20px; height: 20px; accent-color: #6366f1; cursor: pointer;"
				/>
			</Setting>

			<div class="modal-footer">
				<button
					type="button"
					class="btn secondary cancel-btn"
					onclick={() => onClose()}
				>
					Cancel
				</button>
				<button
					type="button"
					class="btn primary submit-btn"
					onclick={() => submitClicked()}
					disabled={submitDisabled || !tagLabel.trim()}
				>
					{submitBtnText}
				</button>
			</div>
		</SettingsList>
	</Modal>
</div>

<style lang="scss">
	.wrap {
		color: $text-color;
	}

	.tag-preview-card {
		margin-bottom: 20px;
		padding: 16px;
		background: rgba(255, 255, 255, 0.04);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 12px;
		display: flex;
		flex-direction: column;
		gap: 12px;

		.preview-header {
			display: flex;
			align-items: center;
			justify-content: space-between;

			.preview-title {
				font-size: 11px;
				font-weight: 700;
				text-transform: uppercase;
				letter-spacing: 0.6px;
				color: rgba(255, 255, 255, 0.5);
			}

			.preview-hint {
				font-size: 11px;
				color: rgba(255, 255, 255, 0.4);
			}
		}

		.preview-stage {
			display: flex;
			align-items: center;
			gap: 12px;
			padding: 12px 16px;
			background: rgba(0, 0, 0, 0.25);
			border-radius: 8px;
			min-height: 48px;
		}
	}

	.name-input-group {
		display: flex;
		align-items: center;
		gap: 8px;
		width: 100%;

		.emoji-indicator {
			display: flex;
			align-items: center;
			justify-content: center;
			width: 42px;
			height: 40px;
			font-size: 20px;
			background: rgba(255, 255, 255, 0.06);
			border: 1px solid rgba(255, 255, 255, 0.1);
			border-radius: 8px;
			flex-shrink: 0;
		}

		.tag-text-input {
			flex: 1;
		}
	}

	.emoji-picker-container {
		margin-top: 10px;
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid rgba(255, 255, 255, 0.06);
		border-radius: 10px;
		padding: 10px;

		.emoji-tabs {
			display: flex;
			align-items: center;
			gap: 6px;
			margin-bottom: 8px;
			flex-wrap: wrap;

			.emoji-tab-btn {
				font-size: 11px;
				font-weight: 600;
				padding: 3px 8px;
				border-radius: 6px;
				color: rgba(255, 255, 255, 0.6);
				background: transparent;
				cursor: pointer;
				transition: all 120ms ease;

				&:hover {
					color: #ffffff;
					background: rgba(255, 255, 255, 0.06);
				}

				&.active {
					color: #ffffff;
					background: rgba(255, 255, 255, 0.12);
				}
			}

			.clear-emoji-btn {
				margin-left: auto;
				font-size: 11px;
				color: #f87171;
				padding: 2px 6px;
				cursor: pointer;

				&:hover {
					text-decoration: underline;
				}
			}
		}

		.emoji-grid {
			display: flex;
			flex-wrap: wrap;
			gap: 6px;

			.emoji-btn {
				display: flex;
				align-items: center;
				justify-content: center;
				width: 34px;
				height: 34px;
				font-size: 18px;
				border-radius: 8px;
				background: rgba(255, 255, 255, 0.04);
				border: 1px solid transparent;
				cursor: pointer;
				transition: all 140ms ease;

				&:hover {
					background: rgba(255, 255, 255, 0.12);
					transform: scale(1.15);
				}

				&.selected {
					background: rgba(99, 102, 241, 0.3);
					border-color: #818cf8;
					transform: scale(1.1);
				}
			}
		}
	}

	.swatches-grid {
		display: grid;
		grid-template-columns: repeat(7, 1fr);
		gap: 8px;
		margin-bottom: 12px;

		@media (max-width: 480px) {
			grid-template-columns: repeat(4, 1fr);
		}

		.swatch-btn {
			display: flex;
			align-items: center;
			justify-content: center;
			height: 32px;
			border-radius: 8px;
			border: 2px solid transparent;
			cursor: pointer;
			transition: all 140ms ease;

			&:hover {
				transform: translateY(-2px);
				filter: brightness(1.15);
			}

			&.active {
				border-color: #ffffff;
				box-shadow: 0 0 8px rgba(255, 255, 255, 0.4);
				transform: scale(1.05);
			}
		}
	}

	.hex-input-row {
		display: flex;
		align-items: center;
		gap: 12px;
		flex-wrap: wrap;

		.hex-field {
			display: flex;
			align-items: center;
			background: rgba(255, 255, 255, 0.05);
			border: 1px solid rgba(255, 255, 255, 0.1);
			border-radius: 8px;
			padding: 2px 8px;
			width: 130px;

			.hex-hash {
				font-size: 11px;
				font-weight: 700;
				color: rgba(255, 255, 255, 0.4);
				margin-right: 6px;
			}

			.hex-text-input {
				background: transparent;
				border: none;
				outline: none;
				color: #ffffff;
				font-family: monospace;
				font-size: 13px;
				width: 100%;
				padding: 4px 0;
			}
		}

		.color-picker-wrap {
			width: 36px;
			height: 36px;
			flex-shrink: 0;
		}

		.auto-contrast-toggle {
			display: flex;
			align-items: center;
			gap: 6px;
			font-size: 12px;
			color: rgba(255, 255, 255, 0.7);
			cursor: pointer;
			user-select: none;
			margin-left: auto;

			input {
				cursor: pointer;
			}
		}
	}

	.modal-footer {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: 10px;
		margin-top: 24px;
		padding-top: 16px;
		border-top: 1px solid rgba(255, 255, 255, 0.06);

		.cancel-btn {
			padding: 8px 16px;
			font-size: 13px;
		}

		.submit-btn {
			padding: 8px 20px;
			font-size: 13px;
			font-weight: 600;
		}
	}
</style>
