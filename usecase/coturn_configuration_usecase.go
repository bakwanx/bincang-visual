package usecase

import (
	"bincang-visual/repository"
)

type CoturnConfigurationUsecase struct {
	coturnRepo repository.CoturnConfigurationRepository
}

func NewCoturnConfigurationUsecase(coturnRepo repository.CoturnConfigurationRepository) *CoturnConfigurationUsecase {
	return &CoturnConfigurationUsecase{coturnRepo: coturnRepo}
}

func (u *CoturnConfigurationUsecase) GetConfiguration() (*string, error) {
	return u.coturnRepo.GetConfiguration()
}
