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
	"fmt"
	"io/ioutil"
	"log"
	"os"

	pkgerrors "github.com/pkg/errors"

	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	appsV1Interface "k8s.io/client-go/kubernetes/typed/apps/v1"
	coreV1Interface "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// Client is the client used to communicate with Kubernetes Reference Deployment
type Client struct {
	deploymentClient ClientDeploymentInterface
	serviceClient    ClientServiceInterface
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
	appsV1Interface.AppsV1Interface
}

// ClientServiceInterface is for Service clients
type ClientServiceInterface interface {
	coreV1Interface.CoreV1Interface
}

// // ClientNamespaceInterface is for Namespace clients
// type ClientNamespaceInterface interface {
// 	coreV1Interface.NamespaceInterface
// }

// NewClient loads Kubernetes local configuration values into a client
func NewClient(kubeconfigPath string) (*Client, error) {
	var deploymentClient ClientDeploymentInterface
	var serviceClient ClientServiceInterface

	deploymentClient, serviceClient, err := GetKubeClient(kubeconfigPath)
	if err != nil {
		return nil, err
	}
	client := &Client{
		deploymentClient: deploymentClient,
		serviceClient:    serviceClient,
	}
	return client, nil
}

// GetKubeClient loads the Kubernetes configuation values stored into the local configuration file
var GetKubeClient = func(configPath string) (ClientDeploymentInterface, ClientServiceInterface, error) {
	var deploy ClientDeploymentInterface
	var service ClientServiceInterface

	if configPath == "" {
		return nil, nil, errors.New("config not passed and is not found in ~/.kube. ")
	}

	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, nil, pkgerrors.Wrap(err, "setConfig: Build config from flags raised an error")
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	deploy = clientset.AppsV1()
	service = clientset.CoreV1()

	return deploy, service, nil
}

/////////////////////////////////////////////////////////////////////////////////

// KubeResourceClient has the signature methods to create Kubernetes reources
type KubeResourceClient interface {
	CreateResource(interface{}, string) (string, error)
	ListResources(string) (*[]string, error)
	DeleteResource(string, string) error
	GetResource(string, string) (string, error)
}

/////////////////////////////////////////////////////////////////////////////////

// VNFInstanceClientInterface has methods to work with VNF Instance resources.
// This interface's signatures matches the methods in the Client struct in krd
// package. This is done so that we can use the Client inside the VNFInstanceService
// above.
type VNFInstanceClientInterface interface {
	CreateDeployment(deployment *appsV1.Deployment, namespace string) (string, error)
	ListDeployment(limit int64, namespace string) (*[]string, error)
	UpdateDeployment(deployment *appsV1.Deployment, namespace string) error
	DeleteDeployment(name string, namespace string) error
	GetDeployment(name string, namespace string) (string, error)

	CreateService(service *coreV1.Service, namespace string) (string, error)
	ListService(limit int64, namespace string) (*[]string, error)
	UpdateService(service *coreV1.Service, namespace string) error
	DeleteService(name string, namespace string) error
	GetService(name string, namespace string) (string, error)

	CreateNamespace(namespace string) error
	IsNamespaceExists(namespace string) (bool, error)
	DeleteNamespace(namespace string) error
}

// CreateDeployment object in a specific Kubernetes Deployment
func (c *Client) CreateDeployment(deployment *appsV1.Deployment, namespace string) (string, error) {
	if namespace == "" {
		namespace = "default"
	}
	result, err := c.deploymentClient.Deployments(namespace).Create(deployment)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Create Deployment error")
	}

	return result.GetObjectMeta().GetName(), nil
}

// ListDeployment of existing deployments hosted in a specific Kubernetes Deployment
func (c *Client) ListDeployment(limit int64, namespace string) (*[]string, error) {
	if namespace == "" {
		namespace = "default"
	}

	opts := metaV1.ListOptions{
		Limit: limit,
	}
	opts.APIVersion = APIVersion
	opts.Kind = "Deployment"

	list, err := c.deploymentClient.Deployments(namespace).List(opts)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Get Deployment list error")
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
func (c *Client) DeleteDeployment(internalVNFID string, namespace string) error {
	if namespace == "" {
		namespace = "default"
	}

	deletePolicy := metaV1.DeletePropagationForeground
	err := c.deploymentClient.Deployments(namespace).Delete(internalVNFID, &metaV1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})

	if err != nil {
		return pkgerrors.Wrap(err, "Delete Deployment error")
	}

	return nil
}

