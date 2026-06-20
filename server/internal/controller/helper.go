package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pvgrid/internal/util"
)

// handleError 统一处理 service 返回的错误
func handleError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if berr, ok := err.(*util.BizError); ok {
		util.Fail(c, berr)
		return
	}
	util.Fail(c, util.NewBizError(util.CodeInternal, err.Error(), http.StatusInternalServerError))
}
