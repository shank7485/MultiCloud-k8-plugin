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
	appsV1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/shank7485/k8-plugin-multicloud/krd"
	"github.com/shank7485/k8-plugin-multicloud/utils"
)

// VNFInstanceService communicates the actions to Kubernetes deployment
type VNFInstanceService struct {
	Client VNFInstanceClientInterface
}

// VNFInstanceClientInterface has methods to work with VNF Instance resources.
type VNFInstanceClientInterface interface {
	Create(deployment *appsV1.Deployment) (string, error)
	List(limit int64) (*appsV1.DeploymentList, error)
}

// NewVNFInstanceService creates a client that comunicates with a Kuberentes Cluster
func NewVNFInstanceService(kubeConfigPath string) (*VNFInstanceService, error) {
	var client VNFInstanceClientInterface

	client, err := krd.NewClient(kubeConfigPath)
	if err != nil {
		return nil, err
	}
	vnfService := &VNFInstanceService{
		Client: client,
	}
	return vnfService, nil
}

// Create a VNF Instance based on the Resquest
func (s *VNFInstanceService) Create(w http.ResponseWriter, r *http.Request) {
	var resource VNFInstanceResource

	err := json.NewDecoder(r.Body).Decode(&resource)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uuid := uuid.NewUUID()
	// Persist in AAI database.
	log.Println(resource.CsarArtificateID + "_" + string(uuid))

	deployment, err := utils.GetDeploymentInfo(resource.CsarArtificateURL)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Get Deployment information error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	name, err := s.Client.Create(deployment)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Create VNF error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	resp := GeneralResponse{
		Response: "Created Deployment:" + name,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Parsing output of new VNF error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
	}
}

// List the existing VNF instances created in a given Kubernetes cluster
func (s *VNFInstanceService) List(w http.ResponseWriter, r *http.Request) {
	_, err := s.Client.List(int64(10)) // TODO (electrocucaracha): export this as configuration value
	if err != nil {
		werr := pkgerrors.Wrap(err, "Get VNF list error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	resp := GeneralResponse{
		Response: "Listing:",
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
