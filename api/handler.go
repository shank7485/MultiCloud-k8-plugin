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
	"os"
	"strings"

	"github.com/gorilla/mux"
	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/shank7485/k8-plugin-multicloud/db"
	"github.com/shank7485/k8-plugin-multicloud/krd"
	"github.com/shank7485/k8-plugin-multicloud/csarParser"
)

// VNFInstanceService communicates the actions to Kubernetes deployment
type VNFInstanceService struct {
	Client krd.VNFInstanceClientInterface
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
var GetVNFClient = func(kubeConfigPath string) (krd.VNFInstanceClientInterface, error) {
	var client krd.VNFInstanceClientInterface

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
		if strings.Contains(b.CloudRegionID, "|") || strings.Contains(b.Namespace, "|") {
			werr := pkgerrors.Wrap(errors.New("Character \"|\" not allowed in CSAR ID"), "CreateVnfRequest bad request")
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
	// err := DownloadKubeConfigFromAAI(resource.CloudRegionID, os.Getenv("KUBE_CONFIG_DIR")
	s, err := NewVNFInstanceService(os.Getenv("KUBE_CONFIG_DIR") + "/" + resource.CloudRegionID)
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

	// uuid
	externalVNFID := string(uuid.NewUUID())

	// cloud1-default-uuid
	internalVNFID := resource.CloudRegionID + "-" + resource.Namespace + "-" + externalVNFID

	// cloud1-default-uuid-sisedeploy
	internalDeploymentName := internalVNFID + "-" + kubeData.Deployment.Name

	// cloud1-default-uuid-sisesvc
	internalServiceName := internalVNFID + "-" + kubeData.Service.Name

	// Persist in AAI database.
	log.Printf("Cloud Region ID: %s, Namespace: %s, VNF ID: %s ", resource.CloudRegionID, resource.Namespace, externalVNFID)

	yamlDeploymentName := kubeData.Deployment.Name
	yamlServiceName := kubeData.Service.Name

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

	// cloud1-default-uuid-sisedeploy|cloud1-default-uuid-sisesvc
	internalCombinedID := internalDeploymentName + "|" + internalServiceName

	// key: cloud1-default-uuid
	// value: cloud1-default-uuid-sisedeploy|cloud1-default-uuid-sisesvc
	err = db.DBconn.CreateEntry(internalVNFID, internalCombinedID)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Create VNF deployment error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	var VNFcomponentList []string

	// TODO: Change this
	VNFcomponentList = append(VNFcomponentList, yamlDeploymentName, yamlServiceName)

	resp := CreateVnfResponse{
		VNFID:         externalVNFID,
		CloudRegionID: resource.CloudRegionID,
		Namespace:     resource.Namespace,
		VNFComponents: VNFcomponentList,
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

	cloudRegionID := vars["cloudRegionID"] // cloud1
	namespace := vars["namespace"]         // default

	prefix := cloudRegionID + "-" + namespace // cloud1-default

	internalVNFIDs, err := db.DBconn.ReadAll(prefix)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Get VNF list error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	if len(internalVNFIDs) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// TODO: There is an edge case where if namespace is passed but is missing some characters
	// trailing, it will print the result with those excluding characters. This is because of
	// the way I am trimming the Prefix. This fix is needed.

	var editedList []string

	for _, id := range internalVNFIDs {
		if len(id) > 0 {
			editedList = append(editedList, strings.TrimPrefix(id, prefix)[1:])
		}
	}

	if len(editedList) == 0 {
		editedList = append(editedList, "")
	}

	resp := ListVnfsResponse{
		VNFs: editedList,
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

	cloudRegionID := vars["cloudRegionID"] // cloud1
	namespace := vars["namespace"]         // default
	externalVNFID := vars["externalVNFID"] // uuid

	// cloud1-default-uuid
	internalVNFID := cloudRegionID + "-" + namespace + "-" + externalVNFID

	// (TODO): Read kubeconfig for specific Cloud Region from local file system
	// if present or download it from AAI
	// err := DownloadKubeConfigFromAAI(resource.CloudRegionID, os.Getenv("KUBE_CONFIG_DIR")
	s, err := NewVNFInstanceService(os.Getenv("KUBE_CONFIG_DIR") + "/" + cloudRegionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	internalCombinedID, found, err := db.DBconn.ReadEntry(internalVNFID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if found == false {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	internalDeploymentName := strings.Split(internalCombinedID, "|")[0]
	internalServiceName := strings.Split(internalCombinedID, "|")[1]

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

	err = db.DBconn.DeleteEntry(internalVNFID)
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

	cloudRegionID := vars["cloudRegionID"] // cloud1
	namespace := vars["namespace"]         // default
	externalVNFID := vars["externalVNFID"] // uuid

	//cloudregion1-testnamespace-1
	//cloudregion1-testnamespace-1-deployName|cloudregion1-testnamespace-1-serviceName"

	// cloud1-default-uuid
	internalVNFID := cloudRegionID + "-" + namespace + "-" + externalVNFID

	deployname := ""
	servicename := ""

	name, found, err := db.DBconn.ReadEntry(internalVNFID)

	if found == false {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if len(name) > 0 {
		deployname = strings.Split(name, "|")[0]
		servicename = strings.Split(name, "|")[1]

		deployname = strings.Split(deployname, internalVNFID)[1][1:]
		servicename = strings.Split(servicename, internalVNFID)[1][1:]
	}

	var VNFcomponentList []string

	// TODO: Change this
	VNFcomponentList = append(VNFcomponentList, deployname, servicename)

	resp := GetVnfResponse{
		VNFID:         externalVNFID,
		CloudRegionID: cloudRegionID,
		Namespace:     namespace,
		VNFComponents: VNFcomponentList,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Parsing output of new VNF error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
	}
}
