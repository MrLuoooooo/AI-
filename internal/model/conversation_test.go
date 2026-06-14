package model

import (
	"encoding/json"
	"testing"
)

func TestCreateConversationRequest_JSON(t *testing.T) {
	req := CreateConversationRequest{Title: "test chat"}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var v CreateConversationRequest
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	if v.Title != "test chat" {
		t.Errorf("Title = %q", v.Title)
	}
}

func TestCreateConversationResponse_JSON(t *testing.T) {
	resp := CreateConversationResponse{
		ConversationID: "conv_1",
		Title:          "hello",
		CreatedAt:      "2025-01-01",
	}
	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var v CreateConversationResponse
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	if v.ConversationID != "conv_1" {
		t.Errorf("ConversationID = %q", v.ConversationID)
	}
}

func TestConversationItem_JSON(t *testing.T) {
	item := ConversationItem{
		ConversationID: "conv_1",
		Title:          "chat",
		MessageCount:   5,
		CreatedAt:      "2025-01-01",
		UpdatedAt:      "2025-01-02",
	}
	b, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	var v ConversationItem
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	if v.MessageCount != 5 {
		t.Errorf("MessageCount = %d", v.MessageCount)
	}
}

func TestListConversationsResponse_JSON(t *testing.T) {
	resp := ListConversationsResponse{
		Total:         2,
		Conversations: []ConversationItem{{ConversationID: "c1"}, {ConversationID: "c2"}},
	}
	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var v ListConversationsResponse
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	if v.Total != 2 {
		t.Errorf("Total = %d", v.Total)
	}
	if len(v.Conversations) != 2 {
		t.Errorf("len = %d", len(v.Conversations))
	}
}

func TestMessageItem_WithToolCalls(t *testing.T) {
	msg := MessageItem{
		Role:    "assistant",
		Content: "let me check",
		ToolCalls: []ToolCall{
			{
				ID:   "call_1",
				Type: "function",
				Function: FunctionCall{
					Name:      "search",
					Arguments: `{"q":"hello"}`,
				},
			},
		},
	}
	b, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	var v MessageItem
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	if len(v.ToolCalls) != 1 {
		t.Fatalf("len(ToolCalls) = %d", len(v.ToolCalls))
	}
	if v.ToolCalls[0].Function.Name != "search" {
		t.Errorf("tool name = %q", v.ToolCalls[0].Function.Name)
	}
	if v.ToolCalls[0].Function.Arguments != `{"q":"hello"}` {
		t.Errorf("tool args = %q", v.ToolCalls[0].Function.Arguments)
	}
}

func TestMessageItem_WithoutToolCalls(t *testing.T) {
	msg := MessageItem{Role: "user", Content: "hi"}
	b, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(b) {
		t.Fatal("invalid JSON")
	}
}

func TestGetMessagesResponse_JSON(t *testing.T) {
	resp := GetMessagesResponse{
		ConversationID: "conv_1",
		Total:          3,
		Messages: []MessageItem{
			{Role: "user", Content: "hi"},
			{Role: "assistant", Content: "hello"},
		},
	}
	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var v GetMessagesResponse
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	if v.Total != 3 {
		t.Errorf("Total = %d", v.Total)
	}
	if len(v.Messages) != 2 {
		t.Errorf("len(Messages) = %d", len(v.Messages))
	}
}

func TestDeleteConversationResponse_JSON(t *testing.T) {
	resp := DeleteConversationResponse{ConversationID: "conv_1"}
	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var v DeleteConversationResponse
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	if v.ConversationID != "conv_1" {
		t.Errorf("ConversationID = %q", v.ConversationID)
	}
}
