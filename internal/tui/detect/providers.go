package detect

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type SubProvider struct {
	Name       string
	ID         string
	Models     []string
	ModelCount int
}

type ProviderInfo struct {
	Name         string
	Binary       string
	Found        bool
	NeedsModel   bool
	Models       []string
	SubProviders []SubProvider
	InstallURL   string
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
	opencodeCachePath  = DefaultOpencodeCachePath
	opencodeAuthPath   = DefaultOpencodeAuthPath
)

func DefaultOpencodeCachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".cache", "opencode", "models.json")
}

func DefaultOpencodeAuthPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".local", "share", "opencode", "auth.json")
}

func DetectProviders(ctx context.Context) []ProviderInfo {
	pctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	specs := []providerSpec{
		{Name: "claude", Binary: "claude", NeedsModel: false, InstallURL: "https://docs.anthropic.com/en/docs/claude-code"},
		{Name: "gemini", Binary: "gemini", NeedsModel: false, InstallURL: "https://github.com/google-gemini/gemini-cli"},
		{Name: "ollama", Binary: "ollama", NeedsModel: true, InstallURL: "https://ollama.com/download"},
		{Name: "codex", Binary: "codex", NeedsModel: false, InstallURL: "https://github.com/openai/codex"},
		{Name: "opencode", Binary: "opencode", NeedsModel: true, InstallURL: "https://opencode.ai"},
		{Name: "github", Binary: "gh", NeedsModel: true, InstallURL: "https://cli.github.com"},
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
		if provider.Found && provider.NeedsModel {
			provider.SubProviders = DiscoverSubProviders(pctx, provider.Name)
			if len(provider.SubProviders) == 0 {
				provider.Models = DiscoverModels(pctx, provider.Name)
			}
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
	subProviders := discoverOllamaSubProviders(ctx)
	if len(subProviders) == 0 {
		return []string{}
	}

	models := make([]string, 0)
	for _, subProvider := range subProviders {
		models = append(models, subProvider.Models...)
	}
	return models
}

func discoverOllamaSubProviders(ctx context.Context) []SubProvider {
	if ctx.Err() != nil {
		return []SubProvider{}
	}

	path, err := lookPath("ollama")
	if err != nil {
		return []SubProvider{}
	}

	octx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	cmd := commandContext(octx, path, "list")
	out, err := cmd.Output()
	if err != nil {
		return []SubProvider{}
	}

	families := make(map[string][]string)
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
		model := fields[0]
		family := model
		if cut := strings.Index(model, ":"); cut >= 0 {
			family = model[:cut]
		}
		families[family] = append(families[family], model)
	}

	if err := s.Err(); err != nil {
		return []SubProvider{}
	}

	familyNames := make([]string, 0, len(families))
	for family := range families {
		familyNames = append(familyNames, family)
	}
	sort.Strings(familyNames)

	subProviders := make([]SubProvider, 0, len(familyNames))
	for _, family := range familyNames {
		models := append([]string(nil), families[family]...)
		sort.Strings(models)
		subProviders = append(subProviders, SubProvider{
			Name:       family,
			ID:         family,
			Models:     models,
			ModelCount: len(models),
		})
	}

	return subProviders
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

func discoverGHModels() []string {
	return []string{
		"gpt-4o",
		"gpt-4.1",
		"claude-sonnet-4-20250514",
		"o3-mini",
	}
}

func discoverOpencodeSubProviders() []SubProvider {
	cachePath := opencodeCachePath()
	if strings.TrimSpace(cachePath) == "" {
		return []SubProvider{}
	}

	content, err := os.ReadFile(cachePath)
	if err != nil {
		return []SubProvider{}
	}

	type opencodeModel struct {
		ID       string `json:"id"`
		ToolCall bool   `json:"tool_call"`
	}

	type opencodeProvider struct {
		Name   string                   `json:"name"`
		Env    []string                 `json:"env"`
		Models map[string]opencodeModel `json:"models"`
	}

	providers := map[string]opencodeProvider{}
	if err := json.Unmarshal(content, &providers); err != nil {
		return []SubProvider{}
	}

	authenticated := loadOpencodeAuthenticatedProviders()
	subProviders := make([]SubProvider, 0, len(providers))
	for id, provider := range providers {
		if !opencodeProviderAvailable(id, provider.Env, authenticated) {
			continue
		}

		models := make([]string, 0, len(provider.Models))
		for key, model := range provider.Models {
			if !model.ToolCall {
				continue
			}
			modelID := strings.TrimSpace(model.ID)
			if modelID == "" {
				modelID = strings.TrimSpace(key)
			}
			if modelID == "" {
				continue
			}
			models = append(models, modelID)
		}
		if len(models) == 0 {
			continue
		}

		sort.Strings(models)
		name := strings.TrimSpace(provider.Name)
		if name == "" {
			name = strings.TrimSpace(id)
		}
		subProviders = append(subProviders, SubProvider{
			Name:       name,
			ID:         strings.TrimSpace(id),
			Models:     models,
			ModelCount: len(models),
		})
	}

	sort.Slice(subProviders, func(i, j int) bool {
		return strings.ToLower(subProviders[i].Name) < strings.ToLower(subProviders[j].Name)
	})

	return subProviders
}

func loadOpencodeAuthenticatedProviders() map[string]struct{} {
	authPath := opencodeAuthPath()
	if strings.TrimSpace(authPath) == "" {
		return map[string]struct{}{}
	}

	content, err := os.ReadFile(authPath)
	if err != nil {
		return map[string]struct{}{}
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(content, &raw); err != nil {
		return map[string]struct{}{}
	}

	providers := map[string]struct{}{}
	if nested, ok := raw["providers"]; ok {
		var nestedProviders map[string]json.RawMessage
		if err := json.Unmarshal(nested, &nestedProviders); err == nil {
			for id := range nestedProviders {
				providers[strings.TrimSpace(id)] = struct{}{}
			}
			return providers
		}
	}

	for id := range raw {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" || trimmed == "providers" {
			continue
		}
		providers[trimmed] = struct{}{}
	}

	return providers
}

func opencodeProviderAvailable(id string, envVars []string, authenticated map[string]struct{}) bool {
	if strings.EqualFold(strings.TrimSpace(id), "opencode") {
		return true
	}
	if _, ok := authenticated[strings.TrimSpace(id)]; ok {
		return true
	}
	if len(envVars) == 0 {
		return false
	}
	for _, name := range envVars {
		if strings.TrimSpace(os.Getenv(name)) == "" {
			return false
		}
	}
	return true
}

func DiscoverSubProviders(ctx context.Context, provider string) []SubProvider {
	if ctx.Err() != nil {
		return []SubProvider{}
	}

	switch strings.TrimSpace(strings.ToLower(provider)) {
	case "opencode":
		return discoverOpencodeSubProviders()
	case "ollama":
		return discoverOllamaSubProviders(ctx)
	default:
		return []SubProvider{}
	}
}

func DiscoverModels(ctx context.Context, provider string) []string {
	if ctx.Err() != nil {
		return []string{}
	}

	switch strings.TrimSpace(strings.ToLower(provider)) {
	case "gh":
		return discoverGHModels()
	case "ollama":
		return discoverOllamaModels(ctx)
	case "lmstudio":
		return discoverLMStudioModels(ctx)
	default:
		return []string{}
	}
}
