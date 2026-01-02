package session

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Session struct {
	Code         string
	OrgID        uuid.UUID
	SharerConn   *websocket.Conn
	ReceiverConn *websocket.Conn
	CreatedAt    time.Time
	Done         chan struct{}
}

type Manager struct {
	sessions map[string]*Session // code -> session
	mu       sync.RWMutex
	ttl      time.Duration
}

func NewManager(ttl time.Duration) *Manager {
	m := &Manager{
		sessions: make(map[string]*Session),
		ttl:      ttl,
	}

	go m.cleanup()
	return m
}

func (m *Manager) Create(orgID uuid.UUID, conn *websocket.Conn) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	code := generateCodeName()
	s := &Session{
		Code:       code,
		OrgID:      orgID,
		SharerConn: conn,
		CreatedAt:  time.Now(),
		Done:       make(chan struct{}),
	}
	m.sessions[code] = s

	return s
}

func (m *Manager) Get(code string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[code]
}

func (m *Manager) GetByOrg(orgID uuid.UUID) []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Session
	for _, s := range m.sessions {
		if s.OrgID == orgID {
			result = append(result, s)
		}
	}

	return result
}

func (m *Manager) Delete(code string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, code)
}

func (m *Manager) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		m.mu.Lock()
		now := time.Now()

		for code, s := range m.sessions {
			if now.Sub(s.CreatedAt) > m.ttl {
				delete(m.sessions, code)
			}
		}

		m.mu.Unlock()
	}
}
