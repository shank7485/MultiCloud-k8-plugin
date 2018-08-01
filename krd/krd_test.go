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

package krd

import (
	"reflect"
	"testing"

	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsV1Interface "k8s.io/client-go/kubernetes/typed/apps/v1"
	coreV1Interface "k8s.io/client-go/kubernetes/typed/core/v1"
)

type mockAppsV1 struct {
	appsV1Interface.AppsV1Interface

	deployments func() *mockDeploymentClient
}

func (c *mockAppsV1) Deployments(namespace string) appsV1Interface.DeploymentInterface {
	if c.deployments != nil {
		return c.deployments()
	}
	return nil
}

type mockDeploymentClient struct {
	appsV1Interface.DeploymentInterface

	create func() (*appsV1.Deployment, error)
	list   func() (*appsV1.DeploymentList, error)
	delete func() error
	update func() (*appsV1.Deployment, error)
	get    func() (*appsV1.Deployment, error)
}

// There mocks are to implement the actual v1.DeploymentInterface
func (c *mockDeploymentClient) Create(deployment *appsV1.Deployment) (*appsV1.Deployment, error) {
	if c.create != nil {
		return c.create()
	}
	return nil, nil
}

func (c *mockDeploymentClient) List(opts metaV1.ListOptions) (*appsV1.DeploymentList, error) {
	if c.list != nil {
		return c.list()
	}
	return nil, nil
}

func (c *mockDeploymentClient) Delete(name string, options *metaV1.DeleteOptions) error {
	if c.delete != nil {
		return c.delete()
	}
	return nil
}

func (c *mockDeploymentClient) Update(deployment *appsV1.Deployment) (*appsV1.Deployment, error) {
	if c.update != nil {
		return c.update()
	}
	return nil, nil
}

func (c *mockDeploymentClient) Get(name string, options metaV1.GetOptions) (*appsV1.Deployment, error) {
	if c.get != nil {
		return c.get()
	}
	return nil, nil
}

type mockCoreV1 struct {
	coreV1Interface.CoreV1Interface

	services func() *mockServiceClient
}

func (c *mockCoreV1) Services(namespace string) coreV1Interface.ServiceInterface {
	if c.services != nil {
		return c.services()
	}
	return nil
}

type mockServiceClient struct {
	coreV1Interface.ServiceInterface

	create func() (*coreV1.Service, error)
	list   func() (*coreV1.ServiceList, error)
	delete func() error
	update func() (*coreV1.Service, error)
	get    func() (*coreV1.Service, error)
}

// There mocks are to implement the actual corev1.ServiceInterface
func (c *mockServiceClient) Create(service *coreV1.Service) (*coreV1.Service, error) {
	if c.create != nil {
		return c.create()
	}
	return nil, nil
}

func (c *mockServiceClient) List(opts metaV1.ListOptions) (*coreV1.ServiceList, error) {
	if c.list != nil {
		return c.list()
	}
	return nil, nil
}

func (c *mockServiceClient) Delete(name string, options *metaV1.DeleteOptions) error {
	if c.delete != nil {
		return c.delete()
	}
	return nil
}

func (c *mockServiceClient) Update(service *coreV1.Service) (*coreV1.Service, error) {
	if c.update != nil {
		return c.update()
	}
	return nil, nil
}

func (c *mockServiceClient) Get(name string, options metaV1.GetOptions) (*coreV1.Service, error) {
	if c.get != nil {
		return c.get()
	}
	return nil, nil
}

func TestClientCreateMethod(t *testing.T) {
	t.Run("Succesful deployment and service creation", func(t *testing.T) {
		expectedDeploy := "sise-deploy"
		expectedService := "sise-svc"
		inputDeploy := &appsV1.Deployment{
			ObjectMeta: metaV1.ObjectMeta{
				Name: expectedDeploy,
			},
		}
		inputService := &coreV1.Service{
			ObjectMeta: metaV1.ObjectMeta{
				Name: expectedService,
			},
		}

		GetKubeClient = func(configPath string) (ClientDeploymentInterface, ClientServiceInterface, error) {
			mockAppsv1 := &mockAppsV1{
				deployments: func() *mockDeploymentClient {
					mockDeploy := &mockDeploymentClient{
						create: func() (*appsV1.Deployment, error) {
							return inputDeploy, nil
						},
					}
					return mockDeploy
				},
			}
			mockCorev1 := &mockCoreV1{
				services: func() *mockServiceClient {
					mockService := &mockServiceClient{
						create: func() (*coreV1.Service, error) {
							return inputService, nil
						},
					}
					return mockService
				},
			}
			return mockAppsv1, mockCorev1, nil
		}

		client, _ := NewClient("")
		resultDeploy, err := client.CreateDeployment(inputDeploy, "")
		if err != nil {
			t.Fatalf("TestClientCreateMethod Deployment returned an error (%s)", err)
		}
		if resultDeploy != expectedDeploy {
			t.Fatalf("TestClientCreateMethod Deployment returned:\n result=%v\n expected=%v", resultDeploy, expectedDeploy)
		}

		resultService, err := client.CreateService(inputService, "")
		if err != nil {
			t.Fatalf("TestClientCreateMethod Service returned an error (%s)", err)
		}
		if resultService != expectedService {
			t.Fatalf("TestClientCreateMethod Service returned:\n result=%v\n expected=%v", resultService, expectedService)
		}
	})
}

