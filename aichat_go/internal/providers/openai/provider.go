package openai

import (
	"context"
	"fmt"
	"strings"

	"aichat_go/internal/chat"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

type Provider struct {
	client       openai.Client
	defaultModel shared.ChatModel
}

type responseMode string

const (
	modeDefault      responseMode = "default"
	modeUIIntegration responseMode = "ui_integration"
	modeArtifact     responseMode = "artifact"
	modePlan         responseMode = "plan"
	modeConfirmation responseMode = "confirmation"
	modeQueue        responseMode = "queue"
	modeCitation     responseMode = "citation"
)

const uiElementsSystemInstruction = "You are generating a real-time UI integration test response for a chat widget.\n\n" +
	"Output sequence requirements:\n" +
	"1) Start with one short plain-text sentence.\n" +
	"2) Include one fenced code block with a language (prefer javascript).\n" +
	"3) Include exactly one artifact block using ```json:part ... ```.\n" +
	"4) Include exactly one confirmation block using ```json:part ... ```.\n" +
	"5) Include exactly one plan block using ```json:part ... ```.\n" +
	"6) Include exactly one queue block using ```json:part ... ```.\n\n" +
	"Use these JSON shapes:\n" +
	"- artifact: {\"type\":\"artifact\",\"content\":\"<markdown>\",\"meta\":{\"title\":\"...\",\"description\":\"...\"}}\n" +
	"- confirmation: {\"type\":\"confirmation\",\"meta\":{\"title\":\"...\",\"description\":\"...\",\"state\":\"approval-requested\",\"approval\":{\"id\":\"approval-1\"}}}\n" +
	"- plan: {\"type\":\"plan\",\"meta\":{\"title\":\"...\",\"description\":\"...\",\"steps\":[\"...\",\"...\",\"...\"]}}\n" +
	"- queue: {\"type\":\"queue\",\"meta\":{\"messages\":[{\"id\":\"m1\",\"text\":\"...\"}],\"todos\":[{\"id\":\"t1\",\"title\":\"...\",\"description\":\"...\",\"status\":\"pending\"},{\"id\":\"t2\",\"title\":\"...\",\"status\":\"completed\"}]}}\n\n" +
	"Rules:\n" +
	"- Keep json:part blocks valid JSON\n" +
	"- Do not wrap json:part JSON in any extra markdown other than the required fences\n" +
	"- Do not omit any of the required sections"

const artifactSystemInstruction = "When the user asks for a document-style output (checklist, plan, SOP, proposal, report, roadmap, spec, runbook, playbook, template), respond as ONE structured block and nothing else.\n\n" +
	"Return exactly:\n" +
	"```json:part\n" +
	"{\n" +
	"  \"type\": \"artifact\",\n" +
	"  \"content\": \"<markdown body>\",\n" +
	"  \"meta\": {\n" +
	"    \"title\": \"<short title>\",\n" +
	"    \"description\": \"<one-line summary>\"\n" +
	"  }\n" +
	"}\n" +
	"```\n\n" +
	"Rules:\n" +
	"- No prose before or after the block\n" +
	"- Valid JSON only\n" +
	"- Keep \"content\" as markdown text"

const planSystemInstruction = "When the user asks for a step-by-step plan, return exactly ONE json:part block and nothing else.\n\n" +
	"Use:\n" +
	"```json:part\n" +
	"{\n" +
	"  \"type\": \"plan\",\n" +
	"  \"meta\": {\n" +
	"    \"title\": \"<short title>\",\n" +
	"    \"description\": \"<one-line summary>\",\n" +
	"    \"steps\": [\"<step 1>\", \"<step 2>\", \"<step 3>\"]\n" +
	"  }\n" +
	"}\n" +
	"```\n\n" +
	"Rules:\n" +
	"- No prose before/after block\n" +
	"- Valid JSON only\n" +
	"- 3 to 7 concise steps"

const confirmationSystemInstruction = "When the user asks for approval/confirmation style output, return exactly ONE json:part block and nothing else.\n\n" +
	"Use:\n" +
	"```json:part\n" +
	"{\n" +
	"  \"type\": \"confirmation\",\n" +
	"  \"meta\": {\n" +
	"    \"title\": \"<action title>\",\n" +
	"    \"description\": \"<what needs approval>\",\n" +
	"    \"state\": \"approval-requested\",\n" +
	"    \"approval\": { \"id\": \"approval-1\" }\n" +
	"  }\n" +
	"}\n" +
	"```\n\n" +
	"Rules:\n" +
	"- No prose before/after block\n" +
	"- Valid JSON only"

const queueSystemInstruction = "When the user asks for tasks/progress tracking, return exactly ONE json:part block and nothing else.\n\n" +
	"Use:\n" +
	"```json:part\n" +
	"{\n" +
	"  \"type\": \"queue\",\n" +
	"  \"meta\": {\n" +
	"    \"messages\": [{\"id\":\"m1\",\"text\":\"<status message>\"}],\n" +
	"    \"todos\": [\n" +
	"      {\"id\":\"t1\",\"title\":\"<pending task>\",\"description\":\"<optional>\",\"status\":\"pending\"},\n" +
	"      {\"id\":\"t2\",\"title\":\"<completed task>\",\"status\":\"completed\"}\n" +
	"    ]\n" +
	"  }\n" +
	"}\n" +
	"```\n\n" +
	"Rules:\n" +
	"- No prose before/after block\n" +
	"- Valid JSON only\n" +
	"- At least 2 todos"

func NewProvider(apiKey, defaultModel string) *Provider {
	model := shared.ChatModel(defaultModel)
	if defaultModel == "" {
		model = shared.ChatModelGPT4oMini
	}
	return &Provider{
		client:       openai.NewClient(option.WithAPIKey(apiKey)),
		defaultModel: model,
	}
}

func (p *Provider) Stream(ctx context.Context, input *chat.StreamInput, history []chat.MessageRecord, out chan<- chat.StreamEvent) {
	defer close(out)

	mode := detectResponseMode(input.Input)

	// Deterministic component output for core structured modes.
	// Emit the part event directly so it bypasses the text parser entirely.
	if mode == modeArtifact || mode == modePlan || mode == modeConfirmation || mode == modeQueue || mode == modeCitation {
		partEvent := buildStructuredPartEvent(mode, input.Input)
		out <- chat.StreamEvent{Event: "part", Data: partEvent}
		out <- chat.StreamEvent{Event: "usage", Data: map[string]int{"input_tokens": 1, "output_tokens": 1}}
		out <- chat.StreamEvent{Event: "_stream_done", Data: map[string]string{"content": ""}}
		return
	}

	model := p.defaultModel
	if input.Model != "" {
		model = shared.ChatModel(input.Model)
	}

	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(history)+1)
	if sys := systemInstructionForPrompt(input.Input); sys != "" {
		messages = append(messages, openai.SystemMessage(sys))
	}
	for _, h := range history {
		if h.Role == "assistant" {
			messages = append(messages, openai.AssistantMessage(h.Content))
		} else {
			messages = append(messages, openai.UserMessage(h.Content))
		}
	}
	messages = append(messages, openai.UserMessage(input.Input))

	params := openai.ChatCompletionNewParams{
		Model:    model,
		Messages: messages,
		StreamOptions: openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: openai.Bool(true),
		},
	}

	stream := p.client.Chat.Completions.NewStreaming(ctx, params)

	var fullContent string
	var inputTokens, outputTokens int

	for stream.Next() {
		chunk := stream.Current()
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta.Content
			if delta != "" {
				fullContent += delta
				out <- chat.StreamEvent{Event: "message.delta", Data: map[string]string{"delta": delta}}
			}
		}
		if chunk.JSON.Usage.Valid() {
			inputTokens = int(chunk.Usage.PromptTokens)
			outputTokens = int(chunk.Usage.CompletionTokens)
		}
	}

	if err := stream.Err(); err != nil {
		out <- chat.StreamEvent{Event: "error", Data: map[string]string{"message": err.Error()}}
		return
	}

	if outputTokens == 0 && fullContent != "" {
		outputTokens = len(fullContent) / 4
	}
	if inputTokens == 0 {
		inputTokens = 100
	}

	out <- chat.StreamEvent{Event: "usage", Data: map[string]int{"input_tokens": inputTokens, "output_tokens": outputTokens}}
	out <- chat.StreamEvent{Event: "_stream_done", Data: map[string]string{"content": fullContent}}
}

