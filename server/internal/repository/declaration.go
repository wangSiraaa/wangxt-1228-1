package repository

import (
	"pvgrid/internal/db"
	"pvgrid/internal/model"
)

type DeclarationRepo struct{}

func (DeclarationRepo) Create(d *model.Declaration) error {
	return db.DB().Create(d).Error
}

func (DeclarationRepo) List(status string, areaID uint64) ([]model.Declaration, error) {
	var list []model.Declaration
	q := db.DB().Order("id DESC")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if areaID > 0 {
		q = q.Where("area_id = ?", areaID)
	}
	err := q.Find(&list).Error
	return list, err
}

func (DeclarationRepo) Get(id uint64) (*model.Declaration, error) {
	var d model.Declaration
	if err := db.DB().First(&d, id).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (DeclarationRepo) Approve(id uint64) error {
	return db.DB().Model(&model.Declaration{}).Where("id = ?", id).
		Update("status", "approved").Error
}

func (DeclarationRepo) Reject(id uint64, reason string) error {
	return db.DB().Model(&model.Declaration{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": "rejected", "reject_reason": reason}).Error
}

// SumApprovedCapacityByArea 求和已审批通过的申报容量
func (DeclarationRepo) SumApprovedCapacityByArea(areaID uint64) (float64, error) {
	var sum float64
	err := db.DB().Model(&model.Declaration{}).
		Where("area_id = ? AND status = ?", areaID, "approved").
		Select("COALESCE(SUM(capacity_kw),0)").Scan(&sum).Error
	return sum, err
}
