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

package main

import (
	"log"
	// 	"net/http"
	// 	"github.com/emicklei/go-restful"
	// 	restfulspec "github.com/emicklei/go-restful-openapi"
	// 	"github.com/go-openapi/spec"
	// 	"github.com/shank7485/k8-plugin-multicloud/api"
)

func main() {
	// vnfService := api.VNFInstanceService{}
	// restful.DefaultContainer.Add(vnfService.NewVNFInstanceService())

	// config := restfulspec.Config{
	// 	WebServices: restful.RegisteredWebServices(),
	// 	APIPath:     "/apidocs.json",
	// 	PostBuildSwaggerObjectHandler: enrichSwaggerObject}
	// restful.DefaultContainer.Add(restfulspec.NewOpenAPIService(config))

	// cors := restful.CrossOriginResourceSharing{
	// 	AllowedHeaders: []string{"Content-Type", "Accept"},
	// 	AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
	// 	CookiesAllowed: false,
	// 	Container:      restful.DefaultContainer}
	// restful.DefaultContainer.Filter(cors.Filter)

	// log.Printf("Get the API using http://localhost:8080/apidocs.json")
	// log.Printf("Open Swagger UI using http://localhost:8080/apidocs/?url=http://localhost:8080/apidocs.json")
	// log.Fatal(http.ListenAndServe(":8080", nil))
	log.Println("Working in progress...")
}

// func enrichSwaggerObject(swo *spec.Swagger) {
// 	https://gist.github.com/shank7485/ce2fd1b8855df2d0cc33db0b15ef0015
// 	https://github.com/emicklei/go-restful/blob/master/examples/restful-openapi.go#L151
// }
