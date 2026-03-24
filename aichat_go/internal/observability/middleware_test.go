package observability

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

func TestRequestLogger_includesRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	app := fiber.New()
	app.Use(requestid.New())
	app.Use(RequestLogger(logger))
	app.Get("/x", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	_, _ = io.ReadAll(resp.Body)

	if !bytes.Contains(buf.Bytes(), []byte(`"request_id"`)) && !bytes.Contains(buf.Bytes(), []byte("request_id")) {
		t.Fatalf("log missing request_id: %s", buf.String())
	}
}
