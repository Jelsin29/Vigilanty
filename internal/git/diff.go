package git

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"unicode/utf8"
)

func StagedFiles() ([]string, error) {
	output, err := exec.Command("git", "diff", "--cached", "--name-only", "--diff-filter=ACMR", "-z").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("list staged files: %w: %s", err, strings.TrimSpace(string(output)))
	}

	raw := strings.Split(string(output), "\x00")
	files := make([]string, 0, len(raw))
	for _, file := range raw {
		if strings.TrimSpace(file) == "" {
			continue
		}
		files = append(files, file)
	}

	return files, nil
}

func StagedDiff() (string, error) {
	files, err := StagedFiles()
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", nil
	}

	binaryFiles, err := stagedBinaryFiles()
	if err != nil {
		return "", err
	}

	textFiles := make([]string, 0, len(files))
	for _, file := range files {
		if _, isBinary := binaryFiles[file]; !isBinary {
			textFiles = append(textFiles, file)
		}
	}

	var builder strings.Builder
	if len(textFiles) > 0 {
		args := []string{"diff", "--cached", "--"}
		args = append(args, textFiles...)

		output, err := exec.Command("git", args...).CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("read staged diff: %w: %s", err, strings.TrimSpace(string(output)))
		}

		builder.Write(output)
	}

	if len(binaryFiles) > 0 {
		if builder.Len() > 0 && !strings.HasSuffix(builder.String(), "\n") {
			builder.WriteString("\n")
		}
		builder.WriteString("Binary staged files excluded from diff:\n")
		for _, file := range files {
			if _, isBinary := binaryFiles[file]; isBinary {
				builder.WriteString("- ")
				builder.WriteString(file)
				builder.WriteString("\n")
			}
		}
	}

	return builder.String(), nil
}

func StagedDiffTruncated(maxBytes int) (string, bool, error) {
	if maxBytes <= 0 {
		return "", false, errors.New("truncate staged diff: maxBytes must be greater than zero")
	}

	diff, err := StagedDiff()
	if err != nil {
		return "", false, err
	}

	if len(diff) <= maxBytes {
		return diff, false, nil
	}

	truncated := diff[:maxBytes]
	for len(truncated) > 0 && !utf8.ValidString(truncated) {
		truncated = truncated[:len(truncated)-1]
	}

	return truncated + "\n... diff truncated ...\n", true, nil
}

func stagedBinaryFiles() (map[string]struct{}, error) {
	output, err := exec.Command("git", "diff", "--cached", "--numstat", "-z").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("inspect staged diff metadata: %w: %s", err, strings.TrimSpace(string(output)))
	}

	parts := strings.Split(string(output), "\x00")
	binaryFiles := make(map[string]struct{})
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}

		fields := strings.Split(part, "\t")
		if len(fields) < 3 {
			continue
		}

		if fields[0] == "-" && fields[1] == "-" {
			binaryFiles[fields[2]] = struct{}{}
		}
	}

	return binaryFiles, nil
}
