package handler

import (
	"fmt"
	"github.com/acikkaynak/backend-api-go/feeds"
	"github.com/acikkaynak/backend-api-go/repository"
	"github.com/gofiber/fiber/v2"
	"strconv"
	"time"
)

func IsValidReason(key string) bool {
	reasons := []string{"", "enkaz", "erzak"}

	for _, reason := range reasons {
		if reason == key {
			return true
		}
	}
	return false
}

// getFeedAreas godoc
// @Summary            Get Feed areas with query strings
// @Tags               Feed
// @Produce            json
// @Success            200 {object} []feeds.Result

// @Param              sw_lat query number true "Sw Lat"
// @Param              sw_lng query number true "Sw Lng"
// @Param              ne_lat query number true "Ne Lat"
// @Param              ne_lng query number true "Ne Lng"
// @Param              time_stamp query integer false "Timestamp"
// @Router             /feeds/areas [GET]
func GetFeedAreas(repo *repository.Repository) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		swLatStr := ctx.Query("sw_lat")
		swLngStr := ctx.Query("sw_lng")
		neLatStr := ctx.Query("ne_lat")
		neLngStr := ctx.Query("ne_lng")
		timeStampStr := ctx.Query("time_stamp")
		reason := ctx.Query("reason", "")

		var timestamp int64
		if timeStampStr == "" {
			timestamp = time.Now().AddDate(-1, -1, -1).Unix()
		} else {
			timeInt, err := strconv.ParseInt(timeStampStr, 10, 64)
			if err != nil {
				timestamp = time.Now().AddDate(-1, -1, -1).Unix()
			} else {
				timestamp = timeInt
			}
		}

		swLat, _ := strconv.ParseFloat(swLatStr, 64)
		swLng, _ := strconv.ParseFloat(swLngStr, 64)
		neLat, _ := strconv.ParseFloat(neLatStr, 64)
		neLng, _ := strconv.ParseFloat(neLngStr, 64)
		if !IsValidReason(reason) {
			return ctx.JSON(fmt.Errorf("invalid reason: %s", reason))
		}

		data, err := repo.GetLocations(swLat, swLng, neLat, neLng, timestamp, reason)
		if err != nil {
			return ctx.JSON(err)
		}

		resp := &feeds.Response{
			Count:   len(data),
			Results: data,
		}

		return ctx.JSON(resp)
	}
}
