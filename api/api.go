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
	"errors"
	"github.com/gorilla/mux"
	"os"
)

// CheckInitialSettings is used to check initial settings required to start api
func CheckInitialSettings() error {
	if os.Getenv("CSAR_DIR") == "" {
		return errors.New("environment variable CSAR_DIR not set")
	}
	return nil
}

// NewRouter creates a router instance that serves the VNFInstance web methods
func NewRouter(kubeconfig string) (s *mux.Router) {
	router := mux.NewRouter()

	vnfInstanceHandler := router.PathPrefix("/v1/vnf_instances").Subrouter()
	vnfInstanceHandler.HandleFunc("/", CreateHandler).Methods("POST").Name("VNFCreation")
	vnfInstanceHandler.HandleFunc("/{namespace}", ListHandler).Methods("GET")
	vnfInstanceHandler.HandleFunc("/{namespace}/{vnfInstanceId}", DeleteHandler).Methods("DELETE")
	vnfInstanceHandler.HandleFunc("/{vnfInstanceId}", UpdateHandler).Methods("PUT")
	vnfInstanceHandler.HandleFunc("/{namespace}/{vnfInstanceId}", GetHandler).Methods("GET")

	return router
}
