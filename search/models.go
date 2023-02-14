package search

// Common Models

type Total struct {
	Value int `json:"value"`
}

type Result[T any] struct {
	Hits Hits[T] `json:"hits"`
}

type Hits[T any] struct {
	Total Total     `json:"total"`
	Hits  []Item[T] `json:"hits"`
}

type Item[T any] struct {
	Index  string `json:"_index"`
	Id     string `json:"_id"`
	Source T      `json:"_source"`
}

// Location Index Specific Models

type Coordinates struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Locations struct {
	Center     Coordinates `json:"center"`
	TopRight   Coordinates `json:"top_right"`
	BottomLeft Coordinates `json:"bottom_left"`
}

type Need struct {
	Label  string `json:"label"`
	Status bool   `json:"status"`
}

type Location struct {
	FormattedAddress   string    `json:"formatted_address"`
	Locations          Locations `json:"locations"`
	RawLocations       Locations `json:"raw_locations"`
	FullText           string    `json:"full_text"`
	Timestamp          *string   `json:"timestamp,omitempty"`
	ExtraParameters    *string   `json:"extra_parameters,omitempty"`
	Channel            []string  `json:"channel,omitempty"`
	Reason             []string  `json:"reason,omitempty"`
	EntryId            int64     `json:"entry_id"`
	Epoch              int64     `json:"epoch"`
	IsLocationVerified bool      `json:"is_location_verified"`
	IsNeedVerified     bool      `json:"is_need_verified"`
	IsDeleted          bool      `json:"is_deleted"`
	Needs              []Need    `json:"needs,omitempty"`
}
