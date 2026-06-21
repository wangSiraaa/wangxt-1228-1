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

type LimitController struct {
	svc *service.LimitService
}

func NewLimitController(svc *service.LimitService) *LimitController {
	return &LimitController{svc: svc}
}

func (ctrl *LimitController) List(c *gin.Context) {
	status := c.Query("status")
	var areaID uint64
	if v := c.Query("area_id"); v != "" {
		areaID, _ = strconv.ParseUint(v, 10, 64)
	}
	list, err := ctrl.svc.List(areaID, status)
	if err != nil {
		handleError(c, err)
		return
	}
	util.OK(c, list)
}

func (ctrl *LimitController) Create(c *gin.Context) {
	var req dto.LimitCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, util.NewBizError(util.CodeBadReq, "invalid request body: "+err.Error(), http.StatusBadRequest))
		return
	}
	cmd, err := ctrl.svc.Create(req, middleware.CurrentUserID(c))
	if err != nil {
		handleError(c, err)
		return
	}
	util.OK(c, cmd)
}

func (ctrl *LimitController) Impact(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, util.NewBizError(util.CodeBadReq, "invalid id", http.StatusBadRequest))
		return
	}
	resp, err := ctrl.svc.Impact(id)
	if err != nil {
		handleError(c, err)
		return
	}
	util.OK(c, resp)
}

func (ctrl *LimitController) CreateRemark(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, util.NewBizError(util.CodeBadReq, "invalid id", http.StatusBadRequest))
		return
	}
	var req dto.LimitRemarkCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, util.NewBizError(util.CodeBadReq, "invalid request body: "+err.Error(), http.StatusBadRequest))
		return
	}
	remark, err := ctrl.svc.CreateRemark(id, req, middleware.CurrentUserID(c))
	if err != nil {
		handleError(c, err)
		return
	}
	util.OK(c, remark)
}

func (ctrl *LimitController) ListRemarks(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, util.NewBizError(util.CodeBadReq, "invalid id", http.StatusBadRequest))
		return
	}
	remarks, err := ctrl.svc.ListRemarks(id)
	if err != nil {
		handleError(c, err)
		return
	}
	util.OK(c, remarks)
}