func TestClientListMethod(t *testing.T) {
	t.Run("Succesful list of all deployments and services", func(t *testing.T) {
		expectedDeploy := &[]string{"testdeploy1-aa-bb-cc-dd-ee-deploy", "testdeploy2-aa-bb-cc-dd-ee-deploy"}
		expectedService := &[]string{"testService1-aa-bb-cc-dd-ee-service", "testService2-aa-bb-cc-dd-ee-service"}

		inputDeploy := &appsV1.DeploymentList{
			Items: []appsV1.Deployment{
				appsV1.Deployment{
					ObjectMeta: metaV1.ObjectMeta{
						Name: "testdeploy1-aa-bb-cc-dd-ee-deploy",
					},
				},
				appsV1.Deployment{
					ObjectMeta: metaV1.ObjectMeta{
						Name: "testdeploy2-aa-bb-cc-dd-ee-deploy",
					},
				},
			},
		}
		inputService := &coreV1.ServiceList{
			Items: []coreV1.Service{
				coreV1.Service{
					ObjectMeta: metaV1.ObjectMeta{
						Name: "testService1-aa-bb-cc-dd-ee-service",
					},
				},
				coreV1.Service{
					ObjectMeta: metaV1.ObjectMeta{
						Name: "testService2-aa-bb-cc-dd-ee-service",
					},
				},
			},
		}

		GetKubeClient = func(configPath string) (ClientDeploymentInterface, ClientServiceInterface, error) {
			mockAppsv1 := &mockAppsV1{
				deployments: func() *mockDeploymentClient {
					mockDeploy := &mockDeploymentClient{
						list: func() (*appsV1.DeploymentList, error) {
							return inputDeploy, nil
						},
					}
					return mockDeploy
				},
			}
			mockCorev1 := &mockCoreV1{
				services: func() *mockServiceClient {
					mockService := &mockServiceClient{
						list: func() (*coreV1.ServiceList, error) {
							return inputService, nil
						},
					}
					return mockService
				},
			}

			return mockAppsv1, mockCorev1, nil
		}
		client, _ := NewClient("")
		resultDeploy, err := client.ListDeployment(10, "")
		if err != nil {
			t.Fatalf("TestClientListMethod Deployment returned an error (%s)", err)
		}
		if !reflect.DeepEqual(expectedDeploy, resultDeploy) {
			t.Fatalf("TestClientListMethod Deployment returned:\n result=%v\n expected=%v", resultDeploy, expectedDeploy)
		}

		resultService, err := client.ListService(10, "")
		if err != nil {
			t.Fatalf("TestClientListMethod Service returned an error (%s)", err)
		}
		if !reflect.DeepEqual(expectedService, resultService) {
			t.Fatalf("TestClientListMethod Service returned:\n result=%v\n expected=%v", resultService, expectedService)
		}

	})
}

func TestClientDeleteMethod(t *testing.T) {
	t.Run("Succesful deployment and service deletion", func(t *testing.T) {

		inputDeploy := &appsV1.DeploymentList{
			Items: []appsV1.Deployment{
				appsV1.Deployment{
					ObjectMeta: metaV1.ObjectMeta{
						Name: "testdeploy1-aa-bb-cc-dd-ee-deploy",
					},
				},
				appsV1.Deployment{
					ObjectMeta: metaV1.ObjectMeta{
						Name: "testdeploy2-aa-bb-cc-dd-ee-deploy",
					},
				},
			},
		}
		inputService := &coreV1.ServiceList{
			Items: []coreV1.Service{
				coreV1.Service{
					ObjectMeta: metaV1.ObjectMeta{
						Name: "testService1-aa-bb-cc-dd-ee-service",
					},
				},
				coreV1.Service{
					ObjectMeta: metaV1.ObjectMeta{
						Name: "testService2-aa-bb-cc-dd-ee-service",
					},
				},
			},
		}

		GetKubeClient = func(configPath string) (ClientDeploymentInterface, ClientServiceInterface, error) {
			mockAppsv1 := &mockAppsV1{
				deployments: func() *mockDeploymentClient {
					mockDeploy := &mockDeploymentClient{
						delete: func() error {
							return nil
						},
						list: func() (*appsV1.DeploymentList, error) {
							return inputDeploy, nil
						},
					}
					return mockDeploy
				},
			}
			mockCorev1 := &mockCoreV1{
				services: func() *mockServiceClient {
					mockService := &mockServiceClient{
						delete: func() error {
							return nil
						},
						list: func() (*coreV1.ServiceList, error) {
							return inputService, nil
						},
					}
					return mockService
				},
			}
			return mockAppsv1, mockCorev1, nil
		}

		client, _ := NewClient("")
		err := client.DeleteDeployment("test", "")
		if err != nil {
			t.Fatalf("TestClientDeleteMethod Deployment returned an error (%s)", err)
		}

		err = client.DeleteService("test", "")
		if err != nil {
			t.Fatalf("TestClientDeleteMethod Service returned an error (%s)", err)
		}
	})
}

