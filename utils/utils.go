package utils

import (
	"net/http"
	"io/ioutil"
	"k8s.io/client-go/kubernetes/scheme"
	appsV1 "k8s.io/api/apps/v1"
)

var (
	GetYAML = DownloadYAML
	ConvertYAMLtoStructs = ConvertToDeployment
)

func DownloadYAML(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rawYAMLbytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return rawYAMLbytes, nil
}

func ConvertToDeployment(rawYAMLbytes []byte) (*appsV1.Deployment, error){
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(rawYAMLbytes, nil, nil)
	if err != nil {
		return nil, err
	}

	return obj.(*appsV1.Deployment), nil
}