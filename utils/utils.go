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
	"archive/zip"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	pkgerrors "github.com/pkg/errors"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

var CSAR FileOperator

// FileOperator is an interface to Download, Extract and Delete files
type FileOperator interface {
	Download(string, string) error
	Unzip(string) error
	Delete(string) error
}

// CSARFile is a concrete type to Download, Extract and Delete CSAR files
type CSARFile struct{}

// Download a CSAR file
func (c *CSARFile) Download(fileName string, url string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return pkgerrors.Wrap(err, "CSAR file create error")
	}
	defer file.Close()

	resp, err := http.Get(url)
	if err != nil {
		return pkgerrors.Wrap(err, "CSAR file download error")
	}

	if resp.StatusCode != http.StatusOK {
		return pkgerrors.Wrap(err, "CSAR file download error")
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return pkgerrors.Wrap(err, "CSAR file encode error")
	}

	return nil
}

// Unzip a CSAR file
func (c *CSARFile) Unzip(filePath string) error {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return pkgerrors.Wrap(err, "CSAR file extracting error")
	}

	for _, file := range reader.File {
		path := filepath.Join(file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return pkgerrors.Wrap(err, "CSAR file extracting error")
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return pkgerrors.Wrap(err, "CSAR file extracting error")
		}
		defer targetFile.Close()

		_, err = io.Copy(targetFile, fileReader)
		if err != nil {
			return pkgerrors.Wrap(err, "CSAR file extracting error")
		}
	}
	return nil
}

// Delete a CSAR file
func (c *CSARFile) Delete(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return pkgerrors.Wrap(err, "CSAR file delete error")
	}
	return nil
}

// CreateKubeObjectsFromCSAR to download the CSAR files from URL and extract it
// to get deployment and service yamls
var CreateKubeObjectsFromCSAR = func(csarID string, csarURL string) (*KubernetesData, error) {

	// 1. Download CSAR file
	err := CSAR.Download("csar_file.zip", csarURL)
	if err != nil {
		return nil, err
	}

	// 2. Unzip CSAR file which will contain a deployment folder containining
	// deployment.yaml and service.yaml.
	err = CSAR.Unzip("csar_file.zip")
	if err != nil {
		return nil, err
	}

	kubeData := &KubernetesData{}

	// 3. Read the deployment.yaml and service.yaml and set the deployment and
	// service structs.
	err = kubeData.ReadDeploymentYAML(csarID + "/deployment.yaml")
	if err != nil {
		return nil, err
	}
	err = kubeData.ReadServiceYAML(csarID + "/service.yaml")
	if err != nil {
		return nil, err
	}

	// 4. Delete extracted directory.
	err = CSAR.Delete(csarID)
	if err != nil {
		return nil, err
	}

	// 5. Delete CSAR file.
	err = CSAR.Delete("csar_file.zip")
	if err != nil {
		return nil, err
	}

	// 6. Return kubeData struct for kubernetes client to spin things up.
	return kubeData, nil
}

// CSARKubeParser is an interface to parse both Deployment and Services
// yaml files
type CSARKubeParser interface {
	ReadDeploymentYAML(string) error
	ReadServiceYAML(string) error
	ParseDeploymentInfo() error
	ParseServiceInfo() error
}

// KubernetesData to store CSAR information including both services and
// deployments
type KubernetesData struct {
	DeploymentData []byte
	ServiceData    []byte
	Deployment     *appsV1.Deployment
	Service        *coreV1.Service
}

// ReadDeploymentYAML reads deployment.yaml and stores in CSARData struct
func (c *KubernetesData) ReadDeploymentYAML(yamlFilePath string) error {
	if _, err := os.Stat(yamlFilePath); err == nil {
		rawBytes, err := ioutil.ReadFile(yamlFilePath)
		if err != nil {
			return pkgerrors.Wrap(err, "Deployment YAML file read error")
		}

		c.DeploymentData = rawBytes

		err = c.ParseDeploymentInfo()
		if err != nil {
			return err
		}
	}
	return nil
}

// ReadServiceYAML reads service.yaml and stores in CSARData struct
func (c *KubernetesData) ReadServiceYAML(yamlFilePath string) error {
	if _, err := os.Stat(yamlFilePath); err == nil {
		rawBytes, err := ioutil.ReadFile(yamlFilePath)
		if err != nil {
			return pkgerrors.Wrap(err, "Service YAML file read error")
		}

		c.ServiceData = rawBytes

		err = c.ParseServiceInfo()
		if err != nil {
			return err
		}
	}
	return nil
}

// ParseDeploymentInfo retrieves the deployment YAML file from a CSAR
func (c *KubernetesData) ParseDeploymentInfo() error {
	if c.DeploymentData != nil {
		log.Println("Decoding deployment YAML")

		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, _, err := decode(c.DeploymentData, nil, nil)
		if err != nil {
			return pkgerrors.Wrap(err, "Deserialize deployment error")
		}

		switch o := obj.(type) {
		case *appsV1.Deployment:
			c.Deployment = o
			return nil
		}
	}
	return nil
}

// ParseServiceInfo retrieves the service YAML file from a CSAR
func (c *KubernetesData) ParseServiceInfo() error {
	if c.ServiceData != nil {
		log.Println("Decoding service YAML")

		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, _, err := decode(c.ServiceData, nil, nil)
		if err != nil {
			return pkgerrors.Wrap(err, "Deserialize deployment error")
		}

		switch o := obj.(type) {
		case *coreV1.Service:
			c.Service = o
			return nil
		}
	}
	return nil
}
