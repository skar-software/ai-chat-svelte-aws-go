<script lang="ts">
	import {
		PromptInput,
		PromptInputActionAddAttachments,
		PromptInputActionMenu,
		PromptInputActionMenuContent,
		PromptInputActionMenuTrigger,
		PromptInputAttachment,
		PromptInputAttachments,
		PromptInputBody,
		type PromptInputMessage,
		PromptInputSubmit,
		PromptInputTextarea,
		PromptInputToolbar,
		PromptInputTools,
	} from "$lib/components/ai-elements/prompt-input/index.js";

	interface ChatInputProps {
		placeholder?: string;
		input: string;
		status: "idle" | "submitted" | "streaming" | "error";
		onSubmit: (message: PromptInputMessage, event?: SubmitEvent) => void;
	}

	let {
		placeholder = "Send a message...",
		input = $bindable(),
		status,
		onSubmit,
	}: ChatInputProps = $props();
</script>

<div class="mx-auto w-full max-w-3xl px-4 pb-4 pt-2 shrink-0">
	<PromptInput globalDrop multiple onSubmit={onSubmit}>
		<PromptInputBody>
			<PromptInputAttachments>
				{#snippet children(attachment)}
					<PromptInputAttachment data={attachment} />
				{/snippet}
			</PromptInputAttachments>
			<PromptInputTextarea
				{placeholder}
				bind:value={input}
				onchange={(e) => (input = (e.target as HTMLTextAreaElement).value)}
			/>
		</PromptInputBody>
		<PromptInputToolbar>
			<PromptInputTools>
				<PromptInputActionMenu>
					<PromptInputActionMenuTrigger />
					<PromptInputActionMenuContent>
						<PromptInputActionAddAttachments />
					</PromptInputActionMenuContent>
				</PromptInputActionMenu>
			</PromptInputTools>
			<PromptInputSubmit {status} />
		</PromptInputToolbar>
	</PromptInput>
	<p class="mt-2 text-center text-[10px] text-muted-foreground">
		AI responses may be inaccurate. Verify important information.
	</p>
</div>
