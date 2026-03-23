// Follows the best practices established in https://shiki.matsu.io/guide/best-performance
import { createJavaScriptRegexEngine } from "shiki/engine/javascript";
import { createHighlighterCore } from "shiki/core";

const bundledLanguages = {
	c: () => import("@shikijs/langs/c"),
	bash: () => import("@shikijs/langs/bash"),
	diff: () => import("@shikijs/langs/diff"),
	javascript: () => import("@shikijs/langs/javascript"),
	json: () => import("@shikijs/langs/json"),
	svelte: () => import("@shikijs/langs/svelte"),
	typescript: () => import("@shikijs/langs/typescript"),
	python: () => import("@shikijs/langs/python"),
	tsx: () => import("@shikijs/langs/tsx"),
	jsx: () => import("@shikijs/langs/jsx"),
	css: () => import("@shikijs/langs/css"),
	cpp: () => import("@shikijs/langs/cpp"),
	"c++": () => import("@shikijs/langs/cpp"),
	text: () => import("@shikijs/langs/markdown"),
};

/** The languages configured for the highlighter */
export type SupportedLanguage = keyof typeof bundledLanguages;

export const supportedLanguageIds = Object.keys(bundledLanguages) as SupportedLanguage[];

/** A preloaded highlighter instance. */
export const highlighter = createHighlighterCore({
	themes: [
		import("@shikijs/themes/github-light-default"),
		import("@shikijs/themes/github-dark-default"),
		import("@shikijs/themes/vesper"),
	],
	langs: Object.entries(bundledLanguages).map(([_, lang]) => lang),
	engine: createJavaScriptRegexEngine(),
});
