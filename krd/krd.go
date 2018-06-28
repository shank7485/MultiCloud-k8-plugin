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
	"errors"

	pkgerrors "github.com/pkg/errors"

	appsV1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// Client is the client used to communicate with Kubernetes Reference Deployment
type Client struct {
	deploymentClient ClientDeploymentInterface
}

// APIVersion supported for the Kubernetes Reference Deployment
const APIVersion = "apps/v1"

// ClientDeploymentInterface having v1.DeploymentInterface inside, tells
// the compiler explicitly that it satisfied the DeploymentInterface without
// having to implement each function below manually.
// Create(*appsV1.Deployment) (*appsV1.Deployment, error)
// List(opts metaV1.ListOptions) (*appsV1.DeploymentList, error)
// Delete(name string, options *metaV1.DeleteOptions) error
// Update(*appsV1.Deployment) (*appsV1.Deployment, error)
// Get(name string, options metaV1.GetOptions) (*appsV1.Deployment, error)
type ClientDeploymentInterface interface {
	v1.DeploymentInterface
}

// NewClient loads Kubernetes local configuration values into a client
func NewClient(kubeconfigPath string) (*Client, error) {
	var deploymentClient ClientDeploymentInterface

	deploymentClient, err := GetKubeClient(kubeconfigPath)
	if err != nil {
		return nil, err
	}
	client := &Client{
		deploymentClient: deploymentClient,
	}
	return client, nil
}

// GetKubeClient loads the Kubernetes configuation values stored into the local configuration file
var GetKubeClient = func(configPath string) (ClientDeploymentInterface, error) {
	var result ClientDeploymentInterface

	if configPath == "" {
		return nil, errors.New("config not passed and is not found in ~/.kube. ")
	}

	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "setConfig: Build config from flags raised an error")
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	result = clientset.AppsV1().Deployments("default")
	return result, nil
}

// The following methods implement the interface VNFInstanceClientInterface.

// CreateDeployment object in a specific Kubernetes Deployment
func (c *Client) CreateDeployment(deployment *appsV1.Deployment) (string, error) {
	result, err := c.deploymentClient.Create(deployment)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Create VNF error")
	}

	return result.GetObjectMeta().GetName(), nil
}

// ListDeployment of existing deployments hosted in a specific Kubernetes Deployment
func (c *Client) ListDeployment(limit int64) (*[]string, error) {
	opts := metaV1.ListOptions{
		Limit: limit,
	}
	opts.APIVersion = APIVersion
	opts.Kind = "Deployment"

	list, err := c.deploymentClient.List(opts)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Get VNF list error")
	}
	result := make([]string, 0, limit)
	if list != nil {
		for _, deployment := range list.Items {
			result = append(result, deployment.Name)
		}
	}
	return &result, nil
}

// DeleteDeployment existing deployments hosting in a specific Kubernetes Deployment
func (c *Client) DeleteDeployment(name string) error {
	deletePolicy := metaV1.DeletePropagationForeground

	err := c.deploymentClient.Delete(name, &metaV1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		return pkgerrors.Wrap(err, "Delete VNF error")
	}
	return nil
}

// UpdateDeployment existing deployments hosting in a specific Kubernetes Deployment
func (c *Client) UpdateDeployment(deployment *appsV1.Deployment) error {
	_, err := c.deploymentClient.Update(deployment)
	if err != nil {
		return pkgerrors.Wrap(err, "Update VNF error")
	}
	return nil
}

// GetDeployment existing deployment hosting in a specific Kubernetes Deployment
func (c *Client) GetDeployment(name string) (string, error) {
	opts := metaV1.GetOptions{}
	opts.APIVersion = APIVersion
	opts.Kind = "Deployment"

	deployment, err := c.deploymentClient.Get(name, opts)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Get VNF error")
	}
	return deployment.Name, nil
}