// UpdateDeployment existing deployments hosting in a specific Kubernetes Deployment
func (c *Client) UpdateDeployment(deployment *appsV1.Deployment, namespace string) error {
	if namespace == "" {
		namespace = "default"
	}
	_, err := c.deploymentClient.Deployments(namespace).Update(deployment)
	if err != nil {
		return pkgerrors.Wrap(err, "Update Deployment error")
	}
	return nil
}

// GetDeployment existing deployment hosting in a specific Kubernetes Deployment
func (c *Client) GetDeployment(internalVNFID string, namespace string) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	opts := metaV1.ListOptions{
		Limit: 10,
	}
	opts.APIVersion = APIVersion
	opts.Kind = "Deployment"

	list, err := c.deploymentClient.Deployments(namespace).List(opts)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Get Deployment error")
	}

	for _, deployment := range list.Items {
		if deployment.Name == internalVNFID {
			return internalVNFID, nil
		}
	}
	return "", nil
}

// CreateService object in a specific Kubernetes Deployment
func (c *Client) CreateService(service *coreV1.Service, namespace string) (string, error) {
	if namespace == "" {
		namespace = "default"
	}
	result, err := c.serviceClient.Services(namespace).Create(service)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Create Service error")
	}

	return result.GetObjectMeta().GetName(), nil
}

// ListService of existing deployments hosted in a specific Kubernetes Deployment
func (c *Client) ListService(limit int64, namespace string) (*[]string, error) {
	if namespace == "" {
		namespace = "default"
	}
	opts := metaV1.ListOptions{
		Limit: limit,
	}
	opts.APIVersion = APIVersion
	opts.Kind = "Service"

	list, err := c.serviceClient.Services(namespace).List(opts)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Get Service list error")
	}
	result := make([]string, 0, limit)
	if list != nil {
		for _, service := range list.Items {
			result = append(result, service.Name)
		}
	}
	return &result, nil
}

// UpdateService updates an existing Kubernetes service
func (c *Client) UpdateService(service *coreV1.Service, namespace string) error {
	if namespace == "" {
		namespace = "default"
	}
	_, err := c.serviceClient.Services(namespace).Update(service)
	if err != nil {
		return pkgerrors.Wrap(err, "Update Service error")
	}
	return nil
}

// DeleteService deletes an existing Kubernetes service
func (c *Client) DeleteService(internalVNFID string, namespace string) error {
	if namespace == "" {
		namespace = "default"
	}

	deletePolicy := metaV1.DeletePropagationForeground
	err := c.serviceClient.Services(namespace).Delete(internalVNFID, &metaV1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Service error")
	}

	return nil
}

// GetService existing service hosting in a specific Kubernetes Service
func (c *Client) GetService(internalVNFID string, namespace string) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	opts := metaV1.ListOptions{
		Limit: 10,
	}
	opts.APIVersion = APIVersion
	opts.Kind = "SErvice"

	list, err := c.serviceClient.Services(namespace).List(opts)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Get Deployment error")
	}

	for _, service := range list.Items {
		if internalVNFID == service.Name {
			return internalVNFID, nil
		}
	}

	return "", nil
}

// CreateNamespace is used to create a new Namespace
func (c *Client) CreateNamespace(namespace string) error {
	namespaceStruct := &coreV1.Namespace{
		ObjectMeta: metaV1.ObjectMeta{
			Name: namespace,
		},
	}

	_, err := c.serviceClient.Namespaces().Create(namespaceStruct)
	if err != nil {
		return pkgerrors.Wrap(err, "Create Namespace error")
	}
	return nil
}

// IsNamespaceExists is used to check if a given namespace actually exists in Kubernetes
func (c *Client) IsNamespaceExists(namespace string) (bool, error) {
	ns, err := c.serviceClient.Namespaces().Get(namespace, metaV1.GetOptions{})
	if err != nil {
		return false, pkgerrors.Wrap(err, "Get Namespace list error")
	}
	return ns != nil, nil
}

