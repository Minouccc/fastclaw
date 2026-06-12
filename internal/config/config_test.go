package config

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestAgentFileConfigUnmarshalModelFallbacks(t *testing.T) {
	var cfg AgentFileConfig
	if err := json.Unmarshal([]byte(`{
		"model": "openai/gpt-4.1",
		"modelFallbacks": ["anthropic/claude-3-7-sonnet", "google/gemini-2.5-pro"]
	}`), &cfg); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	want := []string{"anthropic/claude-3-7-sonnet", "google/gemini-2.5-pro"}
	if !reflect.DeepEqual(cfg.ModelFallbacks, want) {
		t.Fatalf("ModelFallbacks = %#v, want %#v", cfg.ModelFallbacks, want)
	}
}
