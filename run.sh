#!/bin/bash

GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}==================================================${NC}"
echo -e "${BLUE}   🛒 Stok & Market Otomasyonu (Go + Rust)        ${NC}"
echo -e "${BLUE}==================================================${NC}"

# Rust Backend Başlat
echo -e "${GREEN}[1/2] 🦀 Rust REST API (:3030) başlatılıyor...${NC}"
(cd backend-rust && cargo run) &
RUST_PID=$!

# Script kapandığında (Ctrl+C yapıldığında) arkadaki Rust sürecini de durdur
cleanup() {
    echo -e "\n${RED}Servisler durduruluyor...${NC}"
    kill $RUST_PID 2>/dev/null
    exit 0
}
trap cleanup SIGINT SIGTERM EXIT

# Rust'ın ayağa kalkması için kısa bir bekleme
sleep 2

# Go Frontend Başlat
echo -e "${GREEN}[2/2] 🐹 Go Web Sunucusu (:8080) başlatılıyor...${NC}"
echo -e "${BLUE}👉 Tarayıcınızdan açın: http://localhost:8080${NC}"
echo -e "${BLUE}==================================================${NC}"
go run main.go
