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
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/shank7485/k8-plugin-multicloud/db"
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
	CreateDeployment(deployment *appsV1.Deployment, namespace string) (string, error)
	ListDeployment(limit int64, namespace string) (*[]string, error)
	UpdateDeployment(deployment *appsV1.Deployment, namespace string) error
	DeleteDeployment(name string, namespace string) error
	GetDeployment(name string, namespace string) (string, error)

	CreateService(service *coreV1.Service, namespace string) (string, error)
	ListService(limit int64, namespace string) (*[]string, error)
	UpdateService(service *coreV1.Service, namespace string) error
	DeleteService(name string, namespace string) error
	GetService(name string, namespace string) (string, error)

	CreateNamespace(namespace string) error
	CheckNamespace(namespace string) (bool, error)
	DeleteNamespace(namespace string) error
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
		if b.CloudRegionID == "" || b.CsarID == "" {
			werr := pkgerrors.Wrap(errors.New("Invalid/Missing Data in POST request"), "CreateVnfRequest bad request")
			return werr
		}
		if strings.Contains(b.CsarID, ":") {
			werr := pkgerrors.Wrap(errors.New("Invalid/Missing Data in POST request"), "CreateVnfRequest bad request")
			return werr
		}
	case UpdateVnfRequest:
		if b.CloudRegionID == "" || b.CsarID == "" {
			werr := pkgerrors.Wrap(errors.New("Invalid/Missing Data in PUT request"), "UpdateVnfRequest bad request")
			return werr
		}
	}
	return nil
}

// CreateHandler is the POST method creates a new VNF instance resource.
func CreateHandler(w http.ResponseWriter, r *http.Request) {
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

	// (TODO): Read kubeconfig for specific Cloud Region from local file system
	// if present or download it from AAI
	s, err := NewVNFInstanceService("../kubeconfig/config")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	present, err := s.Client.CheckNamespace(resource.Namespace)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Check namespace exists error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	if present == false {
		err = s.Client.CreateNamespace(resource.Namespace)
		if err != nil {
			werr := pkgerrors.Wrap(err, "Create new namespace error")
			http.Error(w, werr.Error(), http.StatusInternalServerError)
			return
		}
	}

	kubeData, err := utils.ReadCSARFromFileSystem(resource.CsarID)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Read Kubernetes Data information error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	// Kubernetes Identifies resources by names. The UID setting doesn't seem to the primary ID.
	// deployment.UID = types.UID(resource.CsarID) + types.UID("_") + uuid
	if kubeData.Deployment == nil {
		werr := pkgerrors.Wrap(err, "Read kubeData.Deployment error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	// Not using "_" since only "." and "-" are allowed for deployment names.
	id := string(uuid.NewUUID())

	externalVNFID := resource.CsarID + "-" + id
	yamlName := kubeData.Deployment.Name

	internalDeploymentName := externalVNFID + "-" + kubeData.Deployment.Name
	internalServiceName := externalVNFID + "-" + kubeData.Service.Name

	// Persist in AAI database.
	log.Println("VNF ID: " + externalVNFID)

	kubeData.Deployment.Namespace = resource.Namespace
	kubeData.Deployment.Name = internalDeploymentName

	if kubeData.Service == nil {
		werr := pkgerrors.Wrap(err, "Read kubeData.Service error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}
	kubeData.Service.Namespace = resource.Namespace
	kubeData.Service.Name = internalServiceName

	// krd.AddNetworkAnnotationsToPod(kubeData, resource.Networks)

	_, err = s.Client.CreateDeployment(kubeData.Deployment, resource.Namespace)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Create VNF deployment error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	_, err = s.Client.CreateService(kubeData.Service, resource.Namespace)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Create VNF service error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	internalID := internalDeploymentName + "|" + internalServiceName

	err = db.DBconn.CreateEntry(resource.Namespace, externalVNFID, internalID)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Create VNF deployment error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	err = db.DBconn.CreateEntry(resource.Namespace, externalVNFID, internalServiceName)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Create VNF deployment error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	resp := CreateVnfResponse{
		DeploymentID: externalVNFID,
		Name:         yamlName,
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
func ListHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	deployments, err := db.DBconn.ReadAll(vars["namespace"])
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
		VNFs: deployments,
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
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	externalVNFID := vars["vnfInstanceId"]
	namespace := vars["namespace"]

	// (TODO): Read kubeconfig for specific Cloud Region from local file system
	// if present or download it from AAI
	s, err := NewVNFInstanceService("../kubeconfig/config")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	internalID, found, err := db.DBconn.ReadEntry(namespace, externalVNFID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if found == false {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	internalDeploymentName := strings.Split(internalID, "|")[0]
	internalServiceName := strings.Split(internalID, "|")[1]

	err = s.Client.DeleteService(internalServiceName, namespace)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Delete VNF error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	err = s.Client.DeleteDeployment(internalDeploymentName, namespace)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Delete VNF error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	err = db.DBconn.DeleteEntry(namespace, externalVNFID)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Delete VNF error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
}

// UpdateHandler method to update a VNF instance.
func UpdateHandler(w http.ResponseWriter, r *http.Request) {
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

	err = validateBody(resource)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	kubeData, err := utils.ReadCSARFromFileSystem(resource.CsarID)

	if kubeData.Deployment == nil {
		werr := pkgerrors.Wrap(err, "Update VNF deployment error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}
	kubeData.Deployment.SetUID(types.UID(id))

	if err != nil {
		werr := pkgerrors.Wrap(err, "Update VNF deployment information error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	// (TODO): Read kubeconfig for specific Cloud Region from local file system
	// if present or download it from AAI
	s, err := NewVNFInstanceService("../kubeconfig/config")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = s.Client.UpdateDeployment(kubeData.Deployment, resource.Namespace)
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
func GetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	externalVNFID := vars["vnfInstanceId"]
	namespace := vars["namespace"]

	name, found, err := db.DBconn.ReadEntry(namespace, externalVNFID)

	if found == false {
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
