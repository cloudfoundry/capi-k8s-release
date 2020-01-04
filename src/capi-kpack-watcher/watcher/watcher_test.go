package watcher

import (
	"testing"

	"capi_kpack_watcher/mocks"
	"capi_kpack_watcher/model"

	corev1 "k8s.io/api/core/v1"
	duckv1alpha1 "knative.dev/pkg/apis/duck/v1alpha1"

	kpack "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"

	"github.com/stretchr/testify/mock"
)

func TestUpdateFunc(t *testing.T) {
	// Create mocks.
	mockCAPI := new(mocks.CAPI)
	mockKubeClient := new(mocks.Kubernetes)

	// Create our object under test. Attach mock clients.
	bw := new(buildWatcher)
	bw.client = mockCAPI
	bw.kubeClient = mockKubeClient

	const (
		guid          = "guid"
		podName       = "fake-pod-name"
		containerName = "fake-container-name"
		fakeLogs      = "ERROR:some error" // Must match regex pattern in UpdateFunc.
	)

	t.Run("successful kpack build", func(t *testing.T) {
		// Mock call to CAPI.
		mockCAPI.On("PATCHBuild", guid, successfulBuildStatus()).Return(nil)

		// Create our simulated inputs.
		oldBuild := &kpack.Build{}
		newBuild := &kpack.Build{
			Status: kpack.BuildStatus{
				PodName: podName,
			},
		}
		setGUIDOnLabel(newBuild, guid)
		markBuildSuccessful(newBuild)

		// Make call to function under test.
		bw.UpdateFunc(oldBuild, newBuild)
	})

	t.Run("failed kpack build", func(t *testing.T) {
		// Mock call to CAPI and Kubernetes.
		mockCAPI.On("PATCHBuild", guid, failedBuildStatus("some error")).Return(nil)
		mockKubeClient.On("GetContainerLogs", podName, containerName).Return([]byte(fakeLogs), nil)

		// Create our simulated inputs.
		oldBuild := &kpack.Build{}
		newBuild := &kpack.Build{
			Status: kpack.BuildStatus{
				PodName: podName,
				// Container name correspond to the Steps in kpack.
				StepsCompleted: []string{containerName},
			},
		}
		setGUIDOnLabel(newBuild, guid)
		markBuildFailed(newBuild)

		// Make call to function under test.
		bw.UpdateFunc(oldBuild, newBuild)
	})

	// Perform assertion.
	mock.AssertExpectationsForObjects(t, mockCAPI, mockKubeClient)
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
