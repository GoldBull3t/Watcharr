import path from "node:path";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";
import { sveltePreprocess } from "svelte-preprocess";

/**
 * Lightweight static SPA adapter that outputs client assets and index.html fallback
 * directly into `build/`, eliminating the need for a Node.js runtime in production.
 * @param {{ out?: string; fallback?: string; precompress?: boolean }} [opts]
 * @returns {import('@sveltejs/kit').Adapter}
 */
function staticSpaAdapter(opts = {}) {
	const { out = "build", fallback = "index.html", precompress = true } = opts;
	return {
		name: "watcharr-static-spa-adapter",
		async adapt(builder) {
			builder.rimraf(out);
			builder.mkdirp(out);

			builder.log.minor("Copying static client assets");
			builder.writeClient(out);
			builder.writePrerendered(out);

			builder.log.minor(`Generating SPA fallback (${fallback})`);
			await builder.generateFallback(path.join(out, fallback));

			if (precompress) {
				builder.log.minor("Precompressing static assets (.gz & .br)");
				await builder.compress(out);
			}
		},
	};
}

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
		adapter: staticSpaAdapter({
			out: "build",
			fallback: "index.html",
			precompress: true,
		}),

		alias: {
			"@": "src",
		},
	},
};

export default config;
