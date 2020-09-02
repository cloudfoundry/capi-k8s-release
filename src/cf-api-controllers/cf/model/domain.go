package model

type Domain struct {
	GUID     string `json:"guid"`
	Name     string `json:"name"`
	Internal bool   `json:"internal"`
}
