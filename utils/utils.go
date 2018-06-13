package utils

import (
	"io/ioutil"
	"net/http"

	pkgerrors "github.com/pkg/errors"
	appsV1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

var (
	GetDeployment = DownloadDeploymentInfo
)

// DownloadDeploymentInfo retrieves the YAML file from an external source
func DownloadDeploymentInfo(url string) (*appsV1.Deployment, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Get YAML file error")
	}
	defer resp.Body.Close()

	rawYAMLbytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Read Body error")
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(rawYAMLbytes, nil, nil)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Deserialize deployment error")
	}

	return obj.(*appsV1.Deployment), nil
}
