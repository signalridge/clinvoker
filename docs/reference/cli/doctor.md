# clinvk doctor

Run environment and configuration diagnostics.

## Synopsis

```bash
clinvk doctor [flags]
```

## Description

`doctor` performs a grouped health check for local runtime readiness and configuration quality.
It is designed for both human troubleshooting and CI gate checks.

Checks include:

- configuration load state
- configuration semantic validation
- session store access
- enabled backend availability
- server auth key status

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output diagnostics as machine-readable JSON |

## Examples

Run diagnostics in text mode:

```bash
clinvk doctor
```

Run diagnostics in JSON mode:

```bash
clinvk doctor --json
```

## Output

Text output example:

```text
Doctor checks:
- [pass] config.load: configuration loaded
- [pass] config.validate: configuration is valid
- [warn] server.auth: API key authentication is disabled (no keys configured)

Summary: pass=4 warn=1 fail=0
```

JSON output example:

```json
{
  "schema_version": "v1",
  "summary": {
    "pass": 4,
    "warn": 1,
    "fail": 0
  },
  "checks": [
    {
      "id": "config.validate",
      "status": "pass",
      "message": "configuration is valid"
    }
  ]
}
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | No failing checks |
| 2 | One or more checks failed |

## See Also

- [config](config.md) - Validate and manage configuration
- [serve](serve.md) - Server runtime configuration