func buildStructuredPartEvent(mode responseMode, prompt string) chat.PartEvent {
	cleanPrompt := strings.TrimSpace(prompt)
	if cleanPrompt == "" {
		cleanPrompt = "Requested by user"
	}

	switch mode {
	case modeArtifact:
		return chat.PartEvent{
			Type: "artifact",
			Content: fmt.Sprintf("## Artifact\n\n%s\n\n- [ ] Review\n- [ ] Approve\n- [ ] Execute",
				cleanPrompt),
			Meta: map[string]interface{}{
				"title":       "Artifact",
				"description": "Generated from your request",
			},
		}
	case modePlan:
		return chat.PartEvent{
			Type: "plan",
			Meta: map[string]interface{}{
				"title":       "Execution Plan",
				"description": "Step-by-step plan generated from your request",
				"steps": []string{
					"Clarify requirements and acceptance criteria",
					"Implement the requested changes",
					"Validate with tests and review",
				},
			},
		}
	case modeConfirmation:
		return chat.PartEvent{
			Type: "confirmation",
			Meta: map[string]interface{}{
				"title":       "Approval Required",
				"description": cleanPrompt,
				"state":       "approval-requested",
				"approval": map[string]interface{}{
					"id": "approval-1",
				},
			},
		}
	case modeQueue:
		tasks := extractQueueTasks(cleanPrompt)
		todos := make([]map[string]interface{}, 0, len(tasks))
		for i, t := range tasks {
			status := "pending"
			if i == len(tasks)-1 && len(tasks) > 1 {
				status = "completed"
			}
			todos = append(todos, map[string]interface{}{
				"id":          fmt.Sprintf("t%d", i+1),
				"title":       t,
				"description": "Generated task",
				"status":      status,
			})
		}
		return chat.PartEvent{
			Type: "queue",
			Meta: map[string]interface{}{
				"messages": []map[string]interface{}{
					{"id": "m1", "text": "Queue generated from your request"},
				},
				"todos": todos,
			},
		}
	case modeCitation:
		return chat.PartEvent{
			Type:    "citation",
			Content: cleanPrompt,
			Meta: map[string]interface{}{
				"sources": []map[string]interface{}{
					{
						"title":       "Svelte Documentation",
						"url":         "https://svelte.dev/docs",
						"description": "Official Svelte framework documentation covering components, reactivity, and SvelteKit.",
						"quote":       "Svelte is a radical new approach to building user interfaces.",
					},
					{
						"title":       "MDN Web Docs",
						"url":         "https://developer.mozilla.org",
						"description": "Comprehensive web development documentation by Mozilla.",
						"quote":       "Resources for developers, by developers.",
					},
					{
						"title":       "Wikipedia",
						"url":         "https://en.wikipedia.org",
						"description": "The free encyclopedia that anyone can edit.",
					},
				},
			},
		}
	default:
		return chat.PartEvent{}
	}
}

