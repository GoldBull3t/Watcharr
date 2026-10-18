import adapter from "@sveltejs/adapter-static";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";
import { sveltePreprocess } from "svelte-preprocess";

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: [
		sveltePreprocess({
			scss: {
				prependData:
					`@use "./src/styles/_vars.scss" as *;` +
					`@use "./src/styles/_mixins.scss" as *;`,
			},
		}),
		vitePreprocess(),
	],

	kit: {
		adapter: adapter({
			pages: "build",
			assets: "build",
			fallback: "index.html",
			precompress: false,
			strict: false,
		}),

		alias: {
			"@": "src",
		},
	},
};

export default config;