package websocketdelivery

import (
	"fmt"
	"sync"

	ds "bincang-visual/datasource"
	"bincang-visual/usecase"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var lock = sync.Mutex{}

type WebSocketHandler interface {
	RegisterWebSocket(app *fiber.App)
	WebSocketSignalingController(c *websocket.Conn)
}

type websocketHandlerImpl struct {
	userUsecase      usecase.UserUsecase
	roomUsecase      usecase.RoomUsecase
	websocketUsecase usecase.WebsocketUsecase
}

func NewWebSocketHandler(userUsecase usecase.UserUsecase, roomUsecase usecase.RoomUsecase, websocketUsecase usecase.WebsocketUsecase) WebSocketHandler {
	return &websocketHandlerImpl{
		userUsecase:      userUsecase,
		roomUsecase:      roomUsecase,
		websocketUsecase: websocketUsecase,
	}
}

func (i websocketHandlerImpl) RegisterWebSocket(app *fiber.App) {
	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Use("/ws", websocket.New(i.WebSocketSignalingController))

}

func (i websocketHandlerImpl) WebSocketSignalingController(c *websocket.Conn) {
	i.websocketUsecase.InitWebsocket(c)
}

func removeUser(roomUsecase usecase.RoomUsecase, userUsecase usecase.UserUsecase, roomId string, userId string) {
	// close connection
	ds.Clients[userId].Conn.Close()

	// remove user in room
	if err := roomUsecase.RemoveUser(roomId, userId); err != nil {
		fmt.Println("[ERROR] remove user", err)
	}

	// remove user in users
	if err := userUsecase.RemoveUser(userId); err != nil {
		fmt.Println("[ERROR] remove user", err)
	}

	// remove user from clients
	delete(ds.Clients, userId)
}

// func WebSocketSignalingController(app *fiber.App) {

// 	app.Use("/ws", func(c *fiber.Ctx) error {
// 		// IsWebSocketUpgrade returns true if the client
// 		// requested upgrade to the WebSocket protocol.
// 		if websocket.IsWebSocketUpgrade(c) {
// 			c.Locals("allowed", true)
// 			return c.Next()
// 		}
// 		return fiber.ErrUpgradeRequired
// 	})

// 	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
// 		var userId = c.Query("userId")
// 		var roomId = c.Query("roomId")

// 		lock.Lock()
// 		ds.Clients[userId] = &model.UserClient{
// 			ID:   userId,
// 			Conn: c,
// 		}
// 		if ds.Rooms[roomId] == nil {
// 			ds.Rooms[roomId] = map[string]model.User{}
// 		}
// 		ds.Rooms[roomId][userId] = ds.Users[userId]
// 		lock.Unlock()

// 		for {
// 			mt, msg, err := c.ReadMessage()
// 			if err != nil {
// 				lock.Lock()
// 				delete(ds.Clients, userId)
// 				delete(ds.Users, userId)

// 				for _, room := range ds.Rooms {
// 					for _, user := range room {
// 						if user.ID == userId {
// 							delete(room, userId)
// 						}
// 					}

// 					// remove roomId
// 					if len(room) == 0 {
// 						delete(ds.Rooms, roomId)
// 					}

// 				}
// 				lock.Unlock()
// 				break
// 			}

// 			lock.Lock()

// 			var webSocketMessage model.WebSocketMessage

// 			err = json.Unmarshal(msg, &webSocketMessage)
// 			if err != nil {
// 				fmt.Println("Error unmarshalling SocketMessage: ", err)
// 			}

// 			switch webSocketMessage.Type {
// 			case "join":
// 				var requestOffering model.RequestOfferingPayload
// 				err = json.Unmarshal(webSocketMessage.Payload, &requestOffering)
// 				if err != nil {
// 					fmt.Println("Error unmarshalling Join: ", err)
// 				}
// 				if ds.Rooms[requestOffering.RoomId] != nil {
// 					if len(ds.Rooms[requestOffering.RoomId]) > 1 {
// 						for _, user := range ds.Rooms[requestOffering.RoomId] {
// 							if user.ID != requestOffering.UserRequest.ID {
// 								if ds.Clients[user.ID].Conn != nil {
// 									if err = ds.Clients[user.ID].Conn.WriteMessage(mt, msg); err != nil {
// 										log.Println("error send message:", err)
// 										ds.Clients[user.ID].Conn.Close()
// 										delete(ds.Clients, user.ID)
// 									}
// 								}
// 							}
// 						}
// 					}
// 				}
// 			case "offer":
// 				var sdpPayload = model.SdpPayload{}
// 				err = json.Unmarshal(webSocketMessage.Payload, &sdpPayload)

// 				if err = ds.Clients[sdpPayload.UserTarget.ID].Conn.WriteMessage(mt, msg); err != nil {
// 					log.Println("error send message:", err)
// 					ds.Clients[sdpPayload.UserTarget.ID].Conn.Close()
// 					delete(ds.Clients, sdpPayload.UserTarget.ID)
// 				}

// 			case "answer":
// 				var sdpPayload = model.SdpPayload{}
// 				err = json.Unmarshal(webSocketMessage.Payload, &sdpPayload)
// 				if err != nil {
// 					fmt.Println("Error unmarshalling sdp payload: ", err)
// 				}
// 				if err = ds.Clients[sdpPayload.UserTarget.ID].Conn.WriteMessage(mt, msg); err != nil {
// 					log.Println("error send message:", err)
// 					ds.Clients[sdpPayload.UserTarget.ID].Conn.Close()
// 					delete(ds.Clients, sdpPayload.UserTarget.ID)
// 				}

// 			case "candidate":
// 				var iceCandidatePayload = model.IceCandidatePayload{}
// 				err = json.Unmarshal(webSocketMessage.Payload, &iceCandidatePayload)
// 				if err != nil {
// 					fmt.Println("Error unmarshalling sdp payload: ", err)
// 				}
// 				if err = ds.Clients[iceCandidatePayload.UserTarget.ID].Conn.WriteMessage(mt, msg); err != nil {
// 					log.Println("error send message:", err)
// 					ds.Clients[iceCandidatePayload.UserTarget.ID].Conn.Close()
// 					delete(ds.Clients, iceCandidatePayload.UserTarget.ID)
// 				}
// 			case "chat":
// 				var chat = model.ChatPayload{}
// 				err = json.Unmarshal(webSocketMessage.Payload, &chat)
// 				if err != nil {
// 					fmt.Println("Error unmarshalling sdp payload: ", err)
// 				}
// 				for _, user := range ds.Rooms[chat.RoomId] {
// 					if user.ID != chat.UserFrom.ID {
// 						if err = ds.Clients[user.ID].Conn.WriteMessage(mt, msg); err != nil {
// 							log.Println("error send message:", err)
// 							ds.Clients[user.ID].Conn.Close()
// 							delete(ds.Clients, user.ID)
// 						}
// 					}
// 				}

// 			case "leave":
// 				var leaveMeeting = model.LeaveMeetingPayload{}
// 				err = json.Unmarshal(webSocketMessage.Payload, &leaveMeeting)
// 				if err != nil {
// 					fmt.Println("Error unmarshalling sdp payload: ", err)
// 				}
// 				if ds.Rooms[leaveMeeting.RoomId] != nil {
// 					for _, user := range ds.Rooms[leaveMeeting.RoomId] {
// 						if user.ID != leaveMeeting.User.ID {
// 							if err = ds.Clients[user.ID].Conn.WriteMessage(mt, msg); err != nil {
// 								log.Println("error send message:", err)
// 								ds.Clients[user.ID].Conn.Close()
// 								delete(ds.Clients, user.ID)
// 							}
// 						}
// 					}
// 					delete(ds.Rooms[leaveMeeting.RoomId], leaveMeeting.User.Username)
// 				}

// 			}

// 			lock.Unlock()
// 		}

// 	}))

// }