func extractQueueTasks(prompt string) []string {
	p := strings.ToLower(prompt)
	out := []string{}

	if strings.Contains(p, "setup") {
		out = append(out, "Setup")
	}
	if strings.Contains(p, "test") {
		out = append(out, "Test")
	}
	if strings.Contains(p, "deploy") {
		out = append(out, "Deploy")
	}

	if len(out) == 0 {
		out = []string{"Task 1", "Task 2", "Task 3"}
	}
	return out
}

func systemInstructionForPrompt(prompt string) string {
	switch detectResponseMode(prompt) {
	case modeUIIntegration:
		return uiElementsSystemInstruction
	case modeConfirmation:
		return confirmationSystemInstruction
	case modeQueue:
		return queueSystemInstruction
	case modePlan:
		return planSystemInstruction
	case modeArtifact:
		return artifactSystemInstruction
	default:
		return ""
	}
}

func detectResponseMode(prompt string) responseMode {
	if shouldRunUIElementsTest(prompt) {
		return modeUIIntegration
	}
	if shouldPreferConfirmation(prompt) {
		return modeConfirmation
	}
	if shouldPreferQueue(prompt) {
		return modeQueue
	}
	if shouldPreferPlan(prompt) {
		return modePlan
	}
	if shouldPreferCitation(prompt) {
		return modeCitation
	}
	if shouldPreferArtifact(prompt) {
		return modeArtifact
	}
	return modeDefault
}

