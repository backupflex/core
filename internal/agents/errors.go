package agents

import "errors"

var (
	// ErrAgentAlreadyExists is returned when trying to register an existing agent.
	ErrAgentAlreadyExists = errors.New("agent already exists")
	// ErrAgentNotFound is returned when an agent was not found.
	ErrAgentNotFound = errors.New("agent not found")
	// ErrInvalidToken is returned when token is unknown or malformed.
	ErrInvalidToken = errors.New("invalid agent token")
	// ErrValidationFailed is returned when agent validation fails.
	ErrValidationFailed = errors.New("validation failed")
)
