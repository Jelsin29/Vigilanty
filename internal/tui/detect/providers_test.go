package detect

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"testing"
	"time"
)

func TestDetectProviderBinarySanity(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	info := detectBinaryProvider(ctx, providerSpec{Name: "go", Binary: "go", NeedsModel: true})
	if info.Name != "go" {
		t.Fatalf("detectBinaryProvider().Name = %q, want %q", info.Name, "go")
	}
	if !info.Found {
		t.Fatal("detectBinaryProvider() reported go as missing")
	}
}

func TestDiscoverOllamaModels(t *testing.T) {
	oldLookPath := lookPath
	oldCommand := commandContext
	t.Cleanup(func() {
		lookPath = oldLookPath
		commandContext = oldCommand
	})

	tests := []struct {
		name     string
		setup    func(t *testing.T)
		expected []string
	}{
		{
			name: "returns empty when ollama is unavailable",
			setup: func(t *testing.T) {
				t.Helper()
				lookPath = func(file string) (string, error) {
					return "", errExecNotFound
				}
			},
			expected: []string{},
		},
		{
			name: "parses model names from command output",
			setup: func(t *testing.T) {
				t.Helper()
				dir := t.TempDir()
				name := "ollama"
				if runtime.GOOS == "windows" {
					name += ".bat"
				}
				path := filepath.Join(dir, name)
				writeExecutable(t, path, "#!/bin/sh\nprintf 'NAME ID SIZE MODIFIED\\nllama3:latest 123 4GB now\\nphi4 456 2GB now\\n'")
				lookPath = func(file string) (string, error) {
					if file != "ollama" {
						return "", errExecNotFound
					}
					return path, nil
				}
				commandContext = realCommandContext
			},
			expected: []string{"llama3:latest", "phi4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)

			got := discoverOllamaModels(context.Background())
			if len(got) != len(tt.expected) {
				t.Fatalf("discoverOllamaModels() = %v, want %v", got, tt.expected)
			}
			for i := range tt.expected {
				if got[i] != tt.expected[i] {
					t.Fatalf("discoverOllamaModels()[%d] = %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestDiscoverOllamaSubProviders(t *testing.T) {
	oldLookPath := lookPath
	oldCommand := commandContext
	t.Cleanup(func() {
		lookPath = oldLookPath
		commandContext = oldCommand
	})

	dir := t.TempDir()
	name := "ollama"
	if runtime.GOOS == "windows" {
		name += ".bat"
	}
	path := filepath.Join(dir, name)
	writeExecutable(t, path, "#!/bin/sh\nprintf 'NAME ID SIZE MODIFIED\\nllama3.2:latest 123 4GB now\\nllama3.2:8b 456 4GB now\\nmistral:latest 789 3GB now\\nphi4 111 2GB now\\n'")
	lookPath = func(file string) (string, error) {
		if file != "ollama" {
			return "", errExecNotFound
		}
		return path, nil
	}
	commandContext = realCommandContext

	got := discoverOllamaSubProviders(context.Background())
	want := []SubProvider{
		{Name: "llama3.2", ID: "llama3.2", Models: []string{"llama3.2:8b", "llama3.2:latest"}, ModelCount: 2},
		{Name: "mistral", ID: "mistral", Models: []string{"mistral:latest"}, ModelCount: 1},
		{Name: "phi4", ID: "phi4", Models: []string{"phi4"}, ModelCount: 1},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("discoverOllamaSubProviders() = %#v, want %#v", got, want)
	}
}

func TestDiscoverLMStudioModels(t *testing.T) {
	oldURL := lmstudioModelsURL
	t.Cleanup(func() {
		lmstudioModelsURL = oldURL
	})

	tests := []struct {
		name     string
		server   func(t *testing.T) *httptest.Server
		expected []string
	}{
		{
			name: "parses model ids from api",
			server: func(t *testing.T) *httptest.Server {
				t.Helper()
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/v1/models" {
						t.Fatalf("unexpected path %q", r.URL.Path)
					}
					_ = json.NewEncoder(w).Encode(map[string]any{
						"data": []map[string]string{{"id": "qwen2.5-coder"}, {"id": "deepseek-r1"}},
					})
				}))
			},
			expected: []string{"qwen2.5-coder", "deepseek-r1"},
		},
		{
			name: "returns empty on connection failure",
			server: func(t *testing.T) *httptest.Server {
				t.Helper()
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
				ts.Close()
				return ts
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := tt.server(t)
			if ts != nil {
				defer ts.Close()
				lmstudioModelsURL = ts.URL + "/v1/models"
			}

			got := discoverLMStudioModels(context.Background())
			if len(got) != len(tt.expected) {
				t.Fatalf("discoverLMStudioModels() = %v, want %v", got, tt.expected)
			}
			for i := range tt.expected {
				if got[i] != tt.expected[i] {
					t.Fatalf("discoverLMStudioModels()[%d] = %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestDiscoverModelsReturnsPredefinedProviderLists(t *testing.T) {
	ctx := context.Background()

	got := DiscoverModels(ctx, "gh")
	want := []string{"gpt-4o", "gpt-4.1", "claude-sonnet-4-20250514", "o3-mini"}
	if !slices.Equal(got, want) {
		t.Fatalf("DiscoverModels(%q) = %v, want %v", "gh", got, want)
	}
}

func TestDiscoverOpencodeSubProviders(t *testing.T) {
	oldCachePath := opencodeCachePath
	oldAuthPath := opencodeAuthPath
	t.Cleanup(func() {
		opencodeCachePath = oldCachePath
		opencodeAuthPath = oldAuthPath
	})

	dir := t.TempDir()
	cacheFile := filepath.Join(dir, "models.json")
	authFile := filepath.Join(dir, "auth.json")

	models := map[string]any{
		"opencode": map[string]any{
			"name": "OpenCode",
			"models": map[string]any{
				"opencode-zen":  map[string]any{"id": "opencode-zen", "tool_call": true},
				"opencode-lite": map[string]any{"id": "opencode-lite", "tool_call": false},
			},
		},
		"anthropic": map[string]any{
			"name": "Anthropic",
			"env":  []string{"ANTHROPIC_API_KEY"},
			"models": map[string]any{
				"claude-sonnet-4": map[string]any{"id": "claude-sonnet-4", "tool_call": true},
				"claude-text":     map[string]any{"id": "claude-text", "tool_call": false},
			},
		},
		"openai": map[string]any{
			"name": "OpenAI",
			"env":  []string{"OPENAI_API_KEY"},
			"models": map[string]any{
				"gpt-4o":  map[string]any{"id": "gpt-4o", "tool_call": true},
				"o3-mini": map[string]any{"id": "o3-mini", "tool_call": true},
			},
		},
		"deepseek": map[string]any{
			"name": "DeepSeek",
			"models": map[string]any{
				"deepseek-chat": map[string]any{"id": "deepseek-chat", "tool_call": true},
			},
		},
	}
	writeJSONFile(t, cacheFile, models)
	writeJSONFile(t, authFile, map[string]any{
		"providers": map[string]any{
			"openai": map[string]any{"token": "secret"},
		},
	})
	t.Setenv("ANTHROPIC_API_KEY", "test-key")

	opencodeCachePath = func() string { return cacheFile }
	opencodeAuthPath = func() string { return authFile }

	got := discoverOpencodeSubProviders()
	want := []SubProvider{
		{Name: "Anthropic", ID: "anthropic", Models: []string{"claude-sonnet-4"}, ModelCount: 1},
		{Name: "OpenAI", ID: "openai", Models: []string{"gpt-4o", "o3-mini"}, ModelCount: 2},
		{Name: "OpenCode", ID: "opencode", Models: []string{"opencode-zen"}, ModelCount: 1},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("discoverOpencodeSubProviders() = %#v, want %#v", got, want)
	}
}

func writeExecutable(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func writeJSONFile(t *testing.T, path string, value any) {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
