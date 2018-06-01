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
	set_config(string) error
	set_client() error
	get_client() *kubernetes.Clientset
}

type ConfigClient struct {
	config *restclient.Config
	client *kubernetes.Clientset
}

var k8 K8ClientInitiator

func InitiateK8client(config_path string) error {
	k8 = &ConfigClient{
		config: &restclient.Config{},
		client: &kubernetes.Clientset{},
	}

	err := k8.set_config(config_path)
	if err != nil {
		return err
	}

	err = k8.set_client()
	if err != nil {
		return err
	}
	return nil
}

func (kc *ConfigClient) set_config(config_path string) error {
	home := homedir.HomeDir()
	var kubeconfig *string

	if config_path == "" && home == "" {
		return errors.New("config not passed and is not found in ~/.kube. ")
	}

	if home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", config_path, "absolute path to the kubeconfig file")
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

func (kc *ConfigClient) set_client() error {
	clientset, err := kubernetes.NewForConfig(kc.config)
	if err != nil {
		return err
	}
	kc.client = clientset
	return nil
}

func (kc *ConfigClient) get_client() *kubernetes.Clientset {
	return kc.client
}
