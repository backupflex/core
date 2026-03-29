package agentauth

import (
	"errors"
	"fmt"

	"github.com/backupflex/core/internal/agents"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
)

type localsKey string

const authenticatedAgentIDKey localsKey = "authenticated_agent_id"

func NewAuth(agentsSvc *agents.Service) fiber.Handler {
	validateAgentToken := func(c *fiber.Ctx, token string) (bool, error) {
		agentID, err := agentsSvc.AuthenticateToken(c.Context(), token)
		if err != nil {
			if errors.Is(err, agents.ErrInvalidToken) {
				return false, nil
			}
			return false, fmt.Errorf("failed to validate agent token: %w", err)
		}

		c.Locals(authenticatedAgentIDKey, agentID)
		return true, nil
	}

	return keyauth.New(keyauth.Config{
		Validator: validateAgentToken,
		ErrorHandler: func(_ *fiber.Ctx, _ error) error {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or missing agent token")
		},
	})
}

func AuthenticatedAgentID(c *fiber.Ctx) string {
	agentID, _ := c.Locals(authenticatedAgentIDKey).(string)

	return agentID
}
