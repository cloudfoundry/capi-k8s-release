package watcher

import (
	"testing"

	"capi_kpack_watcher/model"
	"capi_kpack_watcher/watcher/mocks"

	kpack "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	duckv1alpha1 "knative.dev/pkg/apis/duck/v1alpha1"
)

func TestUpdateFunc(t *testing.T) {
	spec.Run(t, "TestUpdateFunc", func(t *testing.T, when spec.G, it spec.S) {
		const (
			guid          = "guid"
			podName       = "fake-pod-name"
			containerName = "fake-container-name"
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
			it.Before(func() {
				buildUpdater.On("UpdateBuild", guid, successfulBuildStatus()).Return(nil)
			})

			it("updates capi with the success status", func() {
				oldBuild := &kpack.Build{}
				newBuild := &kpack.Build{
					Status: kpack.BuildStatus{
						PodName: podName,
					},
				}
				setGUIDOnLabel(newBuild, guid)
				markBuildSuccessful(newBuild)

				bw.UpdateFunc(oldBuild, newBuild)
			})
		})

		when("build fails", func() {
			it.Before(func() {
				buildUpdater.On("UpdateBuild", guid, failedBuildStatus("some error")).Return(nil)
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

	labels[buildGUIDLabel] = guid

	b.SetLabels(labels)
}

func markBuildSuccessful(b *kpack.Build) {
	b.Status.Conditions = duckv1alpha1.Conditions{
		duckv1alpha1.Condition{
			Type:   duckv1alpha1.ConditionSucceeded,
			Status: corev1.ConditionTrue,
		},
	}
}

func markBuildFailed(b *kpack.Build) {
	b.Status.Conditions = duckv1alpha1.Conditions{
		duckv1alpha1.Condition{
			Type:   duckv1alpha1.ConditionSucceeded,
			Status: corev1.ConditionFalse,
		},
	}
}

func successfulBuildStatus() model.BuildStatus {
	return model.BuildStatus{
		State: buildStagedState,
	}
}

func failedBuildStatus(msg string) model.BuildStatus {
	return model.BuildStatus{
		State: buildFailedState,
		Error: msg,
	}
}
