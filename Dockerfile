FROM golang:1.24-bookworm AS builder

RUN apt-get update && apt-get install -y \
    gcc \
    libsqlite3-dev \
    libsqlcipher-dev \
    make \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY . .

RUN swag fmt && swag init -g cmd/api/main.go -o docs

ARG ENABLE_DB_ENCRYPTION=false
RUN mkdir -p bin && \
    if [ "$ENABLE_DB_ENCRYPTION" = "true" ]; then \
        echo "Building with SQLCipher encryption"; \
        CGO_ENABLED=1 go build -tags sqlcipher -o bin/api cmd/api/main.go; \
        CGO_ENABLED=1 go build -tags sqlcipher -o bin/migrate cmd/migrate/main.go; \
        CGO_ENABLED=1 go build -tags sqlcipher -o bin/qr-worker cmd/qr-worker/main.go; \
        CGO_ENABLED=1 go build -tags sqlcipher -o bin/worker cmd/worker/main.go; \
    else \
        echo "Building with standard SQLite (unencrypted)"; \
        CGO_ENABLED=1 go build -o bin/api cmd/api/main.go; \
        CGO_ENABLED=1 go build -o bin/migrate cmd/migrate/main.go; \
        CGO_ENABLED=1 go build -o bin/qr-worker cmd/qr-worker/main.go; \
        CGO_ENABLED=1 go build -o bin/worker cmd/worker/main.go; \
    fi

FROM golang:1.24-bookworm

RUN apt-get update && apt-get install -y \
    ca-certificates \
    libsqlite3-0 \
    libsqlcipher0 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/bin/api /app/api
COPY --from=builder /app/bin/migrate /app/migrate
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/default.env /app/default.env

RUN mkdir -p /app/data

EXPOSE 8080

CMD ["./api"]


