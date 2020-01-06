package kubernetes

// Kubernetes provides the ability to interact with the Kubernetes master /
// API. This interface allows mocking out the interactions to the actual API.
type Kubernetes interface {
	GetContainerLogs(podName, containerName string) ([]byte, error)
}
