# wypas-lib

Shared Go module for the Wypas project. Provides common packages used by `wypas-api`, `wypas-creator`, and other Go services.

## Module

```
github.com/codefatherllc/wypas-lib
```

Go 1.22+

## Packages

| Package | Purpose |
|---------|---------|
| `config` | Env var helpers (`GetEnv`, `GetEnvInt`, `SplitCSV`, `ConvertDSN`) |
| `db` | MySQL connection pool with configured limits and lifetime |
| `jwt` | JWT (HS256) parsing, signing, middleware, admin guard |
| `middleware` | CORS middleware for `net/http` |
| `gamedata` | DB-backed item types and world data (structs + store functions) |
| `otb` | Legacy file loading (OTB+XML to gamedata.ItemType, OTBM+XMLs to gamedata world types) |
| `otbm` | Raw OTBM/OTB binary format parser (internal — `otb` is the high-level loader) |
| `ratelimit` | In-memory IP rate limiter (sliding window) |
| `response` | JSON response helpers (`JSON`, `Error`) |
| `sprite` | .dat/.spr appearance and sprite parser, cache, item/outfit rendering with color palette |
| `worlds` | Game world list loader from JSON (`worlds.json`) |

## Install

```bash
go get github.com/codefatherllc/wypas-lib@latest
```

For local development with a sibling repo, add a replace directive:

```go
replace github.com/codefatherllc/wypas-lib => ../wypas-lib
```

## Release

Tags trigger CI. Push a semver tag to create a GitHub Release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

CI runs `go vet` + `go test` before publishing the release.

## Dependencies

- `github.com/go-sql-driver/mysql` — MySQL driver
- `github.com/golang-jwt/jwt/v5` — JWT

## License

Part of the Wypas project.
