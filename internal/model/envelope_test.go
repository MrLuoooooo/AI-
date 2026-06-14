package model

import (
	"encoding/json"
	"testing"
)

func TestOK_NilData(t *testing.T) {
	env := OK(nil)
	if env.Code != 0 {
		t.Errorf("Code = %d, want 0", env.Code)
	}
	if env.Message != "success" {
		t.Errorf("Message = %q, want success", env.Message)
	}
	if env.Data != nil {
		t.Errorf("Data = %v, want nil", env.Data)
	}
}

func TestOK_WithData(t *testing.T) {
	env := OK(map[string]string{"key": "val"})
	if env.Code != 0 {
		t.Errorf("Code = %d, want 0", env.Code)
	}
	if env.Message != "success" {
		t.Errorf("Message = %q, want success", env.Message)
	}
}

func TestErr(t *testing.T) {
	env := Err(404, "not found")
	if env.Code != 404 {
		t.Errorf("Code = %d, want 404", env.Code)
	}
	if env.Message != "not found" {
		t.Errorf("Message = %q, want not found", env.Message)
	}
	if env.Data != nil {
		t.Errorf("Data = %v, want nil", env.Data)
	}
}

func TestOK_JSONMarshal(t *testing.T) {
	env := OK("hello")
	b, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m["code"].(float64) != 0 {
		t.Errorf("code = %v", m["code"])
	}
	if m["message"] != "success" {
		t.Errorf("message = %v", m["message"])
	}
	if m["data"] != "hello" {
		t.Errorf("data = %v", m["data"])
	}
}

func TestErr_JSONMarshal(t *testing.T) {
	env := Err(500, "boom")
	b, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m["code"].(float64) != 500 {
		t.Errorf("code = %v", m["code"])
	}
	if _, ok := m["data"]; ok {
		t.Error("data should be omitted for error envelope")
	}
}
