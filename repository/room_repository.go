package repository

type RoomRepository interface {
	AddRoom() string
}

type RoomRepositoryImpl struct {
	rooms map[string]map[string]string
}

func NewInMemoryRoomRepository() RoomRepository {
	return &RoomRepositoryImpl{
		rooms: make(map[string]map[string]string),
	}
}

func (r *RoomRepositoryImpl) AddRoom() string {
	r.AddRoom()
	return ""
}
