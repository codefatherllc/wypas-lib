module github.com/codefatherllc/wypas-lib

go 1.22.3

require (
	github.com/codefatherllc/wypas-proto v0.0.0
	github.com/go-sql-driver/mysql v1.8.1
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/google/flatbuffers v25.2.10+incompatible
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	golang.org/x/image v0.18.0
)

replace github.com/codefatherllc/wypas-proto => ../wypas-proto
