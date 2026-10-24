package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go_market_otomasyon/middleware"
	"go_market_otomasyon/models"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func getRustAPIBaseURL() string {
	url := os.Getenv("RUST_API_URL")
	if url == "" {
		return "http://localhost:3030"
	}
	return strings.TrimRight(url, "/")
}

func getTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(cookie.Value), nil
}

// callRustAPI Rust backend servisine JWT token ile HTTP isteği gönderen ortak yardımcı fonksiyon
func callRustAPI(method, endpoint, token string, payload interface{}) ([]byte, int, error) {
	var bodyReader io.Reader
	if payload != nil {
		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, 0, fmt.Errorf("json marshal error: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBytes)
	}

	fullURL := getRustAPIBaseURL() + endpoint
	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("create request error: %w", err)
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("send request error: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response error: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

func renderTemplate(w http.ResponseWriter, tmplPath string, data interface{}) {
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Printf("Unable to load template %s: %v", tmplPath, err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Failed to execute template %s: %v", tmplPath, err)
	}
}

// --- User CRUD Handlers ---

func ShowUserCrud(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "templates/user_crud.html", nil)
}

func ShowUserCrudAdd(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "templates/user_crud_add.html", nil)
}

func ShowUserCrudDelete(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "templates/user_crud_delete.html", nil)
}

func ShowUserCrudUpdate(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "templates/user_crud_update.html", nil)
}

func ShowUserCrudList(w http.ResponseWriter, r *http.Request) {
	token, err := getTokenFromCookie(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	body, status, err := callRustAPI(http.MethodGet, "/list_users", token, nil)
	if err != nil || status != http.StatusOK {
		http.Error(w, "Failed to fetch users from Rust API", http.StatusInternalServerError)
		return
	}

	var users []models.User
	if err := json.Unmarshal(body, &users); err != nil {
		http.Error(w, "Failed to parse users response from Rust API", http.StatusInternalServerError)
		return
	}

	renderTemplate(w, "templates/user_crud_list.html", users)
}

// --- Product CRUD Handlers ---

func ShowProductCrud(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "templates/product_crud.html", nil)
}

func ShowProductCrudAdd(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "templates/product_crud_add.html", nil)
}

func ShowProductCrudDelete(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "templates/product_crud_delete.html", nil)
}

func ShowProductCrudUpdate(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "templates/product_crud_update.html", nil)
}

