package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	model "bincang-visual/models"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var clients = make(map[string]*websocket.Conn)
var rooms = make(map[string]map[string]string)
var lock = sync.Mutex{}

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

		lock.Lock()
		clients[username] = c
		if rooms[roomId] == nil {
			rooms[roomId] = map[string]string{}
		}
		rooms[roomId][username] = username
		lock.Unlock()

		for {

			mt, msg, err := c.ReadMessage()
			if err != nil {
				lock.Lock()
				delete(clients, username)
				for i, room := range rooms {
					for _, usrName := range room {
						if usrName == username {
							delete(room, usrName)
						}
					}

					// remove roomId
					if len(room) == 0 {
						delete(rooms, room[i])
					}

				}
				fmt.Println("pesan", len(rooms))

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
				var requestOffering model.RequestOffering
				err = json.Unmarshal(webSocketMessage.Payload, &requestOffering)
				if err != nil {
					fmt.Println("Error unmarshalling Join: ", err)
				}
				if rooms[requestOffering.RoomId] != nil {
					if len(rooms[requestOffering.RoomId]) > 1 {
						for _, username := range rooms[requestOffering.RoomId] {
							if username != requestOffering.UsernameRequest {
								if clients[username] != nil {
									if err = clients[username].WriteMessage(mt, msg); err != nil {
										log.Println("error send message:", err)
										clients[username].Close()
										delete(clients, username)
									}
								}
							}
						}
					}
				}
			case "offer":
				if err = clients[webSocketMessage.To].WriteMessage(mt, msg); err != nil {
					log.Println("error send message:", err)
					clients[username].Close()
					delete(clients, username)
				}

			case "answer":
				var sdpPayload = model.SdpPayload{}
				err = json.Unmarshal(webSocketMessage.Payload, &sdpPayload)
				if err != nil {
					fmt.Println("Error unmarshalling sdp payload: ", err)
				}
				if err = clients[webSocketMessage.To].WriteMessage(mt, msg); err != nil {
					log.Println("error send message:", err)
					clients[username].Close()
					delete(clients, username)
				}

			case "candidate":
				var iceCandidatePayload = model.IceCandidatePayload{}
				err = json.Unmarshal(webSocketMessage.Payload, &iceCandidatePayload)
				if err != nil {
					fmt.Println("Error unmarshalling sdp payload: ", err)
				}
				if err = clients[webSocketMessage.To].WriteMessage(mt, msg); err != nil {
					log.Println("error send message:", err)
					clients[username].Close()
					delete(clients, username)
				}

			case "leave":
				var leaveMeeting = model.LeaveMeeting{}
				err = json.Unmarshal(webSocketMessage.Payload, &leaveMeeting)
				if err != nil {
					fmt.Println("Error unmarshalling sdp payload: ", err)
				}
				if rooms[leaveMeeting.RoomId] != nil {
					for _, username := range rooms[leaveMeeting.RoomId] {
						if username != leaveMeeting.Username {
							if err = clients[username].WriteMessage(mt, msg); err != nil {
								log.Println("error send message:", err)
								clients[username].Close()
								delete(clients, username)
							}
						}
					}
					delete(rooms[leaveMeeting.RoomId], leaveMeeting.Username)
				}

			}

			lock.Unlock()
		}

	}))

}
