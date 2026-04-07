# Architecture Overview

This page is for contributors and advanced users who want to understand how Vigilanty is structured internally.

## See also

- [README](../README.md)
- [CLI reference](./cli.md)
- [Configuration](./configuration.md)
- [`run --json`](./run-json.md)
- [Rules file](./rules-file.md)

## High-level flow

```text
CLI command
   |
   v
config resolution + git diff resolution
   |
   v
pipeline construction
   |
   +--> shell checker steps
   |
   +--> AI review step
   |
   v
human output or stable JSON output
```

## Main packages

| Package | Responsibility |
| --- | --- |
| `cmd/` | Cobra commands such as `init`, `run`, `install`, `config`, and cache operations |
| `internal/config/` | YAML loading, defaults, validation, and preset generation |
| `internal/pipeline/` | Step orchestration, result formatting, cache integration, JSON payload generation |
| `internal/checker/` | Shell checker and AI review checker implementations |
| `internal/git/` | Repo detection, staged diff, PR diff, CI diff, and hook installation |
| `internal/wizard/` | Interactive init result generation |
| `internal/tui/` | Bubble Tea-based onboarding experience for interactive setup |
| `internal/rules/` | `AGENTS.md` generation and rules discovery |
| `internal/cache/` | Persistent step cache management |
| `internal/ui/` | Output styling for banners and terminal presentation |

## Command layer

The `cmd/` package is the entry point from Cobra. It handles:

- CLI argument parsing
- mode validation such as `--pr-mode` vs `--ci`
- config source resolution
- printing human or JSON output

This keeps the application shell thin while pushing actual behavior into internal packages.

## Configuration layer

`internal/config/` owns:

- schema structs
- YAML decoding with known-field validation
- default application
- preset generation for `vigilanty init`

This is important because configuration is effectively the product API. If the config contract is sloppy, the whole tool becomes sloppy.

## Wizard and TUI

Interactive setup is split into:

- `internal/wizard/` for orchestration and defaults
- `internal/tui/` for Bubble Tea presentation and step flow

That separation makes the user experience richer without entangling the command layer with terminal UI details.

## Pipeline engine

`internal/pipeline/` is the center of execution. It is responsible for:

- ordering steps
- caching step results
- tracking timing and final status
- deciding how results are formatted for humans vs JSON

Conceptually, this is the verification hub promised in the README.

## Checker model

Vigilanty currently revolves around two checker families:

### Shell checker

Runs deterministic commands such as:

- linters
- typecheckers
- build commands
- tests
- security scanners

### AI review checker

Builds a review prompt from:

- built-in instructions
- `AGENTS.md` or `rules_file`
- custom prompt text
- the active diff

Then it calls the configured provider CLI and interprets the response through pass/fail regex patterns.

## Git integration

`internal/git/` supports the three primary diff sources:

- staged diff for local development
- PR diff vs base branch
- last commit diff for CI

It also installs and removes the Git pre-commit hook used by `vigilanty install` and `vigilanty uninstall`.

## Output modes

Vigilanty intentionally has two output contracts:

1. **human-readable default output** for local developer workflows
2. **stable JSON output** for CI and automation

That split is deliberate. Trying to make one output format serve both audiences usually makes both worse.

## Design philosophy

The architecture favors:

- thin CLI surface
- explicit configuration
- provider-agnostic AI integration
- git-aware diff selection
- strong separation between local human UX and machine-readable automation

If you are contributing, keep those boundaries sharp.
