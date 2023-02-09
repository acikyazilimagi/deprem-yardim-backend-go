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
	FormattedAddress string   `json:"formatted_address,omitempty"`
}

type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Result struct {
	ID               int64     `json:"id"`
	Loc              []float64 `json:"loc"`
	Entry_ID	     int64 	   `json:"entry_id"`
}

type LiteResult struct {
	ID               int64     `json:"id"`
	Loc              []float64 `json:"loc"`
}

type Response struct {
	Count   int      `json:"count"`
	Results []Result `json:"results"`
}
