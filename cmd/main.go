package main

import (
	"log"

	"github.com/shank7485/k8-plugin-multicloud/cmd/deployment"
)

func main() {
	d := deployment.Deploy{}

	d.InitiateDeploymentClient("default")
	log.Println("[INFO] Deploying")
	d.CreateDeployment()
	d.GetDeployment()
	d.DeleteDeployment()

	// k8client, err := client.getClient()
}
