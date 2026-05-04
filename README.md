# etcd-ticket

## 專案目錄

1. root
````
etcd-ticket/
├── backend/        # Go (API + etcd + WebSocket)
├── frontend/       # React
├── docker/         # (選用) etcd / 部署
├── scripts/        # 測試 & 壓測
├── README.md

````

2. backend (Go)
```
backend/
├── cmd/
│   └── server/
│       └── main.go          # 入口點
│
├── internal/
│   ├── api/                 # HTTP handler
│   │   ├── handler.go
│   │   ├── ticket.go        # 搶票 API
│   │   └── ws.go            # WebSocket
│   │
│   ├── service/             # 商業邏輯
│   │   └── ticket_service.go
│   │
│   ├── repository/          # etcd 操作封裝
│   │   └── etcd_repo.go
│   │
│   ├── etcd/                # etcd client 初始化
│   │   └── client.go
│   │
│   ├── lock/                # 分散式鎖
│   │   └── mutex.go
│   │
│   ├── watcher/             # watch 機制
│   │   └── ticket_watcher.go
│   │
│   └── model/               # 資料結構
│       └── ticket.go
│
├── pkg/                     # 可重用工具
│   └── response/
│       └── response.go
│
├── configs/
│   └── config.yaml
│
├── go.mod
└── go.sum
```
### 目錄說明
- api/ 👉 接收 request（Gin handler）
- service/ 👉 搶票邏輯（核心）
- repository/ 👉 操作 etcd
- lock/ 👉 分散式鎖（重點🔥）
- watcher/ 👉 etcd watch → WebSocket 推播


3. frontend (React)
```
frontend/
├── public/
│
├── src/
│   ├── api/                 # 呼叫後端
│   │   └── ticketApi.js
│   │
│   ├── components/
│   │   ├── TicketPanel.jsx  # 票數顯示
│   │   ├── BuyButton.jsx
│   │   └── StatusBoard.jsx  # 即時狀態
│   │
│   ├── hooks/
│   │   └── useWebSocket.js  # WS 封裝
│   │
│   ├── pages/
│   │   └── Home.jsx
│   │
│   ├── context/             # 全域狀態
│   │   └── TicketContext.jsx
│   │
│   ├── App.jsx
│   └── main.jsx
│
├── package.json
└── vite.config.js
```

### 目錄說明
- api/ 👉 REST API
- hooks/useWebSocket 👉 接 etcd watch 推播（⭐）
- context/ 👉 管理票數狀態
- components/ 👉 UI 分離

4. docker
```
docker/
├── docker-compose.yml   # etcd + backend
└── etcd.env
```