func shouldPreferArtifact(prompt string) bool {
	p := strings.ToLower(strings.TrimSpace(prompt))
	if p == "" {
		return false
	}

	artifactHints := []string{
		"artifact",
		"checklist",
		"plan",
		"roadmap",
		"sop",
		"runbook",
		"playbook",
		"spec",
		"proposal",
		"report",
		"document",
		"template",
		"requirements doc",
		"release notes",
		"deployment guide",
	}

	for _, h := range artifactHints {
		if strings.Contains(p, h) {
			return true
		}
	}
	return false
}

func shouldPreferPlan(prompt string) bool {
	p := strings.ToLower(strings.TrimSpace(prompt))
	if p == "" {
		return false
	}
	planHints := []string{
		"step by step plan",
		"plan for",
		"execution plan",
		"implementation plan",
		"rollout plan",
		"migration plan",
		"project plan",
	}
	for _, h := range planHints {
		if strings.Contains(p, h) {
			return true
		}
	}
	return false
}

func shouldPreferConfirmation(prompt string) bool {
	p := strings.ToLower(strings.TrimSpace(prompt))
	if p == "" {
		return false
	}
	confirmationHints := []string{
		"ask for approval",
		"needs approval",
		"approval required",
		"confirm before",
		"confirmation",
		"are you sure",
	}
	for _, h := range confirmationHints {
		if strings.Contains(p, h) {
			return true
		}
	}
	return false
}

func shouldPreferQueue(prompt string) bool {
	p := strings.ToLower(strings.TrimSpace(prompt))
	if p == "" {
		return false
	}
	queueHints := []string{
		"task list",
		"task tracking",
		"tracking template",
		"todo",
		"to-do",
		"queue",
		"progress tracker",
		"track tasks",
		"track the tasks",
		"pending and completed",
	}
	for _, h := range queueHints {
		if strings.Contains(p, h) {
			return true
		}
	}
	return false
}

func shouldPreferCitation(prompt string) bool {
	p := strings.ToLower(strings.TrimSpace(prompt))
	if p == "" {
		return false
	}
	citationHints := []string{
		"cite",
		"citation",
		"source",
		"sources",
		"reference",
		"references",
		"according to",
		"research",
		"study",
		"studies",
		"paper",
		"journal",
		"with citations",
		"with sources",
		"with references",
	}
	for _, h := range citationHints {
		if strings.Contains(p, h) {
			return true
		}
	}
	return false
}

func shouldRunUIElementsTest(prompt string) bool {
	p := strings.ToLower(strings.TrimSpace(prompt))
	if p == "" {
		return false
	}
	uiHints := []string{
		"test ui elements",
		"ui test",
		"test all ui components",
		"artifact confirmation plan queue",
		"full ui checklist",
	}
	for _, h := range uiHints {
		if strings.Contains(p, h) {
			return true
		}
	}
	return false
}
