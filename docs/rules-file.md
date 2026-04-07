# Rules File Guide

Vigilanty gets dramatically better when AI review has explicit project rules. That is what `AGENTS.md` and `rules_file` are for.

## See also

- [README](../README.md)
- [Configuration](./configuration.md)
- [Providers](./providers.md)
- [Architecture](./architecture.md)

## What `AGENTS.md` does

During AI review, Vigilanty builds a prompt from four ingredients:

1. built-in review framing
2. coding standards from `rules_file` or root `AGENTS.md`
3. your custom `prompt`
4. the diff under review

That means `AGENTS.md` is not decoration. It is part of the actual review contract.

## Interactive generation

The setup wizard can generate `AGENTS.md` for you during `vigilanty init`.

<p align="center">
  <img width="1400" alt="Vigilanty AGENTS.md generation step" src="https://github.com/user-attachments/assets/8a8f4187-7be2-457c-8f75-59e191ad667e" />
</p>

Generated content includes:

- file include and exclude scope
- language-specific reject/require/prefer rules
- detected tooling hints when known configs exist
- a strict response format for PASS / FAIL reviews

## Resolution order

When AI review runs, Vigilanty looks for rules in this order:

1. explicit `rules_file` from the step config
2. `AGENTS.md` in the project root

If neither exists, AI review still runs, but without repository-specific coding standards.

## Example step with explicit rules file

```yaml
- name: ai-review
  checker: ai-review
  provider: claude
  rules_file: "docs/review-rules.md"
  prompt: "Review this diff for correctness and maintainability."
```

## When to keep the default `AGENTS.md`

Keep the default when:

- the repository has one main coding standard source
- you want the review rules to live at the root
- you want the wizard-generated behavior with minimal friction

## When to use a custom `rules_file`

Use a custom file when:

- you want multiple rule sets across repos or environments
- the root already has another `AGENTS.md` convention
- you want tighter control over AI review docs organization

Tradeoff: custom `rules_file` is more explicit, but root `AGENTS.md` is easier for most teams to discover.

## What a strong rules file looks like

Good rules files are:

- concrete
- enforceable from a diff
- phrased as reject/require/prefer guidance
- aligned with actual project tooling

Bad rules files are vague philosophy without operational consequences.

## What the generated file includes

Today, generated rules content is shaped by:

- detected preset or language
- selected include/exclude patterns
- detected tools like ESLint, Prettier, Ruff, or golangci-lint

That gives the AI review step more repo awareness without requiring a huge custom prompt.

## Recommendation

Start with generated `AGENTS.md`, review it like any other contract file, then tighten the rules as your team learns what kinds of review feedback are actually valuable.
