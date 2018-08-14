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

package csarparser

import (
	"encoding/json"
	"github.com/shank7485/k8-plugin-multicloud/plugins"
	"io/ioutil"
	"log"
	"os"

	pkgerrors "github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/shank7485/k8-plugin-multicloud/krd"
)

// CreateVNF reads the CSAR files from the files system and creates them one by one
var CreateVNF = func(csarID string, cloudRegionID string, namespace string) (string, map[string][]string, error) {
	var path string

	// uuid
	externalVNFID := string(uuid.NewUUID())

	// cloud1-default-uuid
	internalVNFID := cloudRegionID + "-" + namespace + "-" + externalVNFID

	csarDirPath := os.Getenv("CSAR_DIR") + "/" + csarID
	sequenceYAMLPath := csarDirPath + "/sequence.yaml"

	seqFile, err := ReadSequenceFile(sequenceYAMLPath)
	if err != nil {
		return "", nil, pkgerrors.Wrap(err, "Error while reading Sequence File: "+sequenceYAMLPath)
	}

	var resourceYAMLNameMap map[string][]string

	for _, resource := range seqFile.ResourceTypePathMap {
		for resourceName, resourceFileNames := range resource {
			switch resourceName {
			case "deployment":
				// Load/Use Deployment data/client
				var deployNameList []string

				for _, filename := range resourceFileNames {
					path = csarDirPath + "/" + filename

					_, err = os.Stat(path)
					if os.IsNotExist(err) {
						return "", nil, pkgerrors.New("File " + path + "does not exists")
					}

					log.Println("Processing file: " + path)

					typePlugin := plugins.LoadedPlugins["deployment"]

					symDeploymentData, err := typePlugin.Lookup("DeploymentData")
					if err != nil {
						return "", nil, pkgerrors.Wrap(err, "Error fetching "+resourceName+"plugin data")
					}

					// Type assert to a concrete plugin type. Should this be taken from
					// plugins or from reference copy of plugin type placed in KRD?
					deploymentData, ok := symDeploymentData.(plugins.KubeDeploymentData)
					if !ok {
						return "", nil, pkgerrors.New("Error loading " + resourceName + " plugin data")
					}

					err = deploymentData.ReadYAML(path)
					if err != nil {
						return "", nil, pkgerrors.Wrap(err, "Error Parsing "+filename+".yaml")
					}

					err = deploymentData.ParseYAML()
					if err != nil {
						return "", nil, pkgerrors.Wrap(err, "Error Reading "+filename+".yaml")
					}

					if deploymentData.Deployment == nil {
						return "", nil, pkgerrors.New("Read deploymentData.Deployment error")
					}

					// cloud1-default-uuid-sisedeploy
					internalDeploymentName := internalVNFID + "-" + deploymentData.Deployment.Name

					deploymentData.Deployment.Namespace = namespace
					deploymentData.Deployment.Name = internalDeploymentName

					symDeploymentClient, err := typePlugin.Lookup("KubeDeploymentClient")
					if err != nil {
						return "", nil, pkgerrors.Wrap(err, "Error fetching "+resourceName+"plugin client")
					}

					deploymentClient, ok := symDeploymentClient.(krd.KubeResourceClient)
					if !ok {
						return "", nil, pkgerrors.New("Error loading " + resourceName + " plugin client")
					}

					_, err = deploymentClient.CreateResource(deploymentData, namespace)
					if err != nil {
						return "", nil, pkgerrors.Wrap(err, "Error creating "+resourceName)
					}

					// ["cloud1-default-uuid-sisedeploy1", "cloud1-default-uuid-sisedeploy2", ... ]
					deployNameList = append(deployNameList, internalDeploymentName)

					/*
						{
							"deployment": ["cloud1-default-uuid-sisedeploy1", "cloud1-default-uuid-sisedeploy2", ... ]
						}
					*/
					resourceYAMLNameMap[resourceName] = deployNameList
				}
			case "service":
				// Load/Use Service data/client
				var serviceNameList []string

				for _, filename := range resourceFileNames {
					path = csarDirPath + "/" + filename

					_, err = os.Stat(path)
					if os.IsNotExist(err) {
						return "", nil, pkgerrors.New("File " + path + "does not exists")
					}

					log.Println("Processing file: " + path)

					typePlugin := plugins.LoadedPlugins["service"]

					symDeploymentData, err := typePlugin.Lookup("ServiceData")
					if err != nil {
						return "", nil, pkgerrors.Wrap(err, "Error fetching "+resourceName+"plugin data")
					}

					serviceData, ok := symDeploymentData.(plugins.KubeServiceData)
					if !ok {
						return "", nil, pkgerrors.New("Error loading " + resourceName + " plugin data")
					}

					err = serviceData.ReadYAML(path)
					if err != nil {
						return "", nil, pkgerrors.Wrap(err, "Error Parsing "+filename+".yaml")
					}

					err = serviceData.ParseYAML()
					if err != nil {
						return "", nil, pkgerrors.Wrap(err, "Error Reading "+filename+".yaml")
					}

					if serviceData.Service == nil {
						return "", nil, pkgerrors.New("Read serviceData.Service error")
					}

					// cloud1-default-uuid-sisesvc
					internalServiceName := internalVNFID + "-" + serviceData.Service.Name

					serviceData.Service.Namespace = namespace
					serviceData.Service.Name = internalServiceName

					symDeploymentClient, err := typePlugin.Lookup("KubeServiceClient")
					if err != nil {
						return "", nil, pkgerrors.Wrap(err, "Error fetching "+resourceName+"plugin client")
					}

					serviceClient, ok := symDeploymentClient.(krd.KubeResourceClient)
					if !ok {
						return "", nil, pkgerrors.New("Error loading " + resourceName + " plugin client")
					}

					_, err = serviceClient.CreateResource(serviceData, namespace)
					if err != nil {
						return "", nil, pkgerrors.Wrap(err, "Error creating "+resourceName)
					}

					// ["cloud1-default-uuid-sisesvc1", "cloud1-default-uuid-sisesvc2", ... ]
					serviceNameList = append(serviceNameList, internalServiceName)

					/*
						{
							"service": ["cloud1-default-uuid-sisesvc1", "cloud1-default-uuid-sisesvc2", ... ]
						}
					*/
					resourceYAMLNameMap[resourceName] = serviceNameList
				}
			default:
				return "", nil, pkgerrors.New(resourceName + " resource type not supported.")
			}
		}
	}

	/*
		uuid,
		{
			"deployment": ["cloud1-default-uuid-sisedeploy1", "cloud1-default-uuid-sisedeploy2", ... ]
			"service": ["cloud1-default-uuid-sisesvc1", "cloud1-default-uuid-sisesvc2", ... ]
		},
		nil
	*/
	return externalVNFID, resourceYAMLNameMap, nil
}

