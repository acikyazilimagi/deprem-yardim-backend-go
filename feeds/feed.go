package feeds

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Feed struct {
	ID               int64      `json:"id,omitempty"`
	FullText         string     `json:"full_text"`
	IsResolved       bool       `json:"is_resolved"`
	Channel          string     `json:"channel,omitempty"`
	Timestamp        time.Time  `json:"timestamp,omitempty"`
	Epoch            int64      `json:"epoch"`
	ExtraParameters  *string    `json:"extra_parameters,omitempty"`
	FormattedAddress string     `json:"formatted_address,omitempty"`
	Reason           *string    `json:"reason,omitempty"`
	Needs            []NeedItem `json:"needs,omitempty"`
}

type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Result struct {
	ID                 int64      `json:"id"`
	Loc                []float64  `json:"loc"`
	Entry_ID           int64      `json:"entry_id"`
	Timestamp          *string    `json:"timestamp,omitempty"`
	Epoch              int64      `json:"epoch,omitempty"`
	Reason             *string    `json:"reason,omitempty"`
	Channel            *string    `json:"channel,omitempty"`
	ExtraParameters    *string    `json:"extra_parameters,omitempty"`
	IsLocationVerified bool       `json:"is_location_verified"`
	IsNeedVerified     bool       `json:"is_need_verified"`
	Needs              []NeedItem `json:"needs,omitempty"`
}

type LiteResult struct {
	ID  int64     `json:"id"`
	Loc []float64 `json:"loc"`
}

type Response struct {
	Count   int      `json:"count"`
	Results []Result `json:"results"`
}

type Location struct {
	ID               int64     `json:"id"`
	FormattedAddress string    `json:"formatted_address"`
	Latitude         float64   `json:"latitude"`
	Longitude        float64   `json:"longitude"`
	NortheastLat     float64   `json:"northeast_lat"`
	NortheastLng     float64   `json:"northeast_lng"`
	SouthwestLat     float64   `json:"southwest_lat"`
	SouthwestLng     float64   `json:"southwest_lng"`
	EntryID          int64     `json:"entry_id"`
	Timestamp        time.Time `json:"timestamp"`
	Epoch            int64     `json:"epoch"`
	Reason           string    `json:"reason"`
	Channel          string    `json:"channel"`
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
	return json.Marshal(n)
}

func (n *NeedItem) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("NeedItem::Scan type assertion to []byte failed")
	}
	return json.Unmarshal(b, &n)
}
