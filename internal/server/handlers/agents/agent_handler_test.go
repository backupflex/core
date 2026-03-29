package agents_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/backupflex/core/internal/agents"
	handler "github.com/backupflex/core/internal/server/handlers/agents"
	"github.com/dgraph-io/badger/v4"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func TestAgentHeartbeatUsesBearerTokenWithoutAgentID(t *testing.T) {
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true).WithLogger(nil))
	if err != nil {
		t.Fatalf("open badger: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	svc := agents.New(agents.NewRepository(db), agents.NewStatusTracker(), zap.NewNop())
	handler := handler.NewAgentHandler(svc)

	app := fiber.New()
	handler.Register(app.Group("/api/v1"))

	payload, _ := json.Marshal(map[string]any{
		"id":      "agent-auth-1",
		"role":    "backup",
		"address": "10.0.0.1:9000",
	})

	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/agents/register", bytes.NewReader(payload))
	registerReq.Header.Set("Content-Type", "application/json")
	registerResp, err := app.Test(registerReq)
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	if registerResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", registerResp.StatusCode)
	}

	var body struct {
		AccessToken string `json:"access_token"`
	}
	if jsonErr := json.NewDecoder(registerResp.Body).Decode(&body); jsonErr != nil {
		t.Fatalf("decode register response: %v", jsonErr)
	}
	if body.AccessToken == "" {
		t.Fatal("expected access_token in register response")
	}

	heartbeatReq := httptest.NewRequest(http.MethodPost, "/api/v1/agents/heartbeat", nil)
	heartbeatReq.Header.Set("Authorization", "Bearer "+body.AccessToken)
	heartbeatResp, err := app.Test(heartbeatReq)
	if err != nil {
		t.Fatalf("heartbeat request failed: %v", err)
	}
	if heartbeatResp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", heartbeatResp.StatusCode)
	}
}
