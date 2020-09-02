package model

type Space struct {
	GUID          string                  `json:"guid"`
	Name          string                  `json:"name"`
	Relationships map[string]Relationship `json:"relationships"`
}
