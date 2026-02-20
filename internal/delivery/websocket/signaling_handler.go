package websocket

import (
	"bincang-visual/internal/domain/entity"
	"bincang-visual/internal/domain/usecase"
	"bincang-visual/repository"
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
)

type SignalingHub struct {
	userRepo    repository.UserRepository
	rooms       map[string]map[string]*Client
	mu          sync.RWMutex
	roomUseCase *usecase.RoomUseCase
	register    chan *Client
	unregister  chan *Client
	broadcast   chan *BroadcastMessage
}

type Client struct {
	UserID      string
	RoomID      string
	DisplayName string
	Conn        *websocket.Conn
	Send        chan []byte
	Hub         *SignalingHub
}

type BroadcastMessage struct {
	RoomID  string
	Message []byte
	Exclude string
}

func NewSignalingHub(roomUseCase *usecase.RoomUseCase) *SignalingHub {
	return &SignalingHub{
		rooms:       make(map[string]map[string]*Client),
		roomUseCase: roomUseCase,
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan *BroadcastMessage, 256),
	}
}

func (h *SignalingHub) Run() {
	log.Println("[Hub] Starting SignalingHub...")
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

func (h *SignalingHub) registerClient(client *Client) {
	if client == nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.rooms[client.RoomID]; !exists {
		h.rooms[client.RoomID] = make(map[string]*Client)
		log.Printf("[Hub] Created new room: %s", client.RoomID)
	}

	h.rooms[client.RoomID][client.UserID] = client

	log.Printf("[Hub] Client registered: %s in room %s (Total: %d)",
		client.UserID, client.RoomID, len(h.rooms[client.RoomID]))

	h.notifyPeerJoined(client)
}

func (h *SignalingHub) unregisterClient(client *Client) {
	if client == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, exists := h.rooms[client.RoomID]; exists {
		if _, exists := room[client.UserID]; exists {
			delete(room, client.UserID)

			select {
			case <-client.Send:
			default:
				close(client.Send)
			}

			log.Printf("[Hub] Client unregistered: %s from room %s (Remaining: %d)",
				client.UserID, client.RoomID, len(room))

			currentSharer, _ := h.roomUseCase.GetScreenSharer(context.Background(), client.RoomID)
			if currentSharer == client.UserID {
				_ = h.roomUseCase.StopScreenShare(context.Background(), client.RoomID, client.UserID)
				notification := entity.SignalMessage{
					Type:   "screen-share",
					From:   client.UserID,
					RoomID: client.RoomID,
					Data: map[string]interface{}{
						"isSharing": false,
					},
				}
				data, _ := json.Marshal(notification)
				h.broadcast <- &BroadcastMessage{
					RoomID:  client.RoomID,
					Message: data,
				}
			}

			if len(room) == 0 {
				delete(h.rooms, client.RoomID)
				log.Printf("[Hub] Empty room deleted: %s", client.RoomID)
			}

			h.notifyPeerLeft(client)
		}
	}

	if client.UserID != "" {
		_ = h.roomUseCase.LeaveRoom(context.Background(), client.RoomID, client.UserID)
	}
}

func (h *SignalingHub) broadcastMessage(msg *BroadcastMessage) {
	if msg == nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()

	if room, exists := h.rooms[msg.RoomID]; exists {
		for clientID, client := range room {
			if clientID == msg.Exclude {
				continue
			}

			select {
			case client.Send <- msg.Message:
			default:
				log.Printf("[Hub] Failed to send to %s, channel full", clientID)
			}
		}
	}
}

func (h *SignalingHub) notifyPeerJoined(client *Client) {
	participants, err := h.roomUseCase.GetParticipants(context.Background(), client.RoomID)
	if err != nil {
		log.Printf("[Hub] Error getting participants: %v", err)
		return
	}

	var newParticipant *entity.Participant
	for _, p := range participants {
		if p.UserID == client.UserID {
			newParticipant = p
			break
		}
	}

	if newParticipant == nil {
		log.Printf("[Hub] New participant not found in DB")
		return
	}

	notification := entity.SignalMessage{
		Type:   "peer-joined",
		From:   client.UserID,
		RoomID: client.RoomID,
		Data: map[string]interface{}{
			"userId":      client.UserID,
			"displayName": client.DisplayName,
			"isHost":      false,
		},
	}

	// notification := entity.SignalMessage{
	// 	Type:   "peer-joined",
	// 	From:   client.ID,
	// 	RoomID: client.RoomID,
	// 	Data: map[string]interface{}{
	// 		"participantId": newParticipant.ID,
	// 		"userId":        newParticipant.UserID,
	// 		"displayName":   newParticipant.DisplayName,
	// 		"isHost":        newParticipant.IsHost,
	// 	},
	// }

	data, _ := json.Marshal(notification)
	h.broadcast <- &BroadcastMessage{
		RoomID:  client.RoomID,
		Message: data,
		Exclude: client.UserID,
	}
}

func (h *SignalingHub) notifyPeerLeft(client *Client) {
	notification := entity.SignalMessage{
		Type:   "peer-left",
		From:   client.UserID,
		RoomID: client.RoomID,
		Data: map[string]interface{}{
			"userId": client.UserID,
		},
	}

	data, _ := json.Marshal(notification)
	h.broadcast <- &BroadcastMessage{
		RoomID:  client.RoomID,
		Message: data,
	}
}

func (h *SignalingHub) HandleWebSocket(c *websocket.Conn, roomID, userID, clientID, displayName string) {
	if c == nil {
		log.Println("[WebSocket] Nil connection received")
		return
	}

	log.Printf("[WebSocket] New connection: clientID=%s, userID=%s, roomID=%s", clientID, userID, roomID)
	// user, err := h.roomUseCase.GetCurrentUser(context.Background(), userID)
	// if err != nil {
	// 	log.Println("[WebSocket] User not found/ unauthenticated")
	// 	return
	// }
	participant, err := h.roomUseCase.JoinRoom(context.Background(), usecase.JoinRoomInput{
		RoomID:      roomID,
		UserID:      userID,
		DisplayName: displayName,
	})

	if err != nil {
		log.Printf("[WebSocket] Error joining room: %v", err)
		c.WriteJSON(map[string]string{"error": "Failed to join room"})
		c.Close()
		return
	}

	log.Printf("[WebSocket] Participant joined: ID=%s", participant.UserID)

	client := &Client{
		UserID:      userID,
		RoomID:      roomID,
		DisplayName: displayName,
		Conn:        c,
		Send:        make(chan []byte, 256),
		Hub:         h,
	}

	h.register <- client

	log.Printf("[WebSocket] Starting pumps for client %s", clientID)

	go client.writePump()
	client.readPump()
}

// reads messages from WebSocket
func (c *Client) readPump() {
	defer func() {
		log.Printf("[WebSocket] ReadPump ending for client %s", c.UserID)
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	// configure connection
	c.Conn.SetReadDeadline(time.Now().Add(70 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		log.Printf("[WebSocket] Pong received from %s", c.UserID)
		c.Conn.SetReadDeadline(time.Now().Add(70 * time.Second))
		return nil
	})

	log.Printf("[WebSocket] ReadPump started for client %s", c.UserID)

	for {
		messageType, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				log.Printf("[WebSocket] Unexpected close for %s: %v", c.UserID, err)
			} else {
				log.Printf("[WebSocket] Client %s disconnected: %v", c.UserID, err)
			}
			break
		}

		if messageType != websocket.TextMessage {
			log.Printf("[WebSocket] Received non-text message from %s, type: %d", c.UserID, messageType)
			continue
		}

		var msg entity.SignalMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("[WebSocket] Failed to parse message from %s: %v. Raw: %s", c.UserID, err, string(message))

			continue
		}

		log.Printf("[WebSocket] Received '%s' from %s", msg.Type, c.UserID)

		msg.From = c.UserID
		msg.RoomID = c.RoomID
		c.handleMessage(&msg)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)

	defer func() {
		ticker.Stop()
		log.Printf("[WebSocket] WritePump ending for client %s", c.UserID)
		c.Conn.Close()
	}()

	log.Printf("[WebSocket] WritePump started for client %s", c.UserID)

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			if !ok {
				// Channel closed
				log.Printf("[WebSocket] Send channel closed for %s", c.UserID)
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("[WebSocket] Write error for %s: %v", c.UserID, err)
				return
			}

		case <-ticker.C:
			// Send ping
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("[WebSocket] Ping error for %s: %v", c.UserID, err)
				return
			}
			log.Printf("[WebSocket] Ping sent to %s", c.UserID)
		}
	}
}

