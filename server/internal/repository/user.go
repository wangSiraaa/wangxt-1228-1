package repository

import (
	"pvgrid/internal/db"
	"pvgrid/internal/model"
)

type UserRepo struct{}

func (UserRepo) FindByPhone(phone string) (*model.User, error) {
	var u model.User
	err := db.DB().Where("phone = ?", phone).First(&u).Error
	return &u, err
}

func (UserRepo) FirstOrCreate(u *model.User) error {
	var exist model.User
	err := db.DB().Where("phone = ?", u.Phone).First(&exist).Error
	if err == nil {
		return nil
	}
	return db.DB().Create(u).Error
}

func (UserRepo) GetByID(id uint64) (*model.User, error) {
	var u model.User
	if err := db.DB().First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
