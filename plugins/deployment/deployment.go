package main

import (
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"log"
	"os"

	pkgerrors "github.com/pkg/errors"

	appsV1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/shank7485/k8-plugin-multicloud/krd"
)

// KubeDeploymentClient is a concrete implementaton of DeploymentInterface and KubeResourceClient
type KubeDeploymentClient struct {
	krd.KubeResourceClient
}

// CreateResource object in a specific Kubernetes Deployment
func (d *KubeDeploymentClient) CreateResource(deployment *appsV1.Deployment, namespace string, kubeclient *kubernetes.Clientset) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	result, err := kubeclient.AppsV1().Deployments(namespace).Create(deployment)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Create Deployment error")
	}

	return result.GetObjectMeta().GetName(), nil
}

// ListResources of existing deployments hosted in a specific Kubernetes Deployment
func (d *KubeDeploymentClient) ListResources(limit int64, namespace string, kubeclient *kubernetes.Clientset) (*[]string, error) {
	if namespace == "" {
		namespace = "default"
	}

	opts := metaV1.ListOptions{
		Limit: limit,
	}
	opts.APIVersion = "apps/v1"
	opts.Kind = "Deployment"

	list, err := kubeclient.AppsV1().Deployments(namespace).List(opts)
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

// DeleteResource existing deployments hosting in a specific Kubernetes Deployment
func (d *KubeDeploymentClient) DeleteResource(internalVNFID string, namespace string, kubeclient *kubernetes.Clientset) error {
	if namespace == "" {
		namespace = "default"
	}

	deletePolicy := metaV1.DeletePropagationForeground
	err := kubeclient.AppsV1().Deployments(namespace).Delete(internalVNFID, &metaV1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})

	if err != nil {
		return pkgerrors.Wrap(err, "Delete Deployment error")
	}

	return nil
}

// GetResource existing deployment hosting in a specific Kubernetes Deployment
func (d *KubeDeploymentClient) GetResource(internalVNFID string, namespace string, kubeclient *kubernetes.Clientset) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	opts := metaV1.ListOptions{
		Limit: 10,
	}
	opts.APIVersion = "apps/v1"
	opts.Kind = "Deployment"

	list, err := kubeclient.AppsV1().Deployments(namespace).List(opts)
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

// KubeDeploymentData is a concrete implemetation of KubeResourceData inteface
type KubeDeploymentData struct {
	DeploymentData []byte
	Deployment     *appsV1.Deployment
	krd.KubeResourceData
}

func CreateKubeData() KubeDeploymentData {
	var res KubeDeploymentData
	return res
}

// ReadYAML reads deployment.yaml and stores in KubeDeploymentData struct
func (c *KubeDeploymentData) ReadYAML(yamlFilePath string) error {
	if _, err := os.Stat(yamlFilePath); err == nil {
		log.Println("Reading deployment YAML")
		rawBytes, err := ioutil.ReadFile(yamlFilePath)
		if err != nil {
			return pkgerrors.Wrap(err, "Deployment YAML file read error")
		}

		c.DeploymentData = rawBytes

		err = c.ParseYAML()
		if err != nil {
			return err
		}
	}
	return nil
}

// ParseYAML retrieves the deployment YAML file from a CSAR
func (c *KubeDeploymentData) ParseYAML() error {
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