func (c *Client) handleMessage(msg *entity.SignalMessage) {
	switch msg.Type {
	case "offer", "answer", "ice":
		c.forwardToPeer(msg)
	case "ping":
		c.handlePing()
	case "chat":
		c.handleChatMessage(msg)
	case "media-state":
		c.handleMediaState(msg)
	case "screen-share":
		c.forwardToPeer(msg)
	case "leave":
		c.handleLeave(msg)
	default:
		log.Printf("[WebSocket] 📢 Broadcasting message type '%s'", msg.Type)
		data, _ := json.Marshal(msg)
		c.Hub.broadcast <- &BroadcastMessage{
			RoomID:  c.RoomID,
			Message: data,
			Exclude: c.UserID,
		}
	}
}

func (c *Client) handleLeave(msg *entity.SignalMessage) {
	log.Printf("[WebSocket] Client %s leaving room %s", c.UserID, c.RoomID)

	notification := entity.SignalMessage{
		Type:   "peer-left",
		From:   c.UserID,
		RoomID: c.RoomID,
		Data: map[string]interface{}{
			"userId": c.UserID,
		},
	}

	data, _ := json.Marshal(notification)
	c.Hub.broadcast <- &BroadcastMessage{
		RoomID:  c.RoomID,
		Message: data,
	}

	_ = c.Hub.roomUseCase.LeaveRoom(context.Background(), c.RoomID, c.UserID)

	c.Hub.unregister <- c
	c.Conn.Close()
}

