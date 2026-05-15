# syntax=docker/dockerfile:1.7
FROM golang:1.26.3-alpine AS builder

ENV TZ=America/Argentina/Buenos_Aires
RUN apk add --no-cache \
   tzdata \
   sqlite \
   sqlite-dev \
   gcc \
   musl-dev \
   git

# Configurar acceso a repos privados de Go
ENV GOPRIVATE=github.com/devpablocristo/*

WORKDIR /app

COPY . .

WORKDIR /app
RUN --mount=type=secret,id=go_modules_token,required=false \
    token="$(cat /run/secrets/go_modules_token 2>/dev/null || true)" && \
    if [ -n "$token" ]; then \
      git config --global url."https://${token}@github.com/".insteadOf "https://github.com/"; \
    fi && \
    git config --global http.version HTTP/1.1 && \
    core_modules="github.com/devpablocristo/core/authn/go github.com/devpablocristo/core/databases/postgres/go github.com/devpablocristo/core/errors/go github.com/devpablocristo/core/governance/go github.com/devpablocristo/core/http/gin/go github.com/devpablocristo/core/http/go github.com/devpablocristo/core/notifications/go github.com/devpablocristo/core/security/go github.com/devpablocristo/core/validate/go" && \
    for module in $core_modules; do \
      go mod download "$module"; \
    done && \
    i=0 && \
    until go mod download; do \
      i=$((i+1)); \
      if [ "$i" -ge 3 ]; then \
        exit 1; \
      fi; \
      echo "go mod download failed, retry $i/3" >&2; \
      rm -rf "$(go env GOMODCACHE)/cache/vcs"; \
      for module in $core_modules; do \
        go mod download "$module"; \
      done; \
      sleep 2; \
    done && \
    go mod verify && \
    rm -f /root/.gitconfig

RUN CGO_ENABLED=1 GOOS=linux go build -o /app/prod_binary ./cmd/api/
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/migrate_binary ./cmd/migrate/


# ---------------------------------------------------
# Etapa 2: Imagen final para prod
# ---------------------------------------------------
FROM alpine:latest


ENV TZ=America/Argentina/Buenos_Aires


WORKDIR /app

COPY --from=builder /app/prod_binary /app/prod_binary
COPY --from=builder /app/migrate_binary /app/migrate_binary
COPY --from=builder /app/migrations_v4 /app/migrations_v4
COPY --from=builder /app/scripts/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh


EXPOSE 8080


ENTRYPOINT ["/app/entrypoint.sh"]


