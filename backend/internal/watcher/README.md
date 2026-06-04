# internal/watcher

etcd 事件監聽橋接層，負責把 etcd 的票況變化轉送到 WebSocket server，再由 wsserver 廣播給前端。

整體資料流：

```
etcd 變化 → service.Watch*() → watcher HTTP POST → wsserver → WebSocket → 前端
```

---

## available_ticket_ws.go

### StartHTTPClient(ctx, areaIDs, websocketHost, websocketPort)

為每個 `areaID` 啟動一個獨立 goroutine，監聽 `service.WatchAvailableTickets()` 回傳的 channel。

每當 etcd 的剩餘票數發生變化，立即用 HTTP POST 送到 wsserver 的 `/broadcast/available-tickets`，payload 格式：

```json
{ "type": "ticket_available", "area_id": 1, "available": 99 }
```

錯誤會丟進 errCh，由 main.go 的 goroutine 印出，不影響主流程。

---

## sold_ticket_ws.go

### StartSoldTicketHTTPClient(ctx, areaIDs, websocketHost, websocketPort)

為每個 `areaID` 啟動一個獨立 goroutine，監聽 `service.WatchSoldTickets()` 回傳的 channel。

每當有使用者成功購票（etcd 寫入 sold key），立即 HTTP POST 到 wsserver 的 `/broadcast/sold-tickets`，payload 格式：

```json
{ "type": "ticket_sold", "user_name": "henry", "phone": "0912345678", "area_id": 1 }
```

前端收到後顯示即時成交跑馬燈（Toast notification）。
