package service

import (
	"net/http"

	"pvgrid/internal/dto"
	"pvgrid/internal/model"
	"pvgrid/internal/repository"
	"pvgrid/internal/util"
)

type AreaService struct {
	areas repository.AreaRepo
	decls repository.DeclarationRepo
}

func NewAreaService() *AreaService {
	return &AreaService{areas: repository.AreaRepo{}, decls: repository.DeclarationRepo{}}
}

func (s *AreaService) Create(req dto.AreaCreateReq) (*model.Area, error) {
	threshold := req.Threshold
	if threshold <= 0 || threshold > 1 {
		threshold = 0.80
	}
	a := &model.Area{
		Name:       req.Name,
		OrgName:    req.OrgName,
		CapacityKW: req.CapacityKW,
		Threshold:  threshold,
	}
	if err := s.areas.Create(a); err != nil {
		return nil, util.NewBizError(util.CodeInternal, err.Error(), http.StatusInternalServerError)
	}
	return a, nil
}

func (s *AreaService) List() ([]model.Area, error) {
	return s.areas.List()
}

func (s *AreaService) Summary(id uint64) (*dto.AreaSummary, error) {
	a, err := s.areas.Get(id)
	if err != nil {
		return nil, util.NewBizError(util.CodeNotFound, "area not found", http.StatusNotFound)
	}
	gridCap, err := s.decls.SumApprovedCapacityByArea(id)
	if err != nil {
		return nil, util.NewBizError(util.CodeInternal, err.Error(), http.StatusInternalServerError)
	}
	allowed := a.CapacityKW * a.Threshold
	return &dto.AreaSummary{
		ID:              a.ID,
		Name:            a.Name,
		CapacityKW:      a.CapacityKW,
		Threshold:       a.Threshold,
		OrgName:         a.OrgName,
		GridCapacityKW:  gridCap,
		AllowedCapacity: allowed,
		RemainingKW:     allowed - gridCap,
	}, nil
}

func (s *AreaService) Update(id uint64, req dto.AreaUpdateReq) (*model.Area, error) {
	if _, err := s.areas.Get(id); err != nil {
		return nil, util.NewBizError(util.CodeNotFound, "area not found", http.StatusNotFound)
	}
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.OrgName != nil {
		updates["org_name"] = *req.OrgName
	}
	if req.CapacityKW != nil {
		updates["capacity_kw"] = *req.CapacityKW
	}
	if req.Threshold != nil {
		if *req.Threshold <= 0 || *req.Threshold > 1 {
			return nil, util.NewBizError(util.CodeBadReq, "threshold must be in (0,1]", http.StatusBadRequest)
		}
		updates["threshold"] = *req.Threshold
	}
	if len(updates) == 0 {
		return s.areas.Get(id)
	}
	if err := s.areas.Update(id, updates); err != nil {
		return nil, util.NewBizError(util.CodeInternal, err.Error(), http.StatusInternalServerError)
	}
	return s.areas.Get(id)
}
