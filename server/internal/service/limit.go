package service

import (
	"math"
	"net/http"
	"time"

	"pvgrid/internal/dto"
	"pvgrid/internal/model"
	"pvgrid/internal/repository"
	"pvgrid/internal/util"
)

type LimitService struct {
	limits repository.LimitRepo
	users  repository.UserRepo
	ts     repository.TimeseriesRepo
}

func NewLimitService() *LimitService {
	return &LimitService{
		limits: repository.LimitRepo{},
		users:  repository.UserRepo{},
		ts:     repository.TimeseriesRepo{},
	}
}

// EstimateLoss 影响估算 = 历史同时段平均发电曲线 × Ratio × 时长(小时)
func (s *LimitService) EstimateLoss(areaID uint64, ratio float64, startAt, endAt time.Time) (float64, float64, int, error) {
	if endAt.Before(startAt) || endAt.Equal(startAt) {
		return 0, 0, 0, util.NewBizError(util.CodeBadReq, "end_at must be after start_at", http.StatusBadRequest)
	}
	durationHours := endAt.Sub(startAt).Hours()
	avgGen, sample, err := s.ts.AvgGenInWindow(areaID, startAt, endAt)
	if err != nil {
		return 0, 0, 0, util.NewBizError(util.CodeInternal, err.Error(), http.StatusInternalServerError)
	}
	estLoss := avgGen * ratio * durationHours
	// 保留两位小数，对齐 DECIMAL(12,2)
	return math.Round(estLoss*100) / 100, math.Round(avgGen*1000) / 1000, sample, nil
}

// Create 发布限发指令，并预计算 est_loss_kwh
func (s *LimitService) Create(req dto.LimitCreateReq, createdBy uint64) (*model.LimitCommand, error) {
	if req.Ratio <= 0 || req.Ratio > 1 {
		return nil, util.NewBizError(util.CodeBadReq, "ratio must be in (0,1]", http.StatusBadRequest)
	}
	if !req.EndAt.After(req.StartAt) {
		return nil, util.NewBizError(util.CodeBadReq, "end_at must be after start_at", http.StatusBadRequest)
	}
	estLoss, _, _, err := s.EstimateLoss(req.AreaID, req.Ratio, req.StartAt, req.EndAt)
	if err != nil {
		return nil, err
	}
	cmd := &model.LimitCommand{
		AreaID:     req.AreaID,
		Ratio:      req.Ratio,
		StartAt:    req.StartAt,
		EndAt:      req.EndAt,
		Status:     "executing",
		EstLossKWh: estLoss,
		CreatedBy:  createdBy,
	}
	if err := s.limits.Create(cmd); err != nil {
		return nil, util.NewBizError(util.CodeInternal, err.Error(), http.StatusInternalServerError)
	}
	return cmd, nil
}

// Impact 返回限发指令的影响估算明细
func (s *LimitService) Impact(id uint64) (*dto.LimitImpactResp, error) {
	cmd, err := s.limits.Get(id)
	if err != nil {
		return nil, util.NewBizError(util.CodeNotFound, "limit command not found", http.StatusNotFound)
	}
	estLoss, avgGen, sample, err := s.EstimateLoss(cmd.AreaID, cmd.Ratio, cmd.StartAt, cmd.EndAt)
	if err != nil {
		return nil, err
	}
	durationHours := cmd.EndAt.Sub(cmd.StartAt).Hours()
	return &dto.LimitImpactResp{
		ID:            cmd.ID,
		AreaID:        cmd.AreaID,
		Ratio:         cmd.Ratio,
		StartAt:       cmd.StartAt,
		EndAt:         cmd.EndAt,
		DurationHours: math.Round(durationHours*100) / 100,
		AvgGenKW:      avgGen,
		EstLossKWh:    estLoss,
		SampleCount:   sample,
	}, nil
}

