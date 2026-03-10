<script lang="ts">
    import {
        PromptInput,
        PromptInputBody,
        PromptInputTextarea,
        PromptInputToolbar,
        PromptInputSubmit,
        type PromptInputMessage,
    } from "$lib/components/ai-elements/prompt-input";

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
    <div class="relative rounded-2xl border border-border bg-card">
        <PromptInput {onSubmit} class="flex items-end gap-2 p-2">
            <PromptInputBody class="bg-transparent border-none flex-1 min-w-0">
                <PromptInputTextarea
                    {placeholder}
                    bind:value={input}
                    class="max-h-[200px] min-h-[40px] resize-none bg-transparent py-2 pl-2 pr-0 text-sm focus-visible:ring-0 leading-normal"
                />
            </PromptInputBody>
            <PromptInputToolbar class="pb-1 pr-1 shrink-0">
                <PromptInputSubmit
                    {status}
                    disabled={!input.trim() && status !== "streaming"}
                    class="size-8 rounded-full bg-primary text-primary-foreground hover:bg-primary/80 disabled:opacity-30 transition-colors"
                />
            </PromptInputToolbar>
        </PromptInput>
    </div>
    <p class="mt-2 text-center text-[10px] text-muted-foreground">
        AI responses may be inaccurate. Verify important information.
    </p>
</div>
