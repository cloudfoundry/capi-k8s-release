package model

// Droplet represents the payload that will be sent to CF API server when an Image
// has been rebased.
type Droplet struct {
	Image string `json:"image"`
}
