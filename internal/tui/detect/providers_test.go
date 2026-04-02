package detect

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
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
				writeExecutable(t, path, "#!/bin/sh\nprintf 'NAME ID SIZE MODIFIED\\nllama3 123 4GB now\\nphi4 456 2GB now\\n'")
				lookPath = func(file string) (string, error) {
					if file != "ollama" {
						return "", errExecNotFound
					}
					return path, nil
				}
				commandContext = realCommandContext
			},
			expected: []string{"llama3", "phi4"},
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

func writeExecutable(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
