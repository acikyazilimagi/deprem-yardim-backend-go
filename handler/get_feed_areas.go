package handler

import (
	"strconv"
	"time"

	"github.com/acikkaynak/backend-api-go/feeds"
	"github.com/acikkaynak/backend-api-go/repository"
	"github.com/gofiber/fiber/v2"
)

// GetFeedAreas godoc
//
//	@Summary	Get Feed areas with query strings
//	@Tags		Feed
//	@Produce	json
//	@Success	200			{object}	[]feeds.Result
//	@Param		sw_lat		query		number	true	"Sw Lat"
//	@Param		sw_lng		query		number	true	"Sw Lng"
//	@Param		ne_lat		query		number	true	"Ne Lat"
//	@Param		ne_lng		query		number	true	"Ne Lng"
//	@Param		time_stamp	query		integer	false	"Timestamp"
//	@Param		reason		query		string	false	"Reason",
//	@Param		channel		query		string	false	"Channel"
//	@Router		/feeds/areas [GET]
func GetFeedAreas(repo *repository.Repository) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		swLatStr := ctx.Query("sw_lat")
		swLngStr := ctx.Query("sw_lng")
		neLatStr := ctx.Query("ne_lat")
		neLngStr := ctx.Query("ne_lng")
		timeStampStr := ctx.Query("time_stamp")
		reason := ctx.Query("reason", "")
		channel := ctx.Query("channel", "")
		extraParams := ctx.Query("extraParams", "")
		isLocationVerified := ctx.Query("is_location_verified", "")
		isNeedVerified := ctx.Query("is_need_verified", "")

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

		extraParamsBool, _ := strconv.ParseBool(extraParams)

		data, err := repo.GetLocations(swLat, swLng, neLat, neLng, timestamp, reason, channel, extraParamsBool, isLocationVerified, isNeedVerified)
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
