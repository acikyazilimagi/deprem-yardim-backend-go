package auth

import (
	"github.com/gofiber/fiber/v2"
	"os"
)

const ApiKeyHeaderName = "X-Api-Key"

func New() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		apiKey := os.Getenv("ApiKey")
		if ctx.Method() == fiber.MethodPost && ctx.GetReqHeaders()[ApiKeyHeaderName] != apiKey {
			return ctx.SendStatus(fiber.StatusUnauthorized)
		}
		return ctx.Next()
	}
}
