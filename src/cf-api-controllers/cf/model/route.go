package model

type Route struct {
	GUID          string                  `json:"guid"`
	Host          string                  `json:"host"`
	Path          string                  `json:"path"`
	URL           string                  `json:"url"`
	Destinations  []Destination           `json:"destinations"`
	Relationships map[string]Relationship `json:"relationships"`
}
