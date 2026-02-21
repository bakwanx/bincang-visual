package usecase

import (
	"bincang-visual/old_code/models"
	"bincang-visual/old_code/repository"
)

type UserUsecase struct {
	userRepo repository.UserRepository
}

func NewUserUsecase(userRepo repository.UserRepository) *UserUsecase {
	return &UserUsecase{
		userRepo: userRepo,
	}
}

func (u *UserUsecase) RegisterUser(username string) (*models.User, error) {
	return u.userRepo.RegisterUser(username)
}

func (u *UserUsecase) GetUser(userId string) (*models.User, error) {
	return u.userRepo.GetUser(userId)
}

func (r *UserUsecase) RemoveUser(userId string) error {
	return r.userRepo.RemoveUser(userId)
}
