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
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/shank7485/k8-plugin-multicloud/utils"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
)

type mockClient struct {
	create func() (string, error)
	list   func() (*[]string, error)
	delete func() error
	update func() error
	get    func() (string, error)
}

func (c *mockClient) CreateDeployment(deployment *appsV1.Deployment) (string, error) {
	if c.create != nil {
		return c.create()
	}
	return "", nil
}

func (c *mockClient) ListDeployment(limit int64) (*[]string, error) {
	if c.list != nil {
		return c.list()
	}
	return nil, nil
}

func (c *mockClient) DeleteDeployment(name string) error {
	if c.delete != nil {
		return c.delete()
	}
	return nil
}

func (c *mockClient) UpdateDeployment(deployment *appsV1.Deployment) error {
	if c.delete != nil {
		return c.delete()
	}
	return nil
}

func (c *mockClient) GetDeployment(name string) (string, error) {
	if c.get != nil {
		return c.get()
	}
	return "", nil
}

func (c *mockClient) CreateService(service *coreV1.Service) (string, error) {
	if c.create != nil {
		return c.create()
	}
	return "", nil
}

func (c *mockClient) ListService(limit int64) (*[]string, error) {
	if c.list != nil {
		return c.list()
	}
	return nil, nil
}

func (c *mockClient) DeleteService(name string) error {
	if c.delete != nil {
		return c.delete()
	}
	return nil
}

func (c *mockClient) UpdateService(service *coreV1.Service) error {
	if c.delete != nil {
		return c.delete()
	}
	return nil
}

func (c *mockClient) GetService(name string) (string, error) {
	if c.get != nil {
		return c.get()
	}
	return "", nil
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	router := NewRouter("")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	return recorder
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestVNFInstanceCreation(t *testing.T) {
	t.Run("Succesful create a VNF", func(t *testing.T) {
		payload := []byte(`{
			"csar_id": "1",
			"csar_url": "url",
			"oof_parameters": {
				"key_values": {
					"key1": "value1",
					"key2": "value2"
				}
			},
			"vnf_instance_name": "test",
			"vnf_instance_description": "vRouter_test_description"
		}`)
		expected := &CreateVnfResponse{
			Name: "vRouter_test",
		}
		var result CreateVnfResponse

		req, _ := http.NewRequest("POST", "/v1/vnf_instances/", bytes.NewBuffer(payload))

		GetVNFClient = func(configPath string) (VNFInstanceClientInterface, error) {
			return &mockClient{
				create: func() (string, error) {
					return "vRouter_test", nil
				},
			}, nil
		}
		utils.CreateKubeObjectsFromCSAR = func(csarID string, csarURL string) (*utils.KubernetesData, error) {
			kubeData := &utils.KubernetesData{
				Deployment: &appsV1.Deployment{},
				Service:    &coreV1.Service{},
			}
			return kubeData, nil
		}

		response := executeRequest(req)
		checkResponseCode(t, http.StatusCreated, response.Code)

		err := json.NewDecoder(response.Body).Decode(&result)
		if err != nil {
			t.Fatalf("TestVNFInstanceCreation returned:\n result=%v\n expected=%v", err, expected.Name)
		}
	})
	t.Run("Missing body failure", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/v1/vnf_instances/", nil)
		response := executeRequest(req)

		checkResponseCode(t, http.StatusBadRequest, response.Code)
	})
	t.Run("Invalid JSON request format", func(t *testing.T) {
		payload := []byte("invalid")
		req, _ := http.NewRequest("POST", "/v1/vnf_instances/", bytes.NewBuffer(payload))
		response := executeRequest(req)
		checkResponseCode(t, http.StatusUnprocessableEntity, response.Code)
	})
	t.Run("Missing parameter failure", func(t *testing.T) {
		payload := []byte(`{
			"csar_url": "https://raw.githubusercontent.com/kubernetes/website/master/content/en/docs/concepts/workloads/controllers/nginx-deployment.yaml",
			"oof_parameters": {
				"key_values": {
					"key1": "value1",
					"key2": "value2"
				}
			},
			"vnf_instance_name": "test",
			"vnf_instance_description": "vRouter_test_description"
		}`)
		req, _ := http.NewRequest("POST", "/v1/vnf_instances/", bytes.NewBuffer(payload))
		response := executeRequest(req)
		checkResponseCode(t, http.StatusUnprocessableEntity, response.Code)
	})
}

