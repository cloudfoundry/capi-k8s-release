package kubernetes

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// NewInClusterClient instantiates an in-cluster client to talk to the
// Kubernetes master. An in-client client is intended to be used within a
// Kubernetes cluster. In other words, it is meant to be used inside a pod.
// This means that it can only access resources within said cluster.
func NewInClusterClient() InClusterClient {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return InClusterClient{clientset}
}

// InClusterClient embeds Clientset as a thin wrapper over the native
// Kubernetes API.
type InClusterClient struct {
	*kubernetes.Clientset
}
