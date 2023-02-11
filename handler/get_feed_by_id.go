package handler

import (
	"github.com/acikkaynak/backend-api-go/feeds"
	"github.com/acikkaynak/backend-api-go/repository"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

// GetFeedById godoc
// @Summary            Get Feeds with given id
// @Tags               Feed
// @Produce            json
// @Success            200 {object} feeds.Feed
// @Param              id path integer true "Feed Id"
// @Router             /feeds/{id} [GET]
func GetFeedById(repo *repository.Repository) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		feedIDStr := ctx.Params("id")

		feedID, err := strconv.ParseInt(feedIDStr, 10, 64)
		if err != nil {
			return ctx.SendStatus(fiber.StatusBadRequest)
		}

		feed, err := repo.GetFeed(feedID)
		if err != nil {
			return ctx.JSON(feeds.ErrorResponse{Message: err.Error()})
		}

		return ctx.JSON(feed)
	}
}
