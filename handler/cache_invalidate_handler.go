package handler

import (
	"github.com/acikkaynak/backend-api-go/cache"
	"github.com/gofiber/fiber/v2"
)

func InvalidateCache() fiber.Handler {
	cacheRepo := cache.NewRedisRepository()

	return func(ctx *fiber.Ctx) error {
		err := cacheRepo.Prune()

		if err != nil {
			ctx.Status(fiber.StatusInternalServerError)
			return ctx.SendString(err.Error())
		}

		return ctx.SendStatus(fiber.StatusOK)
	}
}
