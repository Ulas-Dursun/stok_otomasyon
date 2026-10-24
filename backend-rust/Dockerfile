# Stage 1: Build
FROM rust:1.83-alpine AS builder
RUN apk add --no-cache musl-dev
WORKDIR /app

# Bağımlılıkları önbellekle
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main() {}" > src/main.rs && cargo build --release && rm -rf src

# Gerçek kaynak kodları kopyala ve derle
COPY src/ ./src/
RUN touch src/main.rs && cargo build --release

# Stage 2: Hafif Çalışma Zamanı
FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/target/release/stok_otomasyon .

EXPOSE 3030
CMD ["./stok_otomasyon"]
