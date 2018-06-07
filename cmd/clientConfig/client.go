package clientConfig

import (
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func InitiateClient() (*kubernetes.Clientset, error) {
	err := initiateK8client("")
	if err != nil {
		return nil, err
	}
	return k8.getClient(), nil
}

/*
The following are just examples on how to use client.
*/
func main() {
	client := k8.getClient()
	PrintAllPods(client)
}

func PrintAllPods(client *kubernetes.Clientset) {
	pods, err := client.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, pod := range pods.Items {
		log.Println("Container Name: " + pod.GetName())
	}
}
