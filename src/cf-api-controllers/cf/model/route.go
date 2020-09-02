package model

type Route struct {
	GUID          string                  `json:"guid"`
	Port          int                     `json:"port"`
	Host          string                  `json:"host"`
	Path          string                  `json:"path"`
	URL           string                  `json:"url"`
	Destinations  []Destination           `json:"destinations"`
	Relationships map[string]Relationship `json:"relationships"`
}
