module github.com/cvele/authentication-service

go 1.19

require (
	github.com/cockroachdb/cockroach-go/v2 v2.2.20
	github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lib/pq v1.10.6
	github.com/pressly/goose/v3 v3.9.0
	github.com/rs/zerolog v1.29.0
	golang.org/x/crypto v0.6.0
)

require golang.org/x/tools v0.6.0 // indirect

require (
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	golang.org/x/sys v0.5.0 // indirect; indirgoect
)
