package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var knownCheckerTypes = map[string]struct{}{
	"ai":        {},
	"ai-review": {},
	"command":   {},
	"prompt":    {},
	"shell":     {},
}

type Config struct {
	Version  string       `yaml:"version"`
	Global   GlobalConfig `yaml:"global"`
	Pipeline []StepConfig `yaml:"pipeline"`
	Source   string       `yaml:"-"`
}

type GlobalConfig struct {
	FailFast        bool     `yaml:"fail_fast"`
	DiffMaxBytes    int      `yaml:"diff_max_bytes"`
	Timeout         string   `yaml:"timeout"`
	Verbose         bool     `yaml:"verbose"`
	FilePatterns    []string `yaml:"file_patterns,omitempty"`
	ExcludePatterns []string `yaml:"exclude_patterns,omitempty"`
}

type StepConfig struct {
	Name            string                 `yaml:"name"`
	Type            string                 `yaml:"type,omitempty"`
	Checker         string                 `yaml:"checker,omitempty"`
	Command         string                 `yaml:"command,omitempty"`
	Provider        string                 `yaml:"provider,omitempty"`
	Prompt          string                 `yaml:"prompt,omitempty"`
	RulesFile       string                 `yaml:"rules_file,omitempty"`
	Model           string                 `yaml:"model,omitempty"`
	Timeout         string                 `yaml:"timeout,omitempty"`
	Enabled         *bool                  `yaml:"enabled,omitempty"`
	Env             map[string]string      `yaml:"env,omitempty"`
	SkipOnEmptyDiff bool                   `yaml:"skip_on_empty_diff,omitempty"`
	MaxDiffLines    int                    `yaml:"max_diff_lines,omitempty"`
	PassPattern     string                 `yaml:"pass_pattern,omitempty"`
	FailPattern     string                 `yaml:"fail_pattern,omitempty"`
	Config          map[string]interface{} `yaml:"config,omitempty"`
}

func Load(path string) (*Config, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("load config: path is required")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve config path %q: %w", path, err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", absPath, err)
	}

	cfg, err := decodeConfig(data)
	if err != nil {
		return nil, fmt.Errorf("parse config %q: %w", absPath, err)
	}

	applyDefaults(cfg)
	cfg.Source = absPath

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config %q: %w", absPath, err)
	}

	return cfg, nil
}

