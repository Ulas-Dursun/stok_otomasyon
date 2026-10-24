.PHONY: all run rust go test clean docker-up docker-down

all: run

# Hem Rust hem Go'yu aynı anda çalıştırır
run:
	@./run.sh

# Sadece Rust servisini çalıştırır (:3030)
rust:
	cd backend-rust && cargo run

# Sadece Go frontend servisini çalıştırır (:8080)
go:
	go run main.go

# Tüm birim testleri (Go ve Rust) çalıştırır
test:
	@echo "🧪 Running Rust unit tests..."
	@cd backend-rust && cargo test --quiet
	@echo "🧪 Running Go unit tests..."
	@go test -v ./...

# Docker ile tüm servisleri ayağa kaldırır
docker-up:
	docker compose up --build -d

# Docker konteynerlerini durdurur
docker-down:
	docker compose down

# Derleme artıklarını temizler
clean:
	cd backend-rust && cargo clean
	rm -f main server
