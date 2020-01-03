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

type MockKubeClient struct {
	mock.Mock
}

func TestUpdateFunc(t *testing.T) {
	// Create mocks for CAPI client.
	mockCAPI := new(mocks.CAPI)

	// Create our object under test. Attach mock client.
	bw := new(buildWatcher)
	bw.client = mockCAPI

	// Mock call to CAPI.
	mockCAPI.On("PATCHBuild", "guid", successfulBuildStatus()).Return(nil)

	// Create our simulated inputs.
	oldBuild := &kpack.Build{}
	newBuild := &kpack.Build{
		Status: kpack.BuildStatus{
			PodName: "fake-pod-name",
		},
	}
	setGUIDOnLabel(newBuild, "guid")
	markBuildSuccessful(newBuild)

	// Make call to function under test.
	bw.UpdateFunc(oldBuild, newBuild)

	// Perform assertion.
	mock.AssertExpectationsForObjects(t, mockCAPI)
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
