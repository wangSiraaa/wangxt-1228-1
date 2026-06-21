package router

import (
	"github.com/gin-gonic/gin"

	"pvgrid/internal/config"
	"pvgrid/internal/controller"
	"pvgrid/internal/middleware"
)

func Setup(cfg *config.Config, auth *controller.AuthController, area *controller.AreaController,
	device *controller.DeviceController, decl *controller.DeclarationController,
	alarm *controller.AlarmController, limit *controller.LimitController,
	ts *controller.TimeseriesController) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS(cfg))

	api := r.Group("/api")
	{
		api.POST("/auth/login", auth.Login)

		authed := api.Group("")
		authed.Use(middleware.JWTAuth(cfg))
		{
			// 台区 CRUD
			authed.GET("/areas", area.List)
			authed.GET("/areas/:id", area.Detail)
			authed.POST("/areas", middleware.RequireRole("admin", "station"), area.Create)
			authed.PUT("/areas/:id", middleware.RequireRole("admin", "station"), area.Update)

			// 设备
			authed.GET("/devices", device.List)
			authed.POST("/devices", middleware.RequireRole("owner", "admin"), device.Create)

			// 申报
			authed.GET("/declarations", decl.List)
			authed.POST("/declarations", middleware.RequireRole("owner", "admin"), decl.Create)
			authed.POST("/declarations/:id/approve", middleware.RequireRole("station", "dispatcher", "admin"), decl.Approve)
			authed.POST("/declarations/:id/reject", middleware.RequireRole("station", "dispatcher", "admin"), decl.Reject)

			// 反送电告警
			authed.GET("/alarms", alarm.List)
			authed.POST("/alarms/:id/handle", middleware.RequireRole("station", "admin"), alarm.Handle)

			// 限发指令
			authed.GET("/limits", limit.List)
			authed.POST("/limits", middleware.RequireRole("dispatcher", "admin"), limit.Create)
			authed.GET("/limits/:id/impact", limit.Impact)
			authed.GET("/limits/:id/remarks", limit.ListRemarks)
			authed.POST("/limits/:id/remarks", middleware.RequireRole("owner", "station", "admin"), limit.CreateRemark)

			// 时序
			authed.GET("/timeseries", ts.Query)
		}
	}
	return r
}
