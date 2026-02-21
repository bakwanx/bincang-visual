package jobs

import (
	"bincang-visual/internal/domain/usecase"
	"log"
	"time"
)

type RoomCleanupJob struct {
	roomUseCase *usecase.RoomUseCase
}

func NewRoomCleanupJob(roomUseCase *usecase.RoomUseCase) *RoomCleanupJob {
	return &RoomCleanupJob{roomUseCase: roomUseCase}
}

func (j *RoomCleanupJob) Start() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			j.cleanup()
		}
	}()
}

func (j *RoomCleanupJob) cleanup() {

	log.Println("[Cleanup] Running room cleanup job")

	// TODO: Get all rooms then
	// delete rooms older than 24 hours with no participants

	// ctx := context.Background()
}
