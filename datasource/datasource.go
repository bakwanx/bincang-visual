package datasource

import (
	model "bincang-visual/models"

	"github.com/gofiber/contrib/websocket"
)

var Clients = make(map[string]*websocket.Conn)
var Rooms = make(map[string]map[string]model.User)
var Users = make(map[string]model.User)
