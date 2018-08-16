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
package csarparser

import (
	"os"
	"reflect"
	"testing"

	"github.com/shank7485/k8-plugin-multicloud/krd"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReadDeploymentYAML(t *testing.T) {
	t.Run("Successfully read deployment YAML", func(t *testing.T) {
		expected := &appsV1.Deployment{
			ObjectMeta: metaV1.ObjectMeta{
				Name: "sise-deploy",
			},
			Spec: appsV1.DeploymentSpec{
				Template: coreV1.PodTemplateSpec{
					ObjectMeta: metaV1.ObjectMeta{
						Labels: map[string]string{"app": "sise"},
					},
					Spec: coreV1.PodSpec{
						Containers: []coreV1.Container{
							{
								Name:  "sise",
								Image: "mhausenblas/simpleservice:0.5.0",
							},
						},
					},
				},
			},
		}
		expected.APIVersion = "apps/v1"
		expected.Kind = "Deployment"

		kubeData := &krd.KubernetesData{}
		err := kubeData.ReadDeploymentYAML("mock_yamls/deployment.yaml")

		if err != nil {
			t.Fatalf("TestReadDeploymentYAML returned an error (%s)", err)
		}

		if !reflect.DeepEqual(expected, kubeData.Deployment) {
			t.Fatalf("TestReadDeploymentYAML returned:\n result=%v\n expected=%v", kubeData.Deployment, expected)
		}
	})
}

func TestReadServiceYAML(t *testing.T) {
	t.Run("Successfully read service YAML", func(t *testing.T) {
		expected := &coreV1.Service{
			ObjectMeta: metaV1.ObjectMeta{
				Name: "sise-svc",
			},
			Spec: coreV1.ServiceSpec{
				Ports: []coreV1.ServicePort{
					{
						Port:     80,
						Protocol: "TCP",
					},
				},
				Selector: map[string]string{"app": "sise"},
			},
		}
		expected.APIVersion = "v1"
		expected.Kind = "Service"

		kubeData := &krd.KubernetesData{}
		err := kubeData.ReadServiceYAML("mock_yamls/service.yaml")

		if err != nil {
			t.Fatalf("TestReadServiceYAML returned an error (%s)", err)
		}

		if !reflect.DeepEqual(expected, kubeData.Service) {
			t.Fatalf("TestReadServiceYAML returned:\n result=%v\n expected=%v", kubeData.Service, expected)
		}
	})
}

func TestReadCSARFromFileSystem(t *testing.T) {
	t.Run("Successfully create Deployment and Service objects", func(t *testing.T) {
		expectedDeployment := &appsV1.Deployment{
			ObjectMeta: metaV1.ObjectMeta{
				Name: "sise-deploy",
			},
			Spec: appsV1.DeploymentSpec{
				Template: coreV1.PodTemplateSpec{
					ObjectMeta: metaV1.ObjectMeta{
						Labels: map[string]string{"app": "sise"},
					},
					Spec: coreV1.PodSpec{
						Containers: []coreV1.Container{
							{
								Name:  "sise",
								Image: "mhausenblas/simpleservice:0.5.0",
							},
						},
					},
				},
			},
		}
		expectedDeployment.APIVersion = "apps/v1"
		expectedDeployment.Kind = "Deployment"

		expectedService := &coreV1.Service{
			ObjectMeta: metaV1.ObjectMeta{
				Name: "sise-svc",
			},
			Spec: coreV1.ServiceSpec{
				Ports: []coreV1.ServicePort{
					{
						Port:     80,
						Protocol: "TCP",
					},
				},
				Selector: map[string]string{"app": "sise"},
			},
		}
		expectedService.APIVersion = "v1"
		expectedService.Kind = "Service"

		os.Setenv("CSAR_DIR", ".")

		kubeData, err := ReadCSARFromFileSystem("mock_yamls")
		if err != nil {
			t.Fatalf("TestReadCSARFromFileSystem returned an error (%s)", err)
		}

		if !reflect.DeepEqual(expectedService, kubeData.Service) {
			t.Fatalf("TestReadCSARFromFileSystem returned:\n result=%v\n expected=%v", kubeData.Service, expectedService)
		}

		if !reflect.DeepEqual(expectedDeployment, kubeData.Deployment) {
			t.Fatalf("TestReadCSARFromFileSystem returned:\n result=%v\n expected=%v", kubeData.Deployment, expectedDeployment)
		}
	})
}
