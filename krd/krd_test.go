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
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type mockClient struct {
	create func() (*appsV1.Deployment, error)
	list   func() (*appsV1.DeploymentList, error)
	delete func() error
	update func() (*appsV1.Deployment, error)
}

func (c *mockClient) Create(deployment *appsV1.Deployment) (*appsV1.Deployment, error) {
	if c.create != nil {
		return c.create()
	}
	return nil, nil
}

func (c *mockClient) List(opts metaV1.ListOptions) (*appsV1.DeploymentList, error) {
	if c.list != nil {
		return c.list()
	}
	return nil, nil
}

func (c *mockClient) Delete(name string, options *metaV1.DeleteOptions) error {
	if c.delete() != nil {
		return c.delete()
	}
	return nil
}

func (c *mockClient) Update(deployment *appsV1.Deployment) (*appsV1.Deployment, error) {
	if c.update != nil {
		return c.update()
	}
	return nil, nil
}

func TestClientCreateMethod(t *testing.T) {
	t.Run("Succesful deployment creation", func(t *testing.T) {
		expected := "sise-deploy"
		input := &appsV1.Deployment{
			ObjectMeta: metaV1.ObjectMeta{
				Name: expected,
			},
		}
		GetKubeClient = func(configPath string) (ClientDeploymentInterface, error) {
			return &mockClient{
				create: func() (*appsV1.Deployment, error) {
					return input, nil
				},
			}, nil
		}
		client, _ := NewClient("")
		result, err := client.Create(input)
		if err != nil {
			t.Fatalf("TestDeploymentCreation returned an error (%s)", err)
		}
		if result != expected {
			t.Fatalf("TestDeploymentCreation returned:\n result=%v\n expected=%v", result, expected)
		}
	})
}

func TestClientListMethod(t *testing.T) {
	t.Run("Succesful list of all deployments", func(t *testing.T) {
		expected := &[]string{"test1", "test2"}
		input := &appsV1.DeploymentList{
			Items: []appsV1.Deployment{
				appsV1.Deployment{
					ObjectMeta: metaV1.ObjectMeta{
						Name: "test1",
					},
				},
				appsV1.Deployment{
					ObjectMeta: metaV1.ObjectMeta{
						Name: "test2",
					},
				},
			},
		}
		GetKubeClient = func(configPath string) (ClientDeploymentInterface, error) {
			return &mockClient{
				list: func() (*appsV1.DeploymentList, error) {
					return input, nil
				},
			}, nil
		}
		client, _ := NewClient("")
		result, err := client.List(10)
		if err != nil {
			t.Fatalf("TestClientListMethod returned an error (%s)", err)
		}
		if !reflect.DeepEqual(expected, result) {
			t.Fatalf("TestClientListMethod returned:\n result=%v\n expected=%v", result, expected)
		}
	})
}

func TestClientDeleteMethod(t *testing.T) {
	t.Run("Succesful deployment deletion", func(t *testing.T) {
		GetKubeClient = func(configPath string) (ClientDeploymentInterface, error) {
			return &mockClient{
				delete: func() error {
					return nil
				},
			}, nil
		}
		client, _ := NewClient("")
		deleteOpts := &metaV1.DeleteOptions{}
		err := client.Delete("test", deleteOpts)
		if err != nil {
			t.Fatalf("TestDeploymentDeletion returned an error (%s)", err)
		}
	})
}

func TestClientUpdateMethod(t *testing.T) {
	t.Run("Succesful deployment update", func(t *testing.T) {
		oldName := "sise-deploy"
		input := &appsV1.Deployment{
			ObjectMeta: metaV1.ObjectMeta{
				Name: oldName,
			},
		}
		GetKubeClient = func(configPath string) (ClientDeploymentInterface, error) {
			return &mockClient{
				update: func() (*appsV1.Deployment, error) {
					return input, nil
				},
			}, nil
		}
		client, _ := NewClient("")
		input.SetName("New-sise-deploy")

		err := client.Update(input)
		if err != nil {
			t.Fatalf("TestDeploymentCreation returned an error (%s)", err)
		}
	})
}
