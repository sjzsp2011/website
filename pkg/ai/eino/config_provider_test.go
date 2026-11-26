package eino

import "testing"

func TestAssistantUsesSavedProvider(t *testing.T) {
	settings := Configs{LastModel: "anqi-flash"}
	if got := settings.DefaultConfig(); got != nil {
		t.Fatalf("unexpected provider with only legacy selection: %+v", got)
	}
	first := &Config{Model: "first"}
	second := &Config{Model: "second"}
	settings.Configs = []*Config{first, second}
	if got := settings.DefaultConfig(); got != first {
		t.Fatalf("expected the first custom provider for a retired selection, got %v", got)
	}
	settings.LastModel = "custom:1"
	if got := settings.DefaultConfig(); got != second {
		t.Fatalf("expected the selected custom provider, got %v", got)
	}
	if got := settings.SelectedConfig("custom:invalid"); got != nil {
		t.Fatalf("invalid selection resolved to a provider: %v", got)
	}
}
