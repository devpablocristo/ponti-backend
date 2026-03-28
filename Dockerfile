FROM golang:1.26.1-alpine AS builder

ARG GO_MODULES_TOKEN

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
RUN if [ -n "$GO_MODULES_TOKEN" ]; then \
      git config --global url."https://${GO_MODULES_TOKEN}@github.com/".insteadOf "https://github.com/"; \
    fi

WORKDIR /app

COPY . .

WORKDIR /app
RUN go mod download && go mod verify


WORKDIR /app/pkg
RUN go mod download && go mod verify


WORKDIR /app


RUN CGO_ENABLED=1 GOOS=linux go build -o /app/prod_binary ./cmd/api/
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/migrate_binary ./cmd/migrate/


# ---------------------------------------------------
# Etapa 2: Imagen final para prod
# ---------------------------------------------------
FROM alpine:latest


ENV TZ=America/Argentina/Buenos_Aires


WORKDIR /app


COPY --from=builder /app/pkg /app/pkg
COPY --from=builder /app/prod_binary /app/prod_binary
COPY --from=builder /app/migrate_binary /app/migrate_binary
COPY --from=builder /app/migrations_v4 /app/migrations_v4
COPY --from=builder /app/scripts/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh


EXPOSE 8080


ENTRYPOINT ["/app/entrypoint.sh"]




