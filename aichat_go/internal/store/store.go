package store

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Conversation as stored.
type Conversation struct {
	ID          string
	TenantID    string
	WorkspaceID string
	UserID      string
	Title       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Message as stored.
type Message struct {
	ID             string
	ConversationID string
	Role           string
	Content        string
	Provider       string
	Model          string
	InputTokens    int
	OutputTokens   int
	CreatedAt      time.Time
}

// InMemoryStore is a simple in-memory store for V1 (replace with Postgres/DynamoDB later).
type InMemoryStore struct {
	mu            sync.RWMutex
	conversations map[string]*Conversation
	messages      map[string][]*Message
	convByTenant  map[string][]string
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		conversations: make(map[string]*Conversation),
		messages:      make(map[string][]*Message),
		convByTenant:  make(map[string][]string),
	}
}

func (s *InMemoryStore) CreateConversation(tenantID, workspaceID, userID string) (*Conversation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := "c_" + uuid.New().String()
	now := time.Now()
	c := &Conversation{
		ID:          id,
		TenantID:    tenantID,
		WorkspaceID: workspaceID,
		UserID:      userID,
		Title:       "New conversation",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.conversations[id] = c
	s.messages[id] = nil
	s.convByTenant[tenantID] = append(s.convByTenant[tenantID], id)
	return c, nil
}

func (s *InMemoryStore) GetConversation(id string) (*Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.conversations[id], nil
}

func (s *InMemoryStore) ListConversations(tenantID string) ([]*Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := s.convByTenant[tenantID]
	out := make([]*Conversation, 0, len(ids))
	for _, id := range ids {
		if c := s.conversations[id]; c != nil {
			out = append(out, c)
		}
	}
	return out, nil
}

func (s *InMemoryStore) AppendMessage(convID, role, content, provider, model string, inputTokens, outputTokens int) (*Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := "m_" + uuid.New().String()
	now := time.Now()
	m := &Message{
		ID:             id,
		ConversationID: convID,
		Role:           role,
		Content:        content,
		Provider:       provider,
		Model:          model,
		InputTokens:    inputTokens,
		OutputTokens:   outputTokens,
		CreatedAt:      now,
	}
	s.messages[convID] = append(s.messages[convID], m)
	if c := s.conversations[convID]; c != nil {
		c.UpdatedAt = now
	}
	return m, nil
}

func (s *InMemoryStore) GetMessages(convID string) ([]*Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]*Message(nil), s.messages[convID]...), nil
}

func (s *InMemoryStore) UpdateConversationTitle(convID, title string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if c := s.conversations[convID]; c != nil {
		c.Title = title
	}
	return nil
}
