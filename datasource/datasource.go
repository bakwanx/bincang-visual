package datasource

import (
	model "bincang-visual/models"
)

var Clients = make(map[string]*model.UserClient)
var Rooms = make(map[string]map[string]model.User)
var Users = make(map[string]model.User)

// {
// 	"roomId1": {
// 		"userId1": {
// 			"id": "id",
// 			"username": "username",
// 		},
// 		"userId1": {
// 			"id": "id",
// 			"username": "username",
// 		},
// 	},
// 	"roomId2": {
// 		"userId1": {
// 			"id": "id",
// 			"username": "username",
// 		},
// 		"userId1": {
// 			"id": "id",
// 			"username": "username",
// 		},
// 	},
// }
