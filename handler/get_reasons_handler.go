package handler

import (
	"github.com/acikkaynak/backend-api-go/repository"
	"github.com/gofiber/fiber/v2"
)

var (
	reasons = []string{
		"barınma",
		"battaniye",
		"ekip",
		"elektrik",
		"elektronik",
		"enkaz",
		"erzak",
		"genel",
		"giyecek",
		"giyim",
		"giysi",
		"guvenli-noktalar",
		"gıda",
		"hayvanlar-icin-tedavi",
		"hijyen",
		"ilaç",
		"ısınma",
		"kefen",
		"kişisel bakım",
		"kişisel ihtiyaç",
		"konaklama",
		"kurtarma",
		"lojistik",
		"operatör",
		"pet",
		"sağlık",
		"su",
		"teçhizat",
		"ulaşım",
		"yakıt",
		"yemek",
		"çadır",
		"çocuk ihtiyaçları",
		"ısınma",
	}
)

type GetReasonResponse struct {
	Reasons []string `json:"reasons"`
}

func GetReasonsHandler(repo *repository.Repository) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		response := GetReasonResponse{Reasons: reasons}
		return ctx.JSON(response)
	}
}
