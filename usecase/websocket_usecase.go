package usecase

import (
	ds "bincang-visual/datasource"
	"bincang-visual/repository"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	model "bincang-visual/models"

	"github.com/gofiber/contrib/websocket"
)

var lock = sync.Mutex{}

type WebsocketUsecase struct {
	userRepo repository.UserRepository
	roomRepo repository.RoomRepository
}

func NewWebsocketUsecase(userRepo repository.UserRepository, roomRepo repository.RoomRepository) *WebsocketUsecase {
	return &WebsocketUsecase{
		userRepo: userRepo,
		roomRepo: roomRepo,
	}
}

func (u *WebsocketUsecase) InitWebsocket(c *websocket.Conn) {
	var userId = c.Query("userId")
	var roomId = c.Query("roomId")
	userModel, err := u.userRepo.GetUser(userId)

	if err != nil {
		fmt.Println(userId)
		fmt.Println("pesan 0", err)
	}

	lock.Lock()
	ds.Clients[userId] = &model.UserClient{
		ID:   userId,
		Conn: c,
	}
	// if ds.Rooms[roomId] == nil {
	// 	ds.Rooms[roomId] = map[string]model.User{}
	// }

	if err := u.roomRepo.AddUser(roomId, *userModel); err != nil {
		fmt.Println("[ERROR] error adding user: ", err)
	}
	// ds.Rooms[roomId][userId] = ds.Users[userId]
	lock.Unlock()

	u.mainLogic(userId, roomId, c)
}

func (u *WebsocketUsecase) mainLogic(userId, roomId string, c *websocket.Conn) {
	for {
		mt, msg, err := c.ReadMessage()
		isBreak := u.onDisconnect(userId, roomId, mt, msg, err)
		if isBreak {
			break
		}

		lock.Lock()

		var webSocketMessage model.WebSocketMessage

		if err := json.Unmarshal(msg, &webSocketMessage); err != nil {
			fmt.Println("Error unmarshalling SocketMessage: ", err)
		}

		switch webSocketMessage.Type {
		case "join":
			u.onJoin(userId, roomId, webSocketMessage, mt, msg, err)
		case "offer":
			u.onOffer(webSocketMessage, mt, msg, err)
		case "answer":
			u.onAnswer(webSocketMessage, mt, msg, err)
		case "candidate":
			u.onCandidate(webSocketMessage, mt, msg, err)
		case "chat":
			u.onChat(webSocketMessage, mt, msg, err)
		case "leave":
			u.onLeave(webSocketMessage, mt, msg, err)
		}

		lock.Unlock()
	}
}

func (u *WebsocketUsecase) onDisconnect(userId, roomId string, mt int, msg []byte, err error) bool {
	if err != nil {
		lock.Lock()
		delete(ds.Clients, userId)

		// remove user in room
		if err := u.roomRepo.RemoveUser(roomId, userId); err != nil {
			fmt.Println("[ERROR] error remove user from room: ", err)
		}

		// delete(ds.Users, userId)
		// remove user in users
		if err := u.userRepo.RemoveUser(userId); err != nil {
			fmt.Println("[ERROR] error remove user from users: ", err)
		}

		lock.Unlock()
		return true
	}
	return false
}

func (u *WebsocketUsecase) onJoin(userId, roomId string, webSocketMessage model.WebSocketMessage, mt int, msg []byte, err error) {
	var requestOffering model.RequestOfferingPayload
	if err := json.Unmarshal(webSocketMessage.Payload, &requestOffering); err != nil {
		fmt.Println("Error unmarshalling Join: ", err)
	}
	room, err := u.roomRepo.GetRoom(requestOffering.RoomId)
	if err != nil {
		fmt.Println(err)
	}
	if len(room.Users) > 1 {
		for _, user := range room.Users {
			if user.ID != requestOffering.UserRequest.ID {
				if ds.Clients[user.ID].Conn != nil {
					if err = ds.Clients[user.ID].Conn.WriteMessage(mt, msg); err != nil {
						log.Println("error send message:", err)
						ds.Clients[user.ID].Conn.Close()
						delete(ds.Clients, user.ID)
					}
				}
			}
		}
	}
}

