package agents

import (
	"errors"

	"github.com/backupflex/core/internal/agents"
	"github.com/gofiber/fiber/v2"
)

func errorsHandler(c *fiber.Ctx) error {
	err := c.Next()
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, agents.ErrAgentAlreadyExists):
		return fiber.NewError(fiber.StatusConflict, err.Error())
	case errors.Is(err, agents.ErrAgentNotFound):
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	case errors.Is(err, agents.ErrInvalidToken):
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	case errors.Is(err, agents.ErrValidationFailed):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return err //nolint:wrapcheck //already wrapped
}
