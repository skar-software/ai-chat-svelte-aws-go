package chat

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"aichat_go/internal/store"
)

type mockProvider struct {
	fn func(ctx context.Context, input *StreamInput, history []MessageRecord, out chan<- StreamEvent)
}

func (m *mockProvider) Stream(ctx context.Context, input *StreamInput, history []MessageRecord, out chan<- StreamEvent) {
	if m.fn != nil {
		m.fn(ctx, input, history, out)
		return
	}
	close(out)
}

func TestService_Stream_simpleText(t *testing.T) {
	st := store.NewInMemoryStore()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mp := &mockProvider{
		fn: func(_ context.Context, _ *StreamInput, _ []MessageRecord, out chan<- StreamEvent) {
			defer close(out)
			out <- StreamEvent{Event: "message.delta", Data: map[string]string{"delta": "Hello"}}
			out <- StreamEvent{Event: "usage", Data: map[string]int{"input_tokens": 10, "output_tokens": 5}}
			out <- StreamEvent{Event: "_stream_done", Data: map[string]string{"content": "Hello"}}
		},
	}
	svc := NewChatService(st, mp, log)

	var events []StreamEvent
	err := svc.Stream(context.Background(), &StreamInput{Input: "user says hi"}, func(ev StreamEvent) {
		events = append(events, ev)
	})
	if err != nil {
		t.Fatal(err)
	}

	var completed string
	for _, ev := range events {
		if ev.Event == "message.completed" {
			if m, ok := ev.Data.(map[string]string); ok {
				completed = m["conversation_id"]
			}
		}
	}
	if completed == "" {
		t.Fatalf("missing message.completed in %d events", len(events))
	}

	msgs, err := st.GetMessages(completed)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 2 {
		t.Fatalf("messages len = %d, want 2", len(msgs))
	}
	if msgs[0].Role != "user" || msgs[0].Content != "user says hi" {
		t.Fatalf("user msg: %+v", msgs[0])
	}
	if msgs[1].Role != "assistant" || msgs[1].Content != "Hello" {
		t.Fatalf("assistant msg: %+v", msgs[1])
	}
	if msgs[1].InputTokens != 10 || msgs[1].OutputTokens != 5 {
		t.Fatalf("tokens: %+v", msgs[1])
	}
}

func TestService_Stream_providerError(t *testing.T) {
	st := store.NewInMemoryStore()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mp := &mockProvider{
		fn: func(_ context.Context, _ *StreamInput, _ []MessageRecord, out chan<- StreamEvent) {
			defer close(out)
			out <- StreamEvent{Event: "error", Data: map[string]string{"message": "boom"}}
		},
	}
	svc := NewChatService(st, mp, log)

	var sawError bool
	_ = svc.Stream(context.Background(), &StreamInput{Input: "x"}, func(ev StreamEvent) {
		if ev.Event == "error" {
			sawError = true
		}
	})
	if !sawError {
		t.Fatal("expected error event")
	}
}

func TestService_Stream_directPartPassthrough(t *testing.T) {
	st := store.NewInMemoryStore()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mp := &mockProvider{
		fn: func(_ context.Context, _ *StreamInput, _ []MessageRecord, out chan<- StreamEvent) {
			defer close(out)
			out <- StreamEvent{Event: "part", Data: PartEvent{Type: "artifact", Content: "x"}}
			out <- StreamEvent{Event: "usage", Data: map[string]int{"input_tokens": 1, "output_tokens": 1}}
			out <- StreamEvent{Event: "_stream_done", Data: map[string]string{"content": ""}}
		},
	}
	svc := NewChatService(st, mp, log)

	var parts int
	_ = svc.Stream(context.Background(), &StreamInput{Input: "ask artifact"}, func(ev StreamEvent) {
		if ev.Event == "part" {
			parts++
		}
	})
	if parts < 1 {
		t.Fatalf("expected part events, parts=%d", parts)
	}
}
