package model

import (
	"encoding/json"
	"testing"
)

func TestVisionRequest_JSON(t *testing.T) {
	req := VisionRequest{
		Type:       VisionMsgTranscript,
		Frame:      "base64data",
		Transcript: "hello",
		SessionID:  "abc123",
	}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var v2 VisionRequest
	if err := json.Unmarshal(b, &v2); err != nil {
		t.Fatal(err)
	}
	if v2.Type != VisionMsgTranscript {
		t.Errorf("Type = %q", v2.Type)
	}
	if v2.Frame != "base64data" {
		t.Errorf("Frame = %q", v2.Frame)
	}
	if v2.Transcript != "hello" {
		t.Errorf("Transcript = %q", v2.Transcript)
	}
	if v2.SessionID != "abc123" {
		t.Errorf("SessionID = %q", v2.SessionID)
	}
}

func TestVisionResponse_JSON(t *testing.T) {
	resp := VisionResponse{
		Type:      VisionMsgReply,
		Content:   "看到了一个人",
		SessionID: "abc123",
		TokenUsed: 150,
		Error:     "",
	}
	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var v2 VisionResponse
	if err := json.Unmarshal(b, &v2); err != nil {
		t.Fatal(err)
	}
	if v2.Type != VisionMsgReply {
		t.Errorf("Type = %q", v2.Type)
	}
	if v2.Content != "看到了一个人" {
		t.Errorf("Content = %q", v2.Content)
	}
	if v2.TokenUsed != 150 {
		t.Errorf("TokenUsed = %d", v2.TokenUsed)
	}
}

func TestVisionResponse_Error(t *testing.T) {
	resp := VisionResponse{
		Type:      VisionMsgError,
		SessionID: "abc",
		Error:     "rate limited",
	}
	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var v2 VisionResponse
	if err := json.Unmarshal(b, &v2); err != nil {
		t.Fatal(err)
	}
	if v2.Error != "rate limited" {
		t.Errorf("Error = %q", v2.Error)
	}
}

func TestVisionMsgConstants(t *testing.T) {
	if VisionMsgFrame != "frame" {
		t.Errorf("VisionMsgFrame = %q", VisionMsgFrame)
	}
	if VisionMsgTranscript != "transcript" {
		t.Errorf("VisionMsgTranscript = %q", VisionMsgTranscript)
	}
	if VisionMsgPing != "ping" {
		t.Errorf("VisionMsgPing = %q", VisionMsgPing)
	}
	if VisionMsgReply != "reply" {
		t.Errorf("VisionMsgReply = %q", VisionMsgReply)
	}
	if VisionMsgStatus != "status" {
		t.Errorf("VisionMsgStatus = %q", VisionMsgStatus)
	}
	if VisionMsgError != "error" {
		t.Errorf("VisionMsgError = %q", VisionMsgError)
	}
	if VisionMsgPong != "pong" {
		t.Errorf("VisionMsgPong = %q", VisionMsgPong)
	}
}

func TestVisionSession_JSON(t *testing.T) {
	sess := VisionSession{
		ID:           "uuid-1",
		CreatedAt:    1700000000,
		LastFrameAt:  1700000010,
		LastSpeechAt: 1700000005,
		FrameCount:   100,
		TokenUsed:    5000,
	}
	b, err := json.Marshal(sess)
	if err != nil {
		t.Fatal(err)
	}
	var s2 VisionSession
	if err := json.Unmarshal(b, &s2); err != nil {
		t.Fatal(err)
	}
	if s2.ID != sess.ID {
		t.Errorf("ID = %q", s2.ID)
	}
	if s2.FrameCount != sess.FrameCount {
		t.Errorf("FrameCount = %d", s2.FrameCount)
	}
	if s2.TokenUsed != sess.TokenUsed {
		t.Errorf("TokenUsed = %d", s2.TokenUsed)
	}
}
