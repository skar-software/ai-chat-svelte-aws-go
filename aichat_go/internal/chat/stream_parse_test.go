package chat

import (
	"strings"
	"testing"
)

func TestStreamParser_Feed_plainText(t *testing.T) {
	p := &StreamParser{state: "text"}
	forward, parts := p.Feed("hello")
	if len(parts) != 0 {
		t.Fatalf("expected no parts, got %v", parts)
	}
	if len(forward) != 1 || forward[0] != "hello" {
		t.Fatalf("forward = %v, want [hello]", forward)
	}
}

func TestStreamParser_Feed_codeFence_emitsPlaceholderThenContent(t *testing.T) {
	p := &StreamParser{state: "text"}
	var allParts []PartEvent

	delta := "```js\nconsole.log(1)\n```"
	forward, parts := p.Feed(delta)
	allParts = append(allParts, parts...)
	_ = forward

	if len(allParts) < 2 {
		t.Fatalf("expected at least placeholder + final code part, got %d parts", len(allParts))
	}
	if allParts[0].Type != "code" || allParts[0].Content != "" {
		t.Fatalf("first part should be empty placeholder, got %+v", allParts[0])
	}
	if lang, _ := allParts[0].Meta["lang"].(string); lang != "js" {
		t.Fatalf("lang = %q, want js", lang)
	}
	final := allParts[len(allParts)-1]
	if final.Type != "code" || final.Content != "console.log(1)\n" {
		t.Fatalf("final code part = %+v", final)
	}
	if partial, _ := final.Meta["partial"].(bool); partial {
		t.Fatal("final code part should not be partial")
	}
}

func TestStreamParser_Flush_unterminatedCode(t *testing.T) {
	p := &StreamParser{state: "text"}
	_, _ = p.Feed("```go\nfmt.Println")
	// Still inside code fence
	if p.state != "code" {
		t.Fatalf("state = %q, want code", p.state)
	}
	out := p.Flush()
	if len(out) != 1 {
		t.Fatalf("Flush len = %d", len(out))
	}
	if !strings.Contains(out[0], "```go") || !strings.Contains(out[0], "fmt.Println") {
		t.Fatalf("unexpected flush: %q", out[0])
	}
	if p.state != "text" {
		t.Fatalf("state after flush = %q", p.state)
	}
}

func TestStreamParser_Flush_unterminatedJSONPart(t *testing.T) {
	p := &StreamParser{state: "text"}
	_, _ = p.Feed("```json:part\n{\"type\":\"artifact\"")
	if p.state != "json_part" {
		t.Fatalf("state = %q", p.state)
	}
	out := p.Flush()
	if len(out) != 1 || !strings.HasPrefix(out[0], "```json:part") {
		t.Fatalf("unexpected: %q", out[0])
	}
}

func TestStreamParser_Feed_jsonPartBlock(t *testing.T) {
	p := &StreamParser{state: "text"}
	raw := "```json:part\n{\"type\":\"plan\",\"meta\":{\"title\":\"T\"}}\n```"
	_, parts := p.Feed(raw)
	if len(parts) != 1 {
		t.Fatalf("parts = %v", parts)
	}
	if parts[0].Type != "plan" {
		t.Fatalf("type = %q", parts[0].Type)
	}
}

func TestStreamParser_Feed_emptyDelta(t *testing.T) {
	p := &StreamParser{state: "text"}
	forward, parts := p.Feed("")
	if forward != nil || parts != nil {
		t.Fatalf("expected nil, got forward=%v parts=%v", forward, parts)
	}
}

func TestStreamParser_Flush_fenceStartIncomplete(t *testing.T) {
	p := &StreamParser{state: "fence_start"}
	p.buf.WriteString("```")
	out := p.Flush()
	if len(out) != 1 || out[0] != "```" {
		t.Fatalf("got %v", out)
	}
}
