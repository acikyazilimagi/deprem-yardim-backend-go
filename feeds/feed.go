package feeds

type Raw struct {
	Channel    string `json:"channel"`
	FullText   string `json:"full_text"`
	ID         int64  `json:"id"`
	IsResolved bool   `json:"is_resolved"`
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
	Raw              Raw       `json:"raw"`
	ViewPort         ViewPort  `json:"view_port"`
}

type Response struct {
	Count   int      `json:"count"`
	Results []Result `json:"results"`
}
