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
	"os/signal"
	"syscall"

	"capi_kpack_watcher/watcher"

	// The import below is needed to authenticate with GCP.

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"

	kpackclient "github.com/pivotal/kpack/pkg/client/clientset/versioned"
)

func main() {
	if os.Getenv("CAPI_HOST") == "" {
		panic("CAPI_HOST environment variable must be set")
	}

	// This environment variable is useful if running locally such as on
	// minikube. Set it if you want to do that. Otherwise, the Kubernetes
	// libraries try to use an incluster config, so it does not need to be set.
	kubeconfig := os.Getenv("KUBECONFIG")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kpackclient.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	log.Printf("Watcher initialized. Listening...\n")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go watcher.NewBuildWatcher(clientset).Run()
	go watcher.NewImageWatcher(clientset).Run()
	select {
	  case <-sigs:
	  	log.Printf("closing")
	  	close(sigs)
	}
}
