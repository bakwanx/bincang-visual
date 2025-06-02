package service

type Service interface {
	CheckAndRemoveData()
}

type ServiceImpl struct{}

func NewService() Service {
	return &ServiceImpl{}
}

func (r *ServiceImpl) CheckAndRemoveData() {
	// TODO: would be great if using redis

	// scheduler := cron.NewScheduler()

	// _, err := scheduler.AddJob("0 * * * *", func() {
	// 	fmt.Println("Check and remove data sources")
	// 	currentTime := time.Now()
	// 	strCurrentDate := currentTime.Format("01-02-2006")
	// 	for _, room := range ds.Rooms {
	// 		for userId := range room {
	// 			if ds.Clients[userId] == nil && room[userId].CreatedAt != strCurrentDate {
	// 				delete(room, userId)
	// 				delete(ds.Users, userId)
	// 			}
	// 		}
	// 	}
	// })

	// if err != nil {
	// 	fmt.Println("Error running job")
	// }
}
