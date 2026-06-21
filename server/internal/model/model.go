package model

import (
	"fmt"
	"time"
)

// User 对应 DDL sys_user
type User struct {
	ID       uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Phone    string `gorm:"uniqueIndex:uk_phone;size:20;not null" json:"phone"`
	Password string `gorm:"size:128;not null" json:"-"`
	Name     string `gorm:"size:50;not null" json:"name"`
	Role     string `gorm:"size:20;not null" json:"role"`
}

func (User) TableName() string { return "sys_user" }

// Area 对应 DDL transformer_area
type Area struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name       string    `gorm:"size:80;not null" json:"name"`
	OrgName    string    `gorm:"column:org_name;size:80;not null;index:idx_area_org" json:"org_name"`
	CapacityKW float64   `gorm:"column:capacity_kw;type:decimal(10,2);not null" json:"capacity_kw"`
	Threshold  float64   `gorm:"type:decimal(3,2);default:0.80;not null" json:"threshold"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Area) TableName() string { return "transformer_area" }

// Device 对应 DDL device
type Device struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	AreaID     uint64    `gorm:"column:area_id;not null;index:idx_device_area" json:"area_id"`
	OwnerID    uint64    `gorm:"column:owner_id;not null;index:idx_device_owner" json:"owner_id"`
	Model      string    `gorm:"size:80;not null" json:"model"`
	RatedKW    float64   `gorm:"column:rated_kw;type:decimal(10,2);not null" json:"rated_kw"`
	Phase      string    `gorm:"size:8;not null;default:ABC" json:"phase"`
	GridStatus string    `gorm:"column:grid_status;size:16;not null;default:pending;index:idx_device_status" json:"grid_status"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Device) TableName() string { return "device" }

// Declaration 对应 DDL owner_declaration
type Declaration struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	AreaID       uint64    `gorm:"column:area_id;not null;index:idx_decl_area_status,priority:1" json:"area_id"`
	DeviceID     uint64    `gorm:"column:device_id;not null" json:"device_id"`
	OwnerID      uint64    `gorm:"column:owner_id;not null;index:idx_decl_owner" json:"owner_id"`
	Type         string    `gorm:"size:8;not null" json:"type"`
	CapacityKW   float64   `gorm:"column:capacity_kw;type:decimal(10,2);not null" json:"capacity_kw"`
	Status       string    `gorm:"size:16;not null;default:pending;index:idx_decl_area_status,priority:2" json:"status"`
	RejectReason string    `gorm:"column:reject_reason;size:200" json:"reject_reason"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Declaration) TableName() string { return "owner_declaration" }

// Alarm 对应 DDL reverse_alarm
type Alarm struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	AreaID    uint64     `gorm:"column:area_id;not null;index:idx_alarm_area_status,priority:1" json:"area_id"`
	DeviceID  uint64     `gorm:"column:device_id;not null" json:"device_id"`
	Level     string     `gorm:"size:8;not null;default:warn" json:"level"`
	ReverseKW float64    `gorm:"column:reverse_kw;type:decimal(10,2);not null" json:"reverse_kw"`
	AlarmTime time.Time  `gorm:"column:alarm_time;not null;index:idx_alarm_time" json:"alarm_time"`
	Status    string     `gorm:"size:8;not null;default:open;index:idx_alarm_area_status,priority:2" json:"status"`
	HandledBy *uint64    `gorm:"column:handled_by" json:"handled_by"`
	HandledAt *time.Time `gorm:"column:handled_at" json:"handled_at"`
	Remark    string     `gorm:"size:200" json:"remark"`
}

func (Alarm) TableName() string { return "reverse_alarm" }

// LimitCommand 对应 DDL limit_command
type LimitCommand struct {
	ID                 uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	AreaID             uint64    `gorm:"column:area_id;not null;index:idx_limit_area_status,priority:1" json:"area_id"`
	Ratio              float64   `gorm:"type:decimal(4,3);not null" json:"ratio"`
	StartAt            time.Time `gorm:"column:start_at;not null;index:idx_limit_time,priority:1" json:"start_at"`
	EndAt              time.Time `gorm:"column:end_at;not null;index:idx_limit_time,priority:2" json:"end_at"`
	Status             string    `gorm:"size:12;not null;default:executing;index:idx_limit_area_status,priority:2" json:"status"`
	EstLossKWh         float64   `gorm:"column:est_loss_kwh;type:decimal(12,2);default:0" json:"est_loss_kwh"`
	RemarkStatus       string    `gorm:"column:remark_status;size:16;not null;default:pending;index:idx_remark_status" json:"remark_status"`
	RemarkedEstLossKWh float64   `gorm:"column:remarked_est_loss_kwh;type:decimal(12,2);default:0" json:"remarked_est_loss_kwh"`
	CreatedBy          uint64    `gorm:"column:created_by;not null" json:"created_by"`
	CreatedAt          time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (LimitCommand) TableName() string { return "limit_command" }

// LimitExecutionRemark 限发指令执行备注，业主/供电所记录执行受阻原因
type LimitExecutionRemark struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	LimitCommandID uint64     `gorm:"column:limit_command_id;not null;index:idx_remark_cmd" json:"limit_command_id"`
	BlockReason    string     `gorm:"column:block_reason;size:200;not null" json:"block_reason"`
	EstLossKWh     float64    `gorm:"column:est_loss_kwh;type:decimal(12,2);not null;default:0" json:"est_loss_kwh"`
	Remark         string     `gorm:"column:remark;size:500" json:"remark"`
	RemarkedBy     uint64     `gorm:"column:remarked_by;not null" json:"remarked_by"`
	RemarkedAt     time.Time  `gorm:"column:remarked_at;autoCreateTime" json:"remarked_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (LimitExecutionRemark) TableName() string { return "limit_execution_remark" }

// PowerReading 对应月度分表 power_readings_yyyymm（复合主键 id+ts）
// 该结构仅用于查询映射，建表使用 db 包中驱动感知的 DDL，不参与 AutoMigrate。
type PowerReading struct {
	ID        uint64    `gorm:"column:id;primaryKey" json:"id"`
	AreaID    uint64    `gorm:"column:area_id" json:"area_id"`
	DeviceID  uint64    `gorm:"column:device_id" json:"device_id"`
	Ts        time.Time `gorm:"column:ts;primaryKey" json:"ts"`
	GenKW     float64   `gorm:"column:gen_kw" json:"gen_kw"`
	ReverseKW float64   `gorm:"column:reverse_kw" json:"reverse_kw"`
}

func (PowerReading) TableName() string { return "power_readings_template" }

// MonthTable 返回 power_readings_yyyymm 表名
func MonthTable(year, month int) string {
	return fmt.Sprintf("power_readings_%04d%02d", year, month)
}

// MonthTableForTime 按时间返回所属月度分表名
func MonthTableForTime(t time.Time) string {
	return MonthTable(t.Year(), int(t.Month()))
}
