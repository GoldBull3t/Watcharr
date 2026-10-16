/**
 * Utility functions and curated palettes for Watcharr tags with emojis and custom colors.
 */

export interface TagColorPreset {
	name: string;
	bgColor: string;
	textColor: string;
}

export const TAG_COLOR_PRESETS: TagColorPreset[] = [
	{ name: "Cinema Amber", bgColor: "#F59E0B", textColor: "#0f172a" },
	{ name: "Sunset Coral", bgColor: "#F97316", textColor: "#ffffff" },
	{ name: "Neon Rose", bgColor: "#F43F5E", textColor: "#ffffff" },
	{ name: "Crimson Velvet", bgColor: "#EF4444", textColor: "#ffffff" },
	{ name: "Fuchsia Glow", bgColor: "#D946EF", textColor: "#ffffff" },
	{ name: "Cyber Violet", bgColor: "#8B5CF6", textColor: "#ffffff" },
	{ name: "Royal Indigo", bgColor: "#6366F1", textColor: "#ffffff" },
	{ name: "Electric Azure", bgColor: "#0EA5E9", textColor: "#ffffff" },
	{ name: "Cyber Cyan", bgColor: "#06B6D4", textColor: "#0f172a" },
	{ name: "Mint Marine", bgColor: "#14B8A6", textColor: "#0f172a" },
	{ name: "Emerald Velvet", bgColor: "#10B981", textColor: "#0f172a" },
	{ name: "Lime Neon", bgColor: "#84CC16", textColor: "#0f172a" },
	{ name: "Slate Graphite", bgColor: "#64748B", textColor: "#ffffff" },
	{ name: "Obsidian Dark", bgColor: "#334155", textColor: "#ffffff" },
];

export interface TagEmojiCategory {
	category: string;
	emojis: string[];
}

export const TAG_EMOJI_CATEGORIES: TagEmojiCategory[] = [
	{
		category: "Cinema & TV",
		emojis: ["🎬", "🍿", "📺", "🎥", "🎞️", "📼", "🎟️", "🛋️"],
	},
	{
		category: "Favorites & Hype",
		emojis: ["⭐", "🌟", "💖", "🏆", "👑", "💎", "🔥", "⚡"],
	},
	{
		category: "Vibes & Seasons",
		emojis: ["🎃", "👻", "🚀", "🕶️", "🕵️", "🍕", "☕", "🌙"],
	},
	{
		category: "Gaming & Worlds",
		emojis: ["🎮", "🕹️", "👾", "⚔️", "🩸", "🤖", "🪄", "🌈"],
	},
];

/**
 * Splits a tag name into an optional leading emoji and the text label.
 * Supports Unicode 15+ emojis, modifiers, skin tones, and zero-width joiners.
 *
 * Example:
 *   "🍿 Movie Night" -> { emoji: "🍿", label: "Movie Night" }
 *   "Comfort Watches" -> { emoji: undefined, label: "Comfort Watches" }
 *   "⭐" -> { emoji: "⭐", label: "⭐" }
 */
export function splitTagEmoji(name: string | undefined): {
	emoji?: string;
	label: string;
} {
	if (!name) return { label: "" };
	const trimmed = name.trim();
	if (!trimmed) return { label: "" };

	// Matches leading emoji character(s)
	const regex =
		/^(\p{Extended_Pictographic}(?:\p{Emoji_Modifier}|\uFE0F|\u200D\p{Extended_Pictographic})*)\s*(.*)$/u;
	const match = trimmed.match(regex);

	if (match && match[1]) {
		const emoji = match[1];
		const label = match[2] ? match[2].trim() : emoji;
		return { emoji, label };
	}

	return { label: trimmed };
}

/**
 * Combines an emoji prefix and clean label into a tag name string.
 */
export function formatTagName(
	emoji: string | undefined,
	label: string | undefined,
): string {
	const cleanEmoji = emoji?.trim() || "";
	const cleanLabel = label?.trim() || "";

	if (cleanEmoji && cleanLabel) {
		return `${cleanEmoji} ${cleanLabel}`;
	}
	return cleanEmoji || cleanLabel;
}

/**
 * Determines whether text should be dark or white based on background color luminance.
 */
export function getContrastTextColor(hexColor: string | undefined): string {
	if (!hexColor || !hexColor.startsWith("#")) return "#ffffff";
	let hex = hexColor.replace("#", "").trim();
	if (hex.length === 3) {
		hex = hex
			.split("")
			.map((c) => c + c)
			.join("");
	}
	if (hex.length !== 6) return "#ffffff";

	const r = parseInt(hex.substring(0, 2), 16);
	const g = parseInt(hex.substring(2, 4), 16);
	const b = parseInt(hex.substring(4, 6), 16);

	if (isNaN(r) || isNaN(g) || isNaN(b)) return "#ffffff";

	// ITU-R BT.709 relative luminance formula
	const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255;
	return luminance > 0.55 ? "#0f172a" : "#ffffff";
}

/**
 * Converts a hex string to rgba with a specified alpha transparency.
 */
export function hexToRgba(
	hexColor: string | undefined,
	alpha: number,
	fallback = "rgba(255, 255, 255, 0.1)",
): string {
	if (!hexColor || !hexColor.startsWith("#")) return fallback;
	let hex = hexColor.replace("#", "").trim();
	if (hex.length === 3) {
		hex = hex
			.split("")
			.map((c) => c + c)
			.join("");
	}
	if (hex.length !== 6) return fallback;

	const r = parseInt(hex.substring(0, 2), 16);
	const g = parseInt(hex.substring(2, 4), 16);
	const b = parseInt(hex.substring(4, 6), 16);

	if (isNaN(r) || isNaN(g) || isNaN(b)) return fallback;
	return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}

/**
 * Validates a HEX color code (e.g. #fff, #10b981).
 */
export function isValidHex(hex: string): boolean {
	return /^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$/.test(hex.trim());
}

