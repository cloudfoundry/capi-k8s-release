package watcher_test

import (
	"testing"

	"capi_kpack_watcher/capi_model"
	. "capi_kpack_watcher/watcher"
	"capi_kpack_watcher/watcher/watcherfakes"

	kpack "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpackcore "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func TestUpdateFunc(t *testing.T) {
	spec.Run(t, "TestUpdateFunc", func(t *testing.T, when spec.G, it spec.S) {
		const (
			guid          = "guid"
			podName       = "fake-pod-name"
			containerName = "fake-container-name"
			imageRef      = "fake-image-ref"
			fakeLogs      = "ERROR:some error" // Must match regex pattern in UpdateFunc.
		)
		var (
			buildUpdater *watcherfakes.FakeBuildUpdater
			kubeClient   *watcherfakes.FakeKubeClient
			bw           *BuildWatcher
		)

		it.Before(func() {
			RegisterTestingT(t)

			kubeClient = &watcherfakes.FakeKubeClient{}
			buildUpdater = &watcherfakes.FakeBuildUpdater{}

			bw = NewBuildWatcher(nil, buildUpdater, kubeClient)
		})

		when("build is successful", func() {
			var newSuccessfulBuild *kpack.Build
			var oldBuild *kpack.Build

			it.Before(func() {
				newSuccessfulBuild = &kpack.Build{
					Status: kpack.BuildStatus{
						LatestImage: imageRef,
					},
				}

				oldBuild = &kpack.Build{}
				setGUIDOnLabel(newSuccessfulBuild, guid)
				markBuildSuccessful(newSuccessfulBuild)
			})

			it("updates capi with the success status", func() {
				bw.UpdateFunc(oldBuild, newSuccessfulBuild)

				Expect(buildUpdater.UpdateBuildCallCount()).To(Equal(1))
				actualGuid, actualBuild := buildUpdater.UpdateBuildArgsForCall(0)
				Expect(actualGuid).To(Equal(guid))
				Expect(actualBuild).To(Equal(capi_model.NewBuild(newSuccessfulBuild)))
			})
		})

		when("build fails", func() {
			var oldBuild *kpack.Build
			var newBuild *kpack.Build

			it.Before(func() {
				kubeClient.GetContainerLogsReturns([]byte(fakeLogs), nil)

				oldBuild = &kpack.Build{}
				newBuild = &kpack.Build{
					Status: kpack.BuildStatus{
						PodName:        podName,
						StepsCompleted: []string{containerName},
					},
				}
				setGUIDOnLabel(newBuild, guid)
				markBuildFailed(newBuild)
			})

			it("updates capi with the failed state and error message", func() {
				bw.UpdateFunc(oldBuild, newBuild)

				Expect(kubeClient.GetContainerLogsCallCount()).To(Equal(1))
				actualPodName, actualContainerName := kubeClient.GetContainerLogsArgsForCall(0)
				Expect(actualPodName).To(Equal(podName))
				Expect(actualContainerName).To(Equal(containerName))

				Expect(buildUpdater.UpdateBuildCallCount()).To(Equal(1))
				actualGuid, actualBuild := buildUpdater.UpdateBuildArgsForCall(0)
				Expect(actualGuid).To(Equal(guid))
				Expect(actualBuild).To(Equal(failedBuild("some error")))
			})
		})

		when("a build does not have a cloudfoundry.org/build_guid", func() {
			it("ignores it", func() {
				oldBuild := &kpack.Build{}
				newBuild := &kpack.Build{
					Status: kpack.BuildStatus{
						PodName: podName,
					},
				}
				newBuild.SetLabels(nil)
				markBuildSuccessful(newBuild)

				bw.UpdateFunc(oldBuild, newBuild)

				Expect(buildUpdater.UpdateBuildCallCount()).To(Equal(0))
			})
		})
	})
}

func setGUIDOnLabel(b *kpack.Build, guid string) {
	labels := b.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[BuildGUIDLabel] = guid

	b.SetLabels(labels)
}

func markBuildSuccessful(b *kpack.Build) {
	b.Status.Conditions = kpackcore.Conditions{
		kpackcore.Condition{
			Type:   kpackcore.ConditionSucceeded,
			Status: corev1.ConditionTrue,
		},
	}
}

func markBuildFailed(b *kpack.Build) {
	b.Status.Conditions = kpackcore.Conditions{
		kpackcore.Condition{
			Type:   kpackcore.ConditionSucceeded,
			Status: corev1.ConditionFalse,
		},
	}
}

func failedBuild(msg string) capi_model.Build {
	return capi_model.Build{
		State: capi_model.BuildFailedState,
		Error: msg,
	}
}
