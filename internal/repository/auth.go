package repository

import (
	"minecrat_go/dto"
	"minecrat_go/helper/utils"
	"minecrat_go/model"

	"gorm.io/gorm"
)

type AuthRepository interface {
	Register(dto *dto.Register) error
	Login(input *dto.Login) (*model.User, error)
	DeleteUser(id uint) error
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db}
}

func (r *authRepository) Register(dto *dto.Register) error {
	user := model.User{
		Email:    dto.Email,
		Password: dto.Password,
		Username: dto.Username,
	}
	if err := r.db.Create(&user).Error; err != nil {
		return err
	}
	return nil
}

func (r *authRepository) Login(input *dto.Login) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", input.Email).First(&user).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrUserNotFound
		}
	}

	return &user, nil
}

func (r *authRepository) DeleteUser(id uint) error {
	err := r.db.Delete(&model.User{}, id)
	if err.RowsAffected == 0 {
		return utils.ErrUserNotFound
	}
	if err.Error != nil {
		return err.Error
	}
	return nil
}