// DeleteNamespace is used to delete a namespace
func (c *Client) DeleteNamespace(namespace string) error {
	deletePolicy := metaV1.DeletePropagationForeground

	err := c.serviceClient.Namespaces().Delete(namespace, &metaV1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})

	if err != nil {
		return pkgerrors.Wrap(err, "Delete Namespace error")
	}
	return nil
}

/////////////////////////////////////////////////////////////////////////////////

// KubeResourceData is an interface to read and parse YAML
type KubeResourceData interface {
	ReadYAML(string) error
	ParseYAML(string) error
}

/////////////////////////////////////////////////////////////////////////////////

// CSARParser is an interface to parse both Deployment and Services
// yaml files
type CSARParser interface {
	ReadDeploymentYAML(string) error
	ReadServiceYAML(string) error
	ParseDeploymentInfo() error
	ParseServiceInfo() error
}

// KubernetesData to store CSAR information including both services and
// deployments
type KubernetesData struct {
	DeploymentData []byte
	ServiceData    []byte
	Deployment     *appsV1.Deployment
	Service        *coreV1.Service
}

// ReadDeploymentYAML reads deployment.yaml and stores in CSARData struct
func (c *KubernetesData) ReadDeploymentYAML(yamlFilePath string) error {
	if _, err := os.Stat(yamlFilePath); err == nil {
		log.Println("Reading deployment YAML")
		rawBytes, err := ioutil.ReadFile(yamlFilePath)
		if err != nil {
			return pkgerrors.Wrap(err, "Deployment YAML file read error")
		}

		c.DeploymentData = rawBytes

		err = c.ParseDeploymentInfo()
		if err != nil {
			return err
		}
	}
	return nil
}

// ReadServiceYAML reads service.yaml and stores in CSARData struct
func (c *KubernetesData) ReadServiceYAML(yamlFilePath string) error {
	if _, err := os.Stat(yamlFilePath); err == nil {
		log.Println("Reading service YAML")
		rawBytes, err := ioutil.ReadFile(yamlFilePath)
		if err != nil {
			return pkgerrors.Wrap(err, "Service YAML file read error")
		}

		c.ServiceData = rawBytes

		err = c.ParseServiceInfo()
		if err != nil {
			return err
		}
	}
	return nil
}

// ParseDeploymentInfo retrieves the deployment YAML file from a CSAR
func (c *KubernetesData) ParseDeploymentInfo() error {
	if c.DeploymentData != nil {
		log.Println("Decoding deployment YAML")

		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, _, err := decode(c.DeploymentData, nil, nil)
		if err != nil {
			return pkgerrors.Wrap(err, "Deserialize deployment error")
		}

		switch o := obj.(type) {
		case *appsV1.Deployment:
			c.Deployment = o
			return nil
		}
	}
	return nil
}

// ParseServiceInfo retrieves the service YAML file from a CSAR
func (c *KubernetesData) ParseServiceInfo() error {
	if c.ServiceData != nil {
		log.Println("Decoding service YAML")

		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, _, err := decode(c.ServiceData, nil, nil)
		if err != nil {
			return pkgerrors.Wrap(err, "Deserialize deployment error")
		}

		switch o := obj.(type) {
		case *coreV1.Service:
			c.Service = o
			return nil
		}
	}
	return nil
}

// AddNetworkAnnotationsToPod adds networks metadata to pods
func AddNetworkAnnotationsToPod(c *KubernetesData, networksList []string) {
	/*
		Example Annotation:

		apiVersion: v1
		kind: Pod
		metadata:
		name: multus-multi-net-poc
		annotations:
			networks: '[
				{ "name": "flannel-conf" },
				{ "name": "sriov-conf"},
				{ "name": "sriov-vlanid-l2enable-conf" }
			]'
		spec:  # specification of the pod's contents
		containers:
		- name: multus-multi-net-poc
			image: "busybox"
			command: ["top"]
			stdin: true
			tty: true
	*/

	deployment := c.Deployment
	var networksString string
	networksString = "["

	for _, network := range networksList {
		val := fmt.Sprintf("{ \"name\": \"%s\" },", network)
		networksString += val
	}

	// Removing the final ","
	if len(networksString) > 0 {
		networksString = networksString[:len(networksString)-1]
	}
	networksString += "]"

	deployment.Spec.Template.ObjectMeta = metaV1.ObjectMeta{
		Annotations: map[string]string{"kubernetes.v1.cni.cncf.io/networks": networksString},
	}
}
