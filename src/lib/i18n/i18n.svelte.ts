import { en, type TranslationKey } from "./locales/en";
import { pt } from "./locales/pt";

export type SupportedLocale = "en" | "pt";

const dictionaries: Record<SupportedLocale, Record<TranslationKey, string>> = {
	en,
	pt,
};

function getInitialLocale(): SupportedLocale {
	if (typeof window === "undefined") return "en";
	try {
		const saved = localStorage.getItem("watcharr_lang") as SupportedLocale;
		if (saved && (saved === "en" || saved === "pt")) {
			return saved;
		}
		const browserLang = navigator.language?.toLowerCase() || "";
		if (browserLang.startsWith("pt")) {
			return "pt";
		}
	} catch {
		// fallback to en if storage not accessible
	}
	return "en";
}

class I18nManager {
	currentLocale = $state<SupportedLocale>(getInitialLocale());

	setLocale(locale: SupportedLocale) {
		this.currentLocale = locale;
		if (typeof window !== "undefined") {
			try {
				localStorage.setItem("watcharr_lang", locale);
				document.documentElement.lang = locale;
			} catch {
				// ignore
			}
		}
	}

	t(key: TranslationKey, params?: Record<string, string | number>): string {
		const dict = dictionaries[this.currentLocale] || en;
		let str: string = dict[key] || en[key] || key;
		if (params) {
			for (const [k, v] of Object.entries(params)) {
				str = str.replace(new RegExp(`\\{${k}\\}`, "g"), String(v));
			}
		}
		return str;
	}
}

export const i18n = new I18nManager();
export const t = (key: TranslationKey, params?: Record<string, string | number>) =>
	i18n.t(key, params);
export type { TranslationKey };

