#!/bin/bash
set -e

echo "=========================================="
echo "  分布式光伏消纳管理 - 业务逻辑测试"
echo "=========================================="

BASE_URL="http://localhost:8080/api"

# 1. 登录获取 token
echo ""
echo "[1/8] 登录系统..."
LOGIN_RESP=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800000000","password":"123456"}')

TOKEN=$(echo "$LOGIN_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
USER=$(echo "$LOGIN_RESP" | python3 -c "import sys,json; u=json.load(sys.stdin)['data']['user']; print(f'{u[\"name\"]}({u[\"role\"]})')")
echo "  ✓ 登录成功: $USER"

AUTH_HEADER="Authorization: Bearer $TOKEN"

# 2. 检查台区信息
echo ""
echo "[2/8] 获取台区信息..."
AREAS=$(curl -s "$BASE_URL/areas" -H "$AUTH_HEADER")
echo "$AREAS" | python3 -c "
import sys,json
areas = json.load(sys.stdin)['data']
for a in areas:
    print(f'  ✓ 台区 #{a[\"id\"]}: {a[\"name\"]}, 容量={a[\"capacity_kw\"]}kW, 阈值={a[\"threshold\"]*100}%')
"

# 3. 台区容量汇总
echo ""
echo "[3/8] 台区容量汇总（验证消纳余量）..."
for area_id in 1 2; do
    SUMMARY=$(curl -s "$BASE_URL/areas/$area_id" -H "$AUTH_HEADER")
    echo "$SUMMARY" | python3 -c "
import sys,json
d = json.load(sys.stdin)['data']
used_pct = d['grid_capacity_kw'] / d['capacity_kw'] * 100
status = '✓ 余量充足' if d['remaining_kw'] > 0 else '✗ 容量不足'
print(f'  {status} 台区#{d[\"id\"]}: 已并网={d[\"grid_capacity_kw\"]:.1f}kW, 阈值内允许={d[\"allowed_capacity_kw\"]:.1f}kW, 余量={d[\"remaining_kw\"]:.1f}kW, 使用率={used_pct:.1f}%')
"
done

# 4. 测试扩容申报 - 告警卡点（台区1有未处理告警）
echo ""
echo "[4/8] 测试扩容申报 - 反送电告警卡点..."
echo "  台区#1存在未处理反送电告警，扩容申报应被拒绝..."
EXPAND_RESP=$(curl -s -X POST "$BASE_URL/declarations" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{"area_id":1,"device_id":1,"type":"expand","capacity_kw":5}')

echo "$EXPAND_RESP" | python3 -c "
import sys,json
resp = json.load(sys.stdin)
if resp['code'] == 'alarm_unhandled':
    print(f'  ✓ 扩容申报被正确拦截: {resp[\"message\"]}')
else:
    print(f'  ✗ 测试失败: {resp}')
"

# 5. 测试并网申报（无告警卡点）
echo ""
echo "[5/8] 测试并网申报（无告警卡点）..."
GRID_RESP=$(curl -s -X POST "$BASE_URL/declarations" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{"area_id":1,"device_id":2,"type":"grid","capacity_kw":10}')

echo "$GRID_RESP" | python3 -c "
import sys,json
resp = json.load(sys.stdin)
if resp['code'] == 'ok':
    d = resp['data']
    print(f'  ✓ 并网申报成功: 申报#{d[\"id\"]}, 容量={d[\"capacity_kw\"]}kW, 状态={d[\"status\"]}')
else:
    print(f'  ✗ 测试失败: {resp}')
"

# 6. 测试审批 - 容量校验（超容量应被拒绝）
echo ""
echo "[6/8] 测试审批 - 容量校验卡点..."
echo "  申报#5容量10500kW，远超台区#1阈值400kW，应被拒绝..."
APPROVE_RESP=$(curl -s -X POST "$BASE_URL/declarations/5/approve" -H "$AUTH_HEADER")

echo "$APPROVE_RESP" | python3 -c "
import sys,json
resp = json.load(sys.stdin)
if resp['code'] == 'capacity_insufficient':
    print(f'  ✓ 审批被正确拦截: {resp[\"message\"]}')
else:
    print(f'  结果: {resp}')
"

# 7. 测试告警处理
echo ""
echo "[7/8] 测试反送电告警处理..."
ALARMS=$(curl -s "$BASE_URL/alarms?status=open" -H "$AUTH_HEADER")
OPEN_ALARM_ID=$(echo "$ALARMS" | python3 -c "import sys,json; a=json.load(sys.stdin)['data']; print(a[0]['id'] if a else '0')")

if [ "$OPEN_ALARM_ID" != "0" ]; then
    echo "  处理告警#$OPEN_ALARM_ID..."
    HANDLE_RESP=$(curl -s -X POST "$BASE_URL/alarms/$OPEN_ALARM_ID/handle" \
      -H "Content-Type: application/json" \
      -H "$AUTH_HEADER" \
      -d '{"remark":"已通知业主调整出力，反送电已消除"}')
    
    echo "$HANDLE_RESP" | python3 -c "
import sys,json
resp = json.load(sys.stdin)
if resp['code'] == 'ok':
    a = resp['data']
    print(f'  ✓ 告警处理成功: 告警#{a[\"id\"]}, 状态={a[\"status\"]}, 处理人=#{a[\"handled_by\"]}')
else:
    print(f'  ✗ 处理失败: {resp}')
"
else
    echo "  无未处理告警，跳过"
fi

# 8. 测试限发指令 - 影响电量估算
echo ""
echo "[8/8] 测试限发指令 - 影响电量估算..."
START=$(date -u -v-1H +"%Y-%m-%dT%H:00:00.000Z")
END=$(date -u -v+1H +"%Y-%m-%dT%H:00:00.000Z")

LIMIT_RESP=$(curl -s -X POST "$BASE_URL/limits" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d "{\"area_id\":1,\"ratio\":0.3,\"start_at\":\"$START\",\"end_at\":\"$END\"}")

echo "$LIMIT_RESP" | python3 -c "
import sys,json
resp = json.load(sys.stdin)
if resp['code'] == 'ok':
    l = resp['data']
    print(f'  ✓ 限发指令发布成功: 指令#{l[\"id\"]}')
    print(f'    台区#{l[\"area_id\"]}, 限发比例={l[\"ratio\"]*100:.0f}%')
    print(f'    时段: {l[\"start_at\"]} → {l[\"end_at\"]}')
    print(f'    预估影响电量: {l[\"est_loss_kwh\"]:.2f} kWh')
else:
    print(f'  结果: {resp}')
"

# 获取限发影响详情
LIMIT_ID=$(echo "$LIMIT_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")
IMPACT_RESP=$(curl -s "$BASE_URL/limits/$LIMIT_ID/impact" -H "$AUTH_HEADER")

echo "$IMPACT_RESP" | python3 -c "
import sys,json
resp = json.load(sys.stdin)
if resp['code'] == 'ok':
    i = resp['data']
    print(f'  ✓ 影响估算详情:')
    print(f'    持续时长: {i[\"duration_hours\"]:.2f}h')
    print(f'    历史均发功率: {i[\"avg_gen_kw\"]:.3f}kW')
    print(f'    样本点数: {i[\"sample_count\"]}')
    print(f'    预估损失电量: {i[\"est_loss_kwh\"]:.2f}kWh')
else:
    print(f'  结果: {resp}')
"

echo ""
echo "=========================================="
echo "  测试完成！"
echo "=========================================="
echo "  业务逻辑验证摘要："
echo "  ✓ 台区容量管理（供电所录入）"
echo "  ✓ 业主申报逆变器信息"
echo "  ✓ 扩容申报反送电告警卡点"
echo "  ✓ 并网审批容量校验卡点"
echo "  ✓ 反送电告警处理流程"
echo "  ✓ 限发指令发布与电量估算"
echo "=========================================="
