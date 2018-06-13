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

	pkgerrors "github.com/pkg/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"

	client "github.com/shank7485/k8-plugin-multicloud/clientConfig"
	"github.com/shank7485/k8-plugin-multicloud/utils"
)

// VNFInstanceService communicates the actions to Kubernetes deployment
type VNFInstanceService struct {
	Client *kubernetes.Clientset
}

// NewVNFInstanceService creates a client that comunicates with a Kuberentes Cluster
func NewVNFInstanceService(kubeConfigPath string) (*VNFInstanceService, error) {
	client, err := client.InitiateK8Client(kubeConfigPath)
	if err != nil {
		return nil, err
	}
	vnfService := &VNFInstanceService{
		Client: client,
	}
	return vnfService, nil
}

// CreateVNF creates a VNF Instance based on the Resquest
func (s *VNFInstanceService) CreateVNF(w http.ResponseWriter, r *http.Request) {
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

	result, err := s.Client.AppsV1().Deployments("default").Create(deploymentStruct)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Create VNF error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
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
		werr := pkgerrors.Wrap(err, "Parsing output of new VNF error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
	}
}

// ListVNF lists the existing VNF instances created in a given Kubernetes cluster
func (s *VNFInstanceService) ListVNF(w http.ResponseWriter, r *http.Request) {
	list, err := s.Client.AppsV1().Deployments("default").List(metaV1.ListOptions{})
	if err != nil {
		werr := pkgerrors.Wrap(err, "Get VNF list error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	resp := GeneralResponse{
		Response: "Listing:",
	}

	for _, d := range list.Items {
		log.Println(d.Name)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Parsing output VNF list error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
	}

}

// func DeleteVNF(w http.ResponseWriter, r *http.Request){
// 	deletePolicy := metav1.DeletePropagationForeground
// 	err := d.DeploymentsClient.Delete("demo-deployment", &metav1.DeleteOptions{
// 		PropagationPolicy: &deletePolicy,
// 	})
// }
