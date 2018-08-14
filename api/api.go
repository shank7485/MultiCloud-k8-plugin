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
	"github.com/gorilla/mux"
	pkgerrors "github.com/pkg/errors"
	"os"
	"plugin"

	"github.com/shank7485/k8-plugin-multicloud/db"
	"github.com/shank7485/k8-plugin-multicloud/krd"
)

func CheckEnvVariables() error {
	if os.Getenv("CSAR_DIR") == "" {
		return pkgerrors.New("environment variable CSAR_DIR not set")
	}

	if os.Getenv("KUBE_CONFIG_DIR") == "" {
		return pkgerrors.New("enviromment variable KUBE_CONFIG_DIR not set")
	}

	if os.Getenv("DATABASE_TYPE") == "" {
		return pkgerrors.New("enviromment variable DATABASE_TYPE not set")
	}

	if os.Getenv("PLUGINS_DIR") == "" {
		return pkgerrors.New("enviromment variable PLUGINS_DIR not set")
	}
	return nil
}

func CheckDatabaseConnection() error {
	err := db.CreateDBClient(os.Getenv("DATABASE_TYPE"))
	if err != nil {
		return pkgerrors.Cause(err)
	}

	err = db.DBconn.InitializeDatabase()
	if err != nil {
		return pkgerrors.Cause(err)
	}

	err = db.DBconn.CheckDatabase()
	if err != nil {
		return pkgerrors.Cause(err)
	}
	return nil
}

func LoadPlugins() error {
	p, err := plugin.Open(os.Getenv("PLUGINS_DIR") + "/deployment/deployment.so")
	if err != nil {
		return pkgerrors.Cause(err)
	}

	krd.LoadedPlugins["deployment"] = p

	p, err = plugin.Open(os.Getenv("PLUGINS_DIR") + "/service/service.so")
	if err != nil {
		return pkgerrors.Cause(err)
	}

	krd.LoadedPlugins["service"] = p

	return nil
}

// CheckInitialSettings is used to check initial settings required to start api
func CheckInitialSettings() error {
	err := CheckEnvVariables()
	if err != nil {
		return pkgerrors.Cause(err)
	}

	err = CheckDatabaseConnection()
	if err != nil {
		return pkgerrors.Cause(err)
	}

	err = LoadPlugins()
	if err != nil {
		return pkgerrors.Cause(err)
	}

	return nil
}

// NewRouter creates a router instance that serves the VNFInstance web methods
func NewRouter(kubeconfig string) (s *mux.Router) {
	router := mux.NewRouter()

	vnfInstanceHandler := router.PathPrefix("/v1/vnf_instances").Subrouter()
	vnfInstanceHandler.HandleFunc("/", CreateHandler).Methods("POST").Name("VNFCreation")
	vnfInstanceHandler.HandleFunc("/{cloudRegionID}/{namespace}", ListHandler).Methods("GET")
	vnfInstanceHandler.HandleFunc("/{cloudRegionID}/{namespace}/{externalVNFID}", DeleteHandler).Methods("DELETE")
	vnfInstanceHandler.HandleFunc("/{cloudRegionID}/{namespace}/{externalVNFID}", GetHandler).Methods("GET")

	// (TODO): Fix update method
	// vnfInstanceHandler.HandleFunc("/{vnfInstanceId}", UpdateHandler).Methods("PUT")

	return router
}
