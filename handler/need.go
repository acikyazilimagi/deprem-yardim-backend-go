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
