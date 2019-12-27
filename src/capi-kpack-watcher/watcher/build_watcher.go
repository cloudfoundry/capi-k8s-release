package watcher

import (
	"log"

	"capi_kpack_watcher/capi"
	"capi_kpack_watcher/model"

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

		if err := bw.client.PATCHBuild(guid, model.BuildStatus{State: buildStagedState}); err != nil {
			log.Fatalf("Failed to send request: %v\n", err)
		}
	} else if status.GetCondition("Succeeded").IsFalse() {
		labels := build.GetLabels()
		guid := labels[buildGUIDLabel]

		// TODO pass real error message from pod creation when Kpack surfaces this
		if err := bw.client.PATCHBuild(guid, model.BuildStatus{State: buildFailedState, Error: "Kpack build failed."}); err != nil {
			log.Fatalf("Failed to send request: %v\n", err)
		}
	}
}

// NewBuildWatcher initializes a Watcher that watches for Builds in Kpack.
func NewBuildWatcher(clientset *kpackclient.Clientset) Watcher {
	factory := kpackinformer.NewSharedInformerFactory(clientset, 0)

	bw := &buildWatcher{
		client: capi.NewCAPIClient(),
		informer: factory.Build().V1alpha1().Builds().Informer(),
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

	// Below are Kubernetes-internal objects for creating Kubernetes Informers.
	// They are in this struct to abstract away the Informer boilerplate.
	informer cache.SharedIndexInformer
}

// TODO: getPodLogs needs permission to talk to the Kubernetes master for pod logs.
//  Currently, we only mount kpack permissions into this watcher.
//func getPodLogs(podName string) ([]byte, error) {
//	config, err := rest.InClusterConfig()
//	if err != nil {
//		panic(err.Error())
//	}
//
//	clientset, err := kubernetes.NewForConfig(config)
//	if err != nil {
//		panic(err.Error())
//	}
// // kubectl log cmd: k logs build-pod -n cf-workloads -c={failedStep} --ignore-errors
//	return clientset.CoreV1().Pods("cf-workloads").GetLogs(podName, &corev1.PodLogOptions{
//		Container: "build",
//	}).Do().Raw()
//}
