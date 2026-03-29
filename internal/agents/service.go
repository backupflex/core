package agents

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

// Service provides use-cases for managing agents.
type Service struct {
	repository    *Repository
	statusTracker *StatusTracker
	logger        *zap.Logger
}

// New creates a new Service instance.
func New(repository *Repository, statusTracker *StatusTracker, logger *zap.Logger) *Service {
	return &Service{repository: repository, statusTracker: statusTracker, logger: logger}
}

// Register validates and registers a new agent.
func (s *Service) Register(ctx context.Context, draft AgentInput) (*RegisteredAgent, error) {
	if err := draft.Validate(); err != nil {
		return nil, err
	}

	token := uuid.NewString()
	agent, err := s.repository.Register(ctx, draft, token)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	status := s.statusTracker.Touch(draft.ID, now)

	s.logger.Info("agent registered", zap.String("agent_id", draft.ID), zap.String("role", string(draft.Role)))
	return &RegisteredAgent{
		AgentWithStatus: AgentWithStatus{
			Agent:         *agent,
			RuntimeStatus: status,
		},
		Token: token,
	}, nil
}

// AuthenticateToken resolves token to agent id.
func (s *Service) AuthenticateToken(ctx context.Context, token string) (string, error) {
	if strings.TrimSpace(token) == "" {
		return "", ErrInvalidToken
	}

	return s.repository.ResolveToken(ctx, token)
}

// Heartbeat marks the agent as online and refreshes timestamps.
func (s *Service) Heartbeat(ctx context.Context, id string, now time.Time) error {
	if err := s.repository.IsExists(ctx, id); err != nil {
		return err
	}

	s.statusTracker.Touch(id, now)

	s.logger.Debug("agent heartbeat", zap.String("agent_id", id))
	return nil
}

// SetOffline marks an agent as offline.
func (s *Service) SetOffline(ctx context.Context, id string, now time.Time) error {
	if err := s.repository.IsExists(ctx, id); err != nil {
		return err
	}

	s.statusTracker.SetOffline(id, now)

	s.logger.Info("agent disconnected", zap.String("agent_id", id))
	return nil
}

// Get returns single agent by id.
func (s *Service) Get(ctx context.Context, id string) (*AgentWithStatus, error) {
	agent, err := s.repository.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	status := s.statusTracker.Get(id)

	return &AgentWithStatus{
		Agent:         *agent,
		RuntimeStatus: status,
	}, nil
}

// Deregister removes an agent from the registry.
func (s *Service) Deregister(ctx context.Context, id string) error {
	if err := s.repository.Remove(ctx, id); err != nil {
		return err
	}

	s.statusTracker.Remove(id)

	s.logger.Info("agent deregistered", zap.String("agent_id", id))
	return nil
}

// List returns all known agents.
func (s *Service) List(ctx context.Context) ([]AgentWithStatus, error) {
	agents, err := s.repository.List(ctx)
	if err != nil {
		return nil, err
	}

	return lo.Map(
		agents,
		func(agent Agent, _ int) AgentWithStatus {
			status := s.statusTracker.Get(agent.ID)
			return AgentWithStatus{
				Agent:         agent,
				RuntimeStatus: status,
			}
		},
	), nil
}
