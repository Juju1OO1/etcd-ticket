# internal/wsserver

獨立的 WebSocket server（port 8888），負責維護前端連線並廣播即時票況事件。

整體架構：
- 前端透過 `ws://localhost:8888/ws` 建立 WebSocket 長連線
- watcher 透過 HTTP POST 送事件進來
- wsserver 把事件廣播給所有在線的前端連線

---

## wsserver.go

### Start(ctx, host, port)

啟動 HTTP server，以 goroutine 非同步執行，掛載三個路由：

| Path | 說明 |
|------|------|
| `GET /ws` | 前端 WebSocket 連線入口 |
| `POST /broadcast/available-tickets` | 接收剩餘票數更新事件 |
| `POST /broadcast/sold-tickets` | 接收成交紀錄事件 |

ctx 取消時自動關閉 server。

### Server struct

| 欄位 | 說明 |
|------|------|
| `clients` | 所有在線前端連線（`map[*websocket.Conn]struct{}`） |
| `mu` | 保護 clients map 的 mutex |
| `upgrader` | Gorilla WebSocket upgrader，允許所有 origin |

### handleWebSocket

把 HTTP 請求升級為 WebSocket 連線，加入 clients map，並進入阻塞讀取迴圈。
前端斷線時自動從 clients map 移除。

### handleAvailableTickets

接收 watcher 送來的剩餘票數事件，呼叫 `broadcast()` 推送給所有前端。
廣播格式：`{ "type": "ticket_available", "area_id": 1, "available": 99 }`

### handleSoldTickets

接收 watcher 送來的成交紀錄事件，呼叫 `broadcast()` 推送給所有前端。
廣播格式：`{ "type": "ticket_sold", "user": "henry", "area": 1, "phone": "0912345678" }`

### broadcast(message)

先在 mutex 保護下複製 clients slice，釋放 mutex 後再逐一寫入。
這樣廣播期間其他 goroutine 仍可新增/移除連線，不會卡住。
每個連線設定 5 秒寫入 deadline，寫入失敗則移除該連線。
