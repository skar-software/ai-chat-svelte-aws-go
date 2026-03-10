<script lang="ts">
	import { cn } from "$lib/utils";
	import type { HTMLAttributes } from "svelte/elements";

	interface Props extends HTMLAttributes<HTMLDivElement> {
		class?: string;
		children?: any;
	}

	let { class: className, children, ...restProps }: Props = $props();
</script>

<div
	class={cn("bg-card flex flex-col overflow-hidden rounded-lg border border-l-2 border-l-primary shadow-sm", className)}
	{...restProps}
>
	{@render children?.()}
</div>
