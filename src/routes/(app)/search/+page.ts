import type { PageLoad } from "./$types";

export const load = (async ({ url }) => {
	return {
		query: url.searchParams.get("query"),
	};
}) satisfies PageLoad;
