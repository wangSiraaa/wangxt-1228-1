package seed

import (
	"log"
	"math"
	"time"

	"gorm.io/gorm"

	"pvgrid/internal/model"
	"pvgrid/internal/repository"
	"pvgrid/internal/service"
)

// Run 写入演示种子数据，幂等：admin 存在则跳过
func Run(db *gorm.DB) {
	userRepo := repository.UserRepo{}
	if _, err := userRepo.FindByPhone("13800000000"); err == nil {
		log.Println("[seed] admin already exists, skip seeding")
		return
	}
	pwd, _ := service.HashPassword("123456")

	users := []model.User{
		{Phone: "13800000000", Password: pwd, Name: "系统管理员", Role: "admin"},
		{Phone: "13800000001", Password: pwd, Name: "业主张三", Role: "owner"},
		{Phone: "13800000002", Password: pwd, Name: "城东供电所", Role: "station"},
		{Phone: "13800000003", Password: pwd, Name: "调度员李四", Role: "dispatcher"},
	}
	for i := range users {
		if err := userRepo.FirstOrCreate(&users[i]); err != nil {
			log.Printf("[seed] user %s fail: %v", users[i].Phone, err)
		}
	}
	var owner model.User
	db.Where("phone = ?", "13800000001").First(&owner)

	// 台区
	areas := []model.Area{
		{Name: "阳光花园台区", OrgName: "城东供电所", CapacityKW: 500, Threshold: 0.80},
		{Name: "工业园区台区", OrgName: "城西供电所", CapacityKW: 1000, Threshold: 0.80},
	}
	for i := range areas {
		db.Create(&areas[i])
	}

	// 设备
	devs := []model.Device{
		{AreaID: areas[0].ID, OwnerID: owner.ID, Model: "华为SUN2000-10KTL", RatedKW: 10, Phase: "ABC", GridStatus: "grid"},
		{AreaID: areas[0].ID, OwnerID: owner.ID, Model: "阳光电源SG10KU", RatedKW: 10, Phase: "ABC", GridStatus: "pending"},
		{AreaID: areas[0].ID, OwnerID: owner.ID, Model: "古瑞瓦特Growatt-8K", RatedKW: 8, Phase: "ABC", GridStatus: "grid"},
		{AreaID: areas[1].ID, OwnerID: owner.ID, Model: "华为SUN2000-50KTL", RatedKW: 50, Phase: "ABC", GridStatus: "grid"},
	}
	for i := range devs {
		db.Create(&devs[i])
	}

	// 申报（已并网容量 = dev1 10 + dev3 8 = 18kW；台区1余量充足）
	decls := []model.Declaration{
		{AreaID: areas[0].ID, DeviceID: devs[0].ID, OwnerID: owner.ID, Type: "grid", CapacityKW: 10, Status: "approved"},
		{AreaID: areas[0].ID, DeviceID: devs[2].ID, OwnerID: owner.ID, Type: "grid", CapacityKW: 8, Status: "approved"},
		{AreaID: areas[1].ID, DeviceID: devs[3].ID, OwnerID: owner.ID, Type: "grid", CapacityKW: 50, Status: "approved"},
		{AreaID: areas[0].ID, DeviceID: devs[1].ID, OwnerID: owner.ID, Type: "grid", CapacityKW: 10, Status: "pending"},
	}
	for i := range decls {
		db.Create(&decls[i])
	}

	// 反送电告警（台区1存在 open 告警，扩容申报将被卡点）
	now := time.Now()
	alarms := []model.Alarm{
		{AreaID: areas[0].ID, DeviceID: devs[0].ID, Level: "warn", ReverseKW: 2.5, AlarmTime: now.Add(-2 * time.Hour), Status: "open"},
		{AreaID: areas[0].ID, DeviceID: devs[2].ID, Level: "danger", ReverseKW: 5.0, AlarmTime: now.Add(-26 * time.Hour), Status: "handled"},
		{AreaID: areas[1].ID, DeviceID: devs[3].ID, Level: "warn", ReverseKW: 3.0, AlarmTime: now.Add(-50 * time.Hour), Status: "closed"},
	}
	for i := range alarms {
		db.Create(&alarms[i])
	}

	// 限发指令（est_loss 留空，影响估算接口实时计算）
	// 时间窗口在白天 10:00-12:00，与发电数据时段匹配
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	limits := []model.LimitCommand{
		{AreaID: areas[0].ID, Ratio: 0.3, StartAt: midnight.Add(10 * time.Hour), EndAt: midnight.Add(12 * time.Hour), Status: "executing", EstLossKWh: 0, CreatedBy: 1},
		{AreaID: areas[0].ID, Ratio: 0.5, StartAt: midnight.AddDate(0, 0, -1).Add(14 * time.Hour), EndAt: midnight.AddDate(0, 0, -1).Add(16 * time.Hour), Status: "done", EstLossKWh: 0, CreatedBy: 1},
	}
	for i := range limits {
		db.Create(&limits[i])
	}

	// 时序功率点：当前月、上月、上上月，按月度分表写入
	// 需要至少3个月数据以满足 AvgGenInWindow 的2个月回溯查询
	tsRepo := repository.TimeseriesRepo{}
	genReadings := []struct {
		areaID, devID uint64
		rated         float64
	}{
		{areas[0].ID, devs[0].ID, 10},
		{areas[0].ID, devs[2].ID, 8},
		{areas[1].ID, devs[3].ID, 50},
	}
	months := collectMonths(now)
	total := 0
	for _, g := range genReadings {
		for _, ym := range months {
			days := daysInMonth(ym.year, ym.month)
			if ym.year == now.Year() && ym.month == int(now.Month()) {
				days = now.Day()
			}
			for d := 1; d <= days; d++ {
				for h := 6; h <= 18; h++ {
					ts := time.Date(ym.year, time.Month(ym.month), d, h, 0, 0, 0, now.Location())
					if ts.After(now) {
						continue
					}
					gen := genCurve(g.rated, h)
					rev := math.Max(0, gen*0.2-0.5)
					if rev < 0 {
						rev = 0
					}
					if err := tsRepo.Insert(model.PowerReading{
						AreaID: g.areaID, DeviceID: g.devID, Ts: ts, GenKW: round3(gen), ReverseKW: round3(rev),
					}); err != nil {
						log.Printf("[seed] insert reading fail: %v", err)
					}
					total++
				}
			}
		}
	}
	log.Printf("[seed] done, inserted %d power readings", total)
}

// genCurve 白天发电曲线：6-19 时正弦，峰值约额定功率
func genCurve(rated float64, hour int) float64 {
	// hour 6 -> 0, 12/13 -> peak, 19 -> 0
	x := (float64(hour) - 6) / 13.0
	if x < 0 || x > 1 {
		return 0
	}
	return rated * math.Sin(math.Pi*x)
}

func round3(v float64) float64 {
	return math.Round(v*1000) / 1000
}

type ym struct{ year, month int }

func collectMonths(now time.Time) []ym {
	res := []ym{{now.Year(), int(now.Month())}}
	prev1 := now.AddDate(0, -1, 0)
	res = append(res, ym{prev1.Year(), int(prev1.Month())})
	prev2 := now.AddDate(0, -2, 0)
	res = append(res, ym{prev2.Year(), int(prev2.Month())})
	return res
}

func daysInMonth(year, month int) int {
	return time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()
}
