# test/etcd

etcd 功能測試，驗證叢集連線、CRUD、Watch 及搶票邏輯是否正常運作。

## 測試範圍

1. **docker-compose** — 建立三節點 etcd 叢集（etcd1 / etcd2 / etcd3）
2. **client 封裝** — 與 etcd server 通訊（Get / Put / Delete / Watch）
3. **ticket_service** — 搶票核心邏輯（Lock_And_Hold / Checkout）

## 前置條件

- 已啟動 etcd 叢集（見 `backend/internal/etcd/README.md`）

## 執行方式

```bash
cd backend
go test ./test/etcd/...
```
