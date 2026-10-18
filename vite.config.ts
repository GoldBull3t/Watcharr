import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vite";
import { readFileSync } from "fs";
import { SvelteKitPWA } from "@vite-pwa/sveltekit";
const pkg = JSON.parse(readFileSync("package.json", "utf8"));

export default defineConfig({
	plugins: [
		sveltekit(),
		SvelteKitPWA({
			strategies: "generateSW",
			registerType: "autoUpdate",
			workbox: {
				globPatterns: ["**/*.{js,css,html,ico,png,svg,webp,woff,woff2}"],
				runtimeCaching: [
					{
						urlPattern: /^https:\/\/www\.themoviedb\.org\/t\/p\/.*/i,
						handler: "CacheFirst",
						options: {
							cacheName: "tmdb-images-cache",
							expiration: {
								maxEntries: 500,
								maxAgeSeconds: 60 * 60 * 24 * 30, // 30 days
							},
							cacheableResponse: {
								statuses: [0, 200],
							},
						},
					},
					{
						urlPattern: /\/api\/img\/.*/i,
						handler: "CacheFirst",
						options: {
							cacheName: "local-images-cache",
							expiration: {
								maxEntries: 200,
								maxAgeSeconds: 60 * 60 * 24 * 30, // 30 days
							},
							cacheableResponse: {
								statuses: [0, 200],
							},
						},
					},
					{
						urlPattern: /\/api\/watched.*/i,
						handler: "NetworkFirst",
						options: {
							cacheName: "watched-api-cache",
							networkTimeoutSeconds: 3,
							expiration: {
								maxEntries: 50,
								maxAgeSeconds: 60 * 60 * 24, // 24 hours
							},
							cacheableResponse: {
								statuses: [0, 200],
							},
						},
					},
				],
			},
			manifest: {
				name: "Watcharr",
				short_name: "Watcharr",
				description: "Your movie and show watched list.",
				background_color: "#f1da83",
				categories: ["entertainment"],
				start_url: "/",
				scope: "/",
				display: "standalone",
				icons: [
					{
						src: "/logo-sqre-144.png",
						sizes: "144x144",
						type: "image/png",
					},
					{
						src: "/logo-sqre-192.png",
						sizes: "192x192",
						type: "image/png",
					},
					{
						src: "/logo-sqre.png",
						sizes: "512x512",
						type: "image/png",
					},
					{
						src: "/logo-sqre.png",
						sizes: "512x512",
						type: "image/png",
						purpose: "any maskable",
					},
				],
			},
			devOptions: {
				enabled: false,
			},
		}),
	],
	server: {
		port: 5173,
		watch: {
			ignored: [
				"**/server/data/**",
				"**/server/**/*.log",
				"**/server/**/*.db*",
				"**/build/**",
				"**/.svelte-kit/**",
			],
		},
	},
	define: {
		__WATCHARR_VERSION__: JSON.stringify(pkg.version),
	},
});
