package agents

import (
	"fmt"
	"time"

	"github.com/backupflex/core/internal/agents"
	"github.com/backupflex/core/internal/server/middlewares/agentauth"
	"github.com/go-core-fx/fiberfx/handler"
	"github.com/gofiber/fiber/v2"
)

// AgentHandler serves agent-side API for self-registration and liveness.
type AgentHandler struct {
	agentsSvc *agents.Service
}

func NewAgentHandler(agents *agents.Service) handler.Handler {
	return &AgentHandler{agentsSvc: agents}
}

func (h *AgentHandler) Register(router fiber.Router) {
	group := router.Group("/agents")

	group.Use(errorsHandler)

	group.Post("/register", h.register)

	group.Use(agentauth.NewAuth(h.agentsSvc))
	group.Post("/heartbeat", h.heartbeat)
}

// Register agent
//
//	@Summary		Register a new agent
//	@Description	Register a new agent with the coordinator and receive an access token
//	@Tags			Agents
//	@Accept			json
//	@Produce		json
//	@Param			request	body		registerRequest			true	"Agent registration request"
//	@Success		201		{object}	registerResponse		"Agent registered successfully"
//	@Failure		400		{object}	fiberfx.ErrorResponse	"Invalid request body or role"
//	@Failure		409		{object}	fiberfx.ErrorResponse	"Agent already exists"
//	@Failure		500		{object}	fiberfx.ErrorResponse	"Internal server error"
//	@Router			/agents/register [post]
func (h *AgentHandler) register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	created, err := h.agentsSvc.Register(
		c.Context(),
		agents.AgentInput{
			ID:           req.ID,
			Role:         req.Role,
			Address:      req.Address,
			Capabilities: req.Capabilities,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to register agent: %w", err)
	}

	return c.Status(fiber.StatusCreated).
		JSON(registerResponse{Agent: newAgent(created.AgentWithStatus), AccessToken: created.Token})
}

// Heartbeat endpoint for agents
//
//	@Summary		Send agent heartbeat
//	@Description	Update agent last seen timestamp to maintain liveness
//	@Tags			Agents
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		204	"No content, heartbeat accepted"
//	@Failure		401	{object}	fiberfx.ErrorResponse	"Invalid or missing bearer token"
//	@Failure		404	{object}	fiberfx.ErrorResponse	"Agent not found"
//	@Failure		500	{object}	fiberfx.ErrorResponse	"Internal server error"
//	@Router			/agents/heartbeat [post]
func (h *AgentHandler) heartbeat(c *fiber.Ctx) error {
	agentID := agentauth.AuthenticatedAgentID(c)
	if agentID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "missing agent identity")
	}

	if err := h.agentsSvc.Heartbeat(c.Context(), agentID, time.Now().UTC()); err != nil {
		return fmt.Errorf("failed to heartbeat agent: %w", err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
