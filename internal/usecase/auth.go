package usecase

import (
	"minecrat_go/dto"
	"minecrat_go/helper/utils"
	"minecrat_go/internal/repository"
)

type AuthUseCase interface {
	Login(dto *dto.Login) (string, error)
	Register(input *dto.Register) error
	DeleteUser(id uint) error
}

type authUseCase struct {
	authRepo repository.AuthRepository
}

func NewAuthUseCase(authRepo repository.AuthRepository) AuthUseCase {
	return &authUseCase{authRepo}
}

func (u *authUseCase) Login(dto *dto.Login) (string, error) {

	if !utils.IsValidEmail(dto.Email) {
		return "", utils.ErrInvalidEmail
	}

	user, err := u.authRepo.Login(dto)
	if err != nil {
		return "", err
	}

	if !utils.ComparePassword(user.Password, dto.Password) {
		return "", err
	}

	token, err := utils.GenerateJWTLogin(user.ID, user.Email)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (u *authUseCase) Register(input *dto.Register) error {
	if !utils.IsValidEmail(input.Email) {
		return utils.ErrInvalidEmail
	}
	hashed, err := utils.HashPassword(input.Password)
	if err != nil {
		return err
	}

	input.Password = hashed

	if err := u.authRepo.Register(input); err != nil {
		return err
	}

	return nil
}

func (u *authUseCase) DeleteUser(id uint) error {
	return u.authRepo.DeleteUser(id)
}
