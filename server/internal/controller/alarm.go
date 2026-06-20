package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"pvgrid/internal/dto"
	"pvgrid/internal/middleware"
	"pvgrid/internal/service"
	"pvgrid/internal/util"
)

type AlarmController struct {
	svc *service.AlarmService
}

func NewAlarmController(svc *service.AlarmService) *AlarmController {
	return &AlarmController{svc: svc}
}

func (ctrl *AlarmController) List(c *gin.Context) {
	status := c.Query("status")
	var areaID uint64
	if v := c.Query("area_id"); v != "" {
		areaID, _ = strconv.ParseUint(v, 10, 64)
	}
	list, err := ctrl.svc.List(status, areaID)
	if err != nil {
		handleError(c, err)
		return
	}
	util.OK(c, list)
}

func (ctrl *AlarmController) Handle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, util.NewBizError(util.CodeBadReq, "invalid id", http.StatusBadRequest))
		return
	}
	var req dto.AlarmHandleReq
	_ = c.ShouldBindJSON(&req)
	handler := middleware.CurrentUserID(c)
	alarm, err := ctrl.svc.Handle(id, handler, req)
	if err != nil {
		handleError(c, err)
		return
	}
	util.OK(c, alarm)
}
