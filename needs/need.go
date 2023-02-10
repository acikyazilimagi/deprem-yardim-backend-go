package needs

import (
	"time"
)

type CreateNeedRequest struct {
	Address     string `validate:"required"`
	Description string `validate:"required"`
}

type Need struct {
	ID               int64     `json:"id,omitempty"`
	Description      string    `json:"description"`
	IsResolved       bool      `json:"is_resolved"`
	Timestamp        time.Time `json:"timestamp"`
	ExtraParameters  *string   `json:"extra_parameters,omitempty"`
	FormattedAddress string    `json:"formatted_address,omitempty"`
	Loc              []float64 `json:"loc"`
}

type LiteNeed struct {
	ID int64 `json:"id,omitempty"`
}

type Response struct {
	Count   int    `json:"count"`
	Results []Need `json:"results"`
}
