# Stage 1: Build
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Bağımlılıkları önbellekle
COPY go.mod go.sum ./
RUN go mod download

# Kaynak kodları kopyala ve derle
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server .

# Stage 2: Hafif Çalışma Zamanı
FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/server .
COPY templates/ ./templates/
COPY static/ ./static/

EXPOSE 8080
CMD ["./server"]
