package services

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
)

type WSMessage struct {
	Type      string          `json:"type"`
	DeviceID  string          `json:"device_id"`
	Data      json.RawMessage `json:"data,omitempty"`
	IsOnline  *bool           `json:"is_online,omitempty"`
	LastSeen  *time.Time      `json:"last_seen_at,omitempty"`
	CommandID string          `json:"command_id,omitempty"`
	Status    string          `json:"status,omitempty"`
	Timestamp *time.Time      `json:"timestamp,omitempty"`
}

type RealtimeService struct {
	mu   sync.RWMutex
	hubs map[string]map[*websocket.Conn]bool
}

func NewRealtimeService() *RealtimeService {
	return &RealtimeService{
		hubs: make(map[string]map[*websocket.Conn]bool),
	}
}

func (s *RealtimeService) Register(deviceID string, conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.hubs[deviceID] == nil {
		s.hubs[deviceID] = make(map[*websocket.Conn]bool)
	}
	s.hubs[deviceID][conn] = true
}

func (s *RealtimeService) Unregister(deviceID string, conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if hub, ok := s.hubs[deviceID]; ok {
		delete(hub, conn)
		if len(hub) == 0 {
			delete(s.hubs, deviceID)
		}
	}
}

func (s *RealtimeService) Broadcast(deviceID string, msg WSMessage) {
	s.mu.RLock()
	conns := make([]*websocket.Conn, 0, len(s.hubs[deviceID]))
	for conn := range s.hubs[deviceID] {
		conns = append(conns, conn)
	}
	s.mu.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("realtime: marshal error: %v", err)
		return
	}

	for _, conn := range conns {
		if err := conn.WriteMessage(1, data); err != nil {
			log.Printf("realtime: write error: %v", err)
		}
	}
}
