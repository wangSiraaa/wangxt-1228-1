package repository

import (
	"pvgrid/internal/db"
	"pvgrid/internal/model"
)

type LimitRepo struct{}

func (LimitRepo) Create(l *model.LimitCommand) error {
	return db.DB().Create(l).Error
}

func (LimitRepo) List(areaID uint64, status string) ([]model.LimitCommand, error) {
	var list []model.LimitCommand
	q := db.DB().Order("id DESC")
	if areaID > 0 {
		q = q.Where("area_id = ?", areaID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Find(&list).Error
	return list, err
}

func (LimitRepo) Get(id uint64) (*model.LimitCommand, error) {
	var l model.LimitCommand
	if err := db.DB().First(&l, id).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (LimitRepo) UpdateEstLoss(id uint64, estLoss float64) error {
	return db.DB().Model(&model.LimitCommand{}).Where("id = ?", id).
		Update("est_loss_kwh", estLoss).Error
}

func (LimitRepo) UpdateRemarkStatus(id uint64, remarkStatus string, remarkedLoss float64) error {
	return db.DB().Model(&model.LimitCommand{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"remark_status":         remarkStatus,
			"remarked_est_loss_kwh": remarkedLoss,
		}).Error
}

func (LimitRepo) CreateRemark(r *model.LimitExecutionRemark) error {
	return db.DB().Create(r).Error
}

func (LimitRepo) ListRemarks(limitCommandID uint64) ([]model.LimitExecutionRemark, error) {
	var list []model.LimitExecutionRemark
	err := db.DB().Where("limit_command_id = ?", limitCommandID).
		Order("id DESC").Find(&list).Error
	return list, err
}

func (LimitRepo) GetRemark(id uint64) (*model.LimitExecutionRemark, error) {
	var r model.LimitExecutionRemark
	if err := db.DB().First(&r, id).Error; err != nil {
		return nil, err
	}
	return &r, nil
}
