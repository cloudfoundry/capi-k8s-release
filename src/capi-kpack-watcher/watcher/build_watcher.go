package watcher

import (
	"log"

	"capi_kpack_watcher/capi"
	"capi_kpack_watcher/kubernetes"
	"capi_kpack_watcher/model"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"

	kpack "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpackclient "github.com/pivotal/kpack/pkg/client/clientset/versioned"
	kpackinformer "github.com/pivotal/kpack/pkg/client/informers/externalversions"

	"github.com/davecgh/go-spew/spew"
)

const buildGUIDLabel = "cloudfoundry.org/build_guid"
const buildStagedState = "STAGED"
const buildFailedState = "FAILED"

// AddFunc handles when new Builds are detected.
func (bw *buildWatcher) AddFunc(obj interface{}) {
	build := obj.(*kpack.Build)

	log.Printf("[AddFunc] New Build: %s\n", build.GetName())
}

// UpdateFunc handles when Builds are updated.
func (bw *buildWatcher) UpdateFunc(oldobj, newobj interface{}) {
	build := newobj.(*kpack.Build)
	status := build.Status

	log.Printf(
		`[UpdateFunc] Update to Build: %s
status: %s
steps:  %+v

`, build.GetName(), spew.Sdump(status.Status), status.StepsCompleted)

	if status.GetCondition("Succeeded").IsTrue() {
		labels := build.GetLabels()
		guid := labels[buildGUIDLabel]

		model := model.BuildStatus{
			State: buildStagedState,
		}

		if err := bw.client.PATCHBuild(guid, model); err != nil {
			log.Fatalf("[UpdateFunc] Failed to send request: %v\n", err)
		}
	} else if status.GetCondition("Succeeded").IsFalse() {
		labels := build.GetLabels()
		guid := labels[buildGUIDLabel]

		model := model.BuildStatus{
			State: buildFailedState,
		}

		// Retrieve the last container's logs. In kpack, the steps correspond
		// to container names, so we want the last container's logs.
		container := status.StepsCompleted[len(status.StepsCompleted)-1]
		logs, err := bw.getContainerLogs(status.PodName, container)
		if err != nil {
			log.Printf("[UpdateFunc] Failed to get pod logs: %v\n", err)

			model.Error = "Kpack build failed"
		} else {
			model.Error = string(logs)
		}

		if err := bw.client.PATCHBuild(guid, model); err != nil {
			log.Fatalf("[UpdateFunc] Failed to send request: %v\n", err)
		}
	}
}

// NewBuildWatcher initializes a Watcher that watches for Builds in Kpack.
func NewBuildWatcher(clientset *kpackclient.Clientset) Watcher {
	factory := kpackinformer.NewSharedInformerFactory(clientset, 0)

	bw := &buildWatcher{
		client:     capi.NewCAPIClient(),
		kubeClient: kubernetes.NewInClusterClient(),
		informer:   factory.Build().V1alpha1().Builds().Informer(),
	}

	// TODO: ignore added builds at watcher startup
	bw.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    bw.AddFunc,
		UpdateFunc: bw.UpdateFunc,
	})

	return bw
}

// Run runs the informer and begins watching for Builds. This can be stopped by
// sending to the stopped channel.
func (bw *buildWatcher) Run() {
	stopper := make(chan struct{})
	defer close(stopper)

	bw.informer.Run(stopper)
}

type buildWatcher struct {
	client capi.CAPI // The watcher uses this client to talk to CAPI.

	// The watcher uses this kubernetes client to talk to the Kubernetes master.
	kubeClient kubernetes.InClusterClient

	// Below are Kubernetes-internal objects for creating Kubernetes Informers.
	// They are in this struct to abstract away the Informer boilerplate.
	informer cache.SharedIndexInformer
}

func (bw *buildWatcher) getContainerLogs(podName, containerName string) ([]byte, error) {
	// kubectl log cmd: k logs build-pod -n cf-workloads -c={failedStep} --ignore-errors
	return bw.kubeClient.CoreV1().Pods("cf-workloads").GetLogs(podName, &v1.PodLogOptions{
		Container: containerName,
	}).Do().Raw()
}
