package watcher

import (
	"capi_kpack_watcher/capi"
	"capi_kpack_watcher/kubernetes"
	"log"
	"os"

	"k8s.io/client-go/tools/cache"

	kpack "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpackclient "github.com/pivotal/kpack/pkg/client/clientset/versioned"
	kpackinformer "github.com/pivotal/kpack/pkg/client/informers/externalversions"

	"github.com/davecgh/go-spew/spew"
)

const AppGUIDLabel = "cloudfoundry.org/app_guid"

type ImageWatcher struct {
	// The watcher uses this kubernetes client to talk to the Kubernetes master.
	kubeClient     KubeClient
	dropletHandler DropletHandler
	logger         *log.Logger
	// Below are Kubernetes-internal objects for creating Kubernetes Informers.
	// They are in this struct to abstract away the Informer boilerplate.
	informer cache.SharedIndexInformer
}

type DropletHandler interface {
	GetCurrentDroplet(guid string) (string, error)
	CreateDropletCopy(dropletGUID, appGUID, image string) error
}

func (iw *ImageWatcher) AddFunc(obj interface{}) {
	image := obj.(*kpack.Image)

	iw.logger.Printf("[AddFunc] New Image: %s\n", image.GetName())
}

// UpdateFunc handles when Images are updated.
func (iw *ImageWatcher) UpdateFunc(oldobj, newobj interface{}) {
	image := newobj.(*kpack.Image)

	iw.logger.Printf(
		`[UpdateFunc] Update to Image: %s
latest image sha: %s`, image.GetName(), spew.Sdump(image.Status.LatestImage))
	updatedImage := image.Status.LatestImage

    statefulSets, err := iw.kubeClient.GetStatefulSet(AppGUIDLabel,
    	image.GetLabels()["cloudfoundry.org/app_guid"], "cf-workloads")
    if err!=nil || len(statefulSets.Items)<1 {
   	  panic(err)
    }

    newStatefulSet := statefulSets.Items[0]
    ctrs := newStatefulSet.Spec.Template.Spec.Containers
	for i , c := range ctrs {
		if c.Name == "opi" {
			if c.Image!=updatedImage {
				ctrs[i].Image = updatedImage
				//iw.handleUpdate(image)
			}
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
		dropletHandler: capi.NewCAPIClient(),
		kubeClient:     kubernetes.NewInClusterClient(),
		logger:         log.New(os.Stdout, "ImageWatcher", log.LstdFlags),
		informer:       factory.Build().V1alpha1().Images().Informer(),
	}

	// TODO: ignore added builds at watcher startup
	iw.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    iw.AddFunc,
		UpdateFunc: iw.UpdateFunc,
	})

	return iw
}

// Run runs the informer and begins watching for Images. This can be stopped by
// sending to the stopped channel.
func (iw *ImageWatcher) Run() {
	stopper := make(chan struct{})
	defer close(stopper)

	iw.informer.Run(stopper)
}

func (iw *ImageWatcher) handleUpdate(image *kpack.Image) {
	labels := image.GetLabels()
	appGUID := labels[AppGUIDLabel]
	iw.logger.Printf(`APP GUID: %s`, appGUID)
    dropletGUID, err := iw.dropletHandler.GetCurrentDroplet(appGUID)
    if err != nil {
		log.Fatalf("[GetDropletFunc] Failed to send request: %v\n", err)
	}

	//create new droplet w updated image reference and relationship with app guid
	iw.dropletHandler.CreateDropletCopy(dropletGUID, appGUID, image.Status.LatestImage)
	//start rolling deploy with  new droplet guid (includes setting current droplet)
}


