#!/bin/bash
# scripts/integration-test.sh
# 後端 end-to-end 整合測試
# 前置：server 已起 (port 8080)、PostgreSQL/Redis/etcd 容器都 Up

set -u  # 不開 -e，要讓失敗 case 繼續跑

BASE_URL="${BASE_URL:-http://localhost:8080}"
SUFFIX=$(date +%s)
PASS=0
FAIL=0

# ── helpers ──────────────────────────────────────────────────────
curl_get()  { curl -s -o /tmp/itest_body -w "%{http_code}" "$BASE_URL$1"; }
curl_post() {
  curl -s -o /tmp/itest_body -w "%{http_code}" \
    -X POST "$BASE_URL$1" \
    -H 'Content-Type: application/json' \
    -d "$2"
}

assert_status() {
  local name=$1 want=$2 got=$3
  local body
  body=$(cat /tmp/itest_body 2>/dev/null || true)
  if [[ "$got" == "$want" ]]; then
    echo "✓ PASS  $name  (status=$got)"
    PASS=$((PASS+1))
  else
    echo "✗ FAIL  $name  want=$want got=$got"
    echo "        body: $body"
    FAIL=$((FAIL+1))
  fi
}

assert_body_contains() {
  local name=$1 needle=$2
  if grep -q -- "$needle" /tmp/itest_body 2>/dev/null; then
    echo "✓ PASS  $name  (body contains \"$needle\")"
    PASS=$((PASS+1))
  else
    echo "✗ FAIL  $name  body missing \"$needle\""
    echo "        body: $(cat /tmp/itest_body 2>/dev/null || echo '<empty>')"
    FAIL=$((FAIL+1))
  fi
}

echo "=== Backend Integration Test (suffix=$SUFFIX) ==="
echo "BASE_URL=$BASE_URL"
echo ""

# ── A. healthz ──────────────────────────────────────────────────
echo "--- A. Healthz ---"
assert_status "A1 GET /healthz" 200 "$(curl_get /healthz)"
assert_body_contains "A2 healthz 含 status:ok" '"status":"ok"'

# ── B. status 查詢 ──────────────────────────────────────────────
echo ""
echo "--- B. Status ---"
assert_status "B1 GET /api/tickets/status" 200 "$(curl_get /api/tickets/status)"
assert_body_contains "B2 status 含 area_id" '"area_id"'
assert_body_contains "B3 status 含 available" '"available"'

# ── C. Happy Path: reserve + checkout ───────────────────────────
echo ""
echo "--- C. Happy Path (reserve → checkout) ---"
USER_A="alice_$SUFFIX"
PAYLOAD_A="{\"user_name\":\"$USER_A\",\"phone_num\":\"0911111111\",\"area\":1}"
assert_status "C1 reserve $USER_A area 1" 200 "$(curl_post /api/tickets/reserve "$PAYLOAD_A")"
assert_body_contains "C2 reserve 回 reserved:true" '"reserved":true'
assert_status "C3 checkout $USER_A area 1" 200 "$(curl_post /api/tickets/checkout "$PAYLOAD_A")"
assert_body_contains "C4 checkout 回 order_published:true" '"order_published":true'

# ── D. Invalid Inputs ───────────────────────────────────────────
echo ""
echo "--- D. Invalid Inputs ---"
assert_status "D1 reserve 空 body" 400 "$(curl_post /api/tickets/reserve '{}')"
assert_status "D2 reserve 非法 JSON" 400 "$(curl_post /api/tickets/reserve 'not json')"
PAYLOAD_99="{\"user_name\":\"x_$SUFFIX\",\"phone_num\":\"0900000000\",\"area\":99}"
assert_status "D3 reserve area 99 (不存在)" 404 "$(curl_post /api/tickets/reserve "$PAYLOAD_99")"

# ── E. Checkout Without Reserve ─────────────────────────────────
echo ""
echo "--- E. Checkout Without Reserve ---"
USER_NO="ghost_$SUFFIX"
PAYLOAD_NO="{\"user_name\":\"$USER_NO\",\"phone_num\":\"0900000000\",\"area\":1}"
assert_status "E1 checkout $USER_NO (未先 reserve)" 410 "$(curl_post /api/tickets/checkout "$PAYLOAD_NO")"

# ── F. Unknown Path ─────────────────────────────────────────────
echo ""
echo "--- F. Unknown Path ---"
assert_status "F1 GET /nope" 404 "$(curl_get /nope)"

# ── G. DB 持久化驗證 ────────────────────────────────────────────
echo ""
echo "--- G. DB Persistence ---"
echo "  ⏳ 等 1 秒讓 order worker 寫 DB..."
sleep 1
DB_COUNT=$(docker exec etcd-ticket-postgres psql -U postgres -d etcd_ticket -t -A -c \
  "SELECT COUNT(*) FROM orders WHERE user_name='$USER_A'" 2>/dev/null | tr -d '[:space:]')
if [[ "$DB_COUNT" == "1" ]]; then
  echo "✓ PASS  G1 DB orders 表含 $USER_A 紀錄"
  PASS=$((PASS+1))
else
  echo "✗ FAIL  G1 DB 預期 1 筆，實際 \"$DB_COUNT\""
  FAIL=$((FAIL+1))
fi

# ── Summary ─────────────────────────────────────────────────────
echo ""
echo "================================="
echo "  PASS=$PASS  FAIL=$FAIL"
echo "================================="
if [[ $FAIL -eq 0 ]]; then
  echo "✓ ALL PASS"
  exit 0
else
  echo "✗ $FAIL TEST(S) FAILED"
  exit 1
fi
