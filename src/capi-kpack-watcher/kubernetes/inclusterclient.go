package kubernetes

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GetContainerLogs retrieves a specifc container's logs from a given pod.
func (c *InClusterClient) GetContainerLogs(podName, containerName string) ([]byte, error) {
	// kubectl log cmd: k logs build-pod -n cf-workloads -c={failedStep} --ignore-errors
	return c.CoreV1().Pods("cf-workloads").GetLogs(podName, &v1.PodLogOptions{
		Container: containerName,
	}).Do().Raw()
}

// NewInClusterClient instantiates an in-cluster client to talk to the
// Kubernetes master. An in-cluster client is intended to be used within a
// Kubernetes cluster. In other words, it is meant to be used inside a pod.
// This means that it can only access resources within said cluster.
func NewInClusterClient() Kubernetes {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return &InClusterClient{clientset}
}

// InClusterClient is a thin wrapper over the standard (client-go) Kubernetes
// API. This implements our Kubernetes interface.
type InClusterClient struct {
	kubernetes.Interface
}
