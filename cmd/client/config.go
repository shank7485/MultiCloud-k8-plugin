package main

import (
	"errors"
	"flag"
	"log"
	"path/filepath"

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

var k8 K8ClientInitiator

func InitiateK8client(configPath string) error {
	k8 = &ConfigClient{
		config: &restclient.Config{},
		client: &kubernetes.Clientset{},
	}

	err := k8.setConfig(configPath)
	if err != nil {
		return err
	}

	err = k8.setClient()
	if err != nil {
		return err
	}
	return nil
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
		return err
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
