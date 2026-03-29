package agents

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
)

const (
	agentKeyPrefix = "agent:"
	tokenKeyPrefix = "agent_token:"
)

// Repository stores and retrieves registered agents.
//
// It persists agent profiles and tokens to BadgerDB. Status tracking
// is handled by a separate StatusTracker component.
type Repository struct {
	db *badger.DB
}

// NewRepository creates persistent storage for agent profiles.
func NewRepository(db *badger.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// Register stores a new agent profile in registry.
func (r *Repository) Register(_ context.Context, input AgentInput, token string) (*Agent, error) {
	now := time.Now().UTC()
	agent := Agent{
		AgentInput:   input,
		RegisteredAt: now,
	}
	err := r.db.Update(func(txn *badger.Txn) error {
		key := storageKey(input.ID)
		if _, err := txn.Get([]byte(key)); err == nil {
			return ErrAgentAlreadyExists
		} else if !errors.Is(err, badger.ErrKeyNotFound) {
			return fmt.Errorf("failed to check existing agent: %w", err)
		}

		payload, err := json.Marshal(toPersistedAgent(agent))
		if err != nil {
			return fmt.Errorf("failed to marshal agent: %w", err)
		}

		if setErr := txn.Set([]byte(key), payload); setErr != nil {
			return fmt.Errorf("failed to persist agent: %w", setErr)
		}

		if tokenErr := r.saveToken(txn, input.ID, token, now); tokenErr != nil {
			return fmt.Errorf("failed to persist agent token: %w", tokenErr)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to register agent: %w", err)
	}

	return &agent, nil
}

// saveToken stores access token mapping in persistent storage.
func (r *Repository) saveToken(txn *badger.Txn, agentID, rawToken string, now time.Time) error {
	payload, err := json.Marshal(tokenModel{AgentID: agentID, CreatedAt: now})
	if err != nil {
		return fmt.Errorf("failed to marshal token payload: %w", err)
	}

	if setErr := txn.Set([]byte(tokenStorageKey(rawToken)), payload); setErr != nil {
		return fmt.Errorf("failed to persist agent token: %w", setErr)
	}

	return nil
}

// ResolveToken resolves agent id by access token.
func (r *Repository) ResolveToken(_ context.Context, rawToken string) (string, error) {
	var agentID string

	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(tokenStorageKey(rawToken)))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return ErrInvalidToken
			}
			return fmt.Errorf("failed to read token: %w", err)
		}

		return item.Value(func(val []byte) error {
			var token tokenModel
			if jsonErr := json.Unmarshal(val, &token); jsonErr != nil {
				return fmt.Errorf("failed to decode token: %w", jsonErr)
			}
			agentID = token.AgentID
			return nil
		})
	})
	if err != nil {
		return "", fmt.Errorf("failed to resolve token: %w", err)
	}

	return agentID, nil
}

// Get returns agent by id.
func (r *Repository) Get(_ context.Context, id string) (*Agent, error) {
	var persisted *agentModel

	err := r.db.View(func(txn *badger.Txn) error {
		model, err := r.readPersisted(txn, id)
		if err != nil {
			return err
		}

		persisted = model

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read agent: %w", err)
	}

	return persisted.toDomain(), nil
}

// List returns all registered agents.
func (r *Repository) List(_ context.Context) ([]Agent, error) {
	result := make([]Agent, 0)
	err := r.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte(agentKeyPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var profile agentModel
				if err := json.Unmarshal(val, &profile); err != nil {
					return fmt.Errorf("failed to decode agent: %w", err)
				}

				result = append(result, *profile.toDomain())
				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to read agent: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	return result, nil
}

// Remove removes agent from registry.
func (r *Repository) Remove(_ context.Context, id string) error {
	if err := r.db.Update(func(txn *badger.Txn) error {
		key := []byte(storageKey(id))
		if _, err := txn.Get(key); err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return ErrAgentNotFound
			}
			return fmt.Errorf("failed to read agent before delete: %w", err)
		}

		if err := txn.Delete(key); err != nil {
			return fmt.Errorf("failed to delete agent: %w", err)
		}

		if err := r.deleteTokensByAgent(txn, id); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to remove agent: %w", err)
	}

	return nil
}

// DeleteTokensByAgent removes all tokens linked to specific agent id.
func (r *Repository) deleteTokensByAgent(txn *badger.Txn, agentID string) error {
	it := txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	prefix := []byte(tokenKeyPrefix)
	toDelete := make([][]byte, 0)

	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		err := item.Value(func(val []byte) error {
			var token tokenModel
			if err := json.Unmarshal(val, &token); err != nil {
				return fmt.Errorf("failed to decode token: %w", err)
			}
			if token.AgentID == agentID {
				toDelete = append(toDelete, append([]byte(nil), item.Key()...))
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to read token: %w", err)
		}
	}

	for _, key := range toDelete {
		if err := txn.Delete(key); err != nil {
			return fmt.Errorf("failed to delete token: %w", err)
		}
	}

	return nil
}

// IsExists validates that agent exists.
func (r *Repository) IsExists(_ context.Context, id string) error {
	err := r.db.View(func(txn *badger.Txn) error {
		_, err := r.readPersisted(txn, id)
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to read agent: %w", err)
	}

	return nil
}

func (r *Repository) readPersisted(txn *badger.Txn, id string) (*agentModel, error) {
	var profile agentModel

	item, err := txn.Get([]byte(storageKey(id)))
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, ErrAgentNotFound
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	err = item.Value(func(val []byte) error {
		if jsonErr := json.Unmarshal(val, &profile); jsonErr != nil {
			return fmt.Errorf("failed to decode agent: %w", jsonErr)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to read agent: %w", err)
	}

	return &profile, nil
}

func toPersistedAgent(agent Agent) agentModel {
	return agentModel{
		ID:           agent.ID,
		Role:         agent.Role,
		Address:      agent.Address,
		Capabilities: agent.Capabilities,
		RegisteredAt: agent.RegisteredAt,
	}
}

func storageKey(id string) string {
	return agentKeyPrefix + id
}

func tokenStorageKey(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))
	return tokenKeyPrefix + hex.EncodeToString(hash[:])
}
