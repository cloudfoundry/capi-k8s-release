package model

type Destination struct {
	GUID   string         `json:"guid"`
	App    DestinationApp `json:"app"`
	Weight *int           `json:"weight"`
	Port   int            `json:"port"`
}

type DestinationApp struct {
	GUID    string             `json:"guid"`
	Process DestinationProcess `json:"process"`
}

type DestinationProcess struct {
	Type string `json:"type"`
}
