package watcher

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	"capi_kpack_watcher/capi_model"
	"capi_kpack_watcher/image_registry"

	imageV1 "github.com/google/go-containerregistry/pkg/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/buildpacks/lifecycle"
	"github.com/davecgh/go-spew/spew"
	kpack "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	conditions "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
)

const BuildGUIDLabel = "cloudfoundry.org/build_guid"

type BuildWatcher struct {
	buildUpdater BuildUpdater // The watcher uses this client to talk to CAPI.

	// The watcher uses this kubernetes client to talk to the Kubernetes master.
	kubeClient KubeClient

	// Below are Kubernetes-internal objects for creating Kubernetes Informers.
	// They are in this struct to abstract away the Informer boilerplate.
	informer cache.SharedIndexInformer

	imageConfigFetcher image_registry.ImageConfigFetcher
}

func NewBuildWatcher(informer cache.SharedIndexInformer, buildUpdater BuildUpdater, kubeClient KubeClient, imageConfigFetcher image_registry.ImageConfigFetcher) *BuildWatcher {
	bw := &BuildWatcher{
		buildUpdater:       buildUpdater,
		kubeClient:         kubeClient,
		informer:           informer,
		imageConfigFetcher: imageConfigFetcher,
	}

	return bw
}

// Run runs the informer and begins watching for Builds. This can be stopped by
// sending to the stopped channel.
func (bw *BuildWatcher) Run() {
	// TODO: ignore added builds at watcher startup
	bw.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    bw.AddFunc,
		UpdateFunc: bw.UpdateFunc,
	})

	stopper := make(chan struct{})
	defer close(stopper)

	bw.informer.Run(stopper)
}

// AddFunc handles when new Builds are detected.
func (bw *BuildWatcher) AddFunc(obj interface{}) {
	build := obj.(*kpack.Build)

	log.Printf("[AddFunc] New Build: %s\n", build.GetName())
}

// UpdateFunc handles when Builds are updated.
func (bw *BuildWatcher) UpdateFunc(oldobj, newobj interface{}) {
	build := newobj.(*kpack.Build)

	log.Printf(
		`[UpdateFunc] Update to Build: %s
status: %s
steps:  %+v

`, build.GetName(), spew.Sdump(build.Status.Status), build.Status.StepsCompleted)

	if bw.isBuildGUIDMissing(build) {
		return
	}

	c := build.Status.GetCondition("Succeeded")
	if c.IsTrue() {
		bw.handleSuccessfulBuild(build)
	} else if c.IsFalse() {
		bw.handleFailedBuild(build)
	} // c.isUnknown() is also available for pending builds
}

func (bw *BuildWatcher) isBuildGUIDMissing(build *kpack.Build) bool {
	labels := build.GetLabels()
	if labels == nil {
		return true
	} else if _, ok := labels[BuildGUIDLabel]; !ok {
		return true
	}

	return false
}

func (bw *BuildWatcher) handleSuccessfulBuild(build *kpack.Build) {
	labels := build.GetLabels()
	guid := labels[BuildGUIDLabel]

	capiBuild := capi_model.NewBuild(build)

	imageConfig, err := bw.imageConfigFetcher.FetchImageConfig(build.Status.LatestImage, build.Spec.ServiceAccount, build.Namespace)
	if err != nil {
		log.Printf("[UpdateFunc] Failed to fetch image config: %v\n", err)
		return
	}
	processTypes, err := bw.extractProcessTypes(imageConfig)
	if err != nil {
		log.Printf("[UpdateFunc] Failed to parse process types info from image config: %v\n", err)
		return
	}
	capiBuild.Lifecycle.Data.ProcessTypes = processTypes

	if err := bw.buildUpdater.UpdateBuild(guid, capiBuild); err != nil {
		log.Printf("[UpdateFunc] Failed to send request: %v\n", err)
	}
}

func (bw *BuildWatcher) extractProcessTypes(config *imageV1.Config) (map[string]string, error) {
	var buildMetadata lifecycle.BuildMetadata
	if err := json.Unmarshal([]byte(config.Labels[lifecycle.BuildMetadataLabel]), &buildMetadata); err != nil {
		return nil, err
	}

	ret := make(map[string]string)
	for _, process := range buildMetadata.Processes {
		ret[process.Type] = process.Command
	}
	return ret, nil
}

func (bw *BuildWatcher) handleFailedBuild(build *kpack.Build) {
	labels := build.GetLabels()
	guid := labels[BuildGUIDLabel]
	capi_model := capi_model.Build{
		State: capi_model.BuildFailedState,
		Lifecycle: capi_model.Lifecycle{
			Type: capi_model.KpackLifecycleType,
			Data: capi_model.LifecycleData{
				Image: build.Status.LatestImage,
			},
		},
	}

	status := build.Status
	condition := status.GetCondition(conditions.ConditionSucceeded)

	// Retrieve the last container's logs. In kpack, the steps correspond
	// to container names, so we want the last container's logs.
	if len(status.StepsCompleted) > 0 {
		container := status.StepsCompleted[len(status.StepsCompleted)-1]

		logs, err := bw.kubeClient.GetContainerLogs(status.PodName, container)
		if err != nil {
			log.Printf("[UpdateFunc] Failed to get pod logs: %v\n", err)

			capi_model.Error = fmt.Sprintf("Kpack build failed. build failure message: '%s'. err: '%v'", condition.Message, err)
		} else {
			// Take the first word character to the end of the line to avoid ANSI color codes
			regex := regexp.MustCompile(`ERROR:[^\w\[]*(\[[0-9]+m)?(\w[^\n]*)`)
			capi_model.Error = string(regex.FindSubmatch(logs)[2])
		}
	} else {
		capi_model.Error = fmt.Sprintf("Kpack build failed. build failure message: '%s'.", condition.Message)
	}

	if err := bw.buildUpdater.UpdateBuild(guid, capi_model); err != nil {
		log.Fatalf("[UpdateFunc] Failed to send request: %v\n", err)
	}
}
