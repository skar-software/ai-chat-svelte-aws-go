<script lang="ts">
	import ChatWidget from "$lib/components/ChatWidget.svelte";
</script>

<div class="h-screen w-full overflow-hidden">
	<ChatWidget
		apiBaseUrl="https://api.example.com/v1/chat"
		tenantId="demo-workspace-001"
		title="AI Assistant"
		placeholder="Ask me anything..."
	/>
</div>
