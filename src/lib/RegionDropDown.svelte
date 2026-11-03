<script lang="ts">
	import DropDown from "./DropDown.svelte";
	import type { DropDownItem, TMDBRegions } from "@/types";
	import Error from "./Error.svelte";
	import { req } from "./util/api";

	interface Props {
		selectedCountry?: string;
		disabled?: boolean;
		onChange: (country: string) => void;
	}

	let {
		selectedCountry = $bindable("US"),
		disabled = false,
		onChange,
	}: Props = $props();

	let countriesPromise = getCountries();

	async function getCountries(): Promise<DropDownItem[]> {
		const c = await req.get<TMDBRegions>(`/content/regions`);
		return (c?.results ?? [])
			.map((cc) => {
				return {
					id: cc.iso_3166_1,
					value: cc.english_name,
				} as DropDownItem;
			})
			.sort((a, b) => a.value.localeCompare(b.value));
	}
</script>

{#await countriesPromise}
	<DropDown
		placeholder="Select a country"
		options={[]}
		disabled={true}
	/>
{:then mappedCountries}
	<DropDown
		placeholder="Select a country"
		bind:active={selectedCountry}
		options={mappedCountries}
		onChange={() => {
			if (selectedCountry) onChange(selectedCountry);
		}}
		isDropDownItem={true}
		showActiveElementsInOptions={true}
		{disabled}
	/>
{:catch err}
	<Error error={err} pretty="Failed to load countries!" />
{/await}
