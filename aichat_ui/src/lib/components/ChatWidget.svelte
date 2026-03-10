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
  import { type ChatMessage } from "./types";
  import { demoMessages, suggestions } from "./mockData";
  import ChatMessageItem from "./ChatMessageItem.svelte";
  import ChatInput from "./ChatInput.svelte";
  import type { PromptInputMessage } from "$lib/components/ai-elements/prompt-input";

  /**
   * ChatWidget — Reusable AI chat interface for SaaS embedding.
   *
   * The ChatWidget is a reusable AI chat interface built with SvelteKit that
   * can be embedded into different parts of a SaaS application—such as
   * customer-facing pages, internal dashboards, or admin tools—to provide
   * an interactive assistant powered by backend AI services.
   *
   * The widget handles conversation UI, message streaming, prompts, and
   * responses while communicating securely with a backend API (for example
   * a Go service using Amazon Bedrock or other AI providers).
   *
   * Its purpose is to offer a consistent, modular chat experience that
   * requires only configuration inputs like apiBaseUrl, tenantId, and
   * authentication context, allowing developers to easily integrate
   * AI-driven help, automation, or knowledge retrieval features across
   * multiple areas of the product without exposing cloud credentials
   * in the browser.
   *
   * Usage:
   *   <ChatWidget
   *     apiBaseUrl="https://api.example.com"
   *     tenantId="tenant-123"
   *     authToken="bearer-token"
   *     mode="dark"
   *   />
   */
  interface ChatWidgetProps {
    /** Base URL for the backend AI API */
    apiBaseUrl?: string;
    /** Tenant or workspace identifier for multi-tenant routing */
    tenantId?: string;
    /** Bearer token for API authentication (optional if using cookie sessions) */
    authToken?: string;
    /** Resume an existing conversation by ID */
    conversationId?: string;
    /** Visual theme — 'dark' or 'light'. If not provided, it will follow the device/system preference via the 'dark' class on the html element. */
    mode?: "dark" | "light";
    /** Header title displayed in the navigation bar */
    title?: string;
    /** Placeholder text for the message input */
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

  // Helper to check if the page is in dark mode
  let isDarkModeActive = $state(false);

  $effect(() => {
    const checkDark = () => {
      isDarkModeActive = document.documentElement.classList.contains("dark");
    };
    checkDark();

    const observer = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        if (
          mutation.type === "attributes" &&
          mutation.attributeName === "class"
        ) {
          checkDark();
        }
      });
    });

    observer.observe(document.documentElement, { attributes: true });
    return () => observer.disconnect();
  });

  let isDark = $derived(mode ? mode === "dark" : isDarkModeActive);
  let shikiTheme = $derived(
    isDark ? "github-dark-default" : "github-light-default",
  );

  let messages = $state<ChatMessage[]>([...demoMessages]);
  let input = $state("");
  let status = $state<"idle" | "submitted" | "streaming" | "error">("idle");

  async function onSubmit(message: PromptInputMessage, event?: SubmitEvent) {
    event?.preventDefault();
    const text = typeof message === "string" ? message : message.text;
    if (!text?.trim() || status !== "idle") return;

    const userMessage: ChatMessage = {
      id: crypto.randomUUID(),
      role: "user",
      content: text,
    };

    messages = [...messages, userMessage];
    input = "";
    status = "submitted";

    setTimeout(() => {
      status = "streaming";
      const assistantMessage: ChatMessage = {
        id: crypto.randomUUID(),
        role: "assistant",
        content: "",
      };
      messages = [...messages, assistantMessage];

      const dummyText = `Thanks for your message! I'm currently in **demo mode**.\n\nIn production, this response would come from your AI backend at \`${apiBaseUrl}\`${tenantId ? ` for tenant \`${tenantId}\`` : ""}.`;
      let i = 0;
      const interval = setInterval(() => {
        if (i < dummyText.length) {
          // Re-assign the array to trigger Svelte 5 reactivity
          const updatedMessages = [...messages];
          const lastMsg = { ...updatedMessages[updatedMessages.length - 1] };
          lastMsg.content += dummyText[i];
          updatedMessages[updatedMessages.length - 1] = lastMsg;
          messages = updatedMessages;
          i++;
        } else {
          clearInterval(interval);
          status = "idle";
        }
      }, 12);
    }, 600);
  }
</script>

<div
  class="chat-widget-root h-full w-full bg-background text-foreground {isDark
    ? 'dark'
    : ''}"
>
  <Sidebar.Provider>
    <ChatSidebar currentChatTitle="AI Elements Demo" />
    <Sidebar.Inset>
      <div class="flex h-full flex-col">
        <!-- Header -->
        <header
          class="flex h-12 items-center gap-2 border-b border-border px-4 shrink-0 bg-card"
        >
          <SidebarTrigger />
          <div class="text-sm font-semibold">{title}</div>
        </header>

        <!-- Conversation -->
        <Conversation class="flex-1 min-h-0">
          <ConversationContent class="mx-auto w-full max-w-3xl px-4 py-6">
            {#if messages.length === 0}
              <ConversationEmptyState
                class="flex h-full flex-col items-center justify-center text-center pt-[15vh]"
              >
                <div class="mb-4">
                  <Sparkles class="size-8 text-primary" />
                </div>
                <p class="text-sm text-muted-foreground max-w-sm mb-8">
                  A reusable AI chat interface. It connects to your backend API
                  and supports multi-tenant workspaces, streaming responses, and
                  rich message elements.
                </p>
                <div
                  class="grid grid-cols-1 sm:grid-cols-2 gap-2 w-full max-w-md"
                >
                  {#each suggestions as suggestion}
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

            {#each messages as msg (msg.id)}
              <ChatMessageItem {msg} {shikiTheme} />
            {/each}

            {#if status === "submitted"}
              <Message from="assistant" class="mb-2">
                <MessageContent variant="flat">
                  <Loader />
                </MessageContent>
              </Message>
            {/if}
          </ConversationContent>
          <ConversationScrollButton />
        </Conversation>

        <!-- Input -->
        <ChatInput bind:input {status} {placeholder} {onSubmit} />
      </div>
    </Sidebar.Inset>
  </Sidebar.Provider>
</div>
