# internal/service

商業邏輯層，包含搶票核心與訂單處理。

---

## ticket_service.go

etcd 搶票的完整實作。

### AreaConfig

初始化時傳入的區域設定，包含區域名稱與票數上限。

### InitEtcd(ctx, areas)

服務啟動時呼叫一次，把各區域的初始狀態寫進 etcd：
- `area{n}/limit/limit` — 總票數
- `area{n}/limit/current_limit` — 剩餘票數
- `area{n}/status` — 售票狀態（on）

### TicketData

搶票請求的資料結構，包含 `UserName`、`PhoneNum`、`Area`。

### Lock_And_Hold(ctx, td)

第一階段搶票：
1. 競爭 etcd mutex（`area{n}/lock`）
2. 取得鎖後讀取 holder 數量 + paying 數量
3. 若有空位 → 建立 `area{n}/holder/{userName}`，綁定 Lease TTL 300 秒
4. 釋放鎖，回傳是否搶到

TTL 到期後 etcd 自動刪除 holder，名額自動釋放，不需要手動清理逾時者。

### Checkout(ctx, td)

第二階段結帳，搶到票的用戶付款完成後呼叫：
1. 原子驗證 holder 存在，同時建立 paying 標記（TTL 30 秒）
2. 讀取 current_limit，執行 CAS Txn：
   - 刪除 holder
   - 刪除 paying 標記
   - current_limit - 1
   - 寫入 `area{n}/sold/{userName}`（值為電話號碼）
   - 若 current_limit 降至 0，同時把 status 設為 off
3. 若被並發修改則自動重試

---

## order_service.go

訂單非同步寫入，搶票成功後把訂單送進 Redis Stream。

### PublishOrder(ctx, td)

在 `Checkout()` 成功後由 handler 呼叫，產生 UUID 作為 `order_id`，透過 `mq.Publish()` 寫進 Redis Stream，立即回傳不阻塞。

### StartOrderWorker(ctx)

main.go 啟動時以 goroutine 執行，持續用 `XREADGROUP` 阻塞等待新訊息：
- 拿到訊息 → `processMessage()` 解析並寫入 PostgreSQL
- 寫入成功 → `mq.Ack()` 確認消費
- 寫入失敗 → 不 Ack，訊息留在 pending list，下次重新消費

### processMessage(ctx, msgID, values)

把 Redis Stream 的 map 轉成 `model.Order`，呼叫 `repository.InsertOrder()` 寫進 DB。
