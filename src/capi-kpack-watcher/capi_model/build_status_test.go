package capi_model

import (
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildToJSON(t *testing.T) {
	spec.Run(t, "TestBuildToJSON", func(t *testing.T, when spec.G, it spec.S) {
		var build = Build{
			State: BuildStagedState,
			Lifecycle: Lifecycle{
				Type: KpackLifecycleType,
				Data: LifecycleData{
					Image: "some-image-ref:tag",
				},
			},
		}

		it.Before(func() {

		})

		it("converts to json", func() {
			result := build.ToJSON()
			expected := `{"state":"STAGED","error":"","lifecycle":{"type":"kpack","data":{"image":"some-image-ref:tag"}}}`

			assert.Equal(t, expected, string(result))
		})
	})

}
