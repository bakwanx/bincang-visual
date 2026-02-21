package models

import "encoding/json"

func UnmarshalCoturnConfiguration(data []byte) (CoturnConfiguration, error) {
	var r CoturnConfiguration
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *CoturnConfiguration) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type CoturnConfiguration struct {
	IceServers []IceServerConfiguration `json:"iceServers"`
}

type IceServerConfiguration struct {
	Urls       []string `json:"urls"`
	Username   *string  `json:"username,omitempty"`
	Credential *string  `json:"credential,omitempty"`
}
