package feeds

import (
	"time"
)

type Feed struct {
	ID              int64     `json:"id,omitempty"`
	FullText        string    `json:"full_text"`
	IsResolved      bool      `json:"is_resolved"`
	Channel         string    `json:"channel,omitempty"`
	Timestamp       time.Time `json:"timestamp"`
	ExtraParameters *string   `json:"extra_parameters,omitempty"`
}

type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type ViewPort struct {
	Northeast LatLng `json:"northeast"`
	Southwest LatLng `json:"southwest"`
}

type Result struct {
	FormattedAddress string    `json:"formatted_address"`
	ID               int64     `json:"id"`
	Loc              []float64 `json:"loc"`
	Raw              Feed      `json:"raw"`
	ViewPort         ViewPort  `json:"view_port"`
}

type Response struct {
	Count   int      `json:"count"`
	Results []Result `json:"results"`
}
