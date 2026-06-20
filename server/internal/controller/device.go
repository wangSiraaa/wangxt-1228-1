package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"pvgrid/internal/dto"
	"pvgrid/internal/service"
	"pvgrid/internal/util"
)

type DeviceController struct {
	svc *service.DeviceService
}

func NewDeviceController(svc *service.DeviceService) *DeviceController {
	return &DeviceController{svc: svc}
}

func (ctrl *DeviceController) List(c *gin.Context) {
	var areaID uint64
	if v := c.Query("area_id"); v != "" {
		areaID, _ = strconv.ParseUint(v, 10, 64)
	}
	list, err := ctrl.svc.List(areaID)
	if err != nil {
		handleError(c, err)
		return
	}
	util.OK(c, list)
}

func (ctrl *DeviceController) Create(c *gin.Context) {
	var req dto.DeviceCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, util.NewBizError(util.CodeBadReq, "invalid request body: "+err.Error(), http.StatusBadRequest))
		return
	}
	d, err := ctrl.svc.Create(req)
	if err != nil {
		handleError(c, err)
		return
	}
	util.OK(c, d)
}
