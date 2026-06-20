package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pvgrid/internal/dto"
	"pvgrid/internal/service"
	"pvgrid/internal/util"
)

type TimeseriesController struct {
	svc *service.TimeseriesService
}

func NewTimeseriesController(svc *service.TimeseriesService) *TimeseriesController {
	return &TimeseriesController{svc: svc}
}

func (ctrl *TimeseriesController) Query(c *gin.Context) {
	var q dto.TimeseriesQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		util.Fail(c, util.NewBizError(util.CodeBadReq, "invalid query: "+err.Error(), http.StatusBadRequest))
		return
	}
	points, err := ctrl.svc.Query(q)
	if err != nil {
		handleError(c, err)
		return
	}
	util.OK(c, points)
}
