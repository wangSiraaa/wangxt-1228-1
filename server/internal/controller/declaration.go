package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"pvgrid/internal/dto"
	"pvgrid/internal/service"
	"pvgrid/internal/util"
)

type DeclarationController struct {
	svc *service.DeclarationService
}

func NewDeclarationController(svc *service.DeclarationService) *DeclarationController {
	return &DeclarationController{svc: svc}
}

func (ctrl *DeclarationController) List(c *gin.Context) {
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

func (ctrl *DeclarationController) Create(c *gin.Context) {
	var req dto.DeclarationCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, util.NewBizError(util.CodeBadReq, "invalid request body: "+err.Error(), http.StatusBadRequest))
		return
	}
	decl, err := ctrl.svc.Create(req)
	if err != nil {
		handleError(c, err)
		return
	}
	util.OK(c, decl)
}

func (ctrl *DeclarationController) Approve(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, util.NewBizError(util.CodeBadReq, "invalid id", http.StatusBadRequest))
		return
	}
	decl, err := ctrl.svc.Approve(id)
	if err != nil {
		handleError(c, err)
		return
	}
	util.OK(c, decl)
}

func (ctrl *DeclarationController) Reject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, util.NewBizError(util.CodeBadReq, "invalid id", http.StatusBadRequest))
		return
	}
	var req dto.DeclarationRejectReq
	_ = c.ShouldBindJSON(&req)
	decl, err := ctrl.svc.Reject(id, req)
	if err != nil {
		handleError(c, err)
		return
	}
	util.OK(c, decl)
}
