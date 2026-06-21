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
	ts     repository.TimeseriesRepo
}

func NewLimitService() *LimitService {
	return &LimitService{
		limits: repository.LimitRepo{},
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
			ID:            cmd.ID,
			AreaID:        cmd.AreaID,
			Ratio:         cmd.Ratio,
			StartAt:       cmd.StartAt,
			EndAt:         cmd.EndAt,
			Status:        cmd.Status,
			EstLossKWh:    estLoss,
			AvgGenKW:      avgGen,
			SampleCount:   sample,
			DurationHours: durationHours,
			CreatedBy:     cmd.CreatedBy,
			CreatedAt:     cmd.CreatedAt,
		})
	}
	return result, nil
}
