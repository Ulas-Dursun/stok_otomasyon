.PHONY: all run rust go clean

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

# Derleme artıklarını temizler
clean:
	cd backend-rust && cargo clean
	rm -f main