func LoadFromReader(reader io.Reader) (*Config, error) {
	if reader == nil {
		return nil, errors.New("load config: reader is required")
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg, err := decodeConfig(data)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	applyDefaults(cfg)
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

func LoadWithDefaults() (*Config, error) {
	defaultPath := filepath.Join(".vigilanty.yml")
	if _, err := os.Stat(defaultPath); err == nil {
		return Load(defaultPath)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("stat config %q: %w", defaultPath, err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home directory: %w", err)
	}

	homeConfigPath := filepath.Join(homeDir, ".config", "vigilanty", "config.yml")
	if _, err := os.Stat(homeConfigPath); err == nil {
		return Load(homeConfigPath)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("stat config %q: %w", homeConfigPath, err)
	}

	cfg, err := decodeConfig([]byte(DefaultConfigYAML()))
	if err != nil {
		return nil, fmt.Errorf("parse default config: %w", err)
	}

	applyDefaults(cfg)
	cfg.Source = "defaults"

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate default config: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c == nil {
		return errors.New("config is nil")
	}

	if strings.TrimSpace(c.Version) != "" && c.Version != "1" {
		return fmt.Errorf("unsupported config version: %s", c.Version)
	}

	if len(c.Pipeline) == 0 {
		return errors.New("pipeline must contain at least one step")
	}

	if c.Global.DiffMaxBytes < 0 {
		return errors.New("global.diff_max_bytes must be zero or greater")
	}

	if err := validateDuration("global.timeout", c.Global.Timeout); err != nil {
		return err
	}

	seenNames := make(map[string]struct{}, len(c.Pipeline))
	for i, step := range c.Pipeline {
		path := fmt.Sprintf("pipeline[%d]", i)

		if strings.TrimSpace(step.Name) == "" {
			return fmt.Errorf("%s.name is required", path)
		}

		if _, exists := seenNames[step.Name]; exists {
			return fmt.Errorf("duplicate pipeline step name %q", step.Name)
		}
		seenNames[step.Name] = struct{}{}

		if strings.TrimSpace(step.Type) == "" {
			return fmt.Errorf("%s.type is required", path)
		}

		if !isKnownCheckerType(step.Type) {
			return fmt.Errorf("%s.type %q is not supported (known: %s)", path, step.Type, strings.Join(sortedKnownCheckerTypes(), ", "))
		}

		switch step.Type {
		case "command", "shell":
			if strings.TrimSpace(step.Command) == "" {
				return fmt.Errorf("%s.command is required for type %q", path, step.Type)
			}
		case "ai", "prompt", "ai-review":
			if strings.TrimSpace(step.Provider) == "" {
				return fmt.Errorf("%s.provider is required for type %q", path, step.Type)
			}
			if strings.TrimSpace(step.Prompt) == "" {
				return fmt.Errorf("%s.prompt is required for type %q", path, step.Type)
			}
		}

		if err := validateDuration(path+".timeout", step.Timeout); err != nil {
			return err
		}

		if step.PassPattern != "" {
			if _, err := regexp.Compile(step.PassPattern); err != nil {
				return fmt.Errorf("%s.pass_pattern: %w", path, err)
			}
		}

		if step.FailPattern != "" {
			if _, err := regexp.Compile(step.FailPattern); err != nil {
				return fmt.Errorf("%s.fail_pattern: %w", path, err)
			}
		}
	}

	return nil
}

func applyDefaults(cfg *Config) {
	if cfg == nil {
		return
	}

	if cfg.Pipeline == nil {
		cfg.Pipeline = []StepConfig{}
	}

	if strings.TrimSpace(cfg.Version) == "" {
		cfg.Version = "1"
	}

	if cfg.Global.DiffMaxBytes == 0 {
		cfg.Global.DiffMaxBytes = 256 * 1024
	}

	if strings.TrimSpace(cfg.Global.Timeout) == "" {
		cfg.Global.Timeout = "2m"
	}

	for i := range cfg.Pipeline {
		if strings.TrimSpace(cfg.Pipeline[i].Type) == "" && strings.TrimSpace(cfg.Pipeline[i].Checker) != "" {
			cfg.Pipeline[i].Type = cfg.Pipeline[i].Checker
		}
		if strings.TrimSpace(cfg.Pipeline[i].Checker) == "" && strings.TrimSpace(cfg.Pipeline[i].Type) != "" {
			cfg.Pipeline[i].Checker = cfg.Pipeline[i].Type
		}
		if strings.TrimSpace(cfg.Pipeline[i].Timeout) == "" {
			cfg.Pipeline[i].Timeout = cfg.Global.Timeout
		}
		if cfg.Pipeline[i].Config == nil {
			cfg.Pipeline[i].Config = map[string]interface{}{}
		}
		if cfg.Pipeline[i].Env == nil {
			cfg.Pipeline[i].Env = map[string]string{}
		}
	}
}

func decodeConfig(data []byte) (*Config, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	if err := ensureSingleDocument(decoder); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func ensureSingleDocument(decoder *yaml.Decoder) error {
	var extra interface{}
	if err := decoder.Decode(&extra); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	if extra == nil {
		return nil
	}
	return errors.New("config must contain a single YAML document")
}

func validateDuration(field string, raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	if _, err := time.ParseDuration(raw); err != nil {
		return fmt.Errorf("%s must be a valid duration: %w", field, err)
	}

	return nil
}

func isKnownCheckerType(name string) bool {
	_, ok := knownCheckerTypes[strings.TrimSpace(name)]
	return ok
}

func sortedKnownCheckerTypes() []string {
	types := make([]string, 0, len(knownCheckerTypes))
	for name := range knownCheckerTypes {
		types = append(types, name)
	}
	sort.Strings(types)
	return types
}
