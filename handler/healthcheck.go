package handler

import "github.com/gofiber/fiber/v2"

// HealthCheck godoc
// @Summary            Show the status of server.
// @Description        get the status of server.
// @Tags               Healthcheck
// @Accept             */*
// @Produce            json
// @Success            200 {string} map[string]interface{}
// @Router             /healthcheck [GET]
func Healtcheck(ctx *fiber.Ctx) error {
	return ctx.SendStatus(fiber.StatusOK)
}
