package repository

import (
	"pvgrid/internal/db"
	"pvgrid/internal/model"
)

type DeviceRepo struct{}

func (DeviceRepo) Create(d *model.Device) error {
	return db.DB().Create(d).Error
}

func (DeviceRepo) List(areaID uint64) ([]model.Device, error) {
	var list []model.Device
	q := db.DB().Order("id ASC")
	if areaID > 0 {
		q = q.Where("area_id = ?", areaID)
	}
	err := q.Find(&list).Error
	return list, err
}

func (DeviceRepo) Get(id uint64) (*model.Device, error) {
	var d model.Device
	if err := db.DB().First(&d, id).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (DeviceRepo) UpdateGridStatus(id uint64, status string) error {
	return db.DB().Model(&model.Device{}).Where("id = ?", id).Update("grid_status", status).Error
}

// SumGridCapacityByArea 通过已并网设备额定容量求和
func (DeviceRepo) SumGridCapacityByArea(areaID uint64) (float64, error) {
	var sum float64
	err := db.DB().Model(&model.Device{}).
		Where("area_id = ? AND grid_status = ?", areaID, "grid").
		Select("COALESCE(SUM(rated_kw),0)").Scan(&sum).Error
	return sum, err
}
