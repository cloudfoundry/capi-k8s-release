package api_model

import (
	kpack_build "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
)

const BuildStagedState = "STAGED"
const BuildFailedState = "FAILED"
const KpackLifecycleType = "kpack"

// Build represents the payload that will be sent to CF API server when a kpack
// Build has been updated.
type Build struct {
	State     string    `json:"state"`
	Error     string    `json:"error"`
	Lifecycle Lifecycle `json:"lifecycle"`
}

type Lifecycle struct {
	Type string        `json:"type"`
	Data LifecycleData `json:"data"`
}
type LifecycleData struct {
	Image        string            `json:"image"`
	ProcessTypes map[string]string `json:"processTypes"`
}

func NewBuild(build *kpack_build.Build) Build {
	return Build{
		State: BuildStagedState,
		Lifecycle: Lifecycle{
			Type: KpackLifecycleType,
			Data: LifecycleData{
				Image: build.Status.LatestImage,
			},
		},
	}
}