func TestClientUpdateMethod(t *testing.T) {
	t.Run("Succesful deployment and service update", func(t *testing.T) {
		expectedOldDeploy := "sise-deploy"
		expectedOldService := "sise-svc"
		inputDeploy := &appsV1.Deployment{
			ObjectMeta: metaV1.ObjectMeta{
				Name: expectedOldDeploy,
			},
		}
		inputService := &coreV1.Service{
			ObjectMeta: metaV1.ObjectMeta{
				Name: expectedOldService,
			},
		}

		GetKubeClient = func(configPath string) (ClientDeploymentInterface, ClientServiceInterface, error) {
			mockAppsv1 := &mockAppsV1{
				deployments: func() *mockDeploymentClient {
					mockDeploy := &mockDeploymentClient{
						update: func() (*appsV1.Deployment, error) {
							return inputDeploy, nil
						},
					}
					return mockDeploy
				},
			}
			mockCorev1 := &mockCoreV1{
				services: func() *mockServiceClient {
					mockService := &mockServiceClient{
						update: func() (*coreV1.Service, error) {
							return inputService, nil
						},
					}
					return mockService
				},
			}
			return mockAppsv1, mockCorev1, nil
		}

		client, _ := NewClient("")
		inputDeploy.SetName("New-sise-deploy")
		inputService.SetName("New-sise-service")

		err := client.UpdateDeployment(inputDeploy, "")
		if err != nil {
			t.Fatalf("TestClientUpdateMethod Deployment returned an error (%s)", err)
		}

		err = client.UpdateService(inputService, "")
		if err != nil {
			t.Fatalf("TestClientUpdateMethod Service returned an error (%s)", err)
		}
	})
}

func TestClientGetMethod(t *testing.T) {
	t.Run("Succesful get deployment and service", func(t *testing.T) {
		inputDeploy := &appsV1.DeploymentList{
			Items: []appsV1.Deployment{
				appsV1.Deployment{
					ObjectMeta: metaV1.ObjectMeta{
						Name: "testVNF-aa-bb-cc-dd-ee-deploy1",
					},
				},
				appsV1.Deployment{
					ObjectMeta: metaV1.ObjectMeta{
						Name: "testVNF-aa-bb-cc-dd-ee-deploy2",
					},
				},
			},
		}
		inputService := &coreV1.ServiceList{
			Items: []coreV1.Service{
				coreV1.Service{
					ObjectMeta: metaV1.ObjectMeta{
						Name: "testVNF-aa-bb-cc-dd-ee-service1",
					},
				},
				coreV1.Service{
					ObjectMeta: metaV1.ObjectMeta{
						Name: "testVNF-aa-bb-cc-dd-ee-service2",
					},
				},
			},
		}

		outputDeploy := &appsV1.Deployment{
			ObjectMeta: metaV1.ObjectMeta{
				Name: "testVNF-aa-bb-cc-dd-ee-deploy1",
			},
		}
		outputService := &coreV1.Service{
			ObjectMeta: metaV1.ObjectMeta{
				Name: "testVNF-aa-bb-cc-dd-ee-service1",
			},
		}

		GetKubeClient = func(configPath string) (ClientDeploymentInterface, ClientServiceInterface, error) {
			mockAppsv1 := &mockAppsV1{
				deployments: func() *mockDeploymentClient {
					mockDeploy := &mockDeploymentClient{
						list: func() (*appsV1.DeploymentList, error) {
							return inputDeploy, nil
						},
						get: func() (*appsV1.Deployment, error) {
							return outputDeploy, nil
						},
					}
					return mockDeploy
				},
			}
			mockCorev1 := &mockCoreV1{
				services: func() *mockServiceClient {
					mockService := &mockServiceClient{
						list: func() (*coreV1.ServiceList, error) {
							return inputService, nil
						},
						get: func() (*coreV1.Service, error) {
							return outputService, nil
						},
					}
					return mockService
				},
			}
			return mockAppsv1, mockCorev1, nil
		}

		client, _ := NewClient("")
		expectedDeploy := "testVNF-aa-bb-cc-dd-ee-deploy1"
		expectedService := "testVNF-aa-bb-cc-dd-ee-service1"

		result, err := client.GetDeployment("testVNF-aa-bb-cc-dd-ee-deploy1", "")
		if err != nil {
			t.Fatalf("TestClientGetMethod Deployment returned an error (%s)", err)
		}
		if !reflect.DeepEqual(expectedDeploy, result) {
			t.Fatalf("TestClientGetMethod Deployment returned:\n result=%v\n expected=%v", result, expectedDeploy)
		}

		result, err = client.GetService("testVNF-aa-bb-cc-dd-ee-service1", "")
		if err != nil {
			t.Fatalf("TestClientGetMethod Service returned an error (%s)", err)
		}
		if !reflect.DeepEqual(expectedService, result) {
			t.Fatalf("TestClientGetMethod Service returned:\n result=%v\n expected=%v", result, expectedService)
		}

	})
}
