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
	"net/http"
	"net/http/httptest"
	"testing"

	appsV1 "k8s.io/api/apps/v1"
)

type mockClient struct {
	create func() (string, error)
	list   func() (*[]string, error)
	delete func() error
}

func (c *mockClient) Create(deployment *appsV1.Deployment) (string, error) {
	if c.create != nil {
		return c.create()
	}
	return "", nil
}

func (c *mockClient) List(limit int64) (*[]string, error) {
	if c.list != nil {
		return c.list()
	}
	return nil, nil
}

func (c *mockClient) Delete(name string) error {
	if c.delete != nil {
		return c.delete()
	}
	return nil
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
			"csar_url": "https://raw.githubusercontent.com/kubernetes/website/master/content/en/docs/concepts/workloads/controllers/nginx-deployment.yaml",
			"id": "100",
			"oof_parameters": {
				"key_values": {
					"key1": "value1",
					"key2": "value2"
				}
			}
		}`)
		expected := `{"response":"Created Deployment:test"}` + "\n"

		req, _ := http.NewRequest("POST", "/v1/vnf_instances/", bytes.NewBuffer(payload))
		GetVNFClient = func(configPath string) (VNFInstanceClientInterface, error) {
			return &mockClient{
				create: func() (string, error) {
					return "test", nil
				},
			}, nil
		}
		response := executeRequest(req)
		checkResponseCode(t, http.StatusCreated, response.Code)

		if result := response.Body.String(); result != expected {
			t.Fatalf("TestVNFInstanceCreation returned:\n result=%v\n expected=%v", result, expected)
		}
	})
	t.Run("Missing parameters failure", func(t *testing.T) {
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
}

func TestVNFInstancesRetrieval(t *testing.T) {
	var client *mockClient
	GetVNFClient = func(configPath string) (VNFInstanceClientInterface, error) {
		return client, nil
	}

	t.Run("Succesful get a list of VNF", func(t *testing.T) {
		expected := `{"response":"Listing:test1,test2"}` + "\n"
		req, _ := http.NewRequest("GET", "/v1/vnf_instances/", nil)
		client = &mockClient{
			list: func() (*[]string, error) {
				return &[]string{"test1", "test2"}, nil
			},
		}
		response := executeRequest(req)
		checkResponseCode(t, http.StatusOK, response.Code)

		if result := response.Body.String(); result != expected {
			t.Fatalf("TestVNFInstancesRetrieval returned:\n result=%v\n expected=%v", result, expected)
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
		checkResponseCode(t, http.StatusNoContent, response.Code)

		if result := response.Body.String(); result != "" {
			t.Fatalf("TestVNFInstanceDeletion returned:\n result=%v\n expected=%v", result, "")
		}
	})
	// t.Run("Malformed delete request", func(t *testing.T) {
	// 	req, _ := http.NewRequest("DELETE", "/v1/vnf_instances/foo", nil)
	// 	response := executeRequest(req)
	// 	checkResponseCode(t, http.StatusBadRequest, response.Code)
	// })
}
