<script lang="ts">
	import { Collapsible } from "$lib/components/ui/collapsible/index.js";
	import { Card } from "$lib/components/ui/card/index.js";
	import { cn } from "$lib/utils";
	import { setPlanContext } from "./plan-context.svelte.js";
	import type { PlanProps } from "./types.js";

	let {
		class: className,
		isStreaming = false,
		children,
		...restProps
	}: PlanProps = $props();

	// Set context for child components
	setPlanContext({ isStreaming });
</script>

<Collapsible data-slot="plan" {...restProps} class={cn("w-full", className)}>
	<Card class={cn("w-full shadow-none", className)}>
		{@render children?.()}
	</Card>
</Collapsible>
