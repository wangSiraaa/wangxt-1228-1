package repository

import (
	"fmt"
	"sort"
	"sync/atomic"
	"time"

	"pvgrid/internal/db"
	"pvgrid/internal/model"
)

var readingSeq uint64

type TimeseriesRepo struct{}

// Insert 写入一条时序功率点，按 ts 自动路由到 power_readings_yyyymm
func (TimeseriesRepo) Insert(r model.PowerReading) error {
	if err := db.EnsureMonthTable(r.Ts.Year(), int(r.Ts.Month())); err != nil {
		return err
	}
	r.ID = atomic.AddUint64(&readingSeq, 1)
	table := model.MonthTableForTime(r.Ts)
	sql := fmt.Sprintf("INSERT INTO %s (id, area_id, device_id, ts, gen_kw, reverse_kw) VALUES (?,?,?,?,?,?)", table)
	return db.DB().Exec(sql, r.ID, r.AreaID, r.DeviceID, r.Ts, r.GenKW, r.ReverseKW).Error
}

// Query 跨月度分表查询时序点，metric: gen|reverse
func (TimeseriesRepo) Query(areaID uint64, metric string, from, to time.Time) ([]model.PowerReading, error) {
	col := "gen_kw"
	if metric == "reverse" {
		col = "reverse_kw"
	}
	_ = col
	var out []model.PowerReading
	for _, ym := range monthsBetween(from, to) {
		if err := db.EnsureMonthTable(ym.year, ym.month); err != nil {
			return nil, err
		}
		table := model.MonthTable(ym.year, ym.month)
		var rows []model.PowerReading
		err := db.DB().Table(table).
			Select("id, area_id, device_id, ts, gen_kw, reverse_kw").
			Where("area_id = ? AND ts >= ? AND ts < ?", areaID, from, to).
			Order("ts ASC").
			Scan(&rows).Error
		if err != nil {
			return nil, err
		}
		out = append(out, rows...)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Ts.Before(out[j].Ts) })
	return out, nil
}

// AvgGenInWindow 历史同时段（同一天内时刻窗口）平均发电功率，用于限发影响估算
func (TimeseriesRepo) AvgGenInWindow(areaID uint64, startAt, endAt time.Time) (float64, int, error) {
	startMin := startAt.Hour()*60 + startAt.Minute()
	endMin := endAt.Hour()*60 + endAt.Minute()
	var total, count float64
	// 扫描限发开始前两个月的历史数据
	base := time.Date(startAt.Year(), startAt.Month(), 1, 0, 0, 0, 0, startAt.Location())
	for i := 1; i <= 2; i++ {
		t := base.AddDate(0, -i, 0)
		y, m := t.Year(), int(t.Month())
		if err := db.EnsureMonthTable(y, m); err != nil {
			return 0, 0, err
		}
		table := model.MonthTable(y, m)
		var rows []model.PowerReading
		err := db.DB().Table(table).
			Select("id, area_id, device_id, ts, gen_kw, reverse_kw").
			Where("area_id = ?", areaID).
			Scan(&rows).Error
		if err != nil {
			return 0, 0, err
		}
		for _, r := range rows {
			rm := r.Ts.Hour()*60 + r.Ts.Minute()
			inWin := false
			if endMin >= startMin {
				inWin = rm >= startMin && rm <= endMin
			} else {
				inWin = rm >= startMin || rm <= endMin
			}
			if inWin {
				total += r.GenKW
				count++
			}
		}
	}
	if count == 0 {
		return 0, 0, nil
	}
	return total / count, int(count), nil
}

type ym struct{ year, month int }

func monthsBetween(from, to time.Time) []ym {
	var res []ym
	cur := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, from.Location())
	end := time.Date(to.Year(), to.Month(), 1, 0, 0, 0, 0, to.Location())
	for !cur.After(end) {
		res = append(res, ym{cur.Year(), int(cur.Month())})
		cur = cur.AddDate(0, 1, 0)
	}
	return res
}
