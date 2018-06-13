package clientConfig

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

var K8 K8ClientInitiator

func InitiateK8Client(kubeconfigPath string ) (*kubernetes.Clientset, error) {
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

/*
The following are just examples on how to use client.
*/
// func main() {
// 	client := k8.getClient()
// 	PrintAllPods(client)
// }

// func PrintAllPods(client *kubernetes.Clientset) {
// 	pods, err := client.CoreV1().Pods("").List(metav1.ListOptions{})
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	for _, pod := range pods.Items {
// 		log.Println("Container Name: " + pod.GetName())
// 	}
// }