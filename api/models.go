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

package api

// CreateVNFRequest contains the fields required for the VNF Creation
type CreateVNFRequest struct {
	CsarArtificateID  string `json:"csar_artificate_id"`
	CsarArtificateURL string
	OOFParams         OOFParameters `json:"oof_parameters"`
	InstanceID        string        `json:"instance_id"`
}

// OOFParameters contains additional information required for the VNF instance
type OOFParameters struct {
	KeyValues map[string]string `json:"key_values"`
}

// GeneralResponse is a generic response
type GeneralResponse struct {
	Response string `json:"response"`
}
