package service

import (
	"net/http"

	"pvgrid/internal/dto"
	"pvgrid/internal/model"
	"pvgrid/internal/repository"
	"pvgrid/internal/util"
)

type DeclarationService struct {
	decls repository.DeclarationRepo
	devs  repository.DeviceRepo
	areas repository.AreaRepo
	alarms repository.AlarmRepo
}

func NewDeclarationService() *DeclarationService {
	return &DeclarationService{
		decls:  repository.DeclarationRepo{},
		devs:   repository.DeviceRepo{},
		areas:  repository.AreaRepo{},
		alarms: repository.AlarmRepo{},
	}
}

// Create 创建申报：扩容类型触发反送电告警卡点
func (s *DeclarationService) Create(req dto.DeclarationCreateReq) (*model.Declaration, error) {
	dev, err := s.devs.Get(req.DeviceID)
	if err != nil {
		return nil, util.NewBizError(util.CodeNotFound, "device not found", http.StatusNotFound)
	}
	areaID := req.AreaID
	if dev.AreaID != areaID {
		return nil, util.NewBizError(util.CodeBadReq, "device does not belong to the area", http.StatusBadRequest)
	}
	if req.Type != "grid" && req.Type != "expand" {
		return nil, util.NewBizError(util.CodeBadReq, "type must be grid or expand", http.StatusBadRequest)
	}
	// 业务规则：扩容申报时检查该台区是否存在未关闭反送电告警
	if req.Type == "expand" {
		has, err := s.alarms.HasUnhandledAlarm(areaID)
		if err != nil {
			return nil, util.NewBizError(util.CodeInternal, err.Error(), http.StatusInternalServerError)
		}
		if has {
			return nil, util.NewBizError(util.CodeAlarmUnhandled,
				"alarm_unhandled: 该台区存在未关闭的反送电告警，无法提交扩容申报", http.StatusUnprocessableEntity)
		}
	}
	decl := &model.Declaration{
		AreaID:     areaID,
		DeviceID:   req.DeviceID,
		OwnerID:    dev.OwnerID,
		Type:       req.Type,
		CapacityKW: req.CapacityKW,
		Status:     "pending",
	}
	if err := s.decls.Create(decl); err != nil {
		return nil, util.NewBizError(util.CodeInternal, err.Error(), http.StatusInternalServerError)
	}
	return decl, nil
}

// Approve 审批通过：触发容量校验 SUM(已并网容量)+申报容量 ≤ 台区容量×阈值
func (s *DeclarationService) Approve(id uint64) (*model.Declaration, error) {
	decl, err := s.decls.Get(id)
	if err != nil {
		return nil, util.NewBizError(util.CodeNotFound, "declaration not found", http.StatusNotFound)
	}
	if decl.Status != "pending" {
		return nil, util.NewBizError(util.CodeBadReq, "declaration is not pending", http.StatusBadRequest)
	}
	area, err := s.areas.Get(decl.AreaID)
	if err != nil {
		return nil, util.NewBizError(util.CodeNotFound, "area not found", http.StatusNotFound)
	}
	existing, err := s.decls.SumApprovedCapacityByArea(decl.AreaID)
	if err != nil {
		return nil, util.NewBizError(util.CodeInternal, err.Error(), http.StatusInternalServerError)
	}
	allowed := area.CapacityKW * area.Threshold
	if existing+decl.CapacityKW > allowed {
		return nil, util.NewBizError(util.CodeCapacityInsufficient,
			"capacity_insufficient: 已并网容量+申报容量超过消纳安全阈值", http.StatusUnprocessableEntity)
	}
	if err := s.decls.Approve(id); err != nil {
		return nil, util.NewBizError(util.CodeInternal, err.Error(), http.StatusInternalServerError)
	}
	// 同步设备并网状态
	_ = s.devs.UpdateGridStatus(decl.DeviceID, "grid")
	decl.Status = "approved"
	return decl, nil
}

// Reject 驳回申报
func (s *DeclarationService) Reject(id uint64, req dto.DeclarationRejectReq) (*model.Declaration, error) {
	decl, err := s.decls.Get(id)
	if err != nil {
		return nil, util.NewBizError(util.CodeNotFound, "declaration not found", http.StatusNotFound)
	}
	if decl.Status != "pending" {
		return nil, util.NewBizError(util.CodeBadReq, "declaration is not pending", http.StatusBadRequest)
	}
	if err := s.decls.Reject(id, req.Reason); err != nil {
		return nil, util.NewBizError(util.CodeInternal, err.Error(), http.StatusInternalServerError)
	}
	decl.Status = "rejected"
	decl.RejectReason = req.Reason
	return decl, nil
}

// List 查询申报列表
func (s *DeclarationService) List(status string, areaID uint64) ([]model.Declaration, error) {
	return s.decls.List(status, areaID)
}
