package model

type Relationship struct {
	Data RelationshipData `json:"data"`
}

type RelationshipData struct {
	GUID string `json:"guid"`
}
