package repository

import (
	"time"

	"pvgrid/internal/db"
	"pvgrid/internal/model"
)

type AlarmRepo struct{}

func (AlarmRepo) List(status string, areaID uint64) ([]model.Alarm, error) {
	var list []model.Alarm
	q := db.DB().Order("alarm_time DESC")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if areaID > 0 {
		q = q.Where("area_id = ?", areaID)
	}
	err := q.Find(&list).Error
	return list, err
}

func (AlarmRepo) Get(id uint64) (*model.Alarm, error) {
	var a model.Alarm
	if err := db.DB().First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (AlarmRepo) Handle(id, handler uint64, remark string) error {
	return db.DB().Model(&model.Alarm{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     "handled",
			"handled_by": handler,
			"handled_at": time.Now(),
			"remark":     remark,
		}).Error
}

// HasUnhandledAlarm 检查台区是否存在未关闭的反送电告警（status<>'closed'）
func (AlarmRepo) HasUnhandledAlarm(areaID uint64) (bool, error) {
	var count int64
	err := db.DB().Model(&model.Alarm{}).
		Where("area_id = ? AND status <> ?", areaID, "closed").
		Count(&count).Error
	return count > 0, err
}
