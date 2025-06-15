package usecase

import (
	"bincang-visual/models"
	"bincang-visual/repository"
)

type RoomUsecase struct {
	roomRepo repository.RoomRepository
}

func NewRoomUsecase(roomRepo repository.RoomRepository) *RoomUsecase {
	return &RoomUsecase{roomRepo: roomRepo}
}

func (u *RoomUsecase) CreateRoom() (*models.Room, error) {
	return u.roomRepo.CreateRoom()
}

func (u *RoomUsecase) GetRoom(roomId string) (*models.Room, error) {
	return u.roomRepo.GetRoom(roomId)
}

func (u RoomUsecase) AddUser(roomId string, userModel models.User) error {
	return u.roomRepo.AddUser(roomId, userModel)
}

func (u RoomUsecase) RemoveUser(roomId, userId string) error {
	return u.roomRepo.RemoveUser(roomId, userId)
}
