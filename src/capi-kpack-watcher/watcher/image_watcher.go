package watcher

import (
	"capi_kpack_watcher/kubernetes"
	"log"

	"k8s.io/client-go/tools/cache"

	kpack "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpackclient "github.com/pivotal/kpack/pkg/client/clientset/versioned"
	kpackinformer "github.com/pivotal/kpack/pkg/client/informers/externalversions"

	"github.com/davecgh/go-spew/spew"
)


func (iw *ImageWatcher) AddFunc(obj interface{}) {
	image := obj.(*kpack.Image)

	log.Printf("[AddFunc] New Image: %s\n", image.GetName())
}

// UpdateFunc handles when Builds are updated.
func (iw *ImageWatcher) UpdateFunc(oldobj, newobj interface{}) {
	image := newobj.(*kpack.Image)

	log.Printf(
		`[UpdateFunc] Update to Image: %s
latest image sha: %s

`, image.GetName(), spew.Sdump(image.Status.LatestImage))
}

func NewImageWatcher(c kpackclient.Interface) *ImageWatcher {
	factory := kpackinformer.NewSharedInformerFactory(c, 0)

	iw := &ImageWatcher{
		kubeClient:   kubernetes.NewInClusterClient(),
		informer:     factory.Build().V1alpha1().Images().Informer(),
	}

	// TODO: ignore added builds at watcher startup
	iw.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    iw.AddFunc,
		UpdateFunc: iw.UpdateFunc,
	})

	return iw
}

// Run runs the informer and begins watching for Builds. This can be stopped by
// sending to the stopped channel.
func (iw *ImageWatcher) Run() {
	stopper := make(chan struct{})
	defer close(stopper)

	iw.informer.Run(stopper)
}

type ImageWatcher struct {
	// The watcher uses this kubernetes client to talk to the Kubernetes master.
	kubeClient KubeClient

	// Below are Kubernetes-internal objects for creating Kubernetes Informers.
	// They are in this struct to abstract away the Informer boilerplate.
	informer cache.SharedIndexInformer
}

