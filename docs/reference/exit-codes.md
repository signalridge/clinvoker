# Exit Codes

Meaning of clinvk exit codes.

## Summary

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (validation or command failure) |
| 124 | Command timed out (if configured) |
| (backend) | Backend exit code (root / resume) |

## Notes

- For `clinvk [prompt]` and `clinvk resume`, the backend process exit code is propagated when non‑zero.
- `parallel`, `chain`, and `compare` return `1` when any task/step fails.
