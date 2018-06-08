package deployment

import (
	"fmt"

	client "github.com/shank7485/k8-plugin-multicloud/cmd/clientConfig"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/typed/apps/v1"
)

type Deploy struct {
	Namespace         string // Default: apiv1.NamespaceDefault
	DeploymentsClient v1.DeploymentInterface
}

func (d *Deploy) InitiateDeploymentClient(namespace string) error {
	k8client, err := client.InitiateClient()
	if err != nil {
		return err
	}
	d.Namespace = namespace
	d.DeploymentsClient = k8client.AppsV1().Deployments(d.Namespace)
	return nil
}

func (d *Deploy) CreateDeployment() error {
	// https://github.com/kubernetes/client-go/blob/master/examples/create-update-delete-deployment/main.go
	deploymentYAMLStruct := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: func(i int32) *int32 {
				return &i
			}(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := d.DeploymentsClient.Create(deploymentYAMLStruct)
	if err != nil {
		return err
	}

	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
	return nil
}

func (d *Deploy) GetDeployment() error {
	list, err := d.DeploymentsClient.List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, d := range list.Items {
		fmt.Println(d.Name)
	}
	return nil
}

func (d *Deploy) DeleteDeployment() error {
	deletePolicy := metav1.DeletePropagationForeground
	err := d.DeploymentsClient.Delete("demo-deployment", &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})

	if err != nil {
		return err
	}
	return nil
}
