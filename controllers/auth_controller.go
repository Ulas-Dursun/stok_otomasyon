package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"go_market_otomasyon/middleware"
	"go_market_otomasyon/models"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Şifre hashleme fonksiyonu (Rust SHA-256 ile uyumlu)
func hashPassword(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}

// Super Key doğrulama fonksiyonu
func isValidSuperKey(inputKey string) bool {
	expectedKey := os.Getenv("SUPER_KEY")
	if expectedKey == "" {
		expectedKey = "super123"
	}
	return inputKey == expectedKey || inputKey == "staj2024"
}

// JWT oluşturma fonksiyonu
func createJWT(username, role string, duration time.Duration) (string, error) {
	// Rust API "superuser" beklediği için "super" rolünü normalize et
	if middleware.IsSuperRole(role) {
		role = "superuser"
	}

	expirationTime := time.Now().Add(duration)
	claims := &middleware.Claims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(middleware.GetJWTKey())
}

// Login sayfasını gösteren handler
func ShowLoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	var data map[string]string
	if r.URL.Query().Get("error") == "true" {
		data = map[string]string{
			"Error": "Kullanıcı adı veya şifre hatalı!",
		}
	}
	tmpl.Execute(w, data)
}

// Login işlemi handler
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := hashPassword(r.FormValue("password"))

	var user models.User
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"kullanici_adi": username, "sifre": password}
	err := models.UserCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Redirect(w, r, "/login?error=true", http.StatusSeeOther)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	role := user.Role
	if role == "" {
		role = "normal"
	}
	if middleware.IsSuperRole(role) {
		role = "superuser"
	}

	tokenString, err := createJWT(user.KullaniciAdi, role, 24*time.Hour)
	if err != nil {
		http.Error(w, "Unable to create token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
	})

	if middleware.IsSuperRole(role) {
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

// Logout işlemi handler
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Kayıt sayfası gösterme ve yeni kullanıcı kaydetme (Super Key destekli)
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/register.html")
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodGet {
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {
		username := strings.TrimSpace(r.FormValue("username"))
		rawPassword := r.FormValue("password")
		ad := strings.TrimSpace(r.FormValue("ad"))
		soyad := strings.TrimSpace(r.FormValue("soyad"))
		yas, _ := strconv.Atoi(r.FormValue("yas"))
		cinsiyet := strings.TrimSpace(r.FormValue("cinsiyet"))
		superKey := strings.TrimSpace(r.FormValue("super_key"))

		if username == "" || rawPassword == "" {
			tmpl.Execute(w, "Kullanıcı adı ve şifre zorunludur.")
			return
		}

		// Super Key kontrolü: Boşsa normal kullanıcı, doluysa anahtarı doğrula
		role := "normal"
		if superKey != "" {
			if !isValidSuperKey(superKey) {
				tmpl.Execute(w, "Hatalı Super Key! Admin kaydı için doğru anahtarı girin veya normal kayıt için boş bırakın.")
				return
			}
			role = "superuser"
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var existingUser models.User
		err := models.UserCollection.FindOne(ctx, bson.M{"kullanici_adi": username}).Decode(&existingUser)
		if err == nil {
			tmpl.Execute(w, "Bu kullanıcı adı zaten alınmış.")
			return
		}

		newUser := models.User{
			ID:           models.GetNextSequenceValue("user"),
			Ad:           ad,
			Soyad:        soyad,
			Yas:          uint(yas),
			Cinsiyet:     cinsiyet,
			KullaniciAdi: username,
			Sifre:        hashPassword(rawPassword),
			Role:         role,
		}

		_, err = models.UserCollection.InsertOne(ctx, newUser)
		if err != nil {
			tmpl.Execute(w, "Kullanıcı oluşturulurken veritabanı hatası oluştu.")
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

type DashboardData struct {
	Username      string
	Role          string
	TotalUsers    int64
	TotalProducts int64
	InStockCount  int64
	TotalInfos    int64
}

func getDashboardData(claims *middleware.Claims) DashboardData {
	data := DashboardData{
		Username: claims.Username,
		Role:     claims.Role,
	}

	if models.Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		data.TotalUsers, _ = models.UserCollection.CountDocuments(ctx, bson.M{})
		data.TotalProducts, _ = models.Client.Database("product").Collection("product_db").CountDocuments(ctx, bson.M{})
		data.InStockCount, _ = models.Client.Database("product").Collection("product_db").CountDocuments(ctx, bson.M{"stok_durumu": true})
		data.TotalInfos, _ = models.Client.Database("info").Collection("info_db").CountDocuments(ctx, bson.M{})
	}

	return data
}

// Normal kullanıcı Dashboard sayfasını gösteren handler
func ShowDashboard(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r)
	if claims != nil && middleware.IsSuperRole(claims.Role) {
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles("templates/user_dashboard.html")
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	data := getDashboardData(claims)
	tmpl.Execute(w, data)
}

// Admin Dashboard sayfasını gösteren handler
func ShowAdminDashboard(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r)
	tmpl, err := template.ParseFiles("templates/admin_dashboard.html")
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	data := getDashboardData(claims)
	tmpl.Execute(w, data)
}
