#!/usr/bin/env bash
set -u
BASE=http://127.0.0.1:8080
TOKEN=$(curl -s -X POST $BASE/api/auth/login -H 'Content-Type: application/json' -d '{"phone":"13800000000","password":"123456"}' | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
echo "TOKEN_PREFIX=${TOKEN:0:20}..."
echo "--- 2) GET /api/areas ---"
curl -s $BASE/api/areas -H "Authorization: Bearer $TOKEN"
echo
echo "--- 2b) GET /api/areas/1 ---"
curl -s $BASE/api/areas/1 -H "Authorization: Bearer $TOKEN"
echo
echo "--- 3) POST /api/declarations (grid cap 10) ---"
DECL=$(curl -s -X POST $BASE/api/declarations -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d '{"area_id":1,"device_id":2,"type":"grid","capacity_kw":10}')
echo "$DECL"
DID=$(echo "$DECL" | sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' | head -1)
echo "DID=$DID"
echo "--- 4) POST /api/declarations/$DID/approve ---"
curl -s -X POST $BASE/api/declarations/$DID/approve -H "Authorization: Bearer $TOKEN"
echo
NOW_S=$(date -u +%s)
START_S=$((NOW_S - 3600))
END_S=$((NOW_S + 3600))
START=$(date -u -r $START_S +%Y-%m-%dT%H:%M:%SZ)
END=$(date -u -r $END_S +%Y-%m-%dT%H:%M:%SZ)
echo "--- 5) POST /api/limits ($START ~ $END ratio 0.3) ---"
LIM=$(curl -s -X POST $BASE/api/limits -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d "{\"area_id\":1,\"ratio\":0.3,\"start_at\":\"$START\",\"end_at\":\"$END\"}")
echo "$LIM"
LID=$(echo "$LIM" | sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' | head -1)
echo "LID=$LID"
echo "--- 6) GET /api/limits/$LID/impact ---"
curl -s $BASE/api/limits/$LID/impact -H "Authorization: Bearer $TOKEN"
echo
MONTH_START=$(date -u +%Y-%m-01T00:00:00Z)
echo "--- 7) GET /api/timeseries (area=1, gen, $MONTH_START ~ $END) ---"
curl -s "$BASE/api/timeseries?area_id=1&metric=gen&from=$MONTH_START&to=$END" -H "Authorization: Bearer $TOKEN" | head -c 700
echo
echo "--- DONE ---"
