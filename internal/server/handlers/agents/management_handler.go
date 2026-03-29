package agents

import (
	"fmt"
	"time"

	"github.com/backupflex/core/internal/agents"
	"github.com/go-core-fx/fiberfx/handler"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
)

// ManagementHandler serves management-side API for operators.
type ManagementHandler struct {
	agents *agents.Service
}

func NewManagementHandler(agents *agents.Service) handler.Handler {
	return &ManagementHandler{agents: agents}
}

func (h *ManagementHandler) Register(router fiber.Router) {
	group := router.Group("/management/agents")

	group.Use(errorsHandler)

	group.Get("/", h.list)
	group.Get("/:id", h.get)
	group.Delete("/:id", h.delete)
	group.Post("/:id/offline", h.setOffline)
}

// List all agents
//
//	@Summary		List all registered agents
//	@Description	Retrieve a list of all registered agents (management only)
//	@Tags			Management
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		agent					"List of agents"
//	@Failure		500	{object}	fiberfx.ErrorResponse	"Internal server error"
//	@Router			/management/agents [get]
func (h *ManagementHandler) list(c *fiber.Ctx) error {
	items, err := h.agents.List(c.Context())
	if err != nil {
		return fmt.Errorf("failed to list agents: %w", err)
	}

	return c.JSON(
		lo.Map(items, func(agent agents.AgentWithStatus, _ int) agent {
			return newAgent(agent)
		}),
	)
}

// Get agent by ID
//
//	@Summary		Get agent details
//	@Description	Retrieve details of a specific agent by its ID (management only)
//	@Tags			Management
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Agent ID"
//	@Success		200	{object}	agent					"Agent details"
//	@Failure		404	{object}	fiberfx.ErrorResponse	"Agent not found"
//	@Failure		500	{object}	fiberfx.ErrorResponse	"Internal server error"
//	@Router			/management/agents/{id} [get]
func (h *ManagementHandler) get(c *fiber.Ctx) error {
	agentEntity, err := h.agents.Get(c.Context(), c.Params("id"))
	if err != nil {
		return fmt.Errorf("failed to get agent: %w", err)
	}

	return c.JSON(newAgent(*agentEntity))
}

// Delete agent
//
//	@Summary		Delete an agent
//	@Description	Permanently remove an agent from the system (management only)
//	@Tags			Management
//	@Accept			json
//	@Produce		json
//	@Param			id	path	string	true	"Agent ID"
//	@Success		204	"Agent deleted successfully"
//	@Failure		404	{object}	fiberfx.ErrorResponse	"Agent not found"
//	@Failure		500	{object}	fiberfx.ErrorResponse	"Internal server error"
//	@Router			/management/agents/{id} [delete]
func (h *ManagementHandler) delete(c *fiber.Ctx) error {
	if err := h.agents.Deregister(c.Context(), c.Params("id")); err != nil {
		return fmt.Errorf("failed to deregister agent: %w", err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// Set agent offline
//
//	@Summary		Mark an agent as offline
//	@Description	Manually set an agent's status to offline (management only)
//	@Tags			Management
//	@Accept			json
//	@Produce		json
//	@Param			id	path	string	true	"Agent ID"
//	@Success		204	"Agent status set to offline"
//	@Failure		404	{object}	fiberfx.ErrorResponse	"Agent not found"
//	@Failure		500	{object}	fiberfx.ErrorResponse	"Internal server error"
//	@Router			/management/agents/{id}/offline [post]
func (h *ManagementHandler) setOffline(c *fiber.Ctx) error {
	if err := h.agents.SetOffline(c.Context(), c.Params("id"), time.Now().UTC()); err != nil {
		return fmt.Errorf("failed to set agent offline: %w", err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