func ShowProductCrudList(w http.ResponseWriter, r *http.Request) {
	token, err := getTokenFromCookie(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	body, status, err := callRustAPI(http.MethodGet, "/list_products", token, nil)
	if err != nil || status != http.StatusOK {
		http.Error(w, "Failed to fetch products from Rust API", http.StatusInternalServerError)
		return
	}

	var products []models.Product
	if err := json.Unmarshal(body, &products); err != nil {
		http.Error(w, "Failed to parse products response from Rust API", http.StatusInternalServerError)
		return
	}

	renderTemplate(w, "templates/product_crud_list.html", products)
}

func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ShowProductCrudUpdate(w, r)
		return
	}

	token, err := getTokenFromCookie(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var product models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	_, status, err := callRustAPI(http.MethodPut, "/update_product/"+product.ID, token, product)
	if err != nil || status != http.StatusOK {
		http.Error(w, "Failed to update product", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/product/crud/list", http.StatusSeeOther)
}

type rustOrderResponse struct {
	Success        bool    `json:"success"`
	Message        string  `json:"message"`
	ProductName    *string `json:"product_name"`
	RemainingStock *uint   `json:"remaining_stock"`
}

// OrderProduct kullanıcının katalogdan 1 adet sipariş vermesini, stoktan düşülmesini ve sipariş geçmişine kaydedilmesini sağlar
func OrderProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/product/crud/list", http.StatusSeeOther)
		return
	}

	token, err := getTokenFromCookie(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	productID := strings.TrimSpace(r.FormValue("id"))
	if productID == "" {
		http.Redirect(w, r, "/product/crud/list", http.StatusSeeOther)
		return
	}

	body, status, err := callRustAPI(http.MethodPost, "/order_product/"+productID, token, nil)
	if err != nil || status != http.StatusOK {
		http.Redirect(w, r, "/product/crud/list?error=siparis_hatasi", http.StatusSeeOther)
		return
	}

	var resp rustOrderResponse
	if err := json.Unmarshal(body, &resp); err == nil && resp.Success {
		// Başarılı siparişi MongoDB orders koleksiyonuna kaydet
		if claims, ok := middleware.GetClaims(r); ok && models.OrderCollection != nil {
			orderID := models.GetNextSequenceValue("order")
			prodName := "Ürün #" + productID
			if resp.ProductName != nil && *resp.ProductName != "" {
				prodName = *resp.ProductName
			}
			newOrder := models.Order{
				ID:          orderID,
				Username:    claims.Username,
				ProductID:   productID,
				ProductName: prodName,
				Quantity:    1,
				CreatedAt:   time.Now(),
				Status:      "Tamamlandı",
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			_, _ = models.OrderCollection.InsertOne(ctx, newOrder)
			cancel()
		}
		http.Redirect(w, r, "/product/crud/list?success=siparis_verildi", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/product/crud/list?error=siparis_hatasi", http.StatusSeeOther)
}

// RestockProduct admin kullanıcısının ürüne +10 adet stok eklemesini sağlar
func RestockProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/product/crud/list", http.StatusSeeOther)
		return
	}

	token, err := getTokenFromCookie(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	productID := strings.TrimSpace(r.FormValue("id"))
	if productID == "" {
		http.Redirect(w, r, "/product/crud/list", http.StatusSeeOther)
		return
	}

	_, status, err := callRustAPI(http.MethodPost, "/restock_product/"+productID, token, nil)
	if err != nil || status != http.StatusOK {
		http.Redirect(w, r, "/product/crud/list?error=stok_artirma_hatasi", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/product/crud/list?success=stok_artirildi", http.StatusSeeOther)
}

// --- Info CRUD Handlers ---

func ShowInfoCrud(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "templates/info_crud.html", nil)
}

func ShowInfoCrudAdd(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "templates/info_crud_add.html", nil)
}

func ShowInfoCrudDelete(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "templates/info_crud_delete.html", nil)
}

func ShowInfoCrudUpdate(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "templates/info_crud_update.html", map[string]string{
		"CustomID":      "",
		"Adres":         "",
		"Telefon":       "",
		"CalisanSayisi": "",
	})
}

func ShowInfoCrudList(w http.ResponseWriter, r *http.Request) {
	token, err := getTokenFromCookie(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	body, status, err := callRustAPI(http.MethodGet, "/list_info", token, nil)
	if err != nil || status != http.StatusOK {
		log.Printf("Failed to fetch info from Rust API (status %d): %v", status, err)
		http.Error(w, "Failed to fetch info from Rust API", http.StatusInternalServerError)
		return
	}

	var infos []models.Info
	if err := json.Unmarshal(body, &infos); err != nil {
		log.Printf("Failed to parse info response from Rust API: %v", err)
		http.Error(w, "Failed to parse response from Rust API", http.StatusInternalServerError)
		return
	}

	renderTemplate(w, "templates/info_crud_list.html", infos)
}

// --- Central CRUD Dispatcher ---

func PerformCrudOperation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	entity := r.FormValue("entity")
	operation := r.FormValue("operation")

	claims, ok := middleware.GetClaims(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if entity == "user" && !middleware.IsSuperRole(claims.Role) {
		http.Error(w, "Unauthorized: Superuser yetkisi gerekli", http.StatusForbidden)
		return
	}

	if (entity == "product" || entity == "info") && (operation == "create" || operation == "delete" || operation == "update") && !middleware.IsSuperRole(claims.Role) {
		http.Error(w, "Unauthorized: Superuser yetkisi gerekli", http.StatusForbidden)
		return
	}

	token, err := getTokenFromCookie(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch entity {
	case "product":
		switch operation {
		case "create":
			miktar, _ := strconv.Atoi(r.FormValue("miktar"))
			product := models.Product{
				Adi:         r.FormValue("adi"),
				Miktar:      uint(miktar),
				StokDurumu:  r.FormValue("stok_durumu") == "true",
				SatisDurumu: r.FormValue("satis_durumu") == "true",
			}
			if _, status, err := callRustAPI(http.MethodPost, "/add_product", token, product); err != nil || status != http.StatusOK {
				http.Error(w, "Failed to add product", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/product/crud/list", http.StatusSeeOther)

		case "delete":
			productID := strings.TrimSpace(r.FormValue("id"))
			if _, status, err := callRustAPI(http.MethodDelete, "/delete_product/"+productID, token, nil); err != nil || status != http.StatusOK {
				http.Error(w, "Failed to delete product", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/product/crud/list", http.StatusSeeOther)

		case "update":
			productID := strings.TrimSpace(r.FormValue("id"))
			miktar, _ := strconv.Atoi(r.FormValue("miktar"))
			product := models.Product{
				ID:          productID,
				Adi:         r.FormValue("adi"),
				Miktar:      uint(miktar),
				StokDurumu:  r.FormValue("stok_durumu") == "true",
				SatisDurumu: r.FormValue("satis_durumu") == "true",
			}
			if _, status, err := callRustAPI(http.MethodPut, "/update_product/"+productID, token, product); err != nil || status != http.StatusOK {
				http.Error(w, "Failed to update product", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/product/crud/list", http.StatusSeeOther)
		}

	case "info":
		switch operation {
		case "create":
			calisanSayisi, _ := strconv.Atoi(r.FormValue("calisan_sayisi"))
			info := models.Info{
				Adres:         r.FormValue("adres"),
				Telefon:       r.FormValue("telefon"),
				CalisanSayisi: uint(calisanSayisi),
			}
			if _, status, err := callRustAPI(http.MethodPost, "/add_info", token, info); err != nil || status != http.StatusOK {
				http.Error(w, "Failed to add info", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/info/crud/list", http.StatusSeeOther)

		case "delete":
			infoID := strings.TrimSpace(r.FormValue("id"))
			if _, status, err := callRustAPI(http.MethodDelete, "/delete_info/"+infoID, token, nil); err != nil || status != http.StatusOK {
				http.Error(w, "Failed to delete info", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/info/crud/list", http.StatusSeeOther)

		case "update":
			infoID := strings.TrimSpace(r.FormValue("id"))
			calisanSayisi, _ := strconv.Atoi(r.FormValue("calisan_sayisi"))
			info := models.Info{
				ID:            infoID,
				Adres:         r.FormValue("adres"),
				Telefon:       r.FormValue("telefon"),
				CalisanSayisi: uint(calisanSayisi),
			}
			if _, status, err := callRustAPI(http.MethodPut, "/update_info/"+infoID, token, info); err != nil || status != http.StatusOK {
				http.Error(w, "Failed to update info", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/info/crud/list", http.StatusSeeOther)
		}

	case "user":
		switch operation {
		case "create":
			yas, _ := strconv.Atoi(r.FormValue("yas"))
			role := r.FormValue("role")
			if role == "" {
				role = "normal"
			}
			if isValidSuperKey(strings.TrimSpace(r.FormValue("super_key"))) {
				role = "superuser"
			}
			user := models.User{
				Ad:           r.FormValue("ad"),
				Soyad:        r.FormValue("soyad"),
				Yas:          uint(yas),
				Cinsiyet:     r.FormValue("cinsiyet"),
				KullaniciAdi: r.FormValue("kullanici_adi"),
				Sifre:        r.FormValue("sifre"),
				Role:         role,
			}
			if _, status, err := callRustAPI(http.MethodPost, "/add_user", token, user); err != nil || status != http.StatusOK {
				http.Error(w, "Failed to add user", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/user/crud/list", http.StatusSeeOther)

		case "delete":
			userID := strings.TrimSpace(r.FormValue("id"))
			if _, status, err := callRustAPI(http.MethodDelete, "/delete_user/"+userID, token, nil); err != nil || status != http.StatusOK {
				http.Error(w, "Failed to delete user", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/user/crud/list", http.StatusSeeOther)

		case "update":
			userID := strings.TrimSpace(r.FormValue("id"))
			yas, _ := strconv.Atoi(r.FormValue("yas"))
			user := models.User{
				ID:           userID,
				Ad:           r.FormValue("ad"),
				Soyad:        r.FormValue("soyad"),
				Yas:          uint(yas),
				Cinsiyet:     r.FormValue("cinsiyet"),
				KullaniciAdi: r.FormValue("kullanici_adi"),
				Sifre:        r.FormValue("sifre"),
				Role:         r.FormValue("role"),
			}
			if _, status, err := callRustAPI(http.MethodPut, "/update_user/"+userID, token, user); err != nil || status != http.StatusOK {
				http.Error(w, "Failed to update user", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/user/crud/list", http.StatusSeeOther)
		}
	}
}
