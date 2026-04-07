# Troubleshooting

When Vigilanty fails, the right mindset is simple: identify whether the problem is **config**, **git context**, **provider runtime**, or **your pipeline step itself**.

## See also

- [README](../README.md)
- [Configuration](./configuration.md)
- [Providers](./providers.md)
- [Examples](./examples.md)
- [`run --json`](./run-json.md)

## Failure output reference

<p align="center">
  <img width="962" height="563" alt="Vigilanty failed run output" src="https://github.com/user-attachments/assets/92b6aa0e-c486-407c-aef1-95bd62e55858" />
</p>

Use failure output to answer three questions first:

1. Which step failed?
2. Did it fail, skip, or error?
3. Is the problem deterministic tooling, AI provider wiring, or repository state?

## Common problems

### `error: no Vigilanty config file found`

Cause:

- there is no `.vigilanty.yml` in the repo
- no global config exists in `~/.config/vigilanty/config.yml`
- `--config` was not supplied

Fix:

```bash
vigilanty init
```

Or point explicitly to a config:

```bash
vigilanty --config path/to/config.yml run
```

### `error: not a git repository`

Cause:

- you ran a git-dependent command outside a repository

Fix:

- run the command from inside the target repo
- verify `.git/` exists

### `error: cannot use both --pr-mode and --ci`

Cause:

- those run modes are mutually exclusive

Fix:

- use exactly one of:
  - `vigilanty run`
  - `vigilanty run --pr-mode --base main`
  - `vigilanty run --ci`

### `error: --base requires --pr-mode`

Cause:

- `--base` only makes sense in PR comparison mode

Fix:

```bash
vigilanty run --pr-mode --base main
```

### AI CLI not found

Cause:

- the configured provider executable is not available in `PATH`

Fix:

- install the provider CLI
- authenticate it if required
- retry the run

Provider install references are listed in [providers.md](./providers.md).

### Ollama model error

Cause:

- `provider: ollama` is configured without a `model`

Fix:

```yaml
- name: ai-review
  checker: ai-review
  provider: ollama
  model: llama3.1
  prompt: "Review this diff for bugs and maintainability issues."
```

### Ambiguous AI response

Cause:

- provider output did not match Vigilanty's pass/fail patterns
- your prompt may be too vague
- the model may be ignoring the required response format

Fixes:

- keep `prompt` concrete and review-focused
- use a strong `AGENTS.md` or explicit `rules_file`
- customize `pass_pattern` / `fail_pattern` only if you have a real reason

### Invalid timeout or config validation errors

Cause:

- unsupported schema values
- invalid duration syntax
- missing required fields such as `provider`, `prompt`, or `command`

Fix:

- inspect the file with `vigilanty config`
- compare it against [configuration.md](./configuration.md)

### Empty diff behavior

Cause:

- there are no staged changes in default mode
- the selected diff source produced no content
- AI review may skip if `skip_on_empty_diff: true`

Fix:

- stage files before local runs
- use `--pr-mode` or `--ci` when that better matches the workflow

## Debugging checklist

When something feels off, do this in order:

1. Run `vigilanty config`
2. Confirm the repo and git mode are correct
3. Confirm the provider CLI exists and is authenticated
4. Confirm your `rules_file` or `AGENTS.md` exists where expected
5. Confirm the diff source actually contains changes

That sequence avoids random guessing. Infrastructure first, then prompt quality.

## When to clear cache

If you suspect stale step reuse:

```bash
vigilanty cache status
vigilanty cache clear
```

Do not clear cache by reflex. First understand whether cache reuse is actually the issue.
