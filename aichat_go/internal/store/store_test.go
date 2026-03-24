package store

import (
	"testing"
)

func TestInMemoryStore_CreateAndGetConversation(t *testing.T) {
	s := NewInMemoryStore()
	c, err := s.CreateConversation("t1", "w1", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if c.TenantID != "t1" || c.WorkspaceID != "w1" || c.UserID != "u1" {
		t.Fatalf("conversation fields: %+v", c)
	}
	if c.ID == "" {
		t.Fatal("expected id")
	}
	got, err := s.GetConversation(c.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.ID != c.ID {
		t.Fatalf("GetConversation = %+v", got)
	}
}

func TestInMemoryStore_ListConversations(t *testing.T) {
	s := NewInMemoryStore()
	a, _ := s.CreateConversation("tenant-a", "", "u")
	b, _ := s.CreateConversation("tenant-a", "", "u")
	_, _ = s.CreateConversation("tenant-b", "", "u")

	list, err := s.ListConversations("tenant-a")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("len = %d, want 2", len(list))
	}
	ids := map[string]bool{a.ID: true, b.ID: true}
	for _, c := range list {
		if !ids[c.ID] {
			t.Fatalf("unexpected id %s", c.ID)
		}
	}
}

func TestInMemoryStore_AppendAndGetMessages(t *testing.T) {
	s := NewInMemoryStore()
	c, _ := s.CreateConversation("t", "w", "u")
	m1, err := s.AppendMessage(c.ID, "user", "hello", "openai", "gpt-4o-mini", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	m2, err := s.AppendMessage(c.ID, "assistant", "hi", "openai", "gpt-4o-mini", 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	msgs, err := s.GetMessages(c.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 2 {
		t.Fatalf("len = %d", len(msgs))
	}
	if msgs[0].ID != m1.ID || msgs[0].Content != "hello" {
		t.Fatalf("first message: %+v", msgs[0])
	}
	if msgs[1].ID != m2.ID || msgs[1].InputTokens != 1 || msgs[1].OutputTokens != 2 {
		t.Fatalf("second message: %+v", msgs[1])
	}
}

func TestInMemoryStore_UpdateConversationTitle(t *testing.T) {
	s := NewInMemoryStore()
	c, _ := s.CreateConversation("t", "w", "u")
	if err := s.UpdateConversationTitle(c.ID, "Renamed"); err != nil {
		t.Fatal(err)
	}
	got, _ := s.GetConversation(c.ID)
	if got.Title != "Renamed" {
		t.Fatalf("title = %q", got.Title)
	}
}
