package handler

import (
	"fmt"
	"github.com/acikkaynak/backend-api-go/feeds"
	"github.com/acikkaynak/backend-api-go/repository"
	"github.com/gofiber/fiber/v2"
	"net/http"
)

// updateFeedLocations godoc
// @Summary            Update feed locations with correct address and location
// @Tags               Feed
// @Accept             json
// @Produce            json
// @Success            202
// @Param              UpdateFeedLocationsRequest body feeds.UpdateFeedLocationsRequest true "RequestBody"
// @Router             /feeds/areas [PATCH]
func UpdateFeedLocationsHandler(repo *repository.Repository) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var req feeds.UpdateFeedLocationsRequest

		if err := ctx.BodyParser(&req); err != nil {
			return fmt.Errorf("failed to decode request. err: %w", err)
		}

		err := repo.UpdateFeedLocations(ctx.Context(), req.FeedLocations)
		if err != nil {
			return ctx.JSON(err)
		}

		return ctx.SendStatus(http.StatusAccepted)
	}
}
