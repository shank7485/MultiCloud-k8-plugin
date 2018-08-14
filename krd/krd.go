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

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// APIVersion supported for the Kubernetes Reference Deployment
const APIVersion = "apps/v1"

// GetKubeClient loads the Kubernetes configuation values stored into the local configuration file
var GetKubeClient = func(configPath string) (kubernetes.Clientset, error) {
	var clientset *kubernetes.Clientset

	if configPath == "" {
		return *clientset, errors.New("config not passed and is not found in ~/.kube. ")
	}

	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return *clientset, pkgerrors.Wrap(err, "setConfig: Build config from flags raised an error")
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return *clientset, err
	}

	return *clientset, nil
}

// CreateNamespace is used to create a new Namespace
func CreateNamespace(namespace string, client *kubernetes.Clientset) error {
	namespaceStruct := &coreV1.Namespace{
		ObjectMeta: metaV1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := client.CoreV1().Namespaces().Create(namespaceStruct)
	if err != nil {
		return pkgerrors.Wrap(err, "Create Namespace error")
	}
	return nil
}

// IsNamespaceExists is used to check if a given namespace actually exists in Kubernetes
func IsNamespaceExists(namespace string, client *kubernetes.Clientset) (bool, error) {
	ns, err := client.CoreV1().Namespaces().Get(namespace, metaV1.GetOptions{})
	if err != nil {
		return false, pkgerrors.Wrap(err, "Get Namespace list error")
	}
	return ns != nil, nil
}

// DeleteNamespace is used to delete a namespace
func DeleteNamespace(namespace string, client *kubernetes.Clientset) error {
	deletePolicy := metaV1.DeletePropagationForeground

	err := client.CoreV1().Namespaces().Delete(namespace, &metaV1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})

	if err != nil {
		return pkgerrors.Wrap(err, "Delete Namespace error")
	}
	return nil
}
