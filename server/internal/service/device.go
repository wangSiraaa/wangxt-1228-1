package service

import (
	"net/http"

	"pvgrid/internal/dto"
	"pvgrid/internal/model"
	"pvgrid/internal/repository"
	"pvgrid/internal/util"
)

type DeviceService struct {
	devs repository.DeviceRepo
}

func NewDeviceService() *DeviceService {
	return &DeviceService{devs: repository.DeviceRepo{}}
}

func (s *DeviceService) Create(req dto.DeviceCreateReq) (*model.Device, error) {
	phase := req.Phase
	if phase == "" {
		phase = "ABC"
	}
	status := req.GridStatus
	if status == "" {
		status = "pending"
	}
	d := &model.Device{
		AreaID:     req.AreaID,
		OwnerID:    req.OwnerID,
		Model:      req.Model,
		RatedKW:    req.RatedKW,
		Phase:      phase,
		GridStatus: status,
	}
	if err := s.devs.Create(d); err != nil {
		return nil, util.NewBizError(util.CodeInternal, err.Error(), http.StatusInternalServerError)
	}
	return d, nil
}

func (s *DeviceService) List(areaID uint64) ([]model.Device, error) {
	return s.devs.List(areaID)
}
