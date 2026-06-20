package repository

import (
	"pvgrid/internal/db"
	"pvgrid/internal/model"
)

type AreaRepo struct{}

func (AreaRepo) Create(a *model.Area) error {
	return db.DB().Create(a).Error
}

func (AreaRepo) List() ([]model.Area, error) {
	var list []model.Area
	err := db.DB().Order("id ASC").Find(&list).Error
	return list, err
}

func (AreaRepo) Get(id uint64) (*model.Area, error) {
	var a model.Area
	if err := db.DB().First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (AreaRepo) Update(id uint64, updates map[string]interface{}) error {
	return db.DB().Model(&model.Area{}).Where("id = ?", id).Updates(updates).Error
}
