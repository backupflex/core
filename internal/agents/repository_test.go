package agents_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/backupflex/core/internal/agents"
	"github.com/dgraph-io/badger/v4"
)

func TestRepositoryTokenIsPersistentAcrossRestart(t *testing.T) {
	dir := t.TempDir()
	const token = "token-persistent-1"

	open := func() *badger.DB {
		db, err := badger.Open(badger.DefaultOptions(filepath.Join(dir, "badger")).WithLogger(nil))
		if err != nil {
			t.Fatalf("open badger: %v", err)
		}
		return db
	}

	db := open()
	repo := agents.NewRepository(db)

	if _, err := repo.Register(
		context.Background(),
		agents.AgentInput{ID: "agent-2", Role: agents.RoleBackup},
		token,
	); err != nil {
		t.Fatalf("register: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("close db: %v", err)
	}

	db = open()
	t.Cleanup(func() { _ = db.Close() })
	repo = agents.NewRepository(db)

	agentID, err := repo.ResolveToken(context.Background(), token)
	if err != nil {
		t.Fatalf("resolve token after restart: %v", err)
	}

	if agentID != "agent-2" {
		t.Fatalf("expected agent-2, got %s", agentID)
	}
}
