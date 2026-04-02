package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode/utf8"
)

func StagedFiles(root string) ([]string, error) {
	output, err := runGit(root, "diff", "--cached", "--name-only", "--diff-filter=ACMR", "-z")
	if err != nil {
		return nil, fmt.Errorf("list staged files: %w: %s", err, strings.TrimSpace(string(output)))
	}

	raw := strings.Split(output, "\x00")
	files := make([]string, 0, len(raw))
	for _, file := range raw {
		if strings.TrimSpace(file) == "" {
			continue
		}
		files = append(files, file)
	}

	return files, nil
}

func StagedDiff(root string) (string, error) {
	files, err := StagedFiles(root)
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", nil
	}

	binaryFiles, err := stagedBinaryFiles(root)
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

		output, err := runGit(root, args...)
		if err != nil {
			return "", fmt.Errorf("read staged diff: %w: %s", err, strings.TrimSpace(string(output)))
		}

		builder.WriteString(output)
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

func PRDiffWithFiles(root, base string) (string, []string, error) {
	resolvedBase := strings.TrimSpace(base)
	if resolvedBase == "" {
		var err error
		resolvedBase, err = DetectBaseBranch(root)
		if err != nil {
			return "", nil, err
		}
	}

	output, err := runGit(root, "diff", fmt.Sprintf("%s...HEAD", resolvedBase))
	if err != nil {
		return "", nil, fmt.Errorf("read PR diff against %q: %w: %s", resolvedBase, err, strings.TrimSpace(output))
	}

	return output, filesFromDiff(output), nil
}

func LastCommitDiffWithFiles(root string) (string, []string, error) {
	revisionRange, err := lastCommitRevisionRange(root)
	if err != nil {
		return "", nil, err
	}

	args := []string{"diff"}
	args = append(args, revisionRange...)
	output, err := runGit(root, args...)
	if err != nil {
		return "", nil, fmt.Errorf("read last commit diff: %w: %s", err, strings.TrimSpace(output))
	}

	return output, filesFromDiff(output), nil
}

func DetectBaseBranch(root string) (string, error) {
	candidates := []string{"main", "master", "origin/main", "origin/master"}
	for _, candidate := range candidates {
		if _, err := runGit(root, "rev-parse", "--verify", candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("detect base branch: no base branch found (tried %s). Use --base to specify one explicitly", strings.Join(candidates, ", "))
}

func DetectCI() (string, bool) {
	checks := []struct {
		name  string
		key   string
		value string
	}{
		{name: "GitHub Actions", key: "GITHUB_ACTIONS", value: "true"},
		{name: "GitLab CI", key: "GITLAB_CI", value: "true"},
		{name: "Jenkins", key: "JENKINS_URL"},
		{name: "CircleCI", key: "CIRCLECI", value: "true"},
		{name: "CI", key: "CI", value: "true"},
	}

	for _, check := range checks {
		current := strings.TrimSpace(os.Getenv(check.key))
		if current == "" {
			continue
		}
		if check.value == "" || strings.EqualFold(current, check.value) {
			return check.name, true
		}
	}

	return "", false
}

func StagedDiffTruncated(root string, maxBytes int) (string, bool, error) {
	if maxBytes <= 0 {
		return "", false, errors.New("truncate staged diff: maxBytes must be greater than zero")
	}

	diff, err := StagedDiff(root)
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

func filesFromDiff(diff string) []string {
	lines := strings.Split(diff, "\n")
	files := make([]string, 0)
	seen := make(map[string]struct{})

	for _, line := range lines {
		if !strings.HasPrefix(line, "diff --git a/") {
			continue
		}

		rest := strings.TrimPrefix(line, "diff --git a/")
		idx := strings.LastIndex(rest, " b/")
		if idx <= 0 {
			continue
		}

		file := strings.TrimSpace(rest[idx+3:])
		if file == "" {
			continue
		}
		if _, ok := seen[file]; ok {
			continue
		}

		seen[file] = struct{}{}
		files = append(files, file)
	}

	return files
}

func lastCommitRevisionRange(root string) ([]string, error) {
	if _, err := runGit(root, "rev-parse", "--verify", "HEAD~1"); err == nil {
		return []string{"HEAD~1..HEAD"}, nil
	}

	if _, err := runGit(root, "rev-parse", "--verify", "HEAD"); err != nil {
		return nil, fmt.Errorf("resolve last commit revision range: %w", err)
	}

	return []string{"4b825dc642cb6eb9a060e54bf8d69288fbee4904", "HEAD"}, nil
}

func stagedBinaryFiles(root string) (map[string]struct{}, error) {
	output, err := runGit(root, "diff", "--cached", "--numstat", "-z")
	if err != nil {
		return nil, fmt.Errorf("inspect staged diff metadata: %w: %s", err, strings.TrimSpace(output))
	}

	parts := strings.Split(output, "\x00")
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

func runGit(root string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if strings.TrimSpace(root) != "" {
		cmd.Dir = root
	}

	output, err := cmd.CombinedOutput()
	return string(output), err
}
