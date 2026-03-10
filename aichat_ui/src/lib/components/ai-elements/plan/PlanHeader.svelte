<script lang="ts">
	import { CardHeader } from "$lib/components/ui/card/index.js";
	import { cn } from "$lib/utils";
	import type { PlanHeaderProps } from "./types.js";

	let {
		class: className,
		children,
		...restProps
	}: PlanHeaderProps = $props();
</script>

<CardHeader
	class={cn(
		"flex flex-row items-center justify-between space-y-0 p-4 pb-2",
		className,
	)}
	data-slot="plan-header"
	{...restProps}
>
	{@render children?.()}
</CardHeader>
