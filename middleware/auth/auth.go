package auth

import (
	"github.com/gofiber/fiber/v2"
	"os"
	"strings"
)

const ApiKeyHeaderName = "X-Api-Key"

var restrictedHttpMethods = map[string]struct{}{
	"POST":   {},
	"DELETE": {},
	"PUT":    {},
	"PATCH":  {},
}

func New() fiber.Handler {
	apiKey := os.Getenv("ApiKey")

	return func(ctx *fiber.Ctx) error {
		apiKeyNeeded := false
		_, restrictedMethod := restrictedHttpMethods[ctx.Method()]
		if strings.Contains(ctx.Path(), "pprof") || restrictedMethod {
			apiKeyNeeded = true
		}

		if apiKeyNeeded && ctx.GetReqHeaders()[ApiKeyHeaderName] != apiKey {
			return ctx.SendStatus(fiber.StatusUnauthorized)
		}

		return ctx.Next()
	}
}
