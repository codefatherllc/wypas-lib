# wypas-lib — Shared Go Module

## Build

```bash
go vet ./...
go test ./...
```

No binary output. This is a library module consumed via `go get`.

## Packages

| Package | Purpose |
|---------|---------|
| `config` | `GetEnv`, `GetEnvInt`, `SplitCSV`, `ConvertDSN` (mysql:// URI to Go DSN) |
| `db` | `Open(dsn, maxOpen, maxIdle)` — MySQL pool with 5min conn lifetime |
| `jwt` | `Claims`, `ParseToken`, `SignToken`, `Middleware`, `RequireAdmin`, `GetClaims` |
| `middleware` | `CORS(origins...)` — Access-Control headers + OPTIONS preflight |
| `otbm` | `ParseOTBM` — binary OTBM/OTB parser (tiles, towns, items) |
| `ratelimit` | `New(max, window)` → `Limiter.Allow(ip)` — in-memory sliding window |
| `response` | `JSON(w, status, data)`, `Error(w, status, msg)` |
| `sprite` | `Cache` (dat+spr loader, PNG render), `DatFile`, `SpriteFile`, outfit color palette |
| `worlds` | `Load(path)` → `WorldList` from JSON, `All()`, `ByID()`, `Default()` |

## Conventions

- Go 1.22, module path `github.com/codefatherllc/wypas-lib`
- No CLI, no main package
- Consumers use `replace` directive for local dev
- Release via git tags (`v*`), CI runs vet + test + GitHub Release

## CI

`.github/workflows/release.yml` — triggered on `v*` tags. Runs `go vet`, `go test`, creates GitHub Release.