func TestVNFInstancesRetrieval(t *testing.T) {
	var client *mockClient
	GetVNFClient = func(configPath string) (VNFInstanceClientInterface, error) {
		return client, nil
	}

	t.Run("Succesful get a list of VNF", func(t *testing.T) {
		expected := &ListVnfsResponse{
			VNFs: []string{"test1", "test2"},
		}
		var result ListVnfsResponse

		req, _ := http.NewRequest("GET", "/v1/vnf_instances/", nil)
		client = &mockClient{
			list: func() (*[]string, error) {
				return &[]string{"test1", "test2"}, nil
			},
		}
		response := executeRequest(req)
		checkResponseCode(t, http.StatusOK, response.Code)

		err := json.NewDecoder(response.Body).Decode(&result)
		if err != nil {
			t.Fatalf("TestVNFInstancesRetrieval returned:\n result=%v\n expected=list", err)
		}
		if !reflect.DeepEqual(*expected, result) {
			t.Fatalf("TestVNFInstancesRetrieval returned:\n result=%v\n expected=%v", result, *expected)
		}
	})
	t.Run("Get empty list", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/vnf_instances/", nil)
		client = &mockClient{}
		response := executeRequest(req)
		checkResponseCode(t, http.StatusNotFound, response.Code)
	})
}

func TestVNFInstanceDeletion(t *testing.T) {
	t.Run("Succesful delete a VNF", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/v1/vnf_instances/1", nil)
		response := executeRequest(req)
		checkResponseCode(t, http.StatusAccepted, response.Code)

		if result := response.Body.String(); result != "" {
			t.Fatalf("TestVNFInstanceDeletion returned:\n result=%v\n expected=%v", result, "")
		}
	})
	// t.Run("Malformed delete request", func(t *testing.T) {
	// 	req, _ := http.NewRequest("DELETE", "/v1/vnf_instances/foo", nil)
	// 	response := executeRqequest(req)
	// 	checkResponseCode(t, http.StatusBadRequest, response.Code)
	// })
}

func TestVNFInstanceUpdate(t *testing.T) {
	t.Run("Succesful update a VNF", func(t *testing.T) {
		payload := []byte(`{
			"csar_id": "1",
			"csar_url": "https://raw.githubusercontent.com/kubernetes/website/master/content/en/docs/concepts/workloads/controllers/nginx-deployment.yaml",
			"oof_parameters": {
				"key_values": {
					"key1": "value1",
					"key2": "value2"
				}
			},
			"vnf_instance_name": "test",
			"vnf_instance_description": "vRouter_test_description"
		}`)
		var result UpdateVnfResponse

		req, _ := http.NewRequest("PUT", "/v1/vnf_instances/1", bytes.NewBuffer(payload))
		response := executeRequest(req)

		expected := &UpdateVnfResponse{
			DeploymentID: "1",
		}

		err := json.NewDecoder(response.Body).Decode(&result)
		if err != nil {
			t.Fatalf("TestVNFInstanceUpdate returned:\n result=%v\n expected=%v", err, expected.DeploymentID)
		}

		if result.DeploymentID != expected.DeploymentID {
			t.Fatalf("TestVNFInstanceUpdate returned:\n result=%v\n expected=%v", result.DeploymentID, expected.DeploymentID)
		}
	})
}

func TestVNFInstanceRetrieval(t *testing.T) {
	var client *mockClient
	GetVNFClient = func(configPath string) (VNFInstanceClientInterface, error) {
		return client, nil
	}

	t.Run("Succesful get a VNF", func(t *testing.T) {
		expected := `{"response":"Got Deployment:1"}` + "\n"
		req, _ := http.NewRequest("GET", "/v1/vnf_instances/1", nil)
		client = &mockClient{
			get: func() (string, error) {
				return "1", nil
			},
		}
		response := executeRequest(req)
		checkResponseCode(t, http.StatusOK, response.Code)

		if result := response.Body.String(); result != expected {
			t.Fatalf("TestVNFInstanceRetrieval returned:\n result=%v\n expected=%v", result, expected)
		}
	})
	t.Run("VNF not found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/vnf_instances/1", nil)
		client = &mockClient{}
		response := executeRequest(req)
		checkResponseCode(t, http.StatusNotFound, response.Code)
	})
}
