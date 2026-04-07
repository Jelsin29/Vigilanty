# Provider Guide

Vigilanty treats AI review as a pipeline step, not a side quest. That means provider setup needs to be understandable, predictable, and easy to swap.

## See also

- [README](../README.md)
- [Configuration](./configuration.md)
- [Examples](./examples.md)
- [Rules file](./rules-file.md)
- [Troubleshooting](./troubleshooting.md)

## Interactive provider selection

During `vigilanty init`, the wizard helps you choose a provider as part of the onboarding flow.

<p align="center">
  <img width="1400" alt="Vigilanty provider selection screen" src="https://github.com/user-attachments/assets/dbb63929-bc77-4b33-a476-e67f8c58f359" />
</p>

That selection is written into the generated `.vigilanty.yml` so your first run is already wired for AI review.

## Supported providers

| Provider | CLI expectation | Model support | Notes |
| --- | --- | --- | --- |
| `claude` | `claude` in PATH | usually no explicit model | Uses stdin prompt mode |
| `gemini` | `gemini` in PATH | optional in config | Uses stdin prompt mode |
| `codex` | `codex` in PATH | usually no explicit model | Executes prompt as positional argument |
| `opencode` | `opencode` in PATH | optional, passed with `--model` | Good fit when your team already uses OpenCode |
| `ollama` | `ollama` in PATH | **required** | Runs `ollama run <model>` locally |
| `lmstudio` | provider CLI in PATH | optional in config | Model may be selected by the runtime |
| `github` | `gh` in PATH | optional in config | Useful for GitHub-centric workflows |

## Important nuance on models

Here is the real behavior today:

- **Ollama requires `model`** in config
- **OpenCode uses `model` when provided**
- **Other providers may work without `model`** because the provider CLI can resolve a default account/model context

The interactive wizard may ask for a model for some providers so the generated config is explicit, but the strict requirement in the checker is only enforced for Ollama.

## Example provider configs

### Claude

```yaml
- name: ai-review
  checker: ai-review
  provider: claude
  rules_file: "AGENTS.md"
  prompt: "Review this diff for bugs, maintainability issues, and security risks."
```

### Gemini

```yaml
- name: ai-review
  checker: ai-review
  provider: gemini
  prompt: "Review this diff and report only concrete issues."
```

### OpenCode

```yaml
- name: ai-review
  checker: ai-review
  provider: opencode
  model: openai/gpt-5.4
  prompt: "Review this diff for correctness, architecture drift, and risky edge cases."
```

### Ollama

```yaml
- name: ai-review
  checker: ai-review
  provider: ollama
  model: llama3.1
  prompt: "Review this diff for bugs and maintainability issues."
```

## Rules file interaction

Provider output is not treated as free-form chat. Vigilanty expects a structured review response and prepends the coding standards from:

1. the explicit `rules_file`, if configured
2. otherwise `AGENTS.md` in the repo root

That means provider quality improves when you give it a strong rules file. Read [rules-file.md](./rules-file.md) for the full flow.

## Provider install links

If the required CLI is missing, Vigilanty surfaces install guidance. Current provider install targets map to:

- Claude: <https://docs.anthropic.com/en/docs/claude-code>
- Gemini: <https://github.com/google-gemini/gemini-cli>
- Ollama: <https://ollama.com/download>
- Codex: <https://platform.openai.com/docs/codex/cli>
- OpenCode: <https://github.com/sst/opencode>
- LM Studio: <https://lmstudio.ai/>
- GitHub CLI: <https://cli.github.com/>

## How Vigilanty calls providers

The provider integration layer is intentionally thin:

- stdin prompt mode for providers like Claude and Gemini
- positional prompt execution for providers like Codex and OpenCode
- local runtime invocation for Ollama

That is a FEATURE. Vigilanty is not trying to replace the provider CLI; it is orchestrating it inside a consistent verification pipeline.

## Choosing the right provider

### Choose hosted providers when you want

- better frontier-model quality
- fast onboarding for cloud-first teams
- fewer local runtime concerns

### Choose local runtimes when you want

- offline or self-hosted workflows
- tighter control over model runtime
- lower marginal review cost

Tradeoff: local control is great, but hosted providers are usually easier to standardize across a team.

## Failure modes to expect

- CLI not installed or not in `PATH`
- provider authentication not completed
- ambiguous response that does not match pass/fail patterns
- model not available locally for Ollama
- custom prompt too vague to produce deterministic pass/fail output

If that happens, go straight to [Troubleshooting](./troubleshooting.md).
