package handler

import (
	"github.com/acikkaynak/backend-api-go/needs"
	"github.com/acikkaynak/backend-api-go/repository"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

type NeedsHandler struct {
	repo *repository.Repository
}

func NewNeedsHandler(repo *repository.Repository) *NeedsHandler {
	return &NeedsHandler{repo: repo}
}

// createNeed godoc
// @Summary            Create Need
// @Tags               Need
// @Produce            json
// @Success            200 {object} needs.LiteNeed
// @Param              body body needs.CreateNeedRequest true "RequestBody"
// @Router             /needs [POST]
func (h *NeedsHandler) HandleCreate(ctx *fiber.Ctx) error {
	req := needs.CreateNeedRequest{}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.JSON(err)
	}

	id, err := h.repo.CreateNeed(req.Address, req.Description)
	if err != nil {
		return ctx.JSON(err)
	}

	resp := &needs.LiteNeed{
		ID: id,
	}
	return ctx.JSON(resp)

}

// getNeeds godoc
// @Summary            Get Needs
// @Tags               Need
// @Produce            json
// @Success            200 {object} needs.Response
// @Param              only_not_resolved query bool true "Is Only Not Resolved"
// @Router             /needs [GET]
func (h *NeedsHandler) HandleList(ctx *fiber.Ctx) error {
	onlyNotResolvedStr := ctx.Query("only_not_resolved")
	onlyNotResolved, _ := strconv.ParseBool(onlyNotResolvedStr)
	data, err := h.repo.GetNeeds(onlyNotResolved)
	if err != nil {
		return ctx.JSON(err)
	}

	resp := &needs.Response{
		Count:   len(data),
		Results: data,
	}

	return ctx.JSON(resp)
}

// HandleTimestampList godoc
// @Summary            Get Needs after the given timestamp
// @Tags               Need
// @Produce            json
// @Success            200 {object} needs.Response
// @Param              only_not_resolved query bool true "Is Only Not Resolved"
// @Router             /needs [GET]
func (h *NeedsHandler) HandleTimestampList(ctx *fiber.Ctx) error {
	onlyNotResolvedStr := ctx.Query("only_not_resolved")
	onlyNotResolved, _ := strconv.ParseBool(onlyNotResolvedStr)

	timestampString := ctx.Params("timestamp")

	timestampInt, err := strconv.ParseInt(timestampString, 10, 64)

    if err != nil {
        timestampInt = 0
    }

	data, err := h.repo.GetTimestampFilteredNeeds(onlyNotResolved, timestampInt)

	if err != nil {
		return ctx.JSON(err)
	}

	resp := &needs.Response{
		Count:   len(data),
		Results: data,
	}

	return ctx.JSON(resp)
}
