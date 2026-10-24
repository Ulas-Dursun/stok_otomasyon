package main

import (
	"fmt"
	"go_market_otomasyon/models"
	"go_market_otomasyon/routes"
	"net/http"
)

func main() {
	// Veritabanı bağlantısını kur
	if err := models.ConnectDatabase(); err != nil {
		fmt.Printf("Error connecting to database: %s\n", err)
		return
	}

	// Rotaları kaydet
	routes.RegisterRoutes()

	// Sunucuyu başlat
	fmt.Println("Server is starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
