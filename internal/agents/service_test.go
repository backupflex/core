package agents_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/backupflex/core/internal/agents"
	"github.com/dgraph-io/badger/v4"
	"go.uber.org/zap"
)

func TestServiceRegisterAndHeartbeat(t *testing.T) {
	svc := agents.New(newTestRepository(t), agents.NewStatusTracker(), zap.NewNop())
	now := time.Date(2026, time.March, 25, 10, 0, 0, 0, time.UTC)

	a, err := svc.Register(
		context.Background(),
		agents.AgentInput{ID: "a-1", Role: agents.RoleBackup, Address: "10.0.0.1:9000"},
	)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if a.Status != agents.StatusOnline {
		t.Fatalf("expected status online, got %s", a.Status)
	}

	heartbeatTime := now.Add(time.Minute)
	if heartErr := svc.Heartbeat(context.Background(), "a-1", heartbeatTime); heartErr != nil {
		t.Fatalf("heartbeat failed: %v", heartErr)
	}

	stored, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if len(stored) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(stored))
	}

	if !stored[0].LastSeenAt.Equal(heartbeatTime) {
		t.Fatalf("unexpected last_seen_at: %s", stored[0].LastSeenAt)
	}
}

func TestServiceRegisterDuplicate(t *testing.T) {
	svc := agents.New(newTestRepository(t), agents.NewStatusTracker(), zap.NewNop())

	_, err := svc.Register(context.Background(), agents.AgentInput{ID: "dup", Role: agents.RoleBackup})
	if err != nil {
		t.Fatalf("initial register failed: %v", err)
	}

	_, err = svc.Register(context.Background(), agents.AgentInput{ID: "dup", Role: agents.RoleBackup})
	if err == nil {
		t.Fatal("expected duplicate registration error")
	}

	if !errors.Is(err, agents.ErrAgentAlreadyExists) {
		t.Fatalf("expected ErrAgentAlreadyExists, got %v", err)
	}
}

func newTestRepository(t *testing.T) *agents.Repository {
	t.Helper()

	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true).WithLogger(nil))
	if err != nil {
		t.Fatalf("open badger: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return agents.NewRepository(db)
}

func TestServiceIssueAndAuthenticateToken(t *testing.T) {
	svc := agents.New(newTestRepository(t), agents.NewStatusTracker(), zap.NewNop())

	registered, err := svc.Register(
		context.Background(),
		agents.AgentInput{ID: "agent-token", Role: agents.RoleBackup},
	)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if registered.Token == "" {
		t.Fatal("expected non-empty token from Register")
	}

	resolvedID, err := svc.AuthenticateToken(context.Background(), registered.Token)
	if err != nil {
		t.Fatalf("authenticate token failed: %v", err)
	}

	if resolvedID != registered.ID {
		t.Fatalf("expected %s, got %s", registered.ID, resolvedID)
	}

	err = svc.Deregister(context.Background(), registered.ID)
	if err != nil {
		t.Fatalf("deregister failed: %v", err)
	}

	_, err = svc.AuthenticateToken(context.Background(), registered.Token)
	if !errors.Is(err, agents.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken after deregister, got %v", err)
	}
}

func TestServiceAuthenticateToken_EmptyToken(t *testing.T) {
	svc := agents.New(newTestRepository(t), agents.NewStatusTracker(), zap.NewNop())

	_, err := svc.AuthenticateToken(context.Background(), "   ")
	if !errors.Is(err, agents.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}