func (c *Client) forwardToPeer(msg *entity.SignalMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[WebSocket] Failed to marshal message: %v", err)
		return
	}

	if msg.To != "" {
		// send to specific peer
		c.Hub.mu.RLock()
		if room, exists := c.Hub.rooms[c.RoomID]; exists {
			if targetClient, exists := room[msg.To]; exists {
				select {
				case targetClient.Send <- data:
					log.Printf("[WebSocket] Forwarded %s to %s", msg.Type, msg.To)
				default:
					log.Printf("[WebSocket] Failed to forward to %s, channel full", msg.To)
				}
			} else {
				log.Printf("[WebSocket] Target client %s not found", msg.To)
			}
		}
		c.Hub.mu.RUnlock()
	} else {
		// broadcast
		c.Hub.broadcast <- &BroadcastMessage{
			RoomID:  c.RoomID,
			Message: data,
			Exclude: c.UserID,
		}
	}
}

func (c *Client) handlePing() {
	response := entity.SignalMessage{
		Type:   "pong",
		From:   "server",
		RoomID: c.RoomID,
	}
	data, _ := json.Marshal(response)
	select {
	case c.Send <- data:
		log.Printf("[WebSocket] Pong sent to %s", c.UserID)
	default:
		log.Printf("[WebSocket] Failed to send pong to %s", c.UserID)
	}
}

