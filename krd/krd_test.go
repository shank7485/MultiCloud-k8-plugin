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

package krd_test

// import (
// 	"testing"

// 	"github.com/shank7485/k8-plugin-multicloud/multicloud"
// 	"github.com/shank7485/k8-plugin-multicloud/krd"
// 	apps_v1 "k8s.io/api/apps/v1"
// 	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/types"
// 	"k8s.io/apimachinery/pkg/watch"
// )

// type FakeClient struct{}

// func (c *FakeClient) Create(deployment *apps_v1.Deployment) (result *apps_v1.Deployment, err error) {
// 	return
// }

// func (c *FakeClient) Update(*apps_v1.Deployment) (*apps_v1.Deployment, error) {
// 	return nil, nil
// }

// func (c *FakeClient) UpdateStatus(*apps_v1.Deployment) (*apps_v1.Deployment, error) {
// 	return nil, nil
// }

// func (c *FakeClient) Delete(name string, options *meta_v1.DeleteOptions) error {
// 	return nil
// }

// func (c *FakeClient) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
// 	return nil
// }

// func (c *FakeClient) Get(name string, options meta_v1.GetOptions) (*apps_v1.Deployment, error) {
// 	return nil, nil
// }

// func (c *FakeClient) List(opts meta_v1.ListOptions) (*apps_v1.DeploymentList, error) {
// 	return nil, nil
// }

// func (c *FakeClient) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
// 	return nil, nil
// }

// func (c *FakeClient) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *apps_v1.Deployment, err error) {
// 	return nil, nil
// }

// type FakeVNFInstanceResource struct {}

// func (f *FakeVNFInstanceResource) DownloadVNFDeployment(url string) {
// 	return nil
// }

// func TestVNFInstanceClientCreate(t *testing.T) {
// 	client := krd.VNFInstanceClient{
// 		Client: &FakeClient{},
// 	}
// 	testVNF := &FakeVNFInstanceResource{}

// 	err := client.Create(testVNF, "CSAR_URL")
// 	if err != nil {
// 		t.Fatalf("TestVNFInstanceClientCreate returned an error (%s)", err)
// 	}
// }