// List 查询限发指令列表，每条指令实时计算影响电量估算
func (s *LimitService) List(areaID uint64, status string) ([]dto.LimitListItem, error) {
	cmds, err := s.limits.List(areaID, status)
	if err != nil {
		return nil, util.NewBizError(util.CodeInternal, err.Error(), http.StatusInternalServerError)
	}
	result := make([]dto.LimitListItem, 0, len(cmds))
	for _, cmd := range cmds {
		estLoss, avgGen, sample, err := s.EstimateLoss(cmd.AreaID, cmd.Ratio, cmd.StartAt, cmd.EndAt)
		if err != nil {
			return nil, err
		}
		durationHours := cmd.EndAt.Sub(cmd.StartAt).Hours()
		result = append(result, dto.LimitListItem{
			ID:                 cmd.ID,
			AreaID:             cmd.AreaID,
			Ratio:              cmd.Ratio,
			StartAt:            cmd.StartAt,
			EndAt:              cmd.EndAt,
			Status:             cmd.Status,
			EstLossKWh:         estLoss,
			AvgGenKW:           avgGen,
			SampleCount:        sample,
			DurationHours:      durationHours,
			RemarkStatus:       cmd.RemarkStatus,
			RemarkedEstLossKWh: cmd.RemarkedEstLossKWh,
			CreatedBy:          cmd.CreatedBy,
			CreatedAt:          cmd.CreatedAt,
		})
	}
	return result, nil
}

// CreateRemark 创建限发指令执行备注
func (s *LimitService) CreateRemark(limitCmdID uint64, req dto.LimitRemarkCreateReq, remarkedBy uint64) (*dto.LimitRemarkResp, error) {
	cmd, err := s.limits.Get(limitCmdID)
	if err != nil {
		return nil, util.NewBizError(util.CodeNotFound, "limit command not found", http.StatusNotFound)
	}
	remark := &model.LimitExecutionRemark{
		LimitCommandID: limitCmdID,
		BlockReason:    req.BlockReason,
		EstLossKWh:     math.Round(req.EstLossKWh*100) / 100,
		Remark:         req.Remark,
		RemarkedBy:     remarkedBy,
	}
	if err := s.limits.CreateRemark(remark); err != nil {
		return nil, util.NewBizError(util.CodeInternal, err.Error(), http.StatusInternalServerError)
	}
	if err := s.limits.UpdateRemarkStatus(limitCmdID, "remarked", remark.EstLossKWh); err != nil {
		return nil, util.NewBizError(util.CodeInternal, err.Error(), http.StatusInternalServerError)
	}
	user, err := s.users.GetByID(remarkedBy)
	if err != nil {
		return nil, util.NewBizError(util.CodeInternal, "failed to get user info", http.StatusInternalServerError)
	}
	return &dto.LimitRemarkResp{
		ID:             remark.ID,
		LimitCommandID: remark.LimitCommandID,
		BlockReason:    remark.BlockReason,
		EstLossKWh:     remark.EstLossKWh,
		Remark:         remark.Remark,
		RemarkedBy:     remark.RemarkedBy,
		RemarkedByName: user.Name,
		RemarkedAt:     remark.RemarkedAt,
	}, nil
}

// ListRemarks 查询限发指令的执行备注列表
func (s *LimitService) ListRemarks(limitCmdID uint64) ([]dto.LimitRemarkResp, error) {
	if _, err := s.limits.Get(limitCmdID); err != nil {
		return nil, util.NewBizError(util.CodeNotFound, "limit command not found", http.StatusNotFound)
	}
	remarks, err := s.limits.ListRemarks(limitCmdID)
	if err != nil {
		return nil, util.NewBizError(util.CodeInternal, err.Error(), http.StatusInternalServerError)
	}
	result := make([]dto.LimitRemarkResp, 0, len(remarks))
	for _, r := range remarks {
		user, err := s.users.GetByID(r.RemarkedBy)
		name := ""
		if err == nil {
			name = user.Name
		}
		result = append(result, dto.LimitRemarkResp{
			ID:             r.ID,
			LimitCommandID: r.LimitCommandID,
			BlockReason:    r.BlockReason,
			EstLossKWh:     r.EstLossKWh,
			Remark:         r.Remark,
			RemarkedBy:     r.RemarkedBy,
			RemarkedByName: name,
			RemarkedAt:     r.RemarkedAt,
		})
	}
	return result, nil
}