func (c *Client) handleChatMessage(msg *entity.SignalMessage) {
	messageText, ok := msg.Data["message"].(string)
	if !ok {
		log.Printf("[WebSocket] Invalid chat message format")
		return
	}

	participant, err := c.Hub.roomUseCase.GetParticipants(context.Background(), c.RoomID)
	if err != nil {
		log.Printf("[WebSocket] Error getting participants: %v", err)
		return
	}

	var userName string
	for _, p := range participant {
		if p.UserID == c.UserID {
			userName = p.DisplayName
			break
		}
	}

	chatMsg := &entity.ChatMessage{
		RoomID:   c.RoomID,
		UserID:   c.UserID,
		UserName: userName,
		Message:  messageText,
		Type:     "text",
	}

	if err := c.Hub.roomUseCase.SendChatMessage(context.Background(), chatMsg); err != nil {
		log.Printf("[WebSocket] Failed to save chat message: %v", err)
	}

	data, _ := json.Marshal(msg)
	c.Hub.broadcast <- &BroadcastMessage{
		RoomID:  c.RoomID,
		Message: data,
	}
}

func (c *Client) handleMediaState(msg *entity.SignalMessage) {
	participants, err := c.Hub.roomUseCase.GetParticipants(context.Background(), c.RoomID)
	if err != nil {
		log.Printf("[WebSocket] Error getting participants: %v", err)
		return
	}

	for _, p := range participants {
		if p.UserID == c.UserID {
			if muted, ok := msg.Data["isMuted"].(bool); ok {
				p.IsMuted = muted
			}
			if videoOff, ok := msg.Data["isVideoOff"].(bool); ok {
				p.IsVideoOff = videoOff
			}
			_ = c.Hub.roomUseCase.UpdateParticipantState(context.Background(), p)
			break
		}
	}

	data, _ := json.Marshal(msg)
	c.Hub.broadcast <- &BroadcastMessage{
		RoomID:  c.RoomID,
		Message: data,
		Exclude: c.UserID,
	}
}

func (h *SignalingHub) GetRoomClients(roomID string) map[string]*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if room, exists := h.rooms[roomID]; exists {
		clients := make(map[string]*Client)
		for k, v := range room {
			clients[k] = v
		}
		return clients
	}
	return nil
}

func (c *Client) handleScreenShare(msg *entity.SignalMessage) {
	isSharing := msg.Data["isSharing"].(bool)

	if isSharing {
		// request to START screen share
		err := c.Hub.roomUseCase.StartScreenShare(
			context.Background(),
			c.RoomID,
			c.UserID,
		)

		if err != nil {
			errorMsg := entity.SignalMessage{
				Type:   "screen-share-error",
				From:   "server",
				RoomID: c.RoomID,
				Data: map[string]interface{}{
					"error": err.Error(),
				},
			}

			data, _ := json.Marshal(errorMsg)
			select {
			case c.Send <- data:
				log.Printf("[WebSocket] Sent screen share error to %s: %s", c.UserID, err.Error())
			default:
			}
			return
		}
	} else {
		// request to STOP screen share
		err := c.Hub.roomUseCase.StopScreenShare(
			context.Background(),
			c.RoomID,
			c.UserID,
		)

		if err != nil {
			log.Printf("[WebSocket] Error stopping screen share: %v", err)
			return
		}
	}

	data, _ := json.Marshal(msg)
	c.Hub.broadcast <- &BroadcastMessage{
		RoomID:  c.RoomID,
		Message: data,
		Exclude: c.UserID,
	}
}

func (h *SignalingHub) Shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	log.Println("[Hub] Shutting down SignalingHub...")

	for roomID, room := range h.rooms {
		for clientID, client := range room {
			select {
			case <-client.Send:
			default:
				close(client.Send)
			}
			log.Printf("[Hub] Closed client %s in room %s", clientID, roomID)
		}
	}

	h.rooms = make(map[string]map[string]*Client)

	close(h.register)
	close(h.unregister)
	close(h.broadcast)

	log.Println("[Hub] SignalingHub shutdown complete")
}
