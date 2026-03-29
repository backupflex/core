package agents

import (
	"fmt"
	"strings"
	"time"
)

// Role describes the functional type of an agent.
type Role string

const (
	// RoleBackup performs backup creation operations.
	RoleBackup Role = "backup"
	// RoleVerification validates created backups.
	RoleVerification Role = "verification"
)

// Status describes current state of an agent.
type Status string

const (
	// StatusOnline means the agent is connected and ready.
	StatusOnline Status = "online"
	// StatusOffline means the agent is disconnected.
	StatusOffline Status = "offline"
)

type AgentInput struct {
	ID           string
	Role         Role
	Address      string
	Capabilities []string
}

// RuntimeStatus holds in-memory liveness information.
type RuntimeStatus struct {
	Status     Status
	LastSeenAt time.Time
	UpdatedAt  time.Time
}

// Agent stores metadata required by the coordinator to route tasks to agents.
type Agent struct {
	AgentInput

	RegisteredAt time.Time
}

type AgentWithStatus struct {
	Agent
	RuntimeStatus
}

type RegisteredAgent struct {
	AgentWithStatus

	Token string
}

// Validate ensures the agent has the minimum required fields.
func (a AgentInput) Validate() error {
	if strings.TrimSpace(a.ID) == "" {
		return fmt.Errorf("%w: agent id is required", ErrValidationFailed)
	}

	switch a.Role {
	case RoleBackup, RoleVerification:
		return nil
	default:
		return fmt.Errorf("%w: invalid role: %s", ErrValidationFailed, a.Role)
	}
}
