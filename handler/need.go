package handler

import (
	"strconv"

	"github.com/acikkaynak/backend-api-go/needs"
	"github.com/acikkaynak/backend-api-go/repository"
	"github.com/gofiber/fiber/v2"
)

type NeedsHandler struct {
	repo *repository.Repository
}

func NewNeedsHandler(repo *repository.Repository) *NeedsHandler {
	return &NeedsHandler{repo: repo}
}

// HandleCreate godoc
//
//	@Summary	Create Need
//	@Tags		Need
//	@Produce	json
//	@Success	200		{object}	needs.LiteNeed
//	@Param		body	body		needs.CreateNeedRequest	true	"RequestBody"
//	@Security	ApiKeyAuth
//	@Router		/needs [POST]
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

// HandleList godoc
//
//	@Summary	Get Needs
//	@Tags		Need
//	@Produce	json
//	@Success	200					{object}	needs.Response
//	@Param		only_not_resolved	query		bool	true	"Is Only Not Resolved"
//	@Router		/needs [GET]
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
