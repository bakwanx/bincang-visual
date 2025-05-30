package models

import "encoding/json"

type WebSocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type OfferPayload struct {
	To   string `json:"to,omitempty"`
	From string `json:"from,omitempty"`
}

type RequestOfferingPayload struct {
	RoomId          string `json:"roomId"`
	UsernameRequest string `json:"usernameRequest"`
}

type SdpPayload struct {
	Sdp     string `json:"sdp"`
	TypeSdp string `json:"typeSdp"`
	To      string `json:"to,omitempty"`
}

type IceCandidatePayload struct {
	Candidate     string `json:"candidate"`
	SdpMLineIndex int    `json:"sdpMLineIndex"`
	To            string `json:"to,omitempty"`
	SdpMid        string `json:"sdpMid"`
}

type LeaveMeetingPayload struct {
	RoomId   string `json:"roomId"`
	Username string `json:"username"`
}
