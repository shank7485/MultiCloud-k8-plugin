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
	"log"

	"github.com/gorilla/mux"
	pkgerrors "github.com/pkg/errors"
)

// NewRouter creates a router instance that serves the VNFInstance web methods
func NewRouter(kubeconfig string) (s *mux.Router) {
	service, err := NewVNFInstanceService(kubeconfig)
	if err != nil {
		log.Panic(pkgerrors.Wrap(err, "Creation of a service error"))
	}
	router := mux.NewRouter()

	vnfInstanceHandler := router.PathPrefix("/v1/vnf_instances").Subrouter()
	vnfInstanceHandler.HandleFunc("/", service.Create).Methods("POST").Name("VNFCreation")
	vnfInstanceHandler.HandleFunc("/", service.List).Methods("GET")
	vnfInstanceHandler.HandleFunc("/{vnfInstanceId:[0-9]+}", service.Delete).Methods("DELETE")

	return router
}
