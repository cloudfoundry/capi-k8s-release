package model

import (
	buildv1alpha1 "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
)

// build states
const (
	BuildStagedState = "STAGED"
	BuildFailedState = "FAILED"
)
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

func NewBuildFromKpackBuild(kpackBuild *buildv1alpha1.Build) Build {
	return Build{
		State: BuildStagedState,
		Lifecycle: Lifecycle{
			Type: KpackLifecycleType,
			Data: LifecycleData{
				Image: kpackBuild.Status.LatestImage,
			},
		},
	}
}
