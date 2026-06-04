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
	mu       sync.Mutex
	clients  map[*websocket.Conn]struct{}
	upgrader websocket.Upgrader
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
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "websocket upgrade failed", http.StatusBadRequest)
		return
	}

	s.addClient(conn)
	defer s.removeClient(conn)

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

func (s *Server) handleSoldTickets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload SoldTicketPayload
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

	s.mu.Lock()
	clients := make([]*websocket.Conn, 0, len(s.clients))
	for conn := range s.clients {
		clients = append(clients, conn)
	}
	s.mu.Unlock()

	for _, conn := range clients {
		_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			s.removeClient(conn)
		}
	}
}
