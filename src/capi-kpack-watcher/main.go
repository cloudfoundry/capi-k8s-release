/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"log"
	"os"

	// The import below is needed to authenticate with GCP.
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	kpack "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpackclient "github.com/pivotal/kpack/pkg/client/clientset/versioned"
	kpackinformer "github.com/pivotal/kpack/pkg/client/informers/externalversions"
)

func main() {
	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kpackclient.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	factory := kpackinformer.NewSharedInformerFactory(clientset, 0)
	informer := factory.Build().V1alpha1().Builds().Informer()
	stopper := make(chan struct{})
	defer close(stopper)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			build := obj.(*kpack.Build)

			log.Printf("New Build added: %s\n\n", build.GetName())
		},
		UpdateFunc: func(oldobj, newobj interface{}) {
			build := newobj.(*kpack.Build)
			status := build.Status

			log.Printf("Update to Build: %s\n", build.GetName())
			log.Printf("  status: %+v\n", status.Status)
			log.Printf("  steps : %+v\n\n", status.StepsCompleted)
		},
	})

	informer.Run(stopper)
}
