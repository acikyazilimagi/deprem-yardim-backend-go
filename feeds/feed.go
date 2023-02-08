package feeds

type Feed struct {
	ID         int64  `json:"id"`
	FullText   string `json:"full_text"`
	IsResolved bool   `json:"is_resolved"`
	Channel    string `json:"channel"`
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
