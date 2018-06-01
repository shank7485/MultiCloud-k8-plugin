package main

import (
	"fmt"
	"log"

	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	err := InitiateK8client("")
	if err != nil {
		log.Panic(err.Error())
	}
}

func main() {
	client := k8.get_client()
	print_all_pods(client)
}

func print_all_pods(client *kubernetes.Clientset) {
	pods, err := client.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, pod := range pods.Items {
		fmt.Println("Container Name: " + pod.GetName())
	}
}

func sendYAML(client *kubernetes.Clientset){
	
}
