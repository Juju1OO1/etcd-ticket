// 接收etcd變化事件，並把剩餘票數透過WebSocket廣播給前端的Server
package wsserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)


type Server struct {
	mu       sync.Mutex // 這裡的 mutex 是用來保護 clients 這個 map 的併發安全，因為多個 goroutine 可能同時訪問和修改 clients。
	clients  map[*websocket.Conn]struct{}  // 用來儲存「當前所有在線上的前端連線」的集合
	upgrader websocket.Upgrader   // Gorilla WebSocket 套件的核心元件。負責把標準的 HTTP 請求「升級」成持久的 WebSocket 長連線
}

func NewServer() *Server {
	return &Server{
		clients: make(map[*websocket.Conn]struct{}),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

type AvailableTicketPayload struct {
	Type      string `json:"type"`
	AreaID    int    `json:"area_id"`
	Available int64  `json:"available"`
}

type SoldTicketPayload struct {
	Type     string `json:"type"`
	UserName string `json:"user_name"`
	Phone    string `json:"phone"`
	AreaID   int    `json:"area_id"`
}

type TicketUpdateEvent struct {
	Type      string `json:"type"`
	AreaID    int    `json:"area_id"`
	Available int64  `json:"available"`
}

type SoldLogEvent struct {
	Type     string `json:"type"`
	UserName string `json:"user"`
	AreaID   int    `json:"area"`
	Phone    string `json:"phone"`
}

// 開兩個 POST endpoint 等 watcher/available_ticket_ws.go 送事件來：
func Start(ctx context.Context, host string, port int) error {
	server := NewServer()
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", server.handleWebSocket)
	mux.HandleFunc("/broadcast/available-tickets", server.handleAvailableTickets)
	mux.HandleFunc("/broadcast/sold-tickets", server.handleSoldTickets)

	addr := fmt.Sprintf("%s:%d", host, port)
	srv := &http.Server{Addr: addr, Handler: mux}

	go func() {
		<-ctx.Done()
		_ = srv.Close()
	}()

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("wsserver serve error: %v\n", err)
		}
	}()

	return nil
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil) // 1. 升級協定，成功後 conn 就是這個前端連線對應的 WebSocket 物件
	if err != nil {
		http.Error(w, "websocket upgrade failed", http.StatusBadRequest)
		return
	}

	s.addClient(conn)      // 2. 將連線加入全域名單，以便後續廣播訊息時能夠找到這個連線
	fmt.Printf("[WS] client connected from %s\n", r.RemoteAddr)
	defer func() {
		s.removeClient(conn)
		fmt.Printf("[WS] client disconnected from %s\n", r.RemoteAddr)
	}()

	// 它會去讀取該 TCP Socket 的 Receive Buffer（接收緩衝區）。
	// 進入阻塞狀態（Blocking）：如果此時前端（客戶端）沒有發送任何 WebSocket Frame（資料幀），作業系統會將目前這個 Goroutine 掛起（Suspend），並移出 CPU 的排程。此時，這個 Goroutine 完全不佔用任何 CPU 算力，它在等待作業系統核心（Kernel）的網路事件通知。
	for {
		if _, _, err := conn.NextReader(); err != nil {
			return
		}
	}
}

func (s *Server) handleAvailableTickets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload AvailableTicketPayload
	// 設計邏輯：使用 json.NewDecoder 進行串流解碼
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	s.broadcast(TicketUpdateEvent{
		Type:      "ticket_available",
		AreaID:    payload.AreaID,
		Available: payload.Available,
	})

	w.WriteHeader(http.StatusNoContent)
}

// 處理賣出日誌  當有使用者成功買到票，負責訂單的系統會 POST 到這裡，用來在前端即時跑馬燈顯示
func (s *Server) handleSoldTickets(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload SoldTicketPayload

	fmt.Printf(
		"[wsserver sold] %+v\n",
		payload,
	)

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	s.broadcast(SoldLogEvent{
		Type:     "ticket_sold",
		UserName: payload.UserName,
		AreaID:   payload.AreaID,
		Phone:    payload.Phone,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) addClient(conn *websocket.Conn) {

	s.mu.Lock()
	s.clients[conn] = struct{}{}
	fmt.Printf(
		"Client Connected, total=%d\n",
		len(s.clients)+1,
	)

	s.mu.Unlock()
}

func (s *Server) removeClient(conn *websocket.Conn) {
	s.mu.Lock()
	delete(s.clients, conn)
	s.mu.Unlock()
	_ = conn.Close()
}

func (s *Server) broadcast(message any) {

	data, err := json.Marshal(message)
	if err != nil {
		return
	}

	fmt.Printf(
		"[broadcast] %s\n",
		string(data),
	)

	s.mu.Lock()
	clients := make([]*websocket.Conn, 0, len(s.clients))  // 這裡的邏輯是：先把目前 clients map 裡的連線複製到一個新的 slice 裡，然後在 mutex 還沒釋放之前就把這個 slice 的內容讀取完畢。這樣可以確保在廣播訊息的過程中，其他 goroutine 還是可以繼續新增或移除 clients，而不會被鎖住。
	for conn := range s.clients {
		clients = append(clients, conn)
	}

	fmt.Printf(
		"Broadcast to %d clients\n",
		len(clients),
	)

	s.mu.Unlock()

	fmt.Printf("[WS] broadcasting %s to %d client(s)\n", string(data), len(clients))

	for _, conn := range clients {
		_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			s.removeClient(conn)
		}
	}
}
