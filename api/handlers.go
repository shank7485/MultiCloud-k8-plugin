/*
Copyright 2018 Intel Corporation.
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

package api

import (
	"encoding/json"
	"log"
	"net/http"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"

	client "github.com/shank7485/k8-plugin-multicloud/clientConfig"
	"github.com/shank7485/k8-plugin-multicloud/utils"
)

var k8client *kubernetes.Clientset

// InitiateK8client creates a client that comunicates with a Kuberentes Cluster
func InitiateK8client(kubeConfigPath string) error {
	k8client, err := client.InitiateK8Client(kubeConfigPath)
	if err != nil {
		return err
	}
	return nil
}

// CreateVNF creates a VNF Instance based on the Resquest
func CreateVNF(w http.ResponseWriter, r *http.Request) {
	var body CreateVNFRequest

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uuid := uuid.NewUUID()
	// Persist in AAI database.
	log.Println(body.CsarArtificateID + "_" + string(uuid))

	rawYAMLbytes, err := utils.GetYAML(body.CsarArtificateURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	deploymentStruct, err := utils.ConvertYAMLtoStructs(rawYAMLbytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := k8client.AppsV1().Deployments("default").Create(deploymentStruct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := GeneralResponse{
		Response: "Created Deployment:" + result.GetObjectMeta().GetName(),
	}
	name := result.GetObjectMeta().GetName()
	log.Println("Created deployment: " + name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ListVNF lists the existing VNF instances created in a given Kubernetes cluster
func ListVNF(w http.ResponseWriter, r *http.Request) {
	list, err := k8client.AppsV1().Deployments("default").List(metaV1.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := GeneralResponse{
		Response: "Listing:",
	}
	log.Println("")

	for _, d := range list.Items {
		log.Println(d.Name)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

// func DeleteVNF(w http.ResponseWriter, r *http.Request){
// 	deletePolicy := metav1.DeletePropagationForeground
// 	err := d.DeploymentsClient.Delete("demo-deployment", &metav1.DeleteOptions{
// 		PropagationPolicy: &deletePolicy,
// 	})
// }
