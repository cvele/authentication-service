# Build stage
FROM golang:1.19-alpine AS build
ENV GOARCH=arm
RUN apk add --no-cache git

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o /app/authentication-service cmd/auth/main.go
RUN go build -o /app/authentication-migrations cmd/migrations/main.go
RUN chmod +x /app/authentication-migrations /app/authentication-service
# Final stage
FROM alpine:3.14
RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /app/authentication-service /usr/local/bin/authentication-service
COPY --from=build /app/authentication-migrations /usr/local/bin/authentication-migrations
COPY --from=build /app/migrations migrations/.
ENV MIGRATIONS_DIR /app/migrations
ENV GOOSE_CUSTOM_BINARY /usr/local/bin/authentication-migrations
CMD /usr/local/bin/authentication-migrations && /usr/local/bin/authentication-service
