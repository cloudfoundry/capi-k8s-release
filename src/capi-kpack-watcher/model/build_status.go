package model

import "encoding/json"

// BuildStatus represents the payload that will be sent to CAPI when a kpack
// Build has been updated.
type BuildStatus struct {
	State string `json:"state"`
	Error string `json:"error"`
}

func (b *BuildStatus) ToJSON() []byte {
	j, _ := json.Marshal(b)
	return j
}

