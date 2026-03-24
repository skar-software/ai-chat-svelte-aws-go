package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"aichat_go/internal/chat"
	"aichat_go/internal/config"
	"aichat_go/internal/store"
)

func TestHealth(t *testing.T) {
	app := fiber.New()
	app.Get("/health", Health)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatal(err)
	}
	if m["ok"] != true {
		t.Fatalf("body %s", body)
	}
}

func TestListConversations(t *testing.T) {
	st := store.NewInMemoryStore()
	_, err := st.CreateConversation("tenant-x", "w", "u")
	if err != nil {
		t.Fatal(err)
	}

	app := fiber.New()
	app.Get("/c", ListConversations(st))
	req := httptest.NewRequest(http.MethodGet, "/c?tenant_id=tenant-x", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	var list []map[string]any
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("len=%d", len(list))
	}
}

func TestGetConversationMessages_unknownConversation(t *testing.T) {
	st := store.NewInMemoryStore()
	app := fiber.New()
	app.Get("/c/:id/messages", GetConversationMessages(st))
	req := httptest.NewRequest(http.MethodGet, "/c/c_unknown/messages", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var msgs []any
	if err := json.Unmarshal(body, &msgs); err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 0 {
		t.Fatalf("expected empty messages for unknown id, got %v", msgs)
	}
}

func TestGetConversationMessages_ok(t *testing.T) {
	st := store.NewInMemoryStore()
	c, _ := st.CreateConversation("t", "w", "u")
	_, _ = st.AppendMessage(c.ID, "user", "hi", "openai", "m", 0, 0)

	app := fiber.New()
	app.Get("/c/:id/messages", GetConversationMessages(st))
	req := httptest.NewRequest(http.MethodGet, "/c/"+c.ID+"/messages", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	var msgs []map[string]any
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &msgs); err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 || msgs[0]["content"] != "hi" {
		t.Fatalf("%v", msgs)
	}
}

type streamMock struct{}

func (streamMock) Stream(_ context.Context, _ *chat.StreamInput, _ []chat.MessageRecord, out chan<- chat.StreamEvent) {
	defer close(out)
	out <- chat.StreamEvent{Event: "message.delta", Data: map[string]string{"delta": "OK"}}
	out <- chat.StreamEvent{Event: "usage", Data: map[string]int{"input_tokens": 1, "output_tokens": 1}}
	out <- chat.StreamEvent{Event: "_stream_done", Data: map[string]string{"content": "OK"}}
}

func TestStreamChat_validation(t *testing.T) {
	st := store.NewInMemoryStore()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := chat.NewChatService(st, streamMock{}, log)
	cfg := &config.Config{}

	app := fiber.New()
	app.Post("/stream", StreamChat(cfg, st, svc))

	req := httptest.NewRequest(http.MethodPost, "/stream", bytes.NewBufferString(`not-json`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", resp.StatusCode)
	}

	body := bytes.NewBufferString(`{"input":""}`)
	req2 := httptest.NewRequest(http.MethodPost, "/stream", body)
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := app.Test(req2)
	if err != nil {
		t.Fatal(err)
	}
	if resp2.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400 empty input, got %d", resp2.StatusCode)
	}
}

func TestStreamChat_SSE(t *testing.T) {
	st := store.NewInMemoryStore()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := chat.NewChatService(st, streamMock{}, log)
	cfg := &config.Config{}

	app := fiber.New()
	app.Post("/stream", StreamChat(cfg, st, svc))

	payload := `{"input":"hello world"}`
	req := httptest.NewRequest(http.MethodPost, "/stream", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
		t.Fatalf("Content-Type = %q", ct)
	}
	raw, _ := io.ReadAll(resp.Body)
	if !bytes.Contains(raw, []byte("event: message.completed")) {
		t.Fatalf("missing completion in body: %s", raw)
	}
}

func TestTenantFromHeader(t *testing.T) {
	app := fiber.New()
	app.Use(TenantFromHeader())
	app.Get("/t", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"tenant":    getString(c, "tenant_id", ""),
			"workspace": getString(c, "workspace_id", ""),
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	req.Header.Set("X-Tenant-ID", "acme")
	req.Header.Set("X-Workspace-ID", "ws1")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	var m map[string]string
	_ = json.Unmarshal(body, &m)
	if m["tenant"] != "acme" || m["workspace"] != "ws1" {
		t.Fatalf("%s", body)
	}
}

func TestErrorHandler_fiberError(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler(log)})
	app.Get("/e", func(c fiber.Ctx) error {
		return fiber.NewError(http.StatusTeapot, "short and stout")
	})
	req := httptest.NewRequest(http.MethodGet, "/e", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusTeapot {
		t.Fatalf("status %d", resp.StatusCode)
	}
}
