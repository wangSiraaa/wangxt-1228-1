package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pvgrid/internal/dto"
	"pvgrid/internal/service"
	"pvgrid/internal/util"
)

type AuthController struct {
	svc *service.AuthService
}

func NewAuthController(svc *service.AuthService) *AuthController {
	return &AuthController{svc: svc}
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var req dto.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, util.NewBizError(util.CodeBadReq, "invalid request body: "+err.Error(), http.StatusBadRequest))
		return
	}
	resp, err := ctrl.svc.Login(req)
	if err != nil {
		handleError(c, err)
		return
	}
	util.OK(c, resp)
}
