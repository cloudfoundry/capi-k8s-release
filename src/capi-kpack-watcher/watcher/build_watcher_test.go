package watcher

import (
	"github.com/stretchr/testify/mock"
	"testing"

	"capi_kpack_watcher/mocks"
	"capi_kpack_watcher/model"

	kpack "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	"github.com/sclevine/spec"
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
			mockCAPI       *mocks.CAPI
			mockKubeClient *mocks.Kubernetes
			bw             *buildWatcher
		)

		it.Before(func() {
			mockCAPI = new(mocks.CAPI)
			mockKubeClient = new(mocks.Kubernetes)

			bw = new(buildWatcher)
			bw.client = mockCAPI
			bw.kubeClient = mockKubeClient
		})

		it.After(func() {
			mock.AssertExpectationsForObjects(t,
				mockCAPI,
				mockKubeClient,
			)
		})

		when("build is successful", func() {
			it.Before(func() {
				mockCAPI.On("PATCHBuild", guid, successfulBuildStatus()).Return(nil)
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
				mockCAPI.On("PATCHBuild", guid, failedBuildStatus("some error")).Return(nil)
				mockKubeClient.On("GetContainerLogs", podName, containerName).Return([]byte(fakeLogs), nil)
			})

			it("updates capi with the failed state and error message", func() {
				oldBuild := &kpack.Build{}
				newBuild := &kpack.Build{
					Status: kpack.BuildStatus{
						PodName: podName,
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

				mockCAPI.AssertNotCalled(t, "PATCHBuild")
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
