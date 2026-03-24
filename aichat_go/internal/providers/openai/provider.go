package openai

import (
	"context"
	"regexp"
	"strconv"
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

type structuredIntent string

const (
	intentNone         structuredIntent = ""
	intentPlan         structuredIntent = "plan"
	intentArtifact     structuredIntent = "artifact"
	intentConfirmation structuredIntent = "confirmation"
	intentQueue        structuredIntent = "queue"
	intentCitation     structuredIntent = "citation"
)

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

	intent := inferIntentFromPrompt(input.Input)

	model := p.defaultModel
	if input.Model != "" {
		model = shared.ChatModel(input.Model)
	}

	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(history)+1)
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
				// For structured intents, buffer and emit a typed part at the end.
				if intent == intentNone {
					out <- chat.StreamEvent{Event: "message.delta", Data: map[string]string{"delta": delta}}
				}
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

	if intent != intentNone {
		if part, ok := buildStructuredPart(intent, input.Input, fullContent); ok {
			out <- chat.StreamEvent{Event: "part", Data: part}
			fullContent = ""
		} else if fullContent != "" {
			// Fallback to plain text if we couldn't construct a valid structured part.
			out <- chat.StreamEvent{Event: "message.delta", Data: map[string]string{"delta": fullContent}}
		}
	}

	out <- chat.StreamEvent{Event: "usage", Data: map[string]int{"input_tokens": inputTokens, "output_tokens": outputTokens}}
	out <- chat.StreamEvent{Event: "_stream_done", Data: map[string]string{"content": fullContent}}
}

func inferIntentFromPrompt(prompt string) structuredIntent {
	p := strings.ToLower(strings.TrimSpace(prompt))
	if p == "" {
		return intentNone
	}

	if containsAny(p, "citation", "citations", "source", "sources", "reference", "references", "resource", "resources") {
		return intentCitation
	}
	if containsAny(p, "task list", "todo", "to-do", "queue", "track", "tracking", "pending", "completed") {
		return intentQueue
	}
	if containsAny(p, "approval", "approve", "confirm", "are you sure") {
		return intentConfirmation
	}
	if containsAny(p, "runbook", "checklist", "sop", "playbook", "spec", "proposal", "report", "template", "artifact") {
		return intentArtifact
	}
	if containsAny(p, "plan", "roadmap", "step-by-step", "step by step", "migration plan", "rollout plan") {
		return intentPlan
	}

	return intentNone
}

func containsAny(s string, hints ...string) bool {
	for _, h := range hints {
		if strings.Contains(s, h) {
			return true
		}
	}
	return false
}

func buildStructuredPart(intent structuredIntent, prompt, content string) (chat.PartEvent, bool) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return chat.PartEvent{}, false
	}

	switch intent {
	case intentPlan:
		steps := extractListItems(trimmed, 8)
		if len(steps) < 3 {
			return chat.PartEvent{}, false
		}
		return chat.PartEvent{
			Type: "plan",
			Meta: map[string]interface{}{
				"title":       "Execution Plan",
				"description": "Auto-selected by prompt intent",
				"steps":       steps,
			},
		}, true
	case intentQueue:
		items := extractListItems(trimmed, 12)
		if len(items) < 2 {
			return chat.PartEvent{}, false
		}
		todos := make([]map[string]interface{}, 0, len(items))
		for i, it := range items {
			todos = append(todos, map[string]interface{}{
				"id":          "t" + strconv.Itoa(i+1),
				"title":       it,
				"description": "",
				"status":      "pending",
			})
		}
		return chat.PartEvent{
			Type: "queue",
			Meta: map[string]interface{}{
				"messages": []map[string]interface{}{
					{"id": "m1", "text": "Task list generated from your request"},
				},
				"todos": todos,
			},
		}, true
	case intentConfirmation:
		desc := trimmed
		if len(desc) > 240 {
			desc = desc[:240]
		}
		return chat.PartEvent{
			Type: "confirmation",
			Meta: map[string]interface{}{
				"title":       "Action Required",
				"description": desc,
				"state":       "approval-requested",
				"approval":    map[string]interface{}{"id": "approval-1"},
			},
		}, true
	case intentArtifact:
		return chat.PartEvent{
			Type:    "artifact",
			Content: trimmed,
			Meta: map[string]interface{}{
				"title":       inferArtifactTitle(trimmed),
				"description": "Auto-selected by prompt intent",
			},
		}, true
	case intentCitation:
		sources := extractSources(trimmed, 8)
		if len(sources) == 0 {
			return chat.PartEvent{}, false
		}
		return chat.PartEvent{
			Type:    "citation",
			Content: trimmed,
			Meta: map[string]interface{}{
				"sources": sources,
			},
		}, true
	default:
		return chat.PartEvent{}, false
	}
}

func extractListItems(text string, max int) []string {
	re := regexp.MustCompile(`(?m)^\s*(?:\d+[\).\s-]+|[-*]\s+)(.+)$`)
	matches := re.FindAllStringSubmatch(text, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		item := strings.TrimSpace(m[1])
		if item == "" {
			continue
		}
		out = append(out, item)
		if len(out) >= max {
			break
		}
	}
	return out
}

func inferArtifactTitle(text string) string {
	re := regexp.MustCompile(`(?m)^#{1,3}\s+(.+)$`)
	if m := re.FindStringSubmatch(text); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	line := strings.TrimSpace(strings.SplitN(text, "\n", 2)[0])
	if line == "" {
		return "Artifact"
	}
	if len(line) > 80 {
		return line[:80]
	}
	return line
}

func extractSources(text string, max int) []map[string]string {
	seen := map[string]bool{}
	out := make([]map[string]string, 0, max)

	mdLinkRe := regexp.MustCompile(`\[([^\]]+)\]\((https?://[^\s)]+)\)`)
	for _, m := range mdLinkRe.FindAllStringSubmatch(text, -1) {
		url := strings.TrimSpace(m[2])
		if url == "" || seen[url] {
			continue
		}
		seen[url] = true
		title := strings.TrimSpace(m[1])
		if title == "" {
			title = "Source"
		}
		out = append(out, map[string]string{"title": title, "url": url})
		if len(out) >= max {
			return out
		}
	}

	urlRe := regexp.MustCompile(`https?://[^\s)]+`)
	for _, u := range urlRe.FindAllString(text, -1) {
		url := strings.TrimSpace(u)
		if url == "" || seen[url] {
			continue
		}
		seen[url] = true
		out = append(out, map[string]string{"title": "Source", "url": url})
		if len(out) >= max {
			return out
		}
	}
	return out
}
