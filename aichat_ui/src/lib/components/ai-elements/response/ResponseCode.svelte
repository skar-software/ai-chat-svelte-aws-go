<script lang="ts">
	import * as Code from "$lib/components/ai-elements/code";
	import type { Tokens } from "marked";
	import type { SupportedLanguage } from "$lib/components/ai-elements/code/shiki";

	let { token, id }: { token: Tokens.Code; id: string } = $props();

	// Some inline code variants might pass different props, though the type says Tokens.Code
	// Just to be safe, we fall back to raw values if it behaves differently.
	let codeText = $derived(token?.text || "");
	let codeLang = $derived.by(() => {
		const raw = String(token?.lang ?? "plaintext").trim().toLowerCase();
		// Marked often uses "plaintext" for code fences without a language.
		if (raw === "plaintext") return "text";
		return raw as SupportedLanguage;
	});
</script>

<div class="my-3 flex w-full flex-col">
	<Code.Root code={codeText} lang={codeLang}>
		<Code.Overflow>
			<Code.CopyButton />
		</Code.Overflow>
	</Code.Root>
</div>
