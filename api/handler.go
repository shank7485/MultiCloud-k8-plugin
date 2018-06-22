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
	"strings"

	"github.com/gorilla/mux"
	pkgerrors "github.com/pkg/errors"
	appsV1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
type VNFInstanceClientInterface interface {
	Create(deployment *appsV1.Deployment) (string, error)
	List(limit int64) (*[]string, error)
	Delete(name string, options *metaV1.DeleteOptions) error
	Update(deployment *appsV1.Deployment) error
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
		if b.CsarID == "" || b.CsarURL == "" || b.ID == "" {
			werr := pkgerrors.Wrap(errors.New("Invalid Data in PUT request"), "UpdateVnfRequest bad request")
			return werr
		}
	case UpdateVnfRequest:
		if b.CsarID == "" || b.CsarURL == "" || b.ID == "" {
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
	uuidName := resource.CsarID + "." + string(uuid.NewUUID())

	// Persist in AAI database.
	log.Println(uuidName)

	deployment, err := utils.GetDeploymentInfo(resource.CsarURL)
	// Kubernetes Identifies resources by names. The UID setting doesn't seem to the primary ID.
	// deployment.UID = types.UID(resource.CsarID) + types.UID("_") + uuid
	deployment.Name = uuidName

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

	resp := CreateVnfResponse{
		DeploymentID: uuidName,
		Name:         name,
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

	deployments, err := s.Client.List(limit)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Get VNF list error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	if deployments == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	resp := GeneralResponse{
		Response: "Listing:" + strings.Join(*deployments, ","),
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
	id := vars["vnfInstanceId"]

	deletePolicy := metaV1.DeletePropagationForeground
	deleteOptions := &metaV1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	err := s.Client.Delete(id, deleteOptions)
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

	deployment, err := utils.GetDeploymentInfo(resource.CsarURL)
	deployment.SetUID(types.UID(id))

	if err != nil {
		werr := pkgerrors.Wrap(err, "Get Deployment information error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	err = s.Client.Update(deployment)
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
