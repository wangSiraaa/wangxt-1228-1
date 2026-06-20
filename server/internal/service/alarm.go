package service

import (
	"net/http"

	"pvgrid/internal/dto"
	"pvgrid/internal/model"
	"pvgrid/internal/repository"
	"pvgrid/internal/util"
)

type AlarmService struct {
	alarms repository.AlarmRepo
}

func NewAlarmService() *AlarmService {
	return &AlarmService{alarms: repository.AlarmRepo{}}
}

func (s *AlarmService) List(status string, areaID uint64) ([]model.Alarm, error) {
	return s.alarms.List(status, areaID)
}

func (s *AlarmService) Handle(id, handler uint64, req dto.AlarmHandleReq) (*model.Alarm, error) {
	a, err := s.alarms.Get(id)
	if err != nil {
		return nil, util.NewBizError(util.CodeNotFound, "alarm not found", http.StatusNotFound)
	}
	if a.Status == "closed" {
		return nil, util.NewBizError(util.CodeBadReq, "alarm already closed", http.StatusBadRequest)
	}
	if err := s.alarms.Handle(id, handler, req.Remark); err != nil {
		return nil, util.NewBizError(util.CodeInternal, err.Error(), http.StatusInternalServerError)
	}
	return s.alarms.Get(id)
}
