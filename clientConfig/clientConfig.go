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

package clientConfig

import (
	"errors"
	"flag"
	"log"
	"path/filepath"

	pkgerrors "github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type K8ClientInitiator interface {
	setConfig(string) error
	setClient() error
	getClient() *kubernetes.Clientset
}

type ConfigClient struct {
	config *restclient.Config
	client *kubernetes.Clientset
}

var K8 K8ClientInitiator

func InitiateK8Client(kubeconfigPath string) (*kubernetes.Clientset, error) {
	K8 = &ConfigClient{
		config: &restclient.Config{},
		client: &kubernetes.Clientset{},
	}

	err := K8.setConfig(kubeconfigPath)
	if err != nil {
		return nil, err
	}

	err = K8.setClient()
	if err != nil {
		return nil, err
	}

	return K8.getClient(), nil
}

func (kc *ConfigClient) setConfig(configPath string) error {
	home := homedir.HomeDir()
	var kubeconfig *string

	if configPath == "" && home == "" {
		return errors.New("config not passed and is not found in ~/.kube. ")
	}

	if home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", configPath, "absolute path to the kubeconfig file")
	}
	log.Println("[INFO] Config: ", *kubeconfig)
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return pkgerrors.Wrap(err, "setConfig: Build config from flags raised an error")
	}

	kc.config = config
	return nil
}

func (kc *ConfigClient) setClient() error {
	clientset, err := kubernetes.NewForConfig(kc.config)
	if err != nil {
		return err
	}
	kc.client = clientset
	return nil
}

func (kc *ConfigClient) getClient() *kubernetes.Clientset {
	return kc.client
}
