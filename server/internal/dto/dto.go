package dto

import "time"

// === 4.1 鉴权 ===
type LoginReq struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginUser struct {
	ID    uint64 `json:"id"`
	Phone string `json:"phone"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

type LoginResp struct {
	Token string    `json:"token"`
	User  LoginUser `json:"user"`
}

// === 4.2 台区 ===
type AreaCreateReq struct {
	Name       string  `json:"name" binding:"required"`
	OrgName    string  `json:"org_name" binding:"required"`
	CapacityKW float64 `json:"capacity_kw" binding:"required"`
	Threshold  float64 `json:"threshold"`
}

type AreaUpdateReq struct {
	Name       *string  `json:"name"`
	OrgName    *string  `json:"org_name"`
	CapacityKW *float64 `json:"capacity_kw"`
	Threshold  *float64 `json:"threshold"`
}

type AreaSummary struct {
	ID              uint64  `json:"id"`
	Name            string  `json:"name"`
	CapacityKW      float64 `json:"capacity_kw"`
	Threshold       float64 `json:"threshold"`
	OrgName         string  `json:"org_name"`
	GridCapacityKW  float64 `json:"grid_capacity_kw"`
	AllowedCapacity float64 `json:"allowed_capacity_kw"`
	RemainingKW     float64 `json:"remaining_kw"`
}

// === 4.3 设备/申报 ===
type DeviceCreateReq struct {
	AreaID    uint64  `json:"area_id" binding:"required"`
	OwnerID   uint64  `json:"owner_id"`
	Model     string  `json:"model" binding:"required"`
	RatedKW   float64 `json:"rated_kw" binding:"required"`
	Phase     string  `json:"phase"`
	GridStatus string `json:"grid_status"`
}

type DeclarationCreateReq struct {
	AreaID     uint64  `json:"area_id" binding:"required"`
	DeviceID   uint64  `json:"device_id" binding:"required"`
	Type       string  `json:"type" binding:"required"`
	CapacityKW float64 `json:"capacity_kw" binding:"required"`
}

type DeclarationRejectReq struct {
	Reason string `json:"reason"`
}

// === 4.4 告警 ===
type AlarmHandleReq struct {
	Remark string `json:"remark"`
}

// === 4.5 限发 ===
type LimitCreateReq struct {
	AreaID  uint64    `json:"area_id" binding:"required"`
	Ratio   float64   `json:"ratio" binding:"required"`
	StartAt time.Time `json:"start_at" binding:"required"`
	EndAt   time.Time `json:"end_at" binding:"required"`
}

type LimitListItem struct {
	ID                 uint64    `json:"id"`
	AreaID             uint64    `json:"area_id"`
	Ratio              float64   `json:"ratio"`
	StartAt            time.Time `json:"start_at"`
	EndAt              time.Time `json:"end_at"`
	Status             string    `json:"status"`
	EstLossKWh         float64   `json:"est_loss_kwh"`
	AvgGenKW           float64   `json:"avg_gen_kw"`
	SampleCount        int       `json:"sample_count"`
	DurationHours      float64   `json:"duration_hours"`
	RemarkStatus       string    `json:"remark_status"`
	RemarkedEstLossKWh float64   `json:"remarked_est_loss_kwh"`
	CreatedBy          uint64    `json:"created_by"`
	CreatedAt          time.Time `json:"created_at"`
}

type LimitImpactResp struct {
	ID             uint64    `json:"id"`
	AreaID         uint64    `json:"area_id"`
	Ratio          float64   `json:"ratio"`
	StartAt        time.Time `json:"start_at"`
	EndAt          time.Time `json:"end_at"`
	DurationHours  float64   `json:"duration_hours"`
	AvgGenKW       float64   `json:"avg_gen_kw"`
	EstLossKWh     float64   `json:"est_loss_kwh"`
	SampleCount    int       `json:"sample_count"`
}

type LimitRemarkCreateReq struct {
	BlockReason string  `json:"block_reason" binding:"required,max=200"`
	EstLossKWh  float64 `json:"est_loss_kwh" binding:"required,min=0"`
	Remark      string  `json:"remark" binding:"max=500"`
}

type LimitRemarkResp struct {
	ID             uint64    `json:"id"`
	LimitCommandID uint64    `json:"limit_command_id"`
	BlockReason    string    `json:"block_reason"`
	EstLossKWh     float64   `json:"est_loss_kwh"`
	Remark         string    `json:"remark"`
	RemarkedBy     uint64    `json:"remarked_by"`
	RemarkedByName string    `json:"remarked_by_name"`
	RemarkedAt     time.Time `json:"remarked_at"`
}

// === 4.6 时序 ===
type Point struct {
	Ts int64   `json:"ts"`
	V  float64 `json:"v"`
}

type TimeseriesQuery struct {
	AreaID uint64 `form:"area_id" binding:"required"`
	Metric string `form:"metric" binding:"required"`
	From   string `form:"from" binding:"required"`
	To     string `form:"to" binding:"required"`
}
