### Before coding

- Ask clarifying questions for ambiguous requirements.

### Dependencies

- Prefer stdlib; introduce deps only with clear payoff.

### Code style

- Enforce `gofmt`, `go vet`
- Avoid stutter in names: `package kv; type Store` (not `KVStore` in `kv`).
- Small interfaces near consumers; prefer composition over inheritance.
- Avoid reflection on hot paths; prefer generics when it clarifies and speeds.

### Errors

- Wrap with `%w` and context: `fmt.Errorf("open %s: %w", p, err)`.
- Use `errors.Is`/`errors.As` for control flow; no string matching.
- Define sentinel errors in the package; document behavior.
- Use `context.WithCancelCause` and `context.Cause` for propagating error causes.

### Testing

- Table‑driven tests; deterministic and hermetic by default.
- Run `-race` in CI; add `t.Cleanup` for teardown.
- Mark safe tests with `t.Parallel()`.

### Logging

- Structured logging (`slog`) with levels and consistent fields.
- Correlate logs/metrics/traces via request IDs from context.

### Performance

- Measure before optimizing: `pprof`, `go test -bench`, `benchstat`.
- Avoid allocations on hot paths; reuse buffers with care; prefer `bytes`/`strings` APIs.

### Configuration

- Config via env/flags; validate on startup; fail fast.
- Treat config as immutable after init; pass explicitly (not via globals).
- Provide sane defaults and clear docs.

### APIs

- Document exported items: `// Foo does …`; keep exported surface minimal.
- Accept interfaces where variation is needed; **return concrete types** unless abstraction is required.
- Keep functions small, orthogonal, and composable.
- Use constructor options pattern for extensibility.

### Security

- Validate inputs; set explicit I/O timeouts; prefer TLS everywhere.
- Never log secrets; manage secrets outside code (env/secret manager).
- Limit filesystem/network access by default; principle of least privilege.
- Add fuzz tests for untrusted inputs.

### Tooling

- Linters: `golangci-lint`, `staticcheck`, `gofumpt`.
- Security: `govulncheck`, dependency scanners.

### Tooling gates

- `go vet ./...` passes.
- `golangci-lint run` passes with project config.
- `go test -race ./...` passes.

## Writing functions best practices

1. Can you read the function and HONESTLY easily follow what it's doing? If yes, then stop here.
2. Does the function have very high cyclomatic complexity? (number of independent paths, or, in a lot of cases, number of nesting if if-else as a proxy). If it does, then it's probably sketchy.
3. Does it have any hidden untested dependencies or any values that can be factored out into the arguments instead? Only care about non-trivial dependencies that can actually change or affect the function.