func (u *WebsocketUsecase) onOffer(webSocketMessage model.WebSocketMessage, mt int, msg []byte, err error) {
	var sdpPayload = model.SdpPayload{}
	err = json.Unmarshal(webSocketMessage.Payload, &sdpPayload)

	if err = ds.Clients[sdpPayload.UserTarget.ID].Conn.WriteMessage(mt, msg); err != nil {
		log.Println("error send message:", err)
		ds.Clients[sdpPayload.UserTarget.ID].Conn.Close()
		delete(ds.Clients, sdpPayload.UserTarget.ID)
	}

}

func (r *WebsocketUsecase) onAnswer(webSocketMessage model.WebSocketMessage, mt int, msg []byte, err error) {
	var sdpPayload = model.SdpPayload{}
	if err = json.Unmarshal(webSocketMessage.Payload, &sdpPayload); err != nil {
		fmt.Println("Error unmarshalling sdp payload: ", err)
	}
	if err := ds.Clients[sdpPayload.UserTarget.ID].Conn.WriteMessage(mt, msg); err != nil {
		log.Println("error send message:", err)
		ds.Clients[sdpPayload.UserTarget.ID].Conn.Close()
		delete(ds.Clients, sdpPayload.UserTarget.ID)
	}
}

func (r *WebsocketUsecase) onCandidate(webSocketMessage model.WebSocketMessage, mt int, msg []byte, err error) {
	var iceCandidatePayload = model.IceCandidatePayload{}
	if err := json.Unmarshal(webSocketMessage.Payload, &iceCandidatePayload); err != nil {
		fmt.Println("Error unmarshalling sdp payload: ", err)
	}
	if err = ds.Clients[iceCandidatePayload.UserTarget.ID].Conn.WriteMessage(mt, msg); err != nil {
		log.Println("error send message:", err)
		ds.Clients[iceCandidatePayload.UserTarget.ID].Conn.Close()
		delete(ds.Clients, iceCandidatePayload.UserTarget.ID)
	}
}

func (r *WebsocketUsecase) onChat(webSocketMessage model.WebSocketMessage, mt int, msg []byte, err error) {

	var chat = model.ChatPayload{}
	if err := json.Unmarshal(webSocketMessage.Payload, &chat); err != nil {
		fmt.Println("Error unmarshalling sdp payload: ", err)
	}

	room, err := r.roomRepo.GetRoom(chat.RoomId)
	if err != nil {
		fmt.Println(err)
	}

	for _, user := range room.Users {
		if user.ID != chat.UserFrom.ID {
			if err = ds.Clients[user.ID].Conn.WriteMessage(mt, msg); err != nil {
				log.Println("error send message:", err)
				ds.Clients[user.ID].Conn.Close()
				delete(ds.Clients, user.ID)
			}
		}
	}

}

func (r *WebsocketUsecase) onLeave(webSocketMessage model.WebSocketMessage, mt int, msg []byte, err error) {

	var leaveMeeting = model.LeaveMeetingPayload{}
	if err := json.Unmarshal(webSocketMessage.Payload, &leaveMeeting); err != nil {
		fmt.Println("Error unmarshalling sdp payload: ", err)
	}

	room, err := r.roomRepo.GetRoom(leaveMeeting.RoomId)
	if err != nil {
		fmt.Println(err)
	}

	if room != nil {
		for _, user := range room.Users {
			if user.ID != leaveMeeting.User.ID {
				if err = ds.Clients[user.ID].Conn.WriteMessage(mt, msg); err != nil {
					log.Println("error send message:", err)
					ds.Clients[user.ID].Conn.Close()
					delete(ds.Clients, user.ID)
				}
			}
		}
		if err := r.roomRepo.RemoveUser(leaveMeeting.RoomId, leaveMeeting.User.ID); err != nil {
			fmt.Println(err)
		}

	}

}
