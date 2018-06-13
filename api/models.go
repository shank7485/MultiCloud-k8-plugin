package api

type CreateVNFRequest struct {
	CsarArtificateID  string `json:"csar_artificate_id"`
	CsarArtificateURL string
	OOFParams         OOFParameters `json:"oof_parameters"`
	InstanceID        string        `json:"instance_id"`
}

type OOFParameters struct {
	KeyValues map[string]string `json:"key_values"`
}

type GeneralResponse struct {
	Response string `json:"response"`
}
