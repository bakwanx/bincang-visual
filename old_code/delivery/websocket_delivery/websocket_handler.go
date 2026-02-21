package websocketdelivery

import (
	"fmt"
	"sync"

	ds "bincang-visual/old_code/datasource"
	"bincang-visual/old_code/usecase"

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
