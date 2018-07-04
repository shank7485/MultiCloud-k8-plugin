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
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	pkgerrors "github.com/pkg/errors"
	appsV1 "k8s.io/api/apps/v1"
	// coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/shank7485/k8-plugin-multicloud/krd"
	"github.com/shank7485/k8-plugin-multicloud/utils"
)

// VNFInstanceService communicates the actions to Kubernetes deployment
type VNFInstanceService struct {
	Client VNFInstanceClientInterface
}

// VNFInstanceClientInterface has methods to work with VNF Instance resources.
// This interface's signatures matches the methods in the Client struct in krd
// package. This is done so that we can use the Client inside the VNFInstanceService
// above.
type VNFInstanceClientInterface interface {
	CreateDeployment(deployment *appsV1.Deployment) (string, error)
	ListDeployment(limit int64) (*[]string, error)
	UpdateDeployment(deployment *appsV1.Deployment) error
	DeleteDeployment(name string) error
	GetDeployment(name string) (string, error)
}

// NewVNFInstanceService creates a client that comunicates with a Kuberentes Cluster
func NewVNFInstanceService(kubeConfigPath string) (*VNFInstanceService, error) {
	client, err := GetVNFClient(kubeConfigPath)
	if err != nil {
		return nil, err
	}
	return &VNFInstanceService{
		Client: client,
	}, nil
}

// GetVNFClient retrieve the client used to communicate with a Kubernetes Cluster
var GetVNFClient = func(kubeConfigPath string) (VNFInstanceClientInterface, error) {
	var client VNFInstanceClientInterface

	client, err := krd.NewClient(kubeConfigPath)
	if err != nil {
		return nil, err
	}
	return client, err
}

func validateBody(body interface{}) error {
	switch b := body.(type) {
	case CreateVnfRequest:
		if b.CsarID == "" || b.CsarURL == "" || b.Name == "" {
			werr := pkgerrors.Wrap(errors.New("Invalid Data in POST request"), "CreateVnfRequest bad request")
			return werr
		}
	case UpdateVnfRequest:
		if b.CsarID == "" || b.CsarURL == "" || b.Name == "" {
			werr := pkgerrors.Wrap(errors.New("Invalid Data in PUT request"), "UpdateVnfRequest bad request")
			return werr
		}
	}
	return nil
}

// CreateHandler is the POST method creates a new VNF instance resource.
func (s *VNFInstanceService) CreateHandler(w http.ResponseWriter, r *http.Request) {
	var resource CreateVnfRequest

	if r.Body == nil {
		http.Error(w, "Body empty", http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&resource)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err = validateBody(resource)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Not using "_" since only "." and "-" are allowed.
	uuidName := resource.Name + "." + string(uuid.NewUUID())

	// Persist in AAI database.
	log.Println(uuidName)

	utils.CSAR = &utils.CSARFile{} // This is ugly. Move things to create better mocks.
	kubeData, err := utils.CreateKubeObjectsFromCSAR(resource.CsarID, resource.CsarURL)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Get Deployment information error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}
	// Kubernetes Identifies resources by names. The UID setting doesn't seem to the primary ID.
	// deployment.UID = types.UID(resource.CsarID) + types.UID("_") + uuid
	if kubeData.Deployment == nil {
		werr := pkgerrors.Wrap(err, "Create VNF deployment error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}
	kubeData.Deployment.Name = uuidName

	name, err := s.Client.CreateDeployment(kubeData.Deployment)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Create VNF deployment error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	resp := CreateVnfResponse{
		DeploymentID: name,
		Name:         resource.Name,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Parsing output of new VNF error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
	}
}

// ListHandler the existing VNF instances created in a given Kubernetes cluster
func (s *VNFInstanceService) ListHandler(w http.ResponseWriter, r *http.Request) {
	limit := int64(10) // TODO (electrocucaracha): export this as configuration value

	deployments, err := s.Client.ListDeployment(limit)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Get VNF list error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	if deployments == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	resp := ListVnfsResponse{
		VNFs: *deployments,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Parsing output VNF list error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
	}

}

// DeleteHandler method terminates an individual VNF instance.
func (s *VNFInstanceService) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	err := s.Client.DeleteDeployment(vars["vnfInstanceId"])
	if err != nil {
		// TODO (electrocucaracha): Determines the existence of the resource
		werr := pkgerrors.Wrap(err, "Delete VNF error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
}

// UpdateHandler method to update a VNF instance.
func (s *VNFInstanceService) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["vnfInstanceId"]

	var resource UpdateVnfRequest

	if r.Body == nil {
		http.Error(w, "Body empty", http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&resource)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	utils.CSAR = &utils.CSARFile{}
	kubeData, err := utils.CreateKubeObjectsFromCSAR(resource.CsarID, resource.CsarURL)
	
	if kubeData.Deployment == nil {
		werr := pkgerrors.Wrap(err, "Update VNF deployment error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}
	kubeData.Deployment.SetUID(types.UID(id))

	if err != nil {
		werr := pkgerrors.Wrap(err, "Get Deployment information error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	err = s.Client.UpdateDeployment(kubeData.Deployment)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Update VNF error")

		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	resp := UpdateVnfResponse{
		DeploymentID: id,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Parsing output of new VNF error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
	}
}

// GetHandler retrieves information about a VNF instance by reading an individual VNF instance resource.
func (s *VNFInstanceService) GetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	name, err := s.Client.GetDeployment(vars["vnfInstanceId"])
	if err != nil {
		werr := pkgerrors.Wrap(err, "Get VNF error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}
	if name == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	resp := GeneralResponse{
		Response: "Got Deployment:" + name,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Parsing output of new VNF error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
	}
}
