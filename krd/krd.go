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
	"k8s.io/client-go/tools/clientcmd"
)

// Client is the client used to communicate with Kubernetes Reference Deployment
type Client struct {
	deploymentClient ClientDeploymentInterface
}

// ClientDeploymentInterface contains a subset of supported methods
type ClientDeploymentInterface interface {
	Create(*appsV1.Deployment) (*appsV1.Deployment, error)
	List(opts metaV1.ListOptions) (*appsV1.DeploymentList, error)
}

// NewClient loads Kubernetes local configuration values into a client
func NewClient(kubeconfigPath string) (*Client, error) {
	var deploymentClient ClientDeploymentInterface

	deploymentClient, err := getKubeClient(kubeconfigPath)
	if err != nil {
		return nil, err
	}
	client := &Client{
		deploymentClient: deploymentClient,
	}
	return client, nil
}

var getKubeClient = func(configPath string) (ClientDeploymentInterface, error) {
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

// Create deployment object in a specific Kubernetes Deployment
func (c *Client) Create(deployment *appsV1.Deployment) (string, error) {
	result, err := c.deploymentClient.Create(deployment)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Create VNF error")
	}

	return result.GetObjectMeta().GetName(), nil
}

// List of existing deployments hosted in a specidf Kubernetes Deployment
func (c *Client) List(limit int64) (*appsV1.DeploymentList, error) {
	opts := metaV1.ListOptions{
		Limit: limit,
	}
	opts.APIVersion = "apps/v1"
	opts.Kind = "Deployment"

	list, err := c.deploymentClient.List(opts)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Get VNF list error")
	}
	return list, nil
}
