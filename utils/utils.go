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
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	pkgerrors "github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/shank7485/k8-plugin-multicloud/krd"
)

// CSAR interface var used in handler
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
func (c *CSARFile) Download(destFileName string, url string) error {
	file, err := os.Create(destFileName)
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

// GetCSARFromURL to download the CSAR files from URL and extract it
// to get deployment and service yamls
// var GetCSARFromURL = func(csarID string, csarURL string) (*krd.KubernetesData, error) {

// 	// 1. Download CSAR file
// 	err := CSAR.Download("csar_file.zip", csarURL)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 2. Unzip CSAR file which will contain a deployment folder containining
// 	// deployment.yaml and service.yaml.
// 	err = CSAR.Unzip("csar_file.zip")
// 	if err != nil {
// 		return nil, err
// 	}

// 	kubeData := &krd.KubernetesData{}

// 	// 3. Read the deployment.yaml and service.yaml and set the deployment and
// 	// service structs.
// 	err = kubeData.ReadDeploymentYAML(csarID + "/deployment.yaml")
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = kubeData.ReadServiceYAML(csarID + "/service.yaml")
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 4. Delete extracted directory.
// 	err = CSAR.Delete(csarID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 5. Delete CSAR file.
// 	err = CSAR.Delete("csar_file.zip")
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 6. Return kubeData struct for kubernetes client to spin things up.
// 	return kubeData, nil
// }

// ReadCSARFromFileSystem reads the CSAR files from the files system
var ReadCSARFromFileSystem = func(csarID string) (*krd.KubernetesData, error) {
	kubeData := &krd.KubernetesData{}
	var path string

	csarDirPath := os.Getenv("CSAR_DIR") + "/" + csarID
	sequenceYAMLPath := csarDirPath + "/sequence.yaml"

	dlist, slist, err := ReadSequenceFile(sequenceYAMLPath)
	if err != nil {
		return nil, errors.New("File " + sequenceYAMLPath + " does not exists")
	}

	for _, name := range dlist {
		path = csarDirPath + "/" + name

		_, err = os.Stat(path)
		if os.IsNotExist(err) {
			return nil, errors.New("File " + path + " does not exists")
		}

		log.Println("Processing file: " + path)
		err = kubeData.ReadDeploymentYAML(path)
		if err != nil {
			return nil, err
		}

	}

	for _, name := range slist {
		path = csarDirPath + "/" + name

		_, err = os.Stat(path)
		if os.IsNotExist(err) {
			return nil, errors.New("File " + path + " does not exists")
		}

		log.Println("Processing file: " + path)
		err = kubeData.ReadServiceYAML(path)
		if err != nil {
			return nil, err
		}

	}

	return kubeData, nil
}

// FileOrder stores the sequence of execution
type FileOrder struct {
	Dlist []string `yaml:"deployment"`
	SList []string `yaml:"service"`
}

// ReadSequenceFile reads the sequence yaml to return the order or reads
func ReadSequenceFile(yamlFilePath string) ([]string, []string, error) {
	var deploymentlist []string
	var servicelist []string

	var f FileOrder

	if _, err := os.Stat(yamlFilePath); err == nil {
		log.Println("Reading sequence YAML: " + yamlFilePath)
		rawBytes, err := ioutil.ReadFile(yamlFilePath)
		if err != nil {
			return deploymentlist, servicelist, pkgerrors.Wrap(err, "Sequence YAML file read error")
		}

		err = yaml.Unmarshal(rawBytes, &f)
		if err != nil {
			return deploymentlist, servicelist, pkgerrors.Wrap(err, "Sequence YAML file read error")
		}

		for _, name := range f.Dlist {
			deploymentlist = append(deploymentlist, name)
		}

		for _, name := range f.SList {
			servicelist = append(servicelist, name)
		}

	}

	return deploymentlist, servicelist, nil
}
