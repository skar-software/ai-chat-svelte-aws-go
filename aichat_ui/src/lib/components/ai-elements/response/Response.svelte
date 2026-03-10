<script lang="ts">
	import { Streamdown, type StreamdownProps } from "svelte-streamdown";
	import Code from "svelte-streamdown/code";
	import { cn } from "$lib/utils";

	import githubLightDefault from "@shikijs/themes/github-light-default";
	import githubDarkDefault from "@shikijs/themes/github-dark-default";

	type Props = StreamdownProps & {
		class?: string;
		theme?: string;
	};

	let {
		class: className,
		theme = "github-dark-default",
		...restProps
	}: Props = $props();
</script>

<Streamdown
	class={cn(
		"size-full [&>*:first-child]:mt-0 [&>*:last-child]:mb-0",
		className,
	)}
	shikiTheme={theme}
	baseTheme="shadcn"
	controls={{ table: false }}
	components={{ code: Code }}
	shikiThemes={{
		"github-light-default": githubLightDefault,
		"github-dark-default": githubDarkDefault,
	}}
	{...restProps}
/>
