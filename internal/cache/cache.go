package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const currentVersion = 1

type Entry struct {
	Status     string    `json:"status"`
	Timestamp  time.Time `json:"timestamp"`
	ConfigHash string    `json:"config_hash"`
}

type fileData struct {
	Version int              `json:"version"`
	Entries map[string]Entry `json:"entries"`
}

type Cache struct {
	path string
	data fileData
}

func New(root string) *Cache {
	return &Cache{
		path: filepath.Join(CacheDir(), ProjectHash(root)+".json"),
		data: fileData{
			Version: currentVersion,
			Entries: map[string]Entry{},
		},
	}
}

func CacheDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".cache", "vigilanty")
	}

	return filepath.Join(homeDir, ".cache", "vigilanty")
}

func ProjectHash(root string) string {
	return sha256Hex([]byte(strings.TrimSpace(root)))
}

func FilesHash(root string, files []string, mode string) string {
	cloned := append([]string(nil), files...)
	sort.Strings(cloned)

	hasher := sha256.New()
	for _, file := range cloned {
		_, _ = hasher.Write([]byte(file))
		_, _ = hasher.Write([]byte{0})

		var (
			content []byte
			err     error
		)

		switch strings.TrimSpace(mode) {
		case "pr", "ci":
			content, err = os.ReadFile(filepath.Join(root, file))
		default:
			content, err = stagedFileContent(root, file)
			if err != nil {
				content, err = os.ReadFile(filepath.Join(root, file))
			}
		}

		if err != nil {
			_, _ = hasher.Write([]byte("__MISSING__"))
			_, _ = hasher.Write([]byte{0})
			continue
		}

		_, _ = hasher.Write(content)
		_, _ = hasher.Write([]byte{0})
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

func ConfigHash(cfg map[string]interface{}) string {
	if cfg == nil {
		cfg = map[string]interface{}{}
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return sha256Hex([]byte("{}"))
	}

	return sha256Hex(data)
}

func (c *Cache) Path() string {
	if c == nil {
		return ""
	}

	return c.path
}

func (c *Cache) Load() error {
	if c == nil {
		return fmt.Errorf("cache is nil")
	}

	data, err := os.ReadFile(c.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.data = fileData{Version: currentVersion, Entries: map[string]Entry{}}
			return nil
		}
		return fmt.Errorf("read cache %q: %w", c.path, err)
	}

	var decoded fileData
	if err := json.Unmarshal(data, &decoded); err != nil {
		return fmt.Errorf("decode cache %q: %w", c.path, err)
	}

	if decoded.Version != currentVersion {
		decoded = fileData{Version: currentVersion, Entries: map[string]Entry{}}
	}
	if decoded.Entries == nil {
		decoded.Entries = map[string]Entry{}
	}

	c.data = decoded
	return nil
}

func (c *Cache) Save() error {
	if c == nil {
		return fmt.Errorf("cache is nil")
	}

	if c.data.Version == 0 {
		c.data.Version = currentVersion
	}
	if c.data.Entries == nil {
		c.data.Entries = map[string]Entry{}
	}

	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}

	data, err := json.MarshalIndent(c.data, "", "  ")
	if err != nil {
		return fmt.Errorf("encode cache %q: %w", c.path, err)
	}

	if err := os.WriteFile(c.path, data, 0o644); err != nil {
		return fmt.Errorf("write cache %q: %w", c.path, err)
	}

	return nil
}

func (c *Cache) Lookup(checkerName string, filesHash string, configHash string) (Entry, bool) {
	if c == nil {
		return Entry{}, false
	}

	entry, ok := c.data.Entries[cacheKey(checkerName, filesHash)]
	if !ok {
		return Entry{}, false
	}

	if entry.Status != "passed" || entry.ConfigHash != configHash {
		return Entry{}, false
	}

	return entry, true
}

func (c *Cache) Store(checkerName string, filesHash string, configHash string) error {
	if c == nil {
		return fmt.Errorf("cache is nil")
	}

	if c.data.Entries == nil {
		c.data.Entries = map[string]Entry{}
	}

	c.data.Entries[cacheKey(checkerName, filesHash)] = Entry{
		Status:     "passed",
		Timestamp:  time.Now().UTC(),
		ConfigHash: configHash,
	}

	return c.Save()
}

func (c *Cache) Remove(checkerName string) error {
	if c == nil {
		return fmt.Errorf("cache is nil")
	}

	prefix := strings.TrimSpace(checkerName) + ":"
	for key := range c.data.Entries {
		if strings.HasPrefix(key, prefix) {
			delete(c.data.Entries, key)
		}
	}

	return c.Save()
}

func (c *Cache) Clear() error {
	if c == nil {
		return fmt.Errorf("cache is nil")
	}

	if err := os.Remove(c.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove cache %q: %w", c.path, err)
	}

	c.data = fileData{Version: currentVersion, Entries: map[string]Entry{}}
	return nil
}

func (c *Cache) EntryCount() int {
	if c == nil {
		return 0
	}

	return len(c.data.Entries)
}

func cacheKey(checkerName string, filesHash string) string {
	return strings.TrimSpace(checkerName) + ":" + strings.TrimSpace(filesHash)
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func stagedFileContent(root string, file string) ([]byte, error) {
	if filepath.IsAbs(file) {
		return nil, fmt.Errorf("absolute path %q cannot be resolved from git index", file)
	}

	cmd := exec.Command("git", "show", ":"+file)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("read staged content for %q: %w: %s", file, err, strings.TrimSpace(string(output)))
	}

	return output, nil
}
