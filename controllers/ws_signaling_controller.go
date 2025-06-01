package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	model "bincang-visual/models"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var Clients = make(map[string]*websocket.Conn)
var Rooms = make(map[string]map[string]string)
var lock = sync.Mutex{}
var Test = sync.Mutex{}

func WebSocketSignalingController(app *fiber.App) {

	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		var username = c.Query("username")
		var roomId = c.Query("roomId")

		userId := uuid.New().String()

		lock.Lock()
		Clients[userId] = c
		if Rooms[roomId] == nil {
			Rooms[roomId] = map[string]string{}
		}

		Rooms[roomId][userId] = username
		lock.Unlock()

		for {

			mt, msg, err := c.ReadMessage()
			if err != nil {
				lock.Lock()
				delete(Clients, username)
				for i, room := range Rooms {
					for _, usrName := range room {
						if usrName == username {
							delete(room, usrName)
						}
					}

					// remove roomId
					if len(room) == 0 {
						delete(Rooms, room[i])
					}

				}
				lock.Unlock()
				break
			}

			lock.Lock()

			var webSocketMessage model.WebSocketMessage

			err = json.Unmarshal(msg, &webSocketMessage)
			if err != nil {
				fmt.Println("Error unmarshalling SocketMessage: ", err)
			}

			switch webSocketMessage.Type {
			case "join":
				var requestOffering model.RequestOfferingPayload
				err = json.Unmarshal(webSocketMessage.Payload, &requestOffering)
				if err != nil {
					fmt.Println("Error unmarshalling Join: ", err)
				}
				if Rooms[requestOffering.RoomId] != nil {
					if len(Rooms[requestOffering.RoomId]) > 1 {
						for _, username := range Rooms[requestOffering.RoomId] {
							if username != requestOffering.UsernameRequest {
								if Clients[username] != nil {
									if err = Clients[username].WriteMessage(mt, msg); err != nil {
										log.Println("error send message:", err)
										Clients[username].Close()
										delete(Clients, username)
									}
								}
							}
						}
					}
				}
			case "offer":
				var offerPayload model.OfferPayload
				err = json.Unmarshal(webSocketMessage.Payload, &offerPayload)

				if err = Clients[offerPayload.To].WriteMessage(mt, msg); err != nil {
					log.Println("error send message:", err)
					Clients[username].Close()
					delete(Clients, username)
				}

			case "answer":
				var sdpPayload = model.SdpPayload{}
				err = json.Unmarshal(webSocketMessage.Payload, &sdpPayload)
				if err != nil {
					fmt.Println("Error unmarshalling sdp payload: ", err)
				}
				if err = Clients[sdpPayload.To].WriteMessage(mt, msg); err != nil {
					log.Println("error send message:", err)
					Clients[username].Close()
					delete(Clients, username)
				}

			case "candidate":
				var iceCandidatePayload = model.IceCandidatePayload{}
				err = json.Unmarshal(webSocketMessage.Payload, &iceCandidatePayload)
				if err != nil {
					fmt.Println("Error unmarshalling sdp payload: ", err)
				}
				if err = Clients[iceCandidatePayload.To].WriteMessage(mt, msg); err != nil {
					log.Println("error send message:", err)
					Clients[username].Close()
					delete(Clients, username)
				}

			case "leave":
				var leaveMeeting = model.LeaveMeetingPayload{}
				err = json.Unmarshal(webSocketMessage.Payload, &leaveMeeting)
				if err != nil {
					fmt.Println("Error unmarshalling sdp payload: ", err)
				}
				if Rooms[leaveMeeting.RoomId] != nil {
					for _, username := range Rooms[leaveMeeting.RoomId] {
						if username != leaveMeeting.Username {
							if err = Clients[username].WriteMessage(mt, msg); err != nil {
								log.Println("error send message:", err)
								Clients[username].Close()
								delete(Clients, username)
							}
						}
					}
					delete(Rooms[leaveMeeting.RoomId], leaveMeeting.Username)
				}

			}

			lock.Unlock()
		}

	}))

}
