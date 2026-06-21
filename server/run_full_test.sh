#!/usr/bin/env bash
set -euo pipefail

cd /Users/mingyuan/workspace/sihuo/wangxtw3/1228/server

echo "=== Step 1: 编译项目 ==="
go build .
echo "编译成功!"

echo ""
echo "=== Step 2: 启动后端服务 ==="
# 先杀掉可能存在的旧进程
pkill -f "go run main.go" 2>/dev/null || true
pkill -f main 2>/dev/null || true
sleep 1

# 启动服务
DB_DRIVER=sqlite go run main.go > server.log 2>&1 &
SERVER_PID=$!
echo "服务已启动，PID=$SERVER_PID"

echo ""
echo "=== Step 3: 等待3秒让服务启动 ==="
sleep 3

# 检查服务是否正常运行
if ps -p $SERVER_PID > /dev/null; then
    echo "服务运行正常"
else
    echo "服务启动失败，查看日志:"
    cat server.log
    exit 1
fi

echo ""
echo "=== Step 4: 执行回归测试 ==="
bash test_estimation_regression.sh
TEST_EXIT_CODE=$?

echo ""
echo "=== Step 5: 停止服务 ==="
kill $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true

echo ""
echo "=== 测试完成，退出码: $TEST_EXIT_CODE ==="
exit $TEST_EXIT_CODE
