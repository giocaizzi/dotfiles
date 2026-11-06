package catalog

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseModelsDev(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/modelsdev.json")
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseModelsDev(data)
	if err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(got))
	for _, m := range got {
		keys = append(keys, m.Provider+"/"+m.ID)
	}
	wantKeys := []string{
		"anthropic/claude-haiku-4-5",
		"anthropic/claude-opus-5-5",
		"github-copilot/claude-opus-5.5",
		"github-copilot/gemini-3.5-flash",
		"github-copilot/gemini-3.6-flash",
		"openai/gpt-6-luna",
		"openai/gpt-realtime-2.1",
	}
	if diff := cmp.Diff(wantKeys, keys); diff != "" {
		t.Fatalf("models (-want +got):\n%s", diff)
	}
	byKey := map[string]ProviderModel{}
	for _, m := range got {
		byKey[m.Provider+"/"+m.ID] = m
	}
	wantOpus := ProviderModel{
		Provider: "anthropic", ID: "claude-opus-5-5", Name: "Claude Opus 5.5",
		CanonicalID: "anthropic/claude-opus-5-5", ReleaseDate: "2026-09-22",
		In: 4, Out: 20, ToolCall: true, TextOutput: true,
	}
	if diff := cmp.Diff(wantOpus, byKey["anthropic/claude-opus-5-5"]); diff != "" {
		t.Errorf("opus (-want +got):\n%s", diff)
	}
	if rt := byKey["openai/gpt-realtime-2.1"]; !rt.ToolCall || !rt.TextOutput {
		t.Errorf("realtime = %+v, want tool call with text output", rt)
	}
}

func TestParseModelsDevDefaults(t *testing.T) {
	t.Parallel()
	data := []byte(`{
	  "openai": {"models": {"o1": {"id": "o1", "name": "O1", "modalities": {"output": ["audio"]}}}},
	  "other": {"models": {"x": {"id": "x"}}}
	}`)
	got, err := ParseModelsDev(data)
	if err != nil {
		t.Fatal(err)
	}
	want := []ProviderModel{{Provider: "openai", ID: "o1", Name: "O1", CanonicalID: "openai/o1"}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}

func TestParseModelsDevInvalid(t *testing.T) {
	t.Parallel()
	if _, err := ParseModelsDev([]byte(`[`)); err == nil {
		t.Error("want error for invalid JSON")
	}
	if _, err := ParseModelsDev([]byte(`{"openai": 3}`)); err == nil {
		t.Error("want error for invalid provider")
	}
}
