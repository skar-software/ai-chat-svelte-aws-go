<script lang="ts">
    import {
        Message,
        MessageContent,
    } from "$lib/components/ai-elements/message";
    import { Response } from "$lib/components/ai-elements/response";
    import * as Artifact from "$lib/components/ai-elements/artifact";
    import * as Code from "$lib/components/ai-elements/code";
    import {
        Confirmation,
        ConfirmationTitle,
        ConfirmationRequest,
        ConfirmationActions,
        ConfirmationAction,
    } from "$lib/components/ai-elements/confirmation";
    import {
        Plan,
        PlanHeader,
        PlanTitle,
        PlanTrigger,
        PlanContent,
        PlanDescription,
    } from "$lib/components/ai-elements/plan";
    import { Code2, CircleAlert } from "@lucide/svelte";
    import type { ChatMessage } from "./types";

    interface ChatMessageItemProps {
        msg: ChatMessage;
        shikiTheme?: string;
    }

    let { msg, shikiTheme = "github-dark-default" }: ChatMessageItemProps =
        $props();
</script>

<Message from={msg.role} class="mb-2">
    <MessageContent variant="flat">
        {#if msg.role === "user"}
            {msg.content}
        {:else if msg.parts && msg.parts.length > 0}
            {#if msg.content}
                <div class="mb-3">
                    <Response content={msg.content} theme={shikiTheme} />
                </div>
            {/if}

            {#each msg.parts as part}
                {#if part.type === "code" && part.content}
                    <div class="my-3">
                        <Code.Root
                            code={part.content}
                            lang={part.meta?.lang ?? "typescript"}
                        >
                            <Code.Overflow>
                                <Code.CopyButton />
                            </Code.Overflow>
                        </Code.Root>
                    </div>
                {:else if part.type === "artifact" && part.content}
                    <div class="my-3">
                        <Artifact.Root>
                            <Artifact.Header
                                class="flex items-center justify-between p-3"
                            >
                                <div class="flex items-center gap-2">
                                    <div
                                        class="flex size-7 items-center justify-center rounded-md bg-secondary text-foreground"
                                    >
                                        <Code2 class="size-4" />
                                    </div>
                                    <div>
                                        <Artifact.Title
                                            class="text-xs font-semibold"
                                            >{part.meta?.title ??
                                                "Artifact"}</Artifact.Title
                                        >
                                        <Artifact.Description
                                            class="text-[10px] text-muted-foreground"
                                            >{part.meta?.description ??
                                                ""}</Artifact.Description
                                        >
                                    </div>
                                </div>
                            </Artifact.Header>
                            <Artifact.Content class="px-3 pb-3">
                                <Response
                                    content={part.content}
                                    theme={shikiTheme}
                                />
                            </Artifact.Content>
                        </Artifact.Root>
                    </div>
                {:else if part.type === "confirmation"}
                    <div class="my-3">
                        <Confirmation
                            state={part.meta?.state ?? "approval-requested"}
                            approval={part.meta?.approval}
                        >
                            <ConfirmationTitle>
                                <div class="flex items-center gap-2">
                                    <CircleAlert class="size-4" />
                                    <span class="font-semibold text-sm"
                                        >{part.meta?.title ??
                                            "Action Required"}</span
                                    >
                                </div>
                            </ConfirmationTitle>
                            <ConfirmationRequest>
                                <p class="text-sm text-muted-foreground">
                                    {part.meta?.description ?? ""}
                                </p>
                            </ConfirmationRequest>
                            <ConfirmationActions>
                                <ConfirmationAction variant="outline"
                                    >Deny</ConfirmationAction
                                >
                                <ConfirmationAction>Approve</ConfirmationAction>
                            </ConfirmationActions>
                        </Confirmation>
                    </div>
                {:else if part.type === "plan"}
                    <div class="my-3 w-full">
                        <Plan class="w-full">
                            <PlanHeader>
                                <div class="flex flex-col gap-1">
                                    <PlanTitle
                                        >{part.meta?.title ?? "Plan"}</PlanTitle
                                    >
                                    <PlanDescription
                                        >{part.meta?.description ??
                                            ""}</PlanDescription
                                    >
                                </div>
                                <PlanTrigger />
                            </PlanHeader>
                            <PlanContent>
                                {#if part.meta?.steps}
                                    <ul
                                        class="ml-3 pb-4 list-decimal space-y-2 text-sm text-foreground"
                                    >
                                        {#each part.meta.steps as step}
                                            <li>
                                                <Response
                                                    content={step}
                                                    theme={shikiTheme}
                                                />
                                            </li>
                                        {/each}
                                    </ul>
                                {/if}
                            </PlanContent>
                        </Plan>
                    </div>
                {/if}
            {/each}
        {/if}
    </MessageContent>
</Message>
