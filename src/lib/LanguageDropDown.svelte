<script lang="ts">
	import DropDown from "./DropDown.svelte";
	import type { DropDownItem } from "@/types";
	import Error from "./Error.svelte";
	import { req } from "./util/api";

	interface Props {
		selectedLanguage?: string;
		disabled?: boolean;
		onChange: (language: string) => void;
	}

	let {
		selectedLanguage = $bindable("en-US"),
		disabled = false,
		onChange,
	}: Props = $props();

	let languagesPromise = getLanguages();

	function getLanguageLabel(code: string): string {
		try {
			const englishNames = new Intl.DisplayNames(["en"], { type: "language" });
			const name = englishNames.of(code);
			return name ? `${name} (${code})` : code;
		} catch {
			return code;
		}
	}

	async function getLanguages(): Promise<DropDownItem[]> {
		const langs = await req.get<string[]>(`/content/languages`);
		return (langs ?? [])
			.map((code) => {
				return {
					id: code,
					value: getLanguageLabel(code),
				} as DropDownItem;
			})
			.sort((a, b) => a.value.localeCompare(b.value));
	}
</script>

{#await languagesPromise}
	<DropDown placeholder="Select a language" options={[]} disabled={true} />
{:then mappedLanguages}
	<DropDown
		placeholder="Select a language"
		bind:active={selectedLanguage}
		options={mappedLanguages}
		onChange={() => {
			if (selectedLanguage) onChange(selectedLanguage);
		}}
		isDropDownItem={true}
		showActiveElementsInOptions={true}
		{disabled}
	/>
{:catch err}
	<Error error={err} pretty="Failed to load languages!" />
{/await}
