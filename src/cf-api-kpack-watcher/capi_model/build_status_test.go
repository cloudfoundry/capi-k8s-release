package capi_model

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/sclevine/spec"
)

func TestBuildModel(t *testing.T) {
	spec.Run(t, "TestBuildModel", func(t *testing.T, when spec.G, it spec.S) {
		it.Before(func() {
			RegisterTestingT(t)
			format.TruncatedDiff = false
		})

		// when("constructing a build model request to send to the cloud controller", func() {
		// 	var build Build

		// 	it.Before(func() {
		// 		build = NewBuild(&kpack_build.Build{
		// 			Status: kpack_build.BuildStatus{
		// 				LatestImage: "foobar:latest",
		// 			},
		// 		})
		// 	})

		// 	it("constructs a build with all of the required lifecycle data", func() {
		// 		Expect(build.Lifecycle.Data.ProcessTypes).To(HaveKeyWithValue("rake", "bundle exec rake"))
		// 		Expect(build.Lifecycle.Data.ProcessTypes).To(HaveKeyWithValue("web", "bundle exec rackup config.ru -p $PORT"))
		// 	})
		// })

		when("serializing a build model object into JSON", func() {
			var (
				build           Build
				serializedBuild []byte
			)

			it.Before(func() {
				build = Build{
					State: BuildStagedState,
					Lifecycle: Lifecycle{
						Type: KpackLifecycleType,
						Data: LifecycleData{
							Image: "some-image-ref:tag",
							ProcessTypes: map[string]string{
								"rake": "bundle exec rake",
								"web":  "bundle exec rackup config.ru -p $PORT",
							},
						},
					},
				}
				serializedBuild = build.ToJSON()
			})

			it("serializes into the expected JSON payload", func() {
				Expect(string(serializedBuild)).To(Equal(`{"state":"STAGED","error":"","lifecycle":{"type":"kpack","data":{"image":"some-image-ref:tag","processTypes":{"rake":"bundle exec rake","web":"bundle exec rackup config.ru -p $PORT"}}}}`))
			})
		})
	})
}
