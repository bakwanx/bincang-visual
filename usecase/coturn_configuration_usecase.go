package usecase

import (
	"bincang-visual/models"
	"bincang-visual/repository"
)

type CoturnConfigurationUsecase struct {
	coturnRepo repository.CoturnConfigurationRepository
}

func NewCoturnConfigurationUsecase(coturnRepo repository.CoturnConfigurationRepository) *CoturnConfigurationUsecase {
	return &CoturnConfigurationUsecase{coturnRepo: coturnRepo}
}

func (u *CoturnConfigurationUsecase) GetConfiguration() (*models.CoturnConfiguration, error) {
	return u.coturnRepo.GetConfiguration()
}
