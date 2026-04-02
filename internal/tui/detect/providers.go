package detect

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type ProviderInfo struct {
	Name       string
	Binary     string
	Found      bool
	NeedsModel bool
	Models     []string
	InstallURL string
}

type providerSpec struct {
	Name       string
	Binary     string
	NeedsModel bool
	InstallURL string
}

var (
	errExecNotFound    = exec.ErrNotFound
	realCommandContext = exec.CommandContext
	lmstudioModelsURL  = "http://localhost:1234/v1/models"
)

func DetectProviders(ctx context.Context) []ProviderInfo {
	pctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	specs := []providerSpec{
		{Name: "claude", Binary: "claude", NeedsModel: false, InstallURL: "https://docs.anthropic.com/en/docs/claude-code"},
		{Name: "gemini", Binary: "gemini", NeedsModel: true, InstallURL: "https://github.com/google-gemini/gemini-cli"},
		{Name: "ollama", Binary: "ollama", NeedsModel: true, InstallURL: "https://ollama.com/download"},
		{Name: "codex", Binary: "codex", NeedsModel: false, InstallURL: "https://github.com/openai/codex"},
		{Name: "opencode", Binary: "opencode", NeedsModel: true, InstallURL: "https://opencode.ai"},
		{Name: "gh", Binary: "gh", NeedsModel: true, InstallURL: "https://cli.github.com"},
	}

	results := make([]ProviderInfo, len(specs)+1)
	var wg sync.WaitGroup

	for i, spec := range specs {
		if pctx.Err() != nil {
			break
		}

		wg.Add(1)
		go func(i int, spec providerSpec) {
			defer wg.Done()
			results[i] = detectBinaryProvider(pctx, spec)
		}(i, spec)
	}

	if pctx.Err() == nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[len(specs)] = detectLMStudioProvider(pctx)
		}()
	}

	wg.Wait()

	providers := make([]ProviderInfo, 0, len(results))
	for _, provider := range results {
		if provider.Name == "" {
			continue
		}
		providers = append(providers, provider)
	}

	return providers
}

func detectBinaryProvider(ctx context.Context, spec providerSpec) ProviderInfo {
	info := ProviderInfo{
		Name:       spec.Name,
		Binary:     spec.Binary,
		NeedsModel: spec.NeedsModel,
		InstallURL: spec.InstallURL,
		Models:     []string{},
	}
	if ctx.Err() != nil {
		return info
	}

	path, err := lookPath(spec.Binary)
	if err != nil {
		return info
	}

	info.Found = true
	if spec.Name != "gh" {
		return info
	}

	info.Found = hasGHCopilotExtension(ctx, path)
	return info
}

func detectLMStudioProvider(ctx context.Context) ProviderInfo {
	info := ProviderInfo{
		Name:       "lmstudio",
		Binary:     "localhost:1234",
		NeedsModel: true,
		InstallURL: "https://lmstudio.ai",
		Models:     []string{},
	}
	if ctx.Err() != nil {
		return info
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, lmstudioModelsURL, nil)
	if err != nil {
		return info
	}

	hc := &http.Client{Timeout: 1500 * time.Millisecond}
	resp, err := hc.Do(req)
	if err != nil {
		return info
	}
	defer resp.Body.Close()

	info.Found = resp.StatusCode >= 200 && resp.StatusCode < 300
	return info
}

func discoverOllamaModels(ctx context.Context) []string {
	if ctx.Err() != nil {
		return []string{}
	}

	path, err := lookPath("ollama")
	if err != nil {
		return []string{}
	}

	octx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := commandContext(octx, path, "list")
	out, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	models := make([]string, 0)
	s := bufio.NewScanner(strings.NewReader(string(out)))
	first := true
	for s.Scan() {
		if first {
			first = false
			continue
		}

		fields := strings.Fields(s.Text())
		if len(fields) == 0 {
			continue
		}
		models = append(models, fields[0])
	}

	if err := s.Err(); err != nil {
		return []string{}
	}

	return models
}

func discoverLMStudioModels(ctx context.Context) []string {
	if ctx.Err() != nil {
		return []string{}
	}

	lctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(lctx, http.MethodGet, lmstudioModelsURL, nil)
	if err != nil {
		return []string{}
	}

	hc := &http.Client{Timeout: 3 * time.Second}
	resp, err := hc.Do(req)
	if err != nil {
		return []string{}
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return []string{}
	}

	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return []string{}
	}

	models := make([]string, 0, len(payload.Data))
	for _, model := range payload.Data {
		if model.ID == "" {
			continue
		}
		models = append(models, model.ID)
	}

	return models
}

func hasGHCopilotExtension(ctx context.Context, binary string) bool {
	if ctx.Err() != nil {
		return false
	}

	ectx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
	defer cancel()

	cmd := commandContext(ectx, binary, "extension", "list")
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(strings.ToLower(line))
		if strings.Contains(line, "copilot") {
			return true
		}
	}

	return false
}