// DestroyVNF deletes VNFs based on data passed
var DestroyVNF = func(data map[string][]string, namespace string) error {
	/*
		{
			"deployment": ["cloud1-default-uuid-sisedeploy1", "cloud1-default-uuid-sisedeploy2", ... ]
			"service": ["cloud1-default-uuid-sisesvc1", "cloud1-default-uuid-sisesvc2", ... ]
		},
	*/
	for resourceName, resourceList := range data {
		switch resourceName {
		case "deployment":
			typePlugin := plugins.LoadedPlugins["deployment"]
			symDeploymentClient, err := typePlugin.Lookup("KubeDeploymentClient")
			if err != nil {
				return pkgerrors.Wrap(err, "Error fetching "+resourceName+"plugin client")
			}

			deploymentClient, ok := symDeploymentClient.(krd.KubeResourceClient)
			if !ok {
				return pkgerrors.New("Error loading " + resourceName + " plugin client")
			}

			for _, deploymentName := range resourceList {
				err = deploymentClient.DeleteResource(deploymentName, namespace)
				if err != nil {
					return pkgerrors.Wrap(err, "Error destroying "+deploymentName)
				}
			}

		case "service":
			typePlugin := plugins.LoadedPlugins["service"]
			symServiceClient, err := typePlugin.Lookup("KubeServiceClient")
			if err != nil {
				return pkgerrors.Wrap(err, "Error fetching "+resourceName+"plugin client")
			}

			serviceClient, ok := symServiceClient.(krd.KubeResourceClient)
			if !ok {
				return pkgerrors.New("Error loading " + resourceName + " plugin client")
			}

			for _, serviceName := range resourceList {
				err = serviceClient.DeleteResource(serviceName, namespace)
				if err != nil {
					return pkgerrors.Wrap(err, "Error destroying "+serviceName)
				}
			}
		default:
			return pkgerrors.New("Error unsupported " + resourceName)
		}
	}
	return nil
}

// SequenceFile stores the sequence of execution
type SequenceFile struct {
	ResourceTypePathMap []map[string][]string `yaml:"resources"`
}

// ReadSequenceFile reads the sequence yaml to return the order or reads
func ReadSequenceFile(yamlFilePath string) (SequenceFile, error) {
	var seqFile SequenceFile

	if _, err := os.Stat(yamlFilePath); err == nil {
		log.Println("Reading sequence YAML: " + yamlFilePath)
		rawBytes, err := ioutil.ReadFile(yamlFilePath)
		if err != nil {
			return seqFile, pkgerrors.Wrap(err, "Sequence YAML file read error")
		}

		err = yaml.Unmarshal(rawBytes, &seqFile)
		if err != nil {
			return seqFile, pkgerrors.Wrap(err, "Sequence YAML file read error")
		}
	}

	return seqFile, nil
}

// SerializeMap serializes map[string][]string into a string
func SerializeMap(data map[string][]string) (string, error) {
	/*
		IP:
		{
			"deployment": ["cloud1-default-uuid-sisedeploy1", "cloud1-default-uuid-sisedeploy2", ... ]
			"service": ["cloud1-default-uuid-sisesvc1", "cloud1-default-uuid-sisesvc2", ... ]
		}
		OP:
		// "{"deployment":<>,"service":<>}"
	*/
	out, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// DeSerializeMap deserializes string into map[string][]string
func DeSerializeMap(serialData string) (map[string][]string, error) {
	/*
		IP:
		// "{"deployment":<>,"service":<>}"
		OP:
		{
			"deployment": ["cloud1-default-uuid-sisedeploy1", "cloud1-default-uuid-sisedeploy2", ... ]
			"service": ["cloud1-default-uuid-sisesvc1", "cloud1-default-uuid-sisesvc2", ... ]
		}
	*/
	var result map[string][]string
	json.Unmarshal([]byte(serialData), result)

	return result, nil
}
