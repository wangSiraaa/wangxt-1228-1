#!/usr/bin/env bash
set -euo pipefail

API_BASE="http://localhost:8080/api"
TMPFILE=$(mktemp)
PASS=0
FAIL=0

log() { echo -e "  \033[1;34m$1\033[0m"; }
ok()  { echo -e "  ✓ \033[1;32m$1\033[0m"; PASS=$((PASS+1)); }
fail(){ echo -e "  ✗ \033[1;31m$1\033[0m"; FAIL=$((FAIL+1)); }
section() { echo -e "\n\033[1;36m=== $1 ===\033[0m"; }

cleanup() { rm -f "$TMPFILE"; }
trap cleanup EXIT

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║    分布式光伏消纳管理 - 限发影响电量估算 回归测试            ║"
echo "╚══════════════════════════════════════════════════════════════╝"

# ========== [1/8] 登录系统 ==========
section "[1/8] 登录系统（调度员 dispatcher）"
RESP=$(curl -s -X POST "$API_BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800000003","password":"123456"}')
TOKEN=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
  ok "登录成功，获取 token"
else
  fail "登录失败: $RESP"
  exit 1
fi
AUTH="Authorization: Bearer $TOKEN"

# ========== [2/8] 获取台区信息 ==========
section "[2/8] 获取台区信息"
RESP=$(curl -s "$API_BASE/areas" -H "$AUTH")
AREA1_ID=$(echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print([x['id'] for x in d['data'] if '阳光花园' in x['name']][0])")
AREA1_CAP=$(echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print([x['capacity_kw'] for x in d['data'] if '阳光花园' in x['name']][0])")
if [ -n "$AREA1_ID" ]; then
  ok "阳光花园台区 ID=$AREA1_ID, 容量=${AREA1_CAP}kW"
else
  fail "获取台区失败"
  exit 1
fi

# ========== [3/8] 获取当前限发指令列表 ==========
section "[3/8] 检查现有执行中指令的实时估算值"
RESP=$(curl -s "$API_BASE/limits?status=executing" -H "$AUTH")
echo "$RESP" > "$TMPFILE"
COUNT=$(python3 -c "import sys,json; d=json.load(sys.stdin); print(len(d['data']))" < "$TMPFILE")
if [ "$COUNT" -gt 0 ]; then
  ok "找到 $COUNT 条执行中指令"
  
  # 检查每条指令的样本数和估算值
  for i in $(seq 0 $((COUNT-1))); do
    ID=$(python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data'][$i]['id'])" < "$TMPFILE")
    SAMPLE=$(python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data'][$i]['sample_count'])" < "$TMPFILE")
    AVG=$(python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data'][$i]['avg_gen_kw'])" < "$TMPFILE")
    EST=$(python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data'][$i]['est_loss_kwh'])" < "$TMPFILE")
    STATUS=$(python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data'][$i]['status'])" < "$TMPFILE")
    
    log "指令#$ID: 状态=$STATUS, 样本=$SAMPLE, 均发=${AVG}kW, 预估损失=${EST}kWh"
    
    # 白天时段应该有样本
    if [ "$SAMPLE" -gt 0 ]; then
      ok "  ✓ 样本数 > 0 ($SAMPLE)"
    else
      fail "  ✗ 样本数为 0，白天时段应该有历史数据"
    fi
    
    # 平均功率应该 > 0
    if python3 -c "import sys; exit(0 if float(sys.argv[1]) > 0 else 1)" "$AVG"; then
      ok "  ✓ 平均功率 > 0 (${AVG}kW)"
    else
      fail "  ✗ 平均功率为 0"
    fi
    
    # 预估损失应该 > 0
    if python3 -c "import sys; exit(0 if float(sys.argv[1]) > 0 else 1)" "$EST"; then
      ok "  ✓ 预估损失 > 0 (${EST}kWh)"
    else
      fail "  ✗ 预估损失为 0"
    fi
  done
else
  log "暂无执行中指令"
fi

# ========== [4/8] 测试白天时段新建限发指令 ==========
section "[4/8] 测试白天时段新建限发指令（10:00-12:00）"

# 计算明天的日期
TOMORROW=$(date -v+1d +"%Y-%m-%d" 2>/dev/null || date -d "+1 day" +"%Y-%m-%d")
START_AT="${TOMORROW}T10:00:00+08:00"
END_AT="${TOMORROW}T12:00:00+08:00"

log "创建限发指令: 台区#$AREA1_ID, 比例=30%, 时段=$START_AT → $END_AT"
RESP=$(curl -s -X POST "$API_BASE/limits" \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d "{\"area_id\":$AREA1_ID,\"ratio\":0.3,\"start_at\":\"$START_AT\",\"end_at\":\"$END_AT\"}")

echo "$RESP"
CMD_ID=$(echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['id'])")
EST_LOSS=$(echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['est_loss_kwh'])")

if [ -n "$CMD_ID" ] && [ "$CMD_ID" != "null" ]; then
  ok "限发指令创建成功: 指令#$CMD_ID"
  log "创建时预计算 est_loss_kwh = ${EST_LOSS} kWh"
  
  # 验证创建时就有估算值
  if python3 -c "import sys; exit(0 if float(sys.argv[1]) > 0 else 1)" "$EST_LOSS"; then
    ok "  ✓ 创建时预计算值 > 0 (${EST_LOSS}kWh)"
  else
    fail "  ✗ 创建时预计算值为 0"
  fi
else
  fail "限发指令创建失败: $RESP"
  exit 1
fi

# ========== [5/8] 验证限发列表的实时估算 ==========
section "[5/8] 验证限发列表实时估算（不依赖创建时固化值）"
RESP=$(curl -s "$API_BASE/limits?status=executing" -H "$AUTH")
echo "$RESP" > "$TMPFILE"

# 找到刚创建的指令
LIST_SAMPLE=$(python3 -c "
import sys,json
d=json.load(sys.stdin)
for item in d['data']:
    if item['id'] == $CMD_ID:
        print(item['sample_count'])
        break
" < "$TMPFILE")

LIST_AVG=$(python3 -c "
import sys,json
d=json.load(sys.stdin)
for item in d['data']:
    if item['id'] == $CMD_ID:
        print(item['avg_gen_kw'])
        break
" < "$TMPFILE")

LIST_EST=$(python3 -c "
import sys,json
d=json.load(sys.stdin)
for item in d['data']:
    if item['id'] == $CMD_ID:
        print(item['est_loss_kwh'])
        break
" < "$TMPFILE")

log "列表查询结果: 样本=$LIST_SAMPLE, 均发=${LIST_AVG}kW, 预估=${LIST_EST}kWh"

# 验证样本数 > 0
if [ "$LIST_SAMPLE" -gt 0 ]; then
  ok "✓ 列表查询样本数 > 0 ($LIST_SAMPLE 个小时样本点)"
else
  fail "✗ 列表查询样本数为 0"
fi

# 验证平均功率 > 0
if python3 -c "import sys; exit(0 if float(sys.argv[1]) > 0 else 1)" "$LIST_AVG"; then
  ok "✓ 列表查询平均功率 > 0 (${LIST_AVG}kW)"
else
  fail "✗ 列表查询平均功率为 0"
fi

# 验证预估损失 > 0
if python3 -c "import sys; exit(0 if float(sys.argv[1]) > 0 else 1)" "$LIST_EST"; then
  ok "✓ 列表查询预估损失 > 0 (${LIST_EST}kWh)"
else
  fail "✗ 列表查询预估损失为 0"
fi

# 验证列表值与创建时预计算值一致（允许微小浮点误差）
MATCH=$(python3 -c "
import sys
a = float('$EST_LOSS')
b = float('$LIST_EST')
print('1' if abs(a - b) < 0.01 else '0')
")
if [ "$MATCH" = "1" ]; then
  ok "✓ 列表估算值与创建时预计算值一致"
else
  log "  注意: 列表值($LIST_EST)与创建值($EST_LOSS)略有差异，这是正常的（实时计算）"
fi

# ========== [6/8] 验证详情页 Impact 接口 ==========
section "[6/8] 验证详情页 Impact 接口"
RESP=$(curl -s "$API_BASE/limits/$CMD_ID/impact" -H "$AUTH")
echo "$RESP" > "$TMPFILE"

IMPACT_SAMPLE=$(python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['sample_count'])" < "$TMPFILE")
IMPACT_AVG=$(python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['avg_gen_kw'])" < "$TMPFILE")
IMPACT_EST=$(python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['est_loss_kwh'])" < "$TMPFILE")
IMPACT_DURATION=$(python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['duration_hours'])" < "$TMPFILE")

log "Impact 接口结果:"
log "  持续时长: ${IMPACT_DURATION}h"
log "  历史均发: ${IMPACT_AVG}kW"
log "  样本点数: $IMPACT_SAMPLE"
log "  预估损失: ${IMPACT_EST}kWh"

# 验证 Impact 接口数据
if [ "$IMPACT_SAMPLE" -gt 0 ]; then
  ok "✓ Impact 接口样本数 > 0 ($IMPACT_SAMPLE)"
else
  fail "✗ Impact 接口样本数为 0"
fi

if python3 -c "import sys; exit(0 if float(sys.argv[1]) > 0 else 1)" "$IMPACT_AVG"; then
  ok "✓ Impact 接口平均功率 > 0 (${IMPACT_AVG}kW)"
else
  fail "✗ Impact 接口平均功率为 0"
fi

if python3 -c "import sys; exit(0 if float(sys.argv[1]) > 0 else 1)" "$IMPACT_EST"; then
  ok "✓ Impact 接口预估损失 > 0 (${IMPACT_EST}kWh)"
else
  fail "✗ Impact 接口预估损失为 0"
fi

# 验证公式: est_loss = avg_gen * ratio * duration
RATIO=0.3
EXPECTED=$(python3 -c "
avg = float('$IMPACT_AVG')
dur = float('$IMPACT_DURATION')
est = avg * $RATIO * dur
print(round(est, 2))
")
FORMULA_CHECK=$(python3 -c "
a = float('$EXPECTED')
b = float('$IMPACT_EST')
print('1' if abs(a - b) < 0.01 else '0')
")
if [ "$FORMULA_CHECK" = "1" ]; then
  ok "✓ 估算公式正确: ${IMPACT_AVG}kW × ${RATIO} × ${IMPACT_DURATION}h = ${EXPECTED}kWh ✓"
else
  fail "✗ 估算公式错误: 期望 ${EXPECTED}kWh, 实际 ${IMPACT_EST}kWh"
fi

# ========== [7/8] 测试傍晚时段（验证跨时段查询） ==========
section "[7/8] 测试傍晚时段（16:00-18:00）"
START_AT2="${TOMORROW}T16:00:00+08:00"
END_AT2="${TOMORROW}T18:00:00+08:00"

log "创建限发指令: 台区#$AREA1_ID, 比例=50%, 时段=$START_AT2 → $END_AT2"
RESP=$(curl -s -X POST "$API_BASE/limits" \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d "{\"area_id\":$AREA1_ID,\"ratio\":0.5,\"start_at\":\"$START_AT2\",\"end_at\":\"$END_AT2\"}")

CMD_ID2=$(echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['id'])")
if [ -n "$CMD_ID2" ] && [ "$CMD_ID2" != "null" ]; then
  ok "傍晚时段限发指令创建成功: 指令#$CMD_ID2"
  
  # 查询 Impact 验证
  RESP=$(curl -s "$API_BASE/limits/$CMD_ID2/impact" -H "$AUTH")
  SAMPLE2=$(echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['sample_count'])")
  AVG2=$(echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['avg_gen_kw'])")
  EST2=$(echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['est_loss_kwh'])")
  
  log "傍晚时段 Impact 结果: 样本=$SAMPLE2, 均发=${AVG2}kW, 预估=${EST2}kWh"
  
  if [ "$SAMPLE2" -gt 0 ]; then
    ok "✓ 傍晚时段样本数 > 0 ($SAMPLE2)"
  else
    fail "✗ 傍晚时段样本数为 0"
  fi
  
  if python3 -c "import sys; exit(0 if float(sys.argv[1]) > 0 else 1)" "$EST2"; then
    ok "✓ 傍晚时段预估损失 > 0 (${EST2}kWh)"
  else
    fail "✗ 傍晚时段预估损失为 0"
  fi
else
  fail "傍晚时段限发指令创建失败: $RESP"
fi

# ========== [8/8] 估算值合理性验证 ==========
section "[8/8] 估算值合理性验证"

# 台区#1有两个并网设备: 10kW + 8kW = 18kW
# 白天发电峰值约 18kW，时段10:00-12:00平均约 15-18kW
# 限发比例30%，时长2小时，预估损失约 16kW × 0.3 × 2h = 9.6kWh

EXPECTED_MIN=5
EXPECTED_MAX=15

RANGE_CHECK=$(python3 -c "
est = float('$IMPACT_EST')
print('1' if $EXPECTED_MIN <= est <= $EXPECTED_MAX else '0')
")

log "台区#1 10:00-12:00 30%限发:"
log "  理论估算范围: ${EXPECTED_MIN}kWh ~ ${EXPECTED_MAX}kWh"
log "  实际估算值: ${IMPACT_EST}kWh"

if [ "$RANGE_CHECK" = "1" ]; then
  ok "✓ 估算值在合理范围内"
else
  log "  ⚠️  估算值可能超出预期范围（但不影响功能正确性）"
fi

# ========== 测试总结 ==========
echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
if [ "$FAIL" -eq 0 ]; then
  echo -e "║  \033[1;32m✅ 回归测试全部通过!  通过: $PASS / 失败: $FAIL\033[0m                          ║"
  EXIT_CODE=0
else
  echo -e "║  \033[1;31m❌ 回归测试存在失败!  通过: $PASS / 失败: $FAIL\033[0m                        ║"
  EXIT_CODE=1
fi
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

exit $EXIT_CODE
