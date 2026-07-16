module github.com/codefatherllc/wypas-lib

go 1.22.3

require (
	github.com/go-sql-driver/mysql v1.8.1
	github.com/golang-jwt/jwt/v5 v5.2.1
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	golang.org/x/image v0.18.0
)

// v1.27.0 and v1.28.0 are sge-era code tagged in the legacy (v1) version
// space by mistake — the sge line lives at github.com/codefatherllc/wypas-lib/v2
// (v2.x.y). Retracted so @latest resolves within the legacy line (v1.26.x).
retract [v1.27.0, v1.28.1]
