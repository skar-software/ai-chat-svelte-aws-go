<script lang="ts">
  import {
    Conversation,
    ConversationContent,
    ConversationEmptyState,
    ConversationScrollButton,
  } from "$lib/components/ai-elements/conversation";
  import { Message, MessageContent } from "$lib/components/ai-elements/message";
  import { Loader } from "$lib/components/ai-elements/loader";
  import * as Sidebar from "$lib/components/ui/sidebar/index.js";
  import ChatSidebar from "./ChatSidebar.svelte";
  import { SidebarTrigger } from "$lib/components/ui/sidebar/index.js";
  import { Sparkles } from "@lucide/svelte";
  import type { DemoPart, ChatMessage } from "./types";
  import ChatMessageItem from "./ChatMessageItem.svelte";
  import {
    Queue,
    QueueItem,
    QueueItemIndicator,
    QueueItemContent,
    QueueList,
    QueueSection,
    QueueSectionTrigger,
    QueueSectionLabel,
    QueueSectionContent,
  } from "$lib/components/ai-elements/queue";
  import {
    Context,
    ContextTrigger,
    ContextContent,
    ContextContentHeader,
    ContextContentBody,
    ContextContentFooter,
    ContextInputUsage,
    ContextOutputUsage,
  } from "$lib/components/ai-elements/context";
  import type { LanguageModelUsage } from "$lib/components/ai-elements/context/context-context.svelte";

  const emptyStateSuggestions = [
    { title: "Ask a question", text: "What can you help me with?" },
    { title: "Get started", text: "How does this work?" },
    {
      title: "Test UI elements",
      text: "Run a full UI elements integration test with code block, artifact, confirmation, plan, and queue in one response.",
    },
  ];
  import ChatInput from "./ChatInput.svelte";
  import type { PromptInputMessage } from "$lib/components/ai-elements/prompt-input";
  import { streamChat, type StreamPart } from "$lib/chat/api";

  /**
   * ChatWidget — Reusable AI chat interface for SaaS embedding.
   *
   * Props:
   *   apiBaseUrl  — Backend AI API base URL
   *   tenantId    — Multi-tenant routing identifier
   *   authToken   — Bearer token (optional if using cookie sessions)
   *   conversationId — Resume an existing conversation
   *   mode        — 'dark' | 'light' (follows system via mode-watcher if omitted)
   *   title       — Header title
   *   placeholder — Input placeholder text
   */
  interface ChatWidgetProps {
    apiBaseUrl?: string;
    tenantId?: string;
    authToken?: string;
    conversationId?: string;
    mode?: "dark" | "light";
    title?: string;
    placeholder?: string;
  }

  let {
    apiBaseUrl = "/api/chat",
    tenantId = "",
    authToken = "",
    conversationId = "",
    mode,
    title = "AI Assistant",
    placeholder = "Send a message...",
  }: ChatWidgetProps = $props();

  let isDark = $state(false);

  $effect(() => {
    if (mode) {
      isDark = mode === "dark";
      return;
    }

    const check = () => {
      isDark = document.documentElement.classList.contains("dark");
    };
    check();

    const observer = new MutationObserver(check);
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ["class"],
    });

    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const onMediaChange = () => check();
    mq.addEventListener("change", onMediaChange);

    return () => {
      observer.disconnect();
      mq.removeEventListener("change", onMediaChange);
    };
  });

  let effectiveDark = $derived(isDark);
  let shikiTheme = $derived(
    effectiveDark ? "github-dark-default" : "github-light-default",
  );

  let input = $state("");
  let status = $state<"idle" | "submitted" | "streaming" | "error">("idle");

  let totalInputTokens = $state(0);
  let totalOutputTokens = $state(0);
  const tokenUsage: LanguageModelUsage = $derived({
    inputTokens: totalInputTokens,
    outputTokens: totalOutputTokens,
  });
  const usedTokens = $derived(totalInputTokens + totalOutputTokens);
  const maxTokens = 128000;

  function handleRetry(assistantMsgId: string) {
    if (status !== "idle") return;
    const msgs = activeConversation?.messages ?? [];
    const assistantIdx = msgs.findIndex((m) => m.id === assistantMsgId);
    if (assistantIdx < 1) return;

    // Find the user message right before this assistant message
    let userMsg: ChatMessage | undefined;
    for (let i = assistantIdx - 1; i >= 0; i--) {
      if (msgs[i].role === "user") {
        userMsg = msgs[i];
        break;
      }
    }
    if (!userMsg) return;

    // Remove the old assistant message
    updateActiveConversation((conv) => ({
      ...conv,
      messages: conv.messages.filter((m) => m.id !== assistantMsgId),
    }));

    // Re-send the same user message
    processMessage(userMsg.content);
  }

  function handleUpdatePart(msgId: string, partIndex: number, updatedPart: DemoPart) {
    updateActiveConversation((conv) => {
      const msgIdx = conv.messages.findIndex((m) => m.id === msgId);
      if (msgIdx < 0) return conv;
      const msg = conv.messages[msgIdx];
      if (!msg.parts || partIndex < 0 || partIndex >= msg.parts.length) return conv;
      const newParts = [...msg.parts];
      newParts[partIndex] = updatedPart;
      const updated = [...conv.messages];
      updated[msgIdx] = { ...msg, parts: newParts };
      return { ...conv, messages: updated };
    });
  }

  type QueuedMessage = {
    id: string;
    text: string;
    status: "pending" | "in_progress" | "completed";
  };

  let messageQueue = $state<QueuedMessage[]>([]);

  type LocalConversation = {
    id: string;
    title: string;
    backendConversationId: string;
    messages: ChatMessage[];
  };

  let conversations = $state<LocalConversation[]>([
    {
      id: crypto.randomUUID(),
      title: "AI Elements Demo",
      backendConversationId: conversationId || "",
      messages: [],
    },
  ]);

  let activeConversationId = $state(conversations[0].id);

  const activeConversation = $derived(
    conversations.find((c) => c.id === activeConversationId) ?? conversations[0],
  );
  const activeMessages = $derived(activeConversation?.messages ?? []);
  const currentChatTitle = $derived(activeConversation?.title ?? "New conversation");
  const currentBackendConversationId = $derived(
    activeConversation?.backendConversationId ?? "",
  );

  function updateActiveConversation(
    updater: (conv: LocalConversation) => LocalConversation,
  ) {
    const idx = conversations.findIndex((c) => c.id === activeConversationId);
    if (idx < 0) return;
    const updated = updater(conversations[idx]);
    conversations = [
      ...conversations.slice(0, idx),
      updated,
      ...conversations.slice(idx + 1),
    ];
  }

  function startNewChat() {
    const newId = crypto.randomUUID();
    conversations = [
      ...conversations,
      {
        id: newId,
        title: "New conversation",
        backendConversationId: "",
        messages: [],
      },
    ];
    activeConversationId = newId;
    input = "";
    status = "idle";
  }

  function selectConversation(id: string) {
    if (status !== "idle") return; // don't switch while streaming
    activeConversationId = id;
    input = "";
    status = "idle";
  }

  const useRealApi = $derived(!!apiBaseUrl && apiBaseUrl.startsWith("http"));

  const lastAssistantMsg = $derived(
    [...activeMessages].reverse().find((m) => m.role === "assistant"),
  );
  const isStreamingEmpty = $derived(
    status === "streaming" && !!lastAssistantMsg && !lastAssistantMsg.content && !(lastAssistantMsg.parts?.length),
  );
  const showLoader = $derived(status === "submitted" || isStreamingEmpty);
  const messagesToShow = $derived(
    isStreamingEmpty
      ? activeMessages.filter((m) => m.id !== lastAssistantMsg?.id)
      : activeMessages,
  );

  function updateQueueItem(id: string, newStatus: QueuedMessage["status"]) {
    messageQueue = messageQueue.map((q) =>
      q.id === id ? { ...q, status: newStatus } : q,
    );
  }

  function cleanCompletedQueue() {
    messageQueue = messageQueue.filter((q) => q.status !== "completed");
  }

  async function processNextInQueue() {
    const next = messageQueue.find((q) => q.status === "pending");
    if (!next) {
      cleanCompletedQueue();
      return;
    }
    updateQueueItem(next.id, "in_progress");
    await processMessage(next.text, next.id);
  }

  async function onSubmit(message: PromptInputMessage, event?: SubmitEvent) {
    event?.preventDefault();
    const text = typeof message === "string" ? message : message.text;
    if (!text?.trim()) return;

    // If busy, enqueue — the user message will appear in chat once processing starts
    if (status !== "idle") {
      messageQueue = [
        ...messageQueue,
        { id: crypto.randomUUID(), text: text.trim(), status: "pending" },
      ];
      input = "";
      return;
    }

    await processMessage(text);
  }

  function updateMessageById(
    msgId: string,
    updater: (msg: ChatMessage) => ChatMessage,
  ) {
    updateActiveConversation((conv) => {
      const idx = conv.messages.findIndex((m) => m.id === msgId);
      if (idx < 0) return conv;
      const updated = [...conv.messages];
      updated[idx] = updater(updated[idx]);
      return { ...conv, messages: updated };
    });
  }

  function inferArtifactTitle(text: string): string {
    const heading = text.match(/^#{1,3}\s+(.+)$/m)?.[1]?.trim();
    if (heading) return heading.slice(0, 80);
    const firstLine = text.trim().split("\n")[0]?.trim() ?? "";
    if (!firstLine) return "Artifact";
    return firstLine.slice(0, 80);
  }

  function inferPlanFromText(text: string): DemoPart | null {
    const trimmed = text.trim();

    const stepMatches = [...trimmed.matchAll(/^\s*(?:\d+[\).\s-]+|[-*]\s+)(.+)$/gm)];
    const steps = stepMatches
      .map((m) => m[1]?.trim())
      .filter((s): s is string => !!s && s.length > 0)
      .slice(0, 8);

    if (steps.length < 3) return null;

    return {
      type: "plan",
      meta: {
        title: "Execution Plan",
        description: "Auto-selected by response format",
        steps,
      },
    };
  }

  function shouldPromoteToArtifact(text: string): boolean {
    const trimmed = text.trim();
    if (!trimmed) return false;
    if (trimmed.length < 280) return false;
    if (trimmed.includes("```json:part")) return false;

    const lower = trimmed.toLowerCase();
    const structuralSignals = [
      /^#{1,3}\s+/m.test(trimmed),
      /^\d+\.\s+/m.test(trimmed),
      /^-\s+\[[ xX]\]/m.test(trimmed),
      /^[-*]\s+/m.test(trimmed),
    ].filter(Boolean).length;

    const planLikeTerms = ["plan", "step-by-step", "step by step", "migration", "rollout"];
    if (planLikeTerms.some((keyword) => lower.includes(keyword))) return false;

    const intentSignals = [
      "runbook",
      "checklist",
      "playbook",
      "sop",
      "deployment",
      "release plan",
      "migration plan",
      "proposal",
      "spec",
      "roadmap",
      "report",
      "template",
    ].filter((keyword) => lower.includes(keyword)).length;

    // Promote long, structured, document-like answers to artifact cards.
    return structuralSignals >= 2 || (structuralSignals >= 1 && intentSignals >= 1);
  }

  function extractCitationSources(text: string): Array<{ title: string; url: string; description?: string }> {
    const sources = new Map<string, { title: string; url: string; description?: string }>();
    const markdownLinks = [...text.matchAll(/\[([^\]]+)\]\((https?:\/\/[^\s)]+)\)/g)];
    for (const m of markdownLinks) {
      const title = (m[1] ?? "").trim();
      const url = (m[2] ?? "").trim();
      if (!url) continue;
      sources.set(url, { title: title || "Source", url });
    }
    return [...sources.values()].slice(0, 8);
  }

  function inferIntentFromPrompt(prompt: string): "plan" | "artifact" | "confirmation" | "queue" | "citation" | null {
    const p = prompt.toLowerCase();
    if (
      p.includes("citation") ||
      p.includes("citations") ||
      p.includes("source") ||
      p.includes("sources") ||
      p.includes("reference") ||
      p.includes("references") ||
      p.includes("resource") ||
      p.includes("resources")
    ) {
      return "citation";
    }
    if (
      p.includes("task list") ||
      p.includes("todo") ||
      p.includes("to-do") ||
      p.includes("queue") ||
      p.includes("track") ||
      p.includes("tracking") ||
      p.includes("pending") ||
      p.includes("completed")
    ) {
      return "queue";
    }
    if (
      p.includes("approval") ||
      p.includes("approve") ||
      p.includes("confirm") ||
      p.includes("are you sure")
    ) {
      return "confirmation";
    }
    if (
      p.includes("runbook") ||
      p.includes("checklist") ||
      p.includes("sop") ||
      p.includes("playbook") ||
      p.includes("spec") ||
      p.includes("proposal") ||
      p.includes("report") ||
      p.includes("template") ||
      p.includes("artifact")
    ) {
      return "artifact";
    }
    if (
      p.includes("plan") ||
      p.includes("roadmap") ||
      p.includes("step-by-step") ||
      p.includes("step by step") ||
      p.includes("migration plan") ||
      p.includes("rollout plan")
    ) {
      return "plan";
    }
    return null;
  }

  function promoteAssistantMessageToStructuredPart(msgId: string, userPrompt: string) {
    updateMessageById(msgId, (msg) => {
      if (msg.role !== "assistant") return msg;
      if (msg.parts?.some((p) => p.type === "artifact" || p.type === "plan" || p.type === "confirmation" || p.type === "queue" || p.type === "citation")) return msg;

      const intent = inferIntentFromPrompt(userPrompt);
      let selectedPart: DemoPart | null = null;

      if (intent === "citation") {
        const sources = extractCitationSources(msg.content);
        if (sources.length > 0) {
          selectedPart = {
            type: "citation",
            content: msg.content.trim(),
            meta: { sources },
          };
        } else {
          // Keep plain text if we can't extract references.
          selectedPart = null;
        }
      } else if (intent === "queue") {
        const taskMatches = [...msg.content.matchAll(/^\s*(?:\d+[\).\s-]+|[-*]\s+)(.+)$/gm)];
        const todos = taskMatches
          .map((m, idx) => ({
            id: `t${idx + 1}`,
            title: (m[1] ?? "").trim(),
            description: "",
            status: "pending",
          }))
          .filter((t) => t.title.length > 0)
          .slice(0, 12);

        if (todos.length > 0) {
          selectedPart = {
            type: "queue",
            meta: {
              messages: [{ id: "m1", text: "Task list auto-generated from response" }],
              todos,
            },
          };
        }
      } else if (intent === "confirmation") {
        selectedPart = {
          type: "confirmation",
          meta: {
            title: "Action Required",
            description: msg.content.trim().slice(0, 240),
            state: "approval-requested",
            approval: { id: crypto.randomUUID() },
          },
        };
      } else if (intent === "plan") {
        selectedPart = inferPlanFromText(msg.content);
      } else if (intent === "artifact") {
        selectedPart = {
          type: "artifact",
          content: msg.content.trim(),
          meta: {
            title: inferArtifactTitle(msg.content),
            description: "Auto-selected by prompt intent",
          },
        };
      } else {
        const planPart = inferPlanFromText(msg.content);
        const artifactPart: DemoPart | null = shouldPromoteToArtifact(msg.content)
          ? {
              type: "artifact",
              content: msg.content.trim(),
              meta: {
                title: inferArtifactTitle(msg.content),
                description: "Auto-selected by response format",
              },
            }
          : null;
        selectedPart = planPart ?? artifactPart;
      }

      if (!selectedPart) return msg;

      return {
        ...msg,
        content: "",
        parts: [...(msg.parts ?? []), selectedPart],
      };
    });
  }

  async function processMessage(text: string, queueItemId?: string) {
    if (status !== "idle" && !queueItemId) return;

    updateActiveConversation((conv) => ({
      ...conv,
      messages: [...conv.messages, { id: crypto.randomUUID(), role: "user" as const, content: text }],
    }));
    input = "";
    status = "submitted";

    const assistantId = crypto.randomUUID();
    const assistantMessage: ChatMessage = {
      id: assistantId,
      role: "assistant",
      content: "",
    };
    updateActiveConversation((conv) => ({
      ...conv,
      messages: [...conv.messages, assistantMessage],
    }));
    status = "streaming";

    let finished = false;
    const finishProcessing = () => {
      if (finished) return;
      finished = true;
      if (queueItemId) {
        updateQueueItem(queueItemId, "completed");
      }
      status = "idle";
      processNextInQueue();
    };

    if (useRealApi) {
      try {
        const backendConversationId = currentBackendConversationId || "";
        await streamChat(
          {
            apiBaseUrl,
            authToken: authToken || undefined,
            tenantId: tenantId || undefined,
          },
          {
            input: text.trim(),
            conversation_id: backendConversationId || undefined,
          },
          {
            onDelta(delta) {
              updateMessageById(assistantId, (msg) => ({
                ...msg,
                content: msg.content + delta,
              }));
            },
            onPart(part: StreamPart) {
              const demoPart = part as unknown as DemoPart;
              updateMessageById(assistantId, (msg) => {
                const parts = [...(msg.parts ?? [])];
                const metaId = demoPart?.meta?.id;
                if (
                  demoPart.type === "code" &&
                  metaId !== undefined &&
                  metaId !== null
                ) {
                  const idx = parts.findIndex(
                    (p) => p.type === "code" && (p?.meta as any)?.id === metaId,
                  );
                  if (idx >= 0) {
                    parts[idx] = { ...parts[idx], ...demoPart };
                    return { ...msg, parts };
                  }
                }
                parts.push(demoPart);
                return { ...msg, parts };
              });
            },
            onUsage(inputTok, outputTok) {
              totalInputTokens += inputTok;
              totalOutputTokens += outputTok;
            },
            onCompleted(convId) {
              updateActiveConversation((conv) => ({
                ...conv,
                backendConversationId: convId,
              }));
              promoteAssistantMessageToStructuredPart(assistantId, text.trim());
              finishProcessing();
            },
            onError(errMessage) {
              updateMessageById(assistantId, (msg) => ({
                ...msg,
                content: msg.content || `Error: ${errMessage}`,
              }));
              finishProcessing();
            },
          }
        );
        finishProcessing();
      } catch (err) {
        const errMessage = err instanceof Error ? err.message : String(err);
        updateMessageById(assistantId, (msg) => ({
          ...msg,
          content: `Error: ${errMessage}`,
        }));
        finishProcessing();
      }
    } else {
      const dummyText = `Thanks for your message! I'm currently in **demo mode**. Here is an example of a generated code block:\n\n\`\`\`javascript\nfunction helloWorld() {\n  console.log("Hello, world!");\n  return true;\n}\n\`\`\`\n\nIn production, this response would come from your AI backend at \`${apiBaseUrl}\`${tenantId ? ` for tenant \`${tenantId}\`` : ""}.`;
      let i = 0;
      const interval = setInterval(() => {
        if (i < dummyText.length) {
          updateMessageById(assistantId, (msg) => ({
            ...msg,
            content: msg.content + dummyText[i],
          }));
          i++;
        } else {
          clearInterval(interval);
          finishProcessing();
        }
      }, 12);
    }
  }
</script>

<div class="chat-widget-root h-full w-full bg-background text-foreground">
  <Sidebar.Provider>
    <ChatSidebar
      conversations={conversations.map((c) => ({ id: c.id, title: c.title }))}
      activeConversationId={activeConversationId}
      onSelectConversation={selectConversation}
      onNewChat={startNewChat}
    />
    <Sidebar.Inset>
      <div class="flex h-full flex-col">
        <!-- Header -->
        <header
          class="flex h-12 items-center gap-2 border-b border-border px-4 shrink-0 bg-card"
        >
          <SidebarTrigger />
          <div class="text-sm font-semibold flex-1">{title}</div>
          {#if usedTokens > 0}
            <Context {usedTokens} {maxTokens} usage={tokenUsage} modelId="gpt-4o">
              <ContextTrigger />
              <ContextContent>
                <ContextContentHeader />
                <ContextContentBody>
                  <ContextInputUsage />
                  <ContextOutputUsage />
                </ContextContentBody>
                <ContextContentFooter />
              </ContextContent>
            </Context>
          {/if}
        </header>

        <!-- Conversation -->
        <Conversation class="flex-1 min-h-0">
          <ConversationContent class="mx-auto w-full max-w-3xl px-4 py-6">
            {#if activeMessages.length === 0}
              <ConversationEmptyState
                class="flex h-full flex-col items-center justify-center text-center pt-[15vh]"
              >
                <div class="mb-4">
                  <Sparkles class="size-8 text-primary" />
                </div>
                <p class="text-lg font-medium text-foreground max-w-sm mb-2">
                  What can I help you with today?
                </p>
                <p class="text-sm text-muted-foreground max-w-sm mb-8">
                  Send a message below to get started.
                </p>
                <div
                  class="grid grid-cols-1 sm:grid-cols-2 gap-2 w-full max-w-md"
                >
                  {#each emptyStateSuggestions as suggestion}
                    <button
                      onclick={() => onSubmit({ text: suggestion.text } as any)}
                      class="text-left p-3 rounded-lg text-xs border border-border bg-card hover:border-primary/50 hover:bg-secondary transition-colors"
                    >
                      {suggestion.title}
                    </button>
                  {/each}
                </div>
              </ConversationEmptyState>
            {/if}

            {#each messagesToShow as msg (msg.id)}
              <ChatMessageItem {msg} {shikiTheme} onUpdatePart={handleUpdatePart} onRetry={handleRetry} />
            {/each}

            {#if showLoader}
              <Message from="assistant" class="mb-2">
                <MessageContent variant="flat">
                  <Loader />
                </MessageContent>
              </Message>
            {/if}
          </ConversationContent>
          <ConversationScrollButton />
        </Conversation>

        <!-- Message Queue -->
        {#if messageQueue.length > 0}
          <div class="mx-auto w-full max-w-3xl px-4 pt-2">
            <Queue>
              {#if messageQueue.filter((q) => q.status === "in_progress").length > 0}
                <QueueList>
                  {#each messageQueue.filter((q) => q.status === "in_progress") as item (item.id)}
                    <QueueItem>
                      <div class="flex items-center gap-2">
                        <QueueItemIndicator />
                        <QueueItemContent>
                          Processing: {item.text.length > 60 ? item.text.slice(0, 60) + "…" : item.text}
                        </QueueItemContent>
                      </div>
                    </QueueItem>
                  {/each}
                </QueueList>
              {/if}

              {#if messageQueue.filter((q) => q.status === "pending").length > 0}
                <QueueSection open>
                  <QueueSectionTrigger>
                    <QueueSectionLabel
                      label="Pending"
                      count={messageQueue.filter((q) => q.status === "pending").length}
                    />
                  </QueueSectionTrigger>
                  <QueueSectionContent>
                    <QueueList>
                      {#each messageQueue.filter((q) => q.status === "pending") as item (item.id)}
                        <QueueItem>
                          <div class="flex items-center gap-2">
                            <QueueItemIndicator completed={false} />
                            <QueueItemContent completed={false}>
                              {item.text.length > 60 ? item.text.slice(0, 60) + "…" : item.text}
                            </QueueItemContent>
                          </div>
                        </QueueItem>
                      {/each}
                    </QueueList>
                  </QueueSectionContent>
                </QueueSection>
              {/if}

              {#if messageQueue.filter((q) => q.status === "completed").length > 0}
                <QueueSection>
                  <QueueSectionTrigger>
                    <QueueSectionLabel
                      label="Completed"
                      count={messageQueue.filter((q) => q.status === "completed").length}
                    />
                  </QueueSectionTrigger>
                  <QueueSectionContent>
                    <QueueList>
                      {#each messageQueue.filter((q) => q.status === "completed") as item (item.id)}
                        <QueueItem>
                          <div class="flex items-center gap-2">
                            <QueueItemIndicator completed={true} />
                            <QueueItemContent completed={true}>
                              {item.text.length > 60 ? item.text.slice(0, 60) + "…" : item.text}
                            </QueueItemContent>
                          </div>
                        </QueueItem>
                      {/each}
                    </QueueList>
                  </QueueSectionContent>
                </QueueSection>
              {/if}
            </Queue>
          </div>
        {/if}

        <!-- Input -->
        <ChatInput bind:input {status} {placeholder} {onSubmit} />
      </div>
    </Sidebar.Inset>
  </Sidebar.Provider>
</div>
