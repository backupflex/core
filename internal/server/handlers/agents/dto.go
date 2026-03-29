package agents

import (
	"time"

	"github.com/backupflex/core/internal/agents"
)

type agent struct {
	ID           string        `json:"id"`
	Role         agents.Role   `json:"role"`
	Address      string        `json:"address"`
	Capabilities []string      `json:"capabilities"`
	Status       agents.Status `json:"status"`
	RegisteredAt time.Time     `json:"registered_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	LastSeenAt   time.Time     `json:"last_seen_at"`
}

func newAgent(a agents.AgentWithStatus) agent {
	return agent{
		ID:           a.ID,
		Role:         a.Role,
		Address:      a.Address,
		Capabilities: a.Capabilities,
		Status:       a.Status,
		RegisteredAt: a.RegisteredAt,
		UpdatedAt:    a.UpdatedAt,
		LastSeenAt:   a.LastSeenAt,
	}
}

// RegisterRequest represents the request body for agent registration.
//
//	@Description	Request body for registering a new agent
type registerRequest struct {
	ID           string      `json:"id"           example:"agent-001"`
	Role         agents.Role `json:"role"         example:"backup"`
	Address      string      `json:"address"      example:"10.0.0.1:9000"`
	Capabilities []string    `json:"capabilities" example:"[\"backup\", \"restore\"]"`
}

// RegisterResponse represents the response body for successful agent registration.
//
//	@Description	Response body after successful agent registration
type registerResponse struct {
	Agent       agent  `json:"agent"`
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}
