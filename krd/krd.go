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
	deploymentsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"

	multicloud "github.com/shank7485/k8-plugin-multicloud"
	"github.com/shank7485/k8-plugin-multicloud/cmd/clientConfig"
)

// VNFInstanceClient consumes the API of Kubernetes Reference Deployment
type VNFInstanceClient struct {
	Client deploymentsv1.DeploymentInterface
}

// VNFInstanceClientInterface has methods to work with VNF Instance resources.
type VNFInstanceClientInterface interface {
	Create(vnfInstance *multicloud.VNFInstanceResource) error
}

// NewVNFInstanceClient instantiate a VNFInstanceClient object
func NewVNFInstanceClient(namespace string) (*VNFInstanceClient, error) {
	config, err := clientConfig.InitiateClient()
	if err != nil {
		return nil, err
	}
	client := VNFInstanceClient{
		Client: config.AppsV1().Deployments(namespace),
	}
	return &client, nil
}

// Create VNFInstance resource in a specific Kubernetes Deployment
func (c *VNFInstanceClient) Create(vnfInstance *multicloud.VNFInstanceResource) error {
	_, err := c.Client.Create(vnfInstance.ConvertToDeployment())
	if err != nil {
		return err
	}
	return nil
}
