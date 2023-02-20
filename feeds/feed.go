package feeds

import (
	"database/sql/driver"
	"errors"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type Feed struct {
	ID               int64     `json:"id,omitempty"`
	FullText         string    `json:"full_text"`
	IsResolved       bool      `json:"is_resolved"`
	Channel          string    `json:"channel,omitempty"`
	Timestamp        time.Time `json:"timestamp,omitempty"`
	Epoch            int64     `json:"epoch"`
	ExtraParameters  *string   `json:"extra_parameters,omitempty"`
	FormattedAddress string    `json:"formatted_address,omitempty"`
	Reason           *string   `json:"reason,omitempty"`
	Lat              *float64  `json:"lat,omitempty"`
	Lng              *float64  `json:"lng,omitempty"`
}

type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type LiteResult struct {
	ID  int64     `json:"id"`
	Loc []float64 `json:"loc"`
}

type Response struct {
	Count   int        `json:"count"`
	Results []Location `json:"results"`
}

type Location struct {
	ID                 int64      `json:"id"`
	FormattedAddress   string     `json:"formatted_address"`
	Latitude           float64    `json:"latitude"`
	Longitude          float64    `json:"longitude"`
	NortheastLat       float64    `json:"northeast_lat,omitempty"`
	NortheastLng       float64    `json:"northeast_lng,omitempty"`
	SouthwestLat       float64    `json:"southwest_lat,omitempty"`
	SouthwestLng       float64    `json:"southwest_lng,omitempty"`
	EntryID            int64      `json:"entry_id"`
	Epoch              int64      `json:"epoch"`
	Reason             *string    `json:"reason,omitempty"`
	Channel            *string    `json:"channel,omitempty"`
	ExtraParameters    *string    `json:"extra_parameters,omitempty"`
	IsLocationVerified bool       `json:"is_location_verified,omitempty"`
	IsNeedVerified     bool       `json:"is_need_verified,omitempty"`
	Needs              []NeedItem `json:"needs,omitempty"`
	Loc                []float64  `json:"loc"`
}

type UpdateFeedLocationsRequest struct {
	FeedLocations []FeedLocation `json:"feed_locations"`
}

type FeedLocation struct {
	EntryID   int64   `json:"entry_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
}

type NeedItem struct {
	Label  string `json:"label"`
	Status bool   `json:"status"`
}

func (n NeedItem) Value() (driver.Value, error) {
	return jsoniter.Marshal(n)
}

func (n *NeedItem) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("NeedItem::Scan type assertion to []byte failed")
	}
	return jsoniter.Unmarshal(b, &n)
}
