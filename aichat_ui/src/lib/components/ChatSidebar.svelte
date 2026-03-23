<script lang="ts">
  import * as Sidebar from "$lib/components/ui/sidebar/index.js";
  import { Button } from "$lib/components/ui/button/index.js";
  import { Plus, MessageSquare } from "@lucide/svelte";

  interface SidebarConversation {
    id: string;
    title: string;
  }

  interface Props {
    conversations: SidebarConversation[];
    activeConversationId?: string;
    onSelectConversation?: (id: string) => void;
    onNewChat?: () => void;
  }

  let {
    conversations,
    activeConversationId,
    onSelectConversation,
    onNewChat,
  }: Props = $props();
</script>

<Sidebar.Root>
  <Sidebar.Header class="p-4">
    <div class="flex items-center justify-between mb-3">
      <div class="flex items-center gap-2">
        <div
          class="size-7 bg-primary rounded-lg flex items-center justify-center"
        >
          <span class="text-primary-foreground font-bold text-xs">AI</span>
        </div>
        <span class="font-semibold text-sm">Chatbot</span>
      </div>
    </div>
    <Button
      variant="outline"
      class="w-full justify-start gap-2 text-xs h-9 hover:border-primary/50 transition-colors"
      onclick={onNewChat}
    >
      <Plus class="size-3.5" />
      <span>New chat</span>
    </Button>
  </Sidebar.Header>

  <Sidebar.Content>
    <Sidebar.Group>
      <Sidebar.GroupLabel class="px-3 text-[10px] uppercase tracking-widest"
        >Current</Sidebar.GroupLabel
      >
      <Sidebar.GroupContent>
        <Sidebar.Menu>
          {#each conversations as conv (conv.id)}
            <Sidebar.MenuItem>
              <Sidebar.MenuButton
                isActive={conv.id === activeConversationId}
                onclick={() => onSelectConversation?.(conv.id)}
              >
                <MessageSquare class="size-4" />
                <span class="truncate text-xs">{conv.title}</span>
              </Sidebar.MenuButton>
            </Sidebar.MenuItem>
          {/each}
        </Sidebar.Menu>
      </Sidebar.GroupContent>
    </Sidebar.Group>
  </Sidebar.Content>

  <Sidebar.Footer class="p-3">
    <!-- Footer intentionally minimal -->
  </Sidebar.Footer>
</Sidebar.Root>
