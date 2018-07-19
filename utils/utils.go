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
	"log"
	"net/http"
	"os"
	"path/filepath"

	pkgerrors "github.com/pkg/errors"

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

	path = os.Getenv("CSAR_DIR") + "/" + csarID + "/deployment.yaml" // Remove utils path

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, errors.New("File " + path + "does not exists")
	}

	log.Println("deployment file: " + path)
	err = kubeData.ReadDeploymentYAML(path)
	if err != nil {
		return nil, err
	}

	path = os.Getenv("CSAR_DIR") + "/" + csarID + "/service.yaml" // Remove utils path

	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		return nil, errors.New("File " + path + " does not exists")
	}

	log.Println("service file: " + path)
	err = kubeData.ReadServiceYAML(path)
	if err != nil {
		return nil, err
	}

	return kubeData, nil
}
