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

package multicloud

import (
	apps_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VNFInstanceResource represents a Virtual Networking Function in a
// Kurbernetes Deployment
type VNFInstanceResource struct {
	Name     string
	Replicas int
	Labels   map[string]string
}

func (v *VNFInstanceResource) String() string {
	return string(v.Name)
}

// ConvertToDeployment converts the properties of VNFInstanceResource
// in a Kubernetes Deployment Resource
func (v *VNFInstanceResource) ConvertToDeployment() *apps_v1.Deployment {
	spec := apps_v1.DeploymentSpec{
		Template: core_v1.PodTemplateSpec{
			Spec: core_v1.PodSpec{
				Containers: []core_v1.Container{
					{
						Name:  "web",
						Image: "nginx:1.12",
						Ports: []core_v1.ContainerPort{
							{
								Name:          "http",
								Protocol:      core_v1.ProtocolTCP,
								ContainerPort: 80,
							},
						},
					},
				},
			},
		}}

	if v.Replicas > 1 {
		spec.Replicas = func(i int32) *int32 {
			return &i
		}(int32(v.Replicas))
	}
	if v.Labels != nil {
		spec.Selector = &meta_v1.LabelSelector{
			MatchLabels: v.Labels,
		}
	}

	result := &apps_v1.Deployment{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: v.Name,
		},
		Spec: spec,
	}
	return result
}
