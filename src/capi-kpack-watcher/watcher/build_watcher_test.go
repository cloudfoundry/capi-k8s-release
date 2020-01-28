package watcher

import (
	"testing"

	"capi_kpack_watcher/capi_model"
	"capi_kpack_watcher/watcher/mocks"

	kpack "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpackcore "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/mock"
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
			buildUpdater   *mocks.BuildUpdater
			mockKubeClient *mocks.KubeClient
			bw             *BuildWatcher
		)

		it.Before(func() {
			buildUpdater = new(mocks.BuildUpdater)
			mockKubeClient = new(mocks.KubeClient)

			bw = new(BuildWatcher)
			bw.buildUpdater = buildUpdater
			bw.kubeClient = mockKubeClient
		})

		it.After(func() {
			mock.AssertExpectationsForObjects(t,
				buildUpdater,
				mockKubeClient,
			)
		})

		when("build is successful", func() {
			var newSuccessfulBuild *kpack.Build
			it.Before(func() {
				newSuccessfulBuild = &kpack.Build{
					Status: kpack.BuildStatus{
						LatestImage: imageRef,
					},
				}
				buildUpdater.
					On("UpdateBuild", guid, capi_model.NewBuild(newSuccessfulBuild)).
					Return(nil).
					Arguments.Assert(t, guid, capi_model.NewBuild(newSuccessfulBuild))
			})

			it("updates capi with the success status", func() {
				oldBuild := &kpack.Build{}
				setGUIDOnLabel(newSuccessfulBuild, guid)
				markBuildSuccessful(newSuccessfulBuild)

				bw.UpdateFunc(oldBuild, newSuccessfulBuild)
			})
		})

		when("build fails", func() {
			it.Before(func() {
				buildUpdater.On("UpdateBuild", guid, failedBuild("some error")).Return(nil)
				mockKubeClient.On("GetContainerLogs", podName, containerName).Return([]byte(fakeLogs), nil)
			})

			it("updates capi with the failed state and error message", func() {
				oldBuild := &kpack.Build{}
				newBuild := &kpack.Build{
					Status: kpack.BuildStatus{
						PodName:        podName,
						StepsCompleted: []string{containerName},
					},
				}
				setGUIDOnLabel(newBuild, guid)
				markBuildFailed(newBuild)

				bw.UpdateFunc(oldBuild, newBuild)
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

				buildUpdater.AssertNotCalled(t, "UpdateBuild")
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
