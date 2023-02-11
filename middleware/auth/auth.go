package auth

import (
	"github.com/gofiber/fiber/v2"
	"os"
	"strings"
)

const ApiKeyHeaderName = "X-Api-Key"

func New() fiber.Handler {
	apiKey := os.Getenv("ApiKey")

	return func(ctx *fiber.Ctx) error {
		apiKeyNeeded := false

		if strings.Contains(ctx.Path(), "pprof") || ctx.Method() == fiber.MethodPost {
			apiKeyNeeded = true
		}

		if apiKeyNeeded && ctx.GetReqHeaders()[ApiKeyHeaderName] != apiKey {
			return ctx.SendStatus(fiber.StatusUnauthorized)
		}

		return ctx.Next()
	}
}
