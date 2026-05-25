#!/bin/bash
# scripts/start-backend.sh — 一鍵啟動整個後端 stack
# OrbStack → PostgreSQL+Redis → etcd cluster → Go server
# 全部 idempotent，重複跑不會壞

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BACKEND="$PROJECT_ROOT/backend"
LOG_FILE=/tmp/etcd-ticket-server.log

echo "=== etcd-ticket backend 一鍵啟動 ==="
echo ""

# ── 1. Docker daemon (OrbStack) ──────────────────────────────────
echo "[1/4] Docker daemon"
if docker info >/dev/null 2>&1; then
  echo "  ✓ already running"
else
  echo "  → 啟動 OrbStack..."
  open -a OrbStack
  echo -n "  ⏳ 等待 daemon ready"
  until docker info >/dev/null 2>&1; do echo -n "."; sleep 1; done
  echo " ✓"
fi

# ── 2. PostgreSQL + Redis ───────────────────────────────────────
echo ""
echo "[2/4] PostgreSQL + Redis"
cd "$BACKEND/internal/db"
docker compose up -d >/dev/null
echo -n "  ⏳ 等 postgres ready"
until docker exec etcd-ticket-postgres pg_isready -U postgres >/dev/null 2>&1; do echo -n "."; sleep 1; done
echo " ✓"
echo -n "  ⏳ 等 redis ready"
until docker exec etcd-ticket-redis redis-cli ping 2>/dev/null | grep -q PONG; do echo -n "."; sleep 1; done
echo " ✓"

# ── 3. etcd cluster (3 節點) ────────────────────────────────────
echo ""
echo "[3/4] etcd cluster (3 節點)"
cd "$BACKEND/internal/etcd/etcd-cluster"
docker compose up -d >/dev/null
echo -n "  ⏳ 等 etcd cluster ready"
until docker exec etcd1 etcdctl --endpoints=http://etcd1:2379 endpoint health >/dev/null 2>&1; do echo -n "."; sleep 1; done
echo " ✓"

# ── 4. Go server ────────────────────────────────────────────────
echo ""
echo "[4/4] Go server (port 8080)"
if lsof -i :8080 >/dev/null 2>&1; then
  PID=$(lsof -ti :8080 | head -1)
  echo "  ✓ already running (PID $PID)"
else
  echo "  → 啟動 server，log 寫到 $LOG_FILE"
  cd "$BACKEND"
  nohup go run ./cmd/server > "$LOG_FILE" 2>&1 &
  disown
  echo -n "  ⏳ 等 port 8080 ready"
  for _ in $(seq 1 60); do
    if lsof -i :8080 >/dev/null 2>&1; then break; fi
    echo -n "."
    sleep 1
  done
  if lsof -i :8080 >/dev/null 2>&1; then
    echo " ✓ (PID $(lsof -ti :8080 | head -1))"
  else
    echo " ✗"
    echo "  ✗ server 60 秒內沒起，看 log：tail $LOG_FILE"
    exit 1
  fi
fi

# ── Summary ─────────────────────────────────────────────────────
echo ""
echo "=========================================="
echo "✓ 全部 ready"
echo "=========================================="
echo "  Server URL:   http://localhost:8080"
echo "  Health:       curl http://localhost:8080/healthz"
echo "  Server log:   tail -f $LOG_FILE"
echo "  整合測試:     ./scripts/integration-test.sh"
echo "  停掉一切:     ./scripts/stop-backend.sh"
