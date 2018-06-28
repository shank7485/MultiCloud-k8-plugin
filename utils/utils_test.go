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
package utils

import (
	"reflect"
	"testing"

	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDownloadDeploymentInfo(t *testing.T) {
	t.Run("Succesful download deployment information", func(t *testing.T) {
		Download = func(url string) ([]byte, error) {
			body := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sise-deploy
spec:
  template:
    metadata:
      labels:
        app: sise
    spec:
      containers:
      - name: sise
        image: mhausenblas/simpleservice:0.5.0
`
			return []byte(body), nil
		}
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
		result, err := DownloadCSAR("http://fakescarserver.com")
		if err != nil {
			t.Fatalf("TestDownloadDeploymentInfo returned an error (%s)", err)
		}
		if result == nil {
			t.Fatal("TestDownloadDeploymentInfo didn't return a result")
		}
		if !reflect.DeepEqual(expected, result.Deployment) {
			t.Fatalf("TestDownloadDeploymentInfo returned:\n result=%v\n expected=%v", result, expected)
		}
	})
}

func TestDownloadServiceInfo(t *testing.T) {
	t.Run("Succesful parse service information", func(t *testing.T) {
		Download = func(url string) ([]byte, error) {
			body := `
apiVersion: v1
kind: Service
metadata:
  name: sise-svc
spec:
  ports:
  - port: 80
    protocol: TCP
  selector:
    app: sise
---
`
			return []byte(body), nil
		}
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

		result, err := DownloadCSAR("http://fakescarserver.com")

		if err != nil {
			t.Fatalf("TestDownloadServiceInfo returned an error (%s)", err)
		}
		if result == nil {
			t.Fatal("TestDownloadServiceInfo didn't return a result")
		}
		if !reflect.DeepEqual(expectedService, result.Service) {
			t.Fatalf("TestDownloadServiceInfo returned:\n result=%v\n expected=%v", result.Service, expectedService)
		}
	})
}
