package kubernetes

import (
	"fmt"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetContainerLogs retrieves a specifc container's logs from a given pod.
func (c *InClusterClient) GetContainerLogs(podName, containerName string) ([]byte, error) {
	// kubectl log cmd: k logs build-pod -n cf-workloads -c={failedStep} --ignore-errors
	return c.CoreV1().Pods("cf-workloads").GetLogs(podName, &v1.PodLogOptions{
		Container: containerName,
	}).Do().Raw()
}

func (c *InClusterClient) GetStatefulSet(labelSelector, labelSelectorValue, namespace string) (*v12.StatefulSetList, error) {
	return c.AppsV1().StatefulSets(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", labelSelector, labelSelectorValue)})
}

func (c *InClusterClient) UpdateStatefulSet(statefulSet *v12.StatefulSet, namespace string) (*v12.StatefulSet, error) {
	return c.AppsV1().StatefulSets(namespace).Update(statefulSet)
}
// NewInClusterClient instantiates an in-cluster client to talk to the
// Kubernetes master. An in-cluster client is intended to be used within a
// Kubernetes cluster. In other words, it is meant to be used inside a pod.
// This means that it can only access resources within said cluster.
func NewInClusterClient() *InClusterClient {
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
	*kubernetes.Clientset
}
