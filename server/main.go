package main

import (
	"log"

	"pvgrid/internal/config"
	"pvgrid/internal/controller"
	"pvgrid/internal/db"
	"pvgrid/internal/router"
	"pvgrid/internal/seed"
	"pvgrid/internal/service"
)

func main() {
	cfg := config.Load()

	gdb := db.Init(cfg)
	seed.Run(gdb)

	authSvc := service.NewAuthService(cfg)
	areaSvc := service.NewAreaService()
	deviceSvc := service.NewDeviceService()
	declSvc := service.NewDeclarationService()
	alarmSvc := service.NewAlarmService()
	limitSvc := service.NewLimitService()
	tsSvc := service.NewTimeseriesService()

	r := router.Setup(cfg,
		controller.NewAuthController(authSvc),
		controller.NewAreaController(areaSvc),
		controller.NewDeviceController(deviceSvc),
		controller.NewDeclarationController(declSvc),
		controller.NewAlarmController(alarmSvc),
		controller.NewLimitController(limitSvc),
		controller.NewTimeseriesController(tsSvc),
	)

	log.Printf("[pvgrid] driver=%s listening on %s, CORS=%s", cfg.DBDriver, cfg.ListenAddr, cfg.CORSAllowOrigin)
	if err := r.Run(cfg.ListenAddr); err != nil {
		log.Fatalf("server start failed: %v", err)
	}
}
