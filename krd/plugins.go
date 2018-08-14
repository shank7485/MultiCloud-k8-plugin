package krd

import (
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"plugin"
)

// LoadedPlugins stores references to the stored plugins
var LoadedPlugins = map[string]*plugin.Plugin{}

// KubeResourceClient has the signature methods to create Kubernetes reources
type KubeResourceClient interface {
	CreateResource(interface{}, string, *kubernetes.Clientset) (string, error)
	ListResources(string, *kubernetes.Clientset) (*[]string, error)
	DeleteResource(string, string, *kubernetes.Clientset) error
	GetResource(string, string, *kubernetes.Clientset) (string, error)
}

// KubeResourceData is an interface to read and parse YAML
type KubeResourceData interface {
	ReadYAML(string) error
	ParseYAML() error
}

///////////////////////////////////////////////////////////////////////////////

// Kubernetes Kind concrete implementations reference for type assertions

///////////////////////////////////////////////////////////////////////////////

// KubeDeploymentData is a concrete implemetation of KubeResourceData inteface
type KubeDeploymentData struct {
	DeploymentData []byte
	Deployment     *appsV1.Deployment
	KubeResourceData
}

// KubeServiceData is a concrete implemetation of KubeResourceData inteface
type KubeServiceData struct {
	ServiceData []byte
	Service     *coreV1.Service
	KubeResourceData
}
