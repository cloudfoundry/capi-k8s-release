package capi_model

import (
	"testing"

	. "github.com/onsi/gomega"
	kpack_build "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
)

func TestBuildModel(t *testing.T) {
	spec.Run(t, "TestBuildModel", func(t *testing.T, when spec.G, it spec.S) {
		it.Before(func() {
			RegisterTestingT(t)
		})

		when("constructing a build model request to send to the cloud controller", func() {
			var build Build

			it.Before(func() {
				build = NewBuild(&kpack_build.Build{
					Status: kpack_build.BuildStatus{
						LatestImage: "foobar:latest",
					},
				}, nil)
			})

			it("constructs a build with all of the required lifecycle data", func() {
				Expect(build.Lifecycle.Data.ProcessTypes).To(HaveKeyWithValue("rake", "bundle exec rake"))
				Expect(build.Lifecycle.Data.ProcessTypes).To(HaveKeyWithValue("web", "bundle exec rackup config.ru -p $PORT"))
			})
		})

		when("serializing a build model object into JSON", func() {
			var build = Build{
				State: BuildStagedState,
				Lifecycle: Lifecycle{
					Type: KpackLifecycleType,
					Data: LifecycleData{
						Image: "some-image-ref:tag",
					},
				},
			}

			it("serializes into the expected JSON payload", func() {
				result := build.ToJSON()
				expected := `{"state":"STAGED","error":"","lifecycle":{"type":"kpack","data":{"image":"some-image-ref:tag","process_types":{"rake":"bundle exec rake","web":"bundle exec rackup config.ru -p $PORT"}}}}`

				assert.Equal(t, expected, string(result))
			})
		})
	})
}
