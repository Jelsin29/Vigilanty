package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	cachepkg "github.com/jelsin29/vigilanty/internal/cache"
	"github.com/spf13/cobra"
)

func newCacheCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "cache",
		Short: "Inspect and manage Vigilanty cache files",
	}

	command.AddCommand(newCacheStatusCommand())
	command.AddCommand(newCacheClearCommand())

	return command
}

func newCacheStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show cache location and usage",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := cachepkg.CacheDir()
			entries, totalSize, err := cacheStatus(dir)
			if err != nil {
				return newExitError(ExitInternalError, "%s", errorText(fmt.Sprintf("error: failed to inspect cache: %v", err)))
			}

			fmt.Fprintf(os.Stdout, "Location: %s\n", dir)
			fmt.Fprintf(os.Stdout, "Entries: %d\n", entries)
			fmt.Fprintf(os.Stdout, "Total size: %s\n", formatSize(totalSize))
			return nil
		},
	}
}

func newCacheClearCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Delete all Vigilanty cache files",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := cachepkg.CacheDir()
			if err := os.RemoveAll(dir); err != nil {
				return newExitError(ExitInternalError, "%s", errorText(fmt.Sprintf("error: failed to clear cache: %v", err)))
			}

			fmt.Fprintf(os.Stdout, "%s\n", successText("cache cleared"))
			return nil
		},
	}
}

func cacheStatus(dir string) (int, int64, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, 0, nil
		}
		return 0, 0, err
	}

	entries := 0
	var totalSize int64
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		info, err := file.Info()
		if err != nil {
			return 0, 0, err
		}
		totalSize += info.Size()

		data, err := os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return 0, 0, err
		}

		var payload struct {
			Entries map[string]json.RawMessage `json:"entries"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			return 0, 0, err
		}

		entries += len(payload.Entries)
	}

	return entries, totalSize, nil
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}
