package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"pvgrid/internal/dto"
	"pvgrid/internal/service"
	"pvgrid/internal/util"
)

type AreaController struct {
	svc *service.AreaService
}

func NewAreaController(svc *service.AreaService) *AreaController {
	return &AreaController{svc: svc}
}

func (ctrl *AreaController) List(c *gin.Context) {
	list, err := ctrl.svc.List()
	if err != nil {
		handleError(c, err)
		return
	}
	util.OK(c, list)
}

func (ctrl *AreaController) Create(c *gin.Context) {
	var req dto.AreaCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, util.NewBizError(util.CodeBadReq, "invalid request body: "+err.Error(), http.StatusBadRequest))
		return
	}
	a, err := ctrl.svc.Create(req)
	if err != nil {
		handleError(c, err)
		return
	}
	util.OK(c, a)
}

func (ctrl *AreaController) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, util.NewBizError(util.CodeBadReq, "invalid id", http.StatusBadRequest))
		return
	}
	summary, err := ctrl.svc.Summary(id)
	if err != nil {
		handleError(c, err)
		return
	}
	util.OK(c, summary)
}

func (ctrl *AreaController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, util.NewBizError(util.CodeBadReq, "invalid id", http.StatusBadRequest))
		return
	}
	var req dto.AreaUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, util.NewBizError(util.CodeBadReq, "invalid request body: "+err.Error(), http.StatusBadRequest))
		return
	}
	a, err := ctrl.svc.Update(id, req)
	if err != nil {
		handleError(c, err)
		return
	}
	util.OK(c, a)
}
