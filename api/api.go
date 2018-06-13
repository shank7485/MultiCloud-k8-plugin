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
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

func init() {
	err := InitiateK8client("")
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/v1/vnf_instances", CreateVNF).Methods("POST")
	router.HandleFunc("/v1/vnf_instances", ListVNF).Methods("GET")

	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	log.Println("[INFO] Started Kubernetes Multicloud API")
	log.Fatal(http.ListenAndServe(":8080", loggedRouter)) // Remove hardcode.
}

// import (
// 	"net/http"

// 	"github.com/emicklei/go-restful"
// 	restfulspec "github.com/emicklei/go-restful-openapi"
// 	multicloud "github.com/shank7485/k8-plugin-multicloud"
// 	"github.com/shank7485/k8-plugin-multicloud/krd"
// )

// // VNFInstanceService exposes a RESTful API webservice
// type VNFInstanceService struct {
// 	restful.WebService
// 	client krd.VNFInstanceClientInterface
// }

// // NewVNFInstanceService creates a RESTful API webservice for exporting VNFInstanceResource
// func (vr *VNFInstanceService) NewVNFInstanceService() *restful.WebService {
// 	ws := new(restful.WebService)
// 	ws.
// 		Path("/vnf_instances").
// 		Consumes(restful.MIME_XML, restful.MIME_JSON).
// 		Produces(restful.MIME_JSON, restful.MIME_XML) // you can specify this per route as well

// 	tags := []string{"vnf_instances"}

// 	ws.Route(ws.PUT("").To(vr.createVNF).
// 		Metadata(restfulspec.KeyOpenAPITags, tags).
// 		Reads(multicloud.VNFInstanceResource{})) // from the request

// 	return ws
// }

// func (vr *VNFInstanceService) createVNF(request *restful.Request, response *restful.Response) {
// 	vnf := multicloud.VNFInstanceResource{
// 		Name: request.PathParameter("name"),
// 	}
// 	if err := vr.client.Create(&vnf); err != nil {
// 		response.WriteError(http.StatusInternalServerError, err)
// 	}
// }
