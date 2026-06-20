package service

import (
	"net/http"
	"time"

	"pvgrid/internal/dto"
	"pvgrid/internal/repository"
	"pvgrid/internal/util"
)

type TimeseriesService struct {
	ts repository.TimeseriesRepo
}

func NewTimeseriesService() *TimeseriesService {
	return &TimeseriesService{ts: repository.TimeseriesRepo{}}
}

func (s *TimeseriesService) Query(q dto.TimeseriesQuery) ([]dto.Point, error) {
	from, err := time.Parse(time.RFC3339, q.From)
	if err != nil {
		return nil, util.NewBizError(util.CodeBadReq, "invalid from time, expect RFC3339", http.StatusBadRequest)
	}
	to, err := time.Parse(time.RFC3339, q.To)
	if err != nil {
		return nil, util.NewBizError(util.CodeBadReq, "invalid to time, expect RFC3339", http.StatusBadRequest)
	}
	if q.Metric != "gen" && q.Metric != "reverse" {
		return nil, util.NewBizError(util.CodeBadReq, "metric must be gen or reverse", http.StatusBadRequest)
	}
	rows, err := s.ts.Query(q.AreaID, q.Metric, from, to)
	if err != nil {
		return nil, util.NewBizError(util.CodeInternal, err.Error(), http.StatusInternalServerError)
	}
	points := make([]dto.Point, 0, len(rows))
	for _, r := range rows {
		v := r.GenKW
		if q.Metric == "reverse" {
			v = r.ReverseKW
		}
		points = append(points, dto.Point{Ts: r.Ts.Unix(), V: v})
	}
	return points, nil
}
