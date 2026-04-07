# `vigilanty run --json`

`--json` keeps the normal human-readable output untouched for default runs, but switches stdout to a stable JSON document for CI and automation.

## See also

- [README](../README.md)
- [CLI reference](./cli.md)
- [Integrations](./integrations.md)
- [Examples](./examples.md)
- [Troubleshooting](./troubleshooting.md)

## Compatibility rules

- `schema_version` is currently `v1`
- additive fields are allowed in the future
- breaking changes require a new schema version
- when `--json` is enabled, stdout is reserved for JSON only
- exit codes stay the same: `0` for pass, `1` for pipeline failure

## Schema

| Field | Type | Description |
| --- | --- | --- |
| `schema_version` | string | Contract version |
| `passed` | boolean | Overall pipeline result |
| `mode` | string | `staged`, `pr`, or `ci` |
| `duration_ms` | integer | Total pipeline duration in milliseconds |
| `files` | string[] | Files included in the evaluated diff |
| `truncated_diff` | boolean | Whether the diff had to be truncated |
| `steps` | array | Per-step results |
| `summary` | string | Human-readable final summary |

### Step object

| Field | Type | Description |
| --- | --- | --- |
| `name` | string | Step name |
| `type` | string | Checker type |
| `status` | string | `passed`, `failed`, `skipped`, or `error` |
| `duration_ms` | integer | Step duration in milliseconds |
| `output` | string | Trimmed checker output when present |
| `details` | array | Optional structured findings |
| `provider` | string | AI provider when relevant |
| `model` | string | Model name when relevant |
| `cached` | boolean | Whether the step was resolved from cache |
| `files` | string[] | Files passed into the step |
| `timeout` | string | Step timeout |

## Example payload

```json
{"schema_version":"v1","passed":false,"mode":"ci","duration_ms":1500,"files":["main.go"],"truncated_diff":true,"steps":[{"name":"lint","type":"shell","status":"passed","duration_ms":100,"files":["main.go"],"timeout":"60s"},{"name":"test","type":"shell","status":"failed","duration_ms":200,"output":"tests failed","files":["main.go"],"timeout":"120s"},{"name":"review","type":"ai-review","status":"skipped","duration_ms":0,"provider":"claude","files":["main.go"],"timeout":"120s"},{"name":"security","type":"shell","status":"error","duration_ms":300,"output":"checker crashed","details":[{"file":"main.go","line":7,"message":"panic","severity":"high"}],"cached":true,"files":["main.go"],"timeout":"90s"}],"summary":"Pipeline failed at \"test\" (1/4 checkers passed) in 1.5s"}
```

The example above is intentionally the same shape used by the golden fixture in `internal/pipeline/testdata/run_json_golden.json`.

## Recommended usage

- Local terminal workflow: `vigilanty run`
- CI workflow: `vigilanty run --ci --json`
- PR diff automation: `vigilanty run --pr-mode --base main --json`
