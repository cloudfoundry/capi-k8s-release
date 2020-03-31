package watcher

import (
	"capi_kpack_watcher/kubernetes"
	"log"
	"os"

	"k8s.io/client-go/tools/cache"

	kpack "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpackclient "github.com/pivotal/kpack/pkg/client/clientset/versioned"
	kpackinformer "github.com/pivotal/kpack/pkg/client/informers/externalversions"

	"github.com/davecgh/go-spew/spew"
)


func (iw *ImageWatcher) AddFunc(obj interface{}) {
	image := obj.(*kpack.Image)

	iw.logger.Printf("[AddFunc] New Image: %s\n", image.GetName())
}

// UpdateFunc handles when Images are updated.
func (iw *ImageWatcher) UpdateFunc(oldobj, newobj interface{}) {
	iw.logger.Printf(`OLD IMAGE: %v+`, oldobj.(*kpack.Image))
	iw.logger.Printf(`NEW IMAGE: %v+`, newobj.(*kpack.Image))
	image := newobj.(*kpack.Image)

	iw.logger.Printf(
		`[UpdateFunc] Update to Image: %s
latest image sha: %s`, image.GetName(), spew.Sdump(image.Status.LatestImage))
	updatedImage := image.Status.LatestImage

    statefulSets, err := iw.kubeClient.GetStatefulSet("cloudfoundry.org/app_guid",
    	image.GetLabels()["cloudfoundry.org/app_guid"], "cf-workloads")
    if err!=nil || len(statefulSets.Items)<1 {
   	  panic(err)
    }

    newStatefulSet := statefulSets.Items[0]
	for _ , c := range newStatefulSet.Spec.Template.Spec.Containers {
		if c.Name == "opi" {
			c.Image = updatedImage
		}
	}
	_ , err = iw.kubeClient.UpdateStatefulSet(&newStatefulSet, "cf-workloads")
	if err!=nil{
		panic(err)
	}
}

func NewImageWatcher(c kpackclient.Interface) *ImageWatcher {
	factory := kpackinformer.NewSharedInformerFactory(c, 0)

	iw := &ImageWatcher{
		kubeClient:   kubernetes.NewInClusterClient(),
		logger: log.New(os.Stdout, "iwPIYALIiw", log.LstdFlags),
		informer:     factory.Build().V1alpha1().Images().Informer(),
	}

	// TODO: ignore added builds at watcher startup
	iw.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    iw.AddFunc,
		UpdateFunc: iw.UpdateFunc,
	})
	iw.logger.Printf(`Image watcher made!!!!!!!`)
	return iw
}

// Run runs the informer and begins watching for Images. This can be stopped by
// sending to the stopped channel.
func (iw *ImageWatcher) Run() {
	stopper := make(chan struct{})
	defer close(stopper)
    iw.logger.Printf(`RUNNNNNN`)
	iw.informer.Run(stopper)
}

type ImageWatcher struct {
	// The watcher uses this kubernetes client to talk to the Kubernetes master.
	kubeClient KubeClient
    logger *log.Logger
	// Below are Kubernetes-internal objects for creating Kubernetes Informers.
	// They are in this struct to abstract away the Informer boilerplate.
	informer cache.SharedIndexInformer
}

