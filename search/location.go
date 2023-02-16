package search

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/acikkaynak/backend-api-go/feeds"
	"github.com/acikkaynak/backend-api-go/repository"
)

type LocationIndex struct {
	connStr   string
	index     *index[Location]
	indexName string
}

func NewLocationIndex() *LocationIndex {
	connStr := os.Getenv("ELASTIC_CONN_STR")
	indexName := "locations"

	if connStr == "" {
		log.Panic("ELASTIC_CONN_STR env variable must be set")
	}

	return &LocationIndex{
		connStr:   connStr,
		index:     NewIndex[Location](indexName),
		indexName: indexName,
	}
}

func (l *LocationIndex) GetLocations(getLocationsQuery *repository.GetLocationsQuery) ([]feeds.Location, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*25)
	defer cancel()

	var filters []map[string]interface{}

	if getLocationsQuery.SwLat != 0.0 || getLocationsQuery.SwLng != 0.0 || getLocationsQuery.NeLat != 0.0 || getLocationsQuery.NeLng != 0.0 {
		filters = append(filters, map[string]interface{}{
			"range": map[string]interface{}{
				"raw_locations.top_right.lat": map[string]interface{}{
					"lte": getLocationsQuery.NeLat,
				},
			},
		})

		filters = append(filters, map[string]interface{}{
			"range": map[string]interface{}{
				"raw_locations.top_right.lng": map[string]interface{}{
					"lte": getLocationsQuery.NeLng,
				},
			},
		})

		filters = append(filters, map[string]interface{}{
			"range": map[string]interface{}{
				"raw_locations.bottom_left.lat": map[string]interface{}{
					"gte": getLocationsQuery.SwLat,
				},
			},
		})

		filters = append(filters, map[string]interface{}{
			"range": map[string]interface{}{
				"raw_locations.bottom_left.lng": map[string]interface{}{
					"gte": getLocationsQuery.SwLng,
				},
			},
		})
	}

	if getLocationsQuery.Timestamp != 0 {
		if getLocationsQuery.Channel != "ahbap_location" {
			filters = append(filters, map[string]interface{}{
				"range": map[string]interface{}{
					"epoch": map[string]interface{}{
						"gte": getLocationsQuery.Timestamp,
					},
				},
			})
		}
	}

	if getLocationsQuery.Reason != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"reason": getLocationsQuery.Reason,
			},
		})
	}

	if getLocationsQuery.Channel != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"channel": getLocationsQuery.Channel,
			},
		})
	}

	if getLocationsQuery.IsLocationVerified != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"is_location_verified": getLocationsQuery.IsLocationVerified,
			},
		})
	}

	if getLocationsQuery.IsNeedVerified != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"is_need_verified": getLocationsQuery.IsNeedVerified,
			},
		})
	}

	query := map[string]interface{}{
		"track_total_hits": true,
		"size":             10000,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": filters,
			},
		},
	}

	res, err := l.index.Search(ctx, query)

	if err != nil {
		return nil, 0, err
	}

	var results []feeds.Location

	for _, hit := range res.Hits.Hits {
		source := hit.Source
		id, _ := strconv.ParseInt(hit.Id, 10, 64)
		reasons := strings.Join(source.Reason, ",")
		channels := strings.Join(source.Channel, ",")

		results = append(results, feeds.Location{
			ID: id,
			Loc: []float64{
				source.RawLocations.Center.Lat,
				source.RawLocations.Center.Lon,
			},
			EntryID:            source.EntryId,
			Epoch:              source.Epoch,
			Reason:             &reasons,
			Channel:            &channels,
			IsLocationVerified: source.IsLocationVerified,
			IsNeedVerified:     source.IsNeedVerified,
			Needs:              source.Needs,
			ExtraParameters:    source.ExtraParameters,
		})
	}

	return results, res.Hits.Total.Value, nil
}

func (l *LocationIndex) CreateFeedLocation(ctx context.Context, fullText string, location feeds.Location) error {
	locations := Locations{
		Center: Coordinates{
			Lat: location.Latitude,
			Lon: location.Longitude,
		},
		TopRight: Coordinates{
			Lat: location.NortheastLat,
			Lon: location.NortheastLng,
		},
		BottomLeft: Coordinates{
			Lat: location.SouthwestLat,
			Lon: location.SouthwestLng,
		},
	}

	reason := strings.Split(*location.Reason, ",")
	channel := strings.Split(*location.Channel, ",")

	item := Item[Location]{
		Index: l.indexName,
		Id:    strconv.FormatInt(location.ID, 10),
		Source: Location{
			FormattedAddress:   location.FormattedAddress,
			Locations:          locations,
			RawLocations:       locations,
			FullText:           fullText,
			ExtraParameters:    location.ExtraParameters,
			Channel:            channel,
			Reason:             reason,
			EntryId:            location.EntryID,
			Epoch:              location.Epoch,
			IsLocationVerified: false,
			IsNeedVerified:     false,
			IsDeleted:          false,
			Needs:              location.Needs,
		},
	}

	return l.index.Bulk(ctx, []Item[Location]{item})
}
