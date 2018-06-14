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

package utils

import (
	"io/ioutil"
	"net/http"

	pkgerrors "github.com/pkg/errors"
	appsV1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

var download = func(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Get Body error")
	}
	defer resp.Body.Close()

	rawBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Read Body error")
	}

	return rawBody, nil
}

// GetDeploymentInfo retrieves the YAML file from an external source
func GetDeploymentInfo(url string) (*appsV1.Deployment, error) {
	rawYAMLbytes, err := download(url)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Get YAML file error")
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(rawYAMLbytes, nil, nil)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Deserialize deployment error")
	}

	return obj.(*appsV1.Deployment), nil
}
