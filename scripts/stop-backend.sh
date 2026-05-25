#!/bin/bash
# scripts/stop-backend.sh — 反向關閉整個後端 stack
# Go server → etcd cluster → PostgreSQL/Redis
# Docker daemon (OrbStack) 不關，留給使用者自己決定

set -u

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BACKEND="$PROJECT_ROOT/backend"

echo "=== etcd-ticket backend 一鍵關閉 ==="
echo ""

# ── 1. Go server ────────────────────────────────────────────────
echo "[1/3] Go server"
PIDS=$(lsof -ti :8080 2>/dev/null || true)
if [ -z "$PIDS" ]; then
  echo "  ✓ not running"
else
  echo "  → kill PIDs: $PIDS"
  echo "$PIDS" | xargs kill 2>/dev/null || true
  sleep 1
  # 還沒死就 SIGKILL
  PIDS_LEFT=$(lsof -ti :8080 2>/dev/null || true)
  if [ -n "$PIDS_LEFT" ]; then
    echo "$PIDS_LEFT" | xargs kill -9 2>/dev/null || true
  fi
  echo "  ✓ stopped"
fi

# ── 2. etcd cluster ─────────────────────────────────────────────
echo ""
echo "[2/3] etcd cluster"
cd "$BACKEND/internal/etcd/etcd-cluster"
if docker info >/dev/null 2>&1; then
  docker compose down >/dev/null 2>&1 || true
  echo "  ✓ stopped"
else
  echo "  ⚠ docker daemon 沒在跑，跳過"
fi

# ── 3. PostgreSQL + Redis ───────────────────────────────────────
echo ""
echo "[3/3] PostgreSQL + Redis"
cd "$BACKEND/internal/db"
if docker info >/dev/null 2>&1; then
  docker compose down >/dev/null 2>&1 || true
  echo "  ✓ stopped"
else
  echo "  ⚠ docker daemon 沒在跑，跳過"
fi

echo ""
echo "=========================================="
echo "✓ 全部 stopped"
echo "=========================================="
echo "  OrbStack 仍在背景跑，要關手動從狀態列退出"
echo "  Volume 資料保留（PostgreSQL 資料、etcd 資料）"
echo "  要清資料重來： docker compose -f backend/internal/db/docker-compose.yaml down -v"
