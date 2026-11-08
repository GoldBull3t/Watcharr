import { decode } from "blurhash";

const blurHashCache = new Map<string, ImageData>();

/**
 * Retrieves decoded BlurHash ImageData from in-memory cache or computes it.
 * Prevents synchronous DCT recomputation on the main thread during scrolling.
 */
export function getOrDecodeBlurHash(
	hash: string,
	width = 32,
	height = 48,
): ImageData | undefined {
	if (!hash) return undefined;
	const key = `${hash}_${width}x${height}`;
	const cached = blurHashCache.get(key);
	if (cached) return cached;

	try {
		const pixels = decode(hash, width, height);
		const imgData = new ImageData(new Uint8ClampedArray(pixels), width, height);
		// Keep cache bounded to prevent memory growth
		if (blurHashCache.size > 300) {
			const firstKey = blurHashCache.keys().next().value;
			if (firstKey) blurHashCache.delete(firstKey);
		}
		blurHashCache.set(key, imgData);
		return imgData;
	} catch (e) {
		console.warn("blurHashCache: Failed to decode blurhash", e);
		return undefined;
	}
}
