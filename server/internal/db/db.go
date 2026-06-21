package db

import (
	"fmt"
	"sync"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"pvgrid/internal/config"
	"pvgrid/internal/model"
)

var (
	globalDB *gorm.DB
	driver   string
	mu       sync.Mutex
	ensured  = map[string]bool{}
)

// Init 初始化数据库连接并 AutoMigrate 业务表
func Init(cfg *config.Config) *gorm.DB {
	var dialector gorm.Dialector
	switch cfg.DBDriver {
	case "sqlite":
		driver = "sqlite"
		dialector = sqlite.Open(cfg.SQLitePath + "?_pragma=journal_mode(WAL)")
	default:
		driver = "mysql"
		dialector = mysql.Open(cfg.MySQLDSN)
	}
	gdb, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database(%s): %v", cfg.DBDriver, err))
	}
	if err := gdb.AutoMigrate(
		&model.User{},
		&model.Area{},
		&model.Device{},
		&model.Declaration{},
		&model.Alarm{},
		&model.LimitCommand{},
		&model.LimitExecutionRemark{},
	); err != nil {
		panic(fmt.Sprintf("automigrate failed: %v", err))
	}
	globalDB = gdb
	return gdb
}

// DB 返回全局 gorm.DB
func DB() *gorm.DB { return globalDB }

// Driver 返回当前数据库驱动名 mysql/sqlite
func Driver() string { return driver }

// EnsureMonthTable 确保 power_readings_yyyymm 表存在，按驱动生成兼容 DDL
func EnsureMonthTable(year, month int) error {
	table := model.MonthTable(year, month)
	mu.Lock()
	defer mu.Unlock()
	if ensured[table] {
		return nil
	}
	var ddl string
	if driver == "sqlite" {
		ddl = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
  id INTEGER NOT NULL,
  area_id INTEGER NOT NULL,
  device_id INTEGER NOT NULL,
  ts DATETIME NOT NULL,
  gen_kw REAL NOT NULL,
  reverse_kw REAL NOT NULL DEFAULT 0,
  PRIMARY KEY (id, ts)
)`, table)
	} else {
		ddl = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  area_id BIGINT UNSIGNED NOT NULL,
  device_id BIGINT UNSIGNED NOT NULL,
  ts DATETIME NOT NULL,
  gen_kw DECIMAL(10,3) NOT NULL,
  reverse_kw DECIMAL(10,3) NOT NULL DEFAULT 0,
  PRIMARY KEY (id, ts),
  KEY idx_area_ts (area_id, ts),
  KEY idx_device_ts (device_id, ts)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`, table)
	}
	if err := globalDB.Exec(ddl).Error; err != nil {
		return err
	}
	if driver == "sqlite" {
		globalDB.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_area_ts ON %s (area_id, ts)", table, table))
		globalDB.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_device_ts ON %s (device_id, ts)", table, table))
	}
	ensured[table] = true
	return nil
}
