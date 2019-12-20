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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	// The import below is needed to authenticate with GCP.

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	kpack "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	kpackclient "github.com/pivotal/kpack/pkg/client/clientset/versioned"
	kpackinformer "github.com/pivotal/kpack/pkg/client/informers/externalversions"
)

var CAPIHost string = os.Getenv("CAPI_HOST")

func main() {
	if CAPIHost == "" {
		panic("CAPI_HOST environment variable must be set")
	}

	// TODO: kubeconfig is really meant for running locally
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

	// TODO: ignore added builds at watcher startup
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			build := obj.(*kpack.Build)

			log.Printf("New Build added: %s\n\n", build.GetName())
		},
		UpdateFunc: update,
	})

	log.Printf("Watcher initialized. Listening...\n")

	informer.Run(stopper)
}

func update(oldobj, newobj interface{}) {
	build := newobj.(*kpack.Build)
	status := build.Status

	log.Printf("Update to Build: %s\n", build.GetName())
	log.Printf("  status: %+v\n", status.Status)
	log.Printf("  steps : %+v\n", status.StepsCompleted)
	log.Printf("  conditions: %+v\n\n", status.Conditions)

	if status.GetCondition("Succeeded").IsTrue() {
		labels := build.GetLabels()
		guid := labels["cloudfoundry.org/build_guid"]

		payload := struct {
			State string `json:"state"`
		}{
			State: "STAGED",
		}
		jsonPayload, _ := json.Marshal(&payload)

		uri := fmt.Sprintf("https://api.%s/v3/internal/builds/%s", CAPIHost, guid)
		req, err := http.NewRequest(http.MethodPatch, uri, bytes.NewReader(jsonPayload))
		if err != nil {
			log.Printf("Failed to create request: %v\n", err)
		}

		req.Header.Add("Content-Type", "application/json")

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}

		log.Printf("  sent payload: %s\n", jsonPayload)

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Failed to send request: %v\n", err)
		}

		log.Printf("CAPI repsonse: %v\n", resp)
	}
}
