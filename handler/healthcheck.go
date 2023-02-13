package handler

import "github.com/gofiber/fiber/v2"

// HealthCheck godoc
//	@Summary		Show the status of server.
//	@Description	get the status of server.
//	@Tags			HealthCheck
//	@Accept			*/*
//	@Produce		json
//	@Success		200	{string}	nil
//	@Router			/healthcheck [GET]
func HealthCheck(ctx *fiber.Ctx) error {
	return ctx.SendStatus(fiber.StatusOK)
}
