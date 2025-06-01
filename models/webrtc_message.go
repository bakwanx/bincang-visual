package models

import "encoding/json"

type WebSocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type OfferPayload struct {
	UserTarget User `json:"userTarget,omitempty"`
	UserFrom   User `json:"userFrom,omitempty"`
}

type RequestOfferingPayload struct {
	RoomId      string `json:"roomId"`
	UserRequest User   `json:"userRequest"`
}

type SdpPayload struct {
	Sdp        string `json:"sdp"`
	TypeSdp    string `json:"typeSdp"`
	UserTarget User   `json:"userTarget,omitempty"`
}

type IceCandidatePayload struct {
	Candidate     string `json:"candidate"`
	SdpMLineIndex int    `json:"sdpMLineIndex"`
	UserTarget    User   `json:"userTarget,omitempty"`
	SdpMid        string `json:"sdpMid"`
}

type LeaveMeetingPayload struct {
	RoomId string `json:"roomId"`
	User   User   `json:"user"`
}

type ChatPayload struct {
	RoomId   string `json:"roomId"`
	UserFrom User   `json:"userFrom,omitempty"`
	Message  string `json:"message,omitempty"`
}
