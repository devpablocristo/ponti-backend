
FROM golang:1.23-alpine AS builder

ENV TZ=America/Argentina/Buenos_Aires
RUN apk add --no-cache \
    tzdata \
    sqlite \
    sqlite-dev \
    gcc \
    musl-dev \
    git

WORKDIR /app

COPY . .

WORKDIR /app/projects/ponti-api
RUN go mod download && go mod verify

WORKDIR /app/pkg
RUN go mod download && go mod verify

WORKDIR /app/projects/ponti-api

RUN CGO_ENABLED=1 GOOS=linux go build -o /app/staging_binary ./cmd/api/

# ---------------------------------------------------
# Etapa 2: Imagen final para staging
# ---------------------------------------------------
FROM alpine:latest

ENV TZ=America/Argentina/Buenos_Aires
RUN apk add --no-cache tzdata sqlite

WORKDIR /app

COPY --from=builder /app/pkg  /app/pkg
COPY --from=builder /app/staging_binary /app/staging_binary

EXPOSE 8080     
    
# Comando de inicio de la aplicación
CMD ["/app/staging_binary"]
