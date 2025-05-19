package models

import "encoding/json"

type WebSocketMessage struct {
	Type    string          `json:"type"`
	To      string          `json:"to,omitempty"`
	From    string          `json:"from,omitempty"`
	Payload json.RawMessage `json:"payload"`
}

type RequestOffering struct {
	RoomId          string `json:"roomId"`
	UsernameRequest string `json:"usernameRequest"`
}

type SdpPayload struct {
	Sdp     string `json:"sdp"`
	TypeSdp string `json:"typeSdp"`
}

type IceCandidatePayload struct {
	Candidate     string `json:"candidate"`
	SdpMLineIndex int    `json:"sdpMLineIndex"`
	SdpMid        string `json:"sdpMid"`
}

type LeaveMeeting struct {
	RoomId   string `json:"roomId"`
	Username string `json:"username"`
}
