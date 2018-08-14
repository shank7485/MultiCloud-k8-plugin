package plugins

import (
	"io/ioutil"
	"log"
	"os"

	pkgerrors "github.com/pkg/errors"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	coreV1Interface "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/shank7485/k8-plugin-multicloud/krd"
)

// KubeServiceClient is a concrete implementaton of ServiceInterface and KubeResourceClient
type KubeServiceClient struct {
	serviceClient ServiceInterface
	krd.KubeResourceClient
}

// ServiceInterface is an interface which wraps Core V1
type ServiceInterface interface {
	coreV1Interface.CoreV1Interface
}

// CreateResource object in a specific Kubernetes Deployment
func (s *KubeServiceClient) CreateResource(service *coreV1.Service, namespace string) (string, error) {
	if namespace == "" {
		namespace = "default"
	}
	result, err := s.serviceClient.Services(namespace).Create(service)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Create Service error")
	}

	return result.GetObjectMeta().GetName(), nil
}

// ListResources of existing deployments hosted in a specific Kubernetes Deployment
func (s *KubeServiceClient) ListResources(limit int64, namespace string) (*[]string, error) {
	if namespace == "" {
		namespace = "default"
	}
	opts := metaV1.ListOptions{
		Limit: limit,
	}
	opts.APIVersion = apiVersion
	opts.Kind = "Service"

	list, err := s.serviceClient.Services(namespace).List(opts)
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

// DeleteResource deletes an existing Kubernetes service
func (s *KubeServiceClient) DeleteResource(internalVNFID string, namespace string) error {
	if namespace == "" {
		namespace = "default"
	}

	deletePolicy := metaV1.DeletePropagationForeground
	err := s.serviceClient.Services(namespace).Delete(internalVNFID, &metaV1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Service error")
	}

	return nil
}

// GetResource existing service hosting in a specific Kubernetes Service
func (s *KubeServiceClient) GetResource(internalVNFID string, namespace string) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	opts := metaV1.ListOptions{
		Limit: 10,
	}
	opts.APIVersion = apiVersion
	opts.Kind = "SErvice"

	list, err := s.serviceClient.Services(namespace).List(opts)
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

// KubeServiceData is a concrete implemetation of KubeResourceData inteface
type KubeServiceData struct {
	ServiceData []byte
	Service     *coreV1.Service
}

// ReadYAML reads service.yaml and stores in KubeServiceData struct
func (c *KubeServiceData) ReadYAML(yamlFilePath string) error {
	if _, err := os.Stat(yamlFilePath); err == nil {
		log.Println("Reading service YAML")
		rawBytes, err := ioutil.ReadFile(yamlFilePath)
		if err != nil {
			return pkgerrors.Wrap(err, "Service YAML file read error")
		}

		c.ServiceData = rawBytes

		err = c.ParseYAML()
		if err != nil {
			return err
		}
	}
	return nil
}

// ParseYAML retrieves the service YAML file from a KubeServiceData
func (c *KubeServiceData) ParseYAML() error {
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
