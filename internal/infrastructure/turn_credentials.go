package webrtc

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"time"
)

type TURNCredentials struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	TTL      int      `json:"ttl"`
	URIs     []string `json:"uris"`
}

func GenerateTURNCredentials(
	username string,
	secret string,
	ttl int,
	turnHost string,
	turnPort int,
) *TURNCredentials {
	timestamp := time.Now().Unix() + int64(ttl)

	turnUsername := fmt.Sprintf("%d:%s", timestamp, username)

	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(turnUsername))
	password := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return &TURNCredentials{
		Username: turnUsername,
		Password: password,
		TTL:      ttl,
		URIs: []string{
			fmt.Sprintf("turn:%s:%d?transport=udp", turnHost, turnPort),
			fmt.Sprintf("turn:%s:%d?transport=tcp", turnHost, turnPort),
		},
	}
}
