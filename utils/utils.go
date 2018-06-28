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
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

// Download the raw CSAR file from a specific URL
var Download = func(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Get CSAR file error")
	}
	defer resp.Body.Close()

	rawBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Read CSAR file error")
	}

	return rawBody, nil
}

// DownloadCSAR to download the CSAR files form URL and parse it
// to get deployment and service yamls
var DownloadCSAR = func(url string) (*CSARData, error) {
	rawFile, err := Download(url)
	if err != nil {
		return nil, err
	}

	// This will be changed to whatever is actually present in a CSAR
	csar := &CSARData{
		CSARdata: rawFile,
	}

	err = csar.ParseDeploymentInfo()
	if err != nil {
		return nil, err
	}

	err = csar.ParseServiceInfo()
	if err != nil {
		return nil, err
	}

	return csar, nil
}

// CSARParser is an interface to parse both Deployment and Services
// yaml files
type CSARParser interface {
	ParseDeploymentInfo() error
	ParseServiceInfo() error
}

// CSARData to store CSAR information including both services and
// deployments
type CSARData struct {
	CSARdata   []byte // Change this to whatever type read CSAR files have
	Deployment *appsV1.Deployment
	Service    *coreV1.Service
}

// ParseDeploymentInfo retrieves the deployment YAML file from a CSAR
func (c *CSARData) ParseDeploymentInfo() error {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(c.CSARdata, nil, nil)
	if err != nil {
		return pkgerrors.Wrap(err, "Deserialize deployment error")
	}

	switch o := obj.(type) {
	case *appsV1.Deployment:
		c.Deployment = o
		return nil
	default:
		return pkgerrors.Wrap(err, "Unknown type in Yaml")
	}
}

// ParseServiceInfo retrieves the service YAML file from a CSAR
func (c *CSARData) ParseServiceInfo() error {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(c.CSARdata, nil, nil)
	if err != nil {
		return pkgerrors.Wrap(err, "Deserialize service error")
	}
	switch o := obj.(type) {
	case *coreV1.Service:
		c.Service = o
		return nil
	default:
		return pkgerrors.Wrap(err, "Unknown type in Yaml")
	}
}
