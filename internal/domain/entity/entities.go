package entity

import (
	"bincang-visual/internal/config"
	"time"
)

type Room struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	HostID          string                 `json:"hostId"`
	CreatedAt       time.Time              `json:"createdAt"`
	MaxParticipants int                    `json:"maxParticipants"`
	IsRecording     bool                   `json:"isRecording"`
	Settings        RoomSettings           `json:"settings"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

type RoomSettings struct {
	AllowScreenShare bool `json:"allowScreenShare"`
	AllowChat        bool `json:"allowChat"`
	WaitingRoom      bool `json:"waitingRoom"`
	RecordingEnabled bool `json:"recordingEnabled"`
	MaxDuration      int  `json:"maxDuration"` // in minutes
}

type Participant struct {
	UserID        string    `json:"userId"`
	RoomID        string    `json:"roomId"`
	DisplayName   string    `json:"displayName"`
	JoinedAt      time.Time `json:"joinedAt"`
	IsHost        bool      `json:"isHost"`
	IsMuted       bool      `json:"isMuted"`
	IsVideoOff    bool      `json:"isVideoOff"`
	IsScreenShare bool      `json:"isScreenShare"`
}

type ChatMessage struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"roomId"`
	UserID    string    `json:"userId"`
	UserName  string    `json:"userName"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // text, file, system
}

type SignalMessage struct {
	Type      string                 `json:"type"` // offer, answer, ice, join, leave, chat, etc.
	From      string                 `json:"from"`
	To        string                 `json:"to,omitempty"`
	RoomID    string                 `json:"roomId"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

type Recording struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"roomId"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime,omitempty"`
	Duration  int       `json:"duration"` // in seconds
	FileURL   string    `json:"fileUrl"`
	Status    string    `json:"status"` // recording, processing, completed, failed
	Chunks    []string  `json:"chunks"`
	Size      int64     `json:"size"` // in bytes
}

type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	PhotoURL    string    `json:"photoUrl,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

type CalendarEvent struct {
	ID            string    `json:"id"`
	RoomID        string    `json:"roomId"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	StartTime     time.Time `json:"startTime"`
	EndTime       time.Time `json:"endTime"`
	Attendees     []string  `json:"attendees"`
	CreatorID     string    `json:"creatorId"`
	GoogleEventID string    `json:"googleEventId,omitempty"`
}

type RoomConfig struct {
	ICEServers       []config.TurnStunConfig `json:"iceServers"`
	MaxBitrate       int                     `json:"maxBitrate"`
	CodecPreferences []string                `json:"codecPreferences,omitempty"`
}
