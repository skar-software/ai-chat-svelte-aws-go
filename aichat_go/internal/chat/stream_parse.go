package chat

import (
	"encoding/json"
	"strings"
)

// PartEvent is a structured block to send as SSE "part" event.
type PartEvent struct {
	Type    string                 `json:"type"`
	Content string                 `json:"content,omitempty"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// StreamParser parses streamed markdown and extracts code blocks and optional json:part blocks.
// It forwards text as content deltas and emits PartEvent when a block completes.
type StreamParser struct {
	state       string // "text" | "code" | "json_part"
	buf         strings.Builder
	codeLang    string
	codeBuf     strings.Builder
	backtickCnt int
	jsonPartBuf strings.Builder

	// currentCodePartID is stable for the lifetime of a single fenced code block.
	// We use it so the frontend can update an in-progress code part instead of
	// waiting for the closing fence.
	currentCodePartID int
	nextPartID        int
}

// Feed processes a delta and returns (content deltas to forward, parts completed in this delta).
func (p *StreamParser) Feed(delta string) (forward []string, parts []PartEvent) {
	if delta == "" {
		return nil, nil
	}
	codeUpdated := false
	for _, r := range delta {
		var part *PartEvent
		switch p.state {
		case "text":
			p.buf.WriteRune(r)
			if r == '`' {
				p.backtickCnt++
				if p.backtickCnt == 3 {
					s := p.buf.String()
					p.buf.Reset()
					p.backtickCnt = 0
					before := s[:len(s)-3]
					if before != "" {
						forward = append(forward, before)
					}
					p.state = "fence_start"
					p.buf.WriteString("```")
				}
			} else {
				p.backtickCnt = 0
			}
		case "fence_start":
			p.buf.WriteRune(r)
			if r == '\n' {
				fence := strings.TrimSpace(p.buf.String())
				p.buf.Reset()
				if strings.HasPrefix(fence, "```json:part") {
					p.state = "json_part"
					p.jsonPartBuf.Reset()
				} else {
					langLine := strings.TrimPrefix(fence, "```")
					langLine = strings.TrimSpace(langLine)
					if idx := strings.IndexAny(langLine, " \t"); idx > 0 {
						langLine = langLine[:idx]
					}
					p.codeLang = langLine
					p.codeBuf.Reset()
					p.currentCodePartID = p.nextPartID
					p.nextPartID++
					p.state = "code"

					// Emit an immediate placeholder part so the UI shows the code block
					// as soon as the opening fence is received.
					parts = append(parts, PartEvent{
						Type:    "code",
						Content: "",
						Meta:    map[string]interface{}{"lang": p.codeLang, "id": p.currentCodePartID, "partial": true},
					})
				}
			} else if r == '`' {
				p.backtickCnt++
				if p.backtickCnt >= 3 {
					p.buf.Reset()
					p.backtickCnt = 0
					p.codeLang = ""
					p.codeBuf.Reset()
					p.currentCodePartID = p.nextPartID
					p.nextPartID++
					p.state = "code"

					// Emit an immediate placeholder part so the UI shows the code block
					// as soon as the opening fence is received.
					parts = append(parts, PartEvent{
						Type:    "code",
						Content: "",
						Meta:    map[string]interface{}{"lang": p.codeLang, "id": p.currentCodePartID, "partial": true},
					})
				}
			} else {
				p.backtickCnt = 0
			}
		case "code":
			if r == '`' {
				p.backtickCnt++
				if p.backtickCnt == 3 {
					lang := p.codeLang
					code := p.codeBuf.String()
					p.codeBuf.Reset()
					p.backtickCnt = 0
					p.state = "text"
					part = &PartEvent{
						Type:    "code",
						Content: code,
						Meta:    map[string]interface{}{"lang": lang, "id": p.currentCodePartID, "partial": false},
					}
					p.codeLang = ""
					p.currentCodePartID = 0
					codeUpdated = false
				}
			} else {
				for p.backtickCnt > 0 {
					p.codeBuf.WriteByte('`')
					p.backtickCnt--
				}
				p.codeBuf.WriteRune(r)
				codeUpdated = true
			}
		case "json_part":
			if r == '`' {
				p.backtickCnt++
				if p.backtickCnt == 3 {
					raw := strings.TrimSpace(p.jsonPartBuf.String())
					p.jsonPartBuf.Reset()
					p.backtickCnt = 0
					p.state = "text"
					var pe PartEvent
					if err := json.Unmarshal([]byte(raw), &pe); err == nil && pe.Type != "" {
						part = &pe
					}
				}
			} else {
				for p.backtickCnt > 0 {
					p.jsonPartBuf.WriteByte('`')
					p.backtickCnt--
				}
				p.jsonPartBuf.WriteRune(r)
			}
		}
		if part != nil {
			parts = append(parts, *part)
		}
	}

	// If we are inside a fenced code block, emit an in-progress code part so
	// the UI doesn't "pause" until the closing ``` arrives.
	if p.state == "code" && codeUpdated && p.codeBuf.Len() > 0 {
		// Send the full buffered code so the frontend can just replace.
		parts = append(parts, PartEvent{
			Type:    "code",
			Content: p.codeBuf.String(),
			Meta:    map[string]interface{}{"lang": p.codeLang, "id": p.currentCodePartID, "partial": true},
		})
	}
	if p.state == "text" && p.buf.Len() > 0 {
		s := p.buf.String()
		p.buf.Reset()
		forward = append(forward, s)
		p.backtickCnt = 0
	}
	return forward, parts
}

// Flush sends any remaining buffered text (e.g. at end of stream).
func (p *StreamParser) Flush() (forward []string) {
	switch p.state {
	case "text":
		if p.buf.Len() > 0 {
			forward = append(forward, p.buf.String())
			p.buf.Reset()
		}
	case "fence_start":
		// Incomplete opening fence: flush raw text instead of dropping.
		if p.buf.Len() > 0 {
			forward = append(forward, p.buf.String())
			p.buf.Reset()
		}
	case "code":
		// Unterminated code fence: degrade gracefully to raw markdown text.
		var b strings.Builder
		b.WriteString("```")
		if p.codeLang != "" {
			b.WriteString(p.codeLang)
		}
		b.WriteString("\n")
		b.WriteString(p.codeBuf.String())
		forward = append(forward, b.String())
		p.codeBuf.Reset()
	case "json_part":
		// Unterminated json:part fence: flush as plain text.
		var b strings.Builder
		b.WriteString("```json:part\n")
		b.WriteString(p.jsonPartBuf.String())
		forward = append(forward, b.String())
		p.jsonPartBuf.Reset()
	}

	p.state = "text"
	p.codeLang = ""
	p.currentCodePartID = 0
	p.backtickCnt = 0
	return forward
}
