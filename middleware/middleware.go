package middleware

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type contextKey string

const ClaimsContextKey contextKey = "claims"

// Claims yapısı, JWT'nin içinde saklanacak olan kullanıcı bilgilerini temsil eder
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GetJWTKey ortam değişkeninden veya varsayılan değerden JWT gizli anahtarını döndürür
func GetJWTKey() []byte {
	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		secret = "your_secret_key"
	}
	return []byte(secret)
}

// IsSuperRole rolün yönetici olup olmadığını kontrol eder ("superuser" ve "super" uyumluluğu)
func IsSuperRole(role string) bool {
	return role == "superuser" || role == "super"
}

// GetClaims context içerisinden kullanıcı Claims bilgisini okur
func GetClaims(r *http.Request) (*Claims, bool) {
	claims, ok := r.Context().Value(ClaimsContextKey).(*Claims)
	return claims, ok
}

// JWTMiddleware, JWT doğrulaması yapan middleware
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		tokenStr := cookie.Value
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return GetJWTKey(), nil
		})

		if err != nil || !token.Valid || (claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now())) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RoleBasedMiddleware, kullanıcı rollerini kontrol eden middleware
func RoleBasedMiddleware(next http.HandlerFunc, roles ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetClaims(r)
		if !ok {
			http.Error(w, "Unauthorized: Role not found", http.StatusForbidden)
			return
		}

		if IsSuperRole(claims.Role) {
			next.ServeHTTP(w, r)
			return
		}

		for _, role := range roles {
			if claims.Role == role {
				next.ServeHTTP(w, r)
				return
			}
		}

		http.Error(w, "Forbidden: Bu işlem için Superuser (Admin) yetkisi gereklidir.", http.StatusForbidden)
	}
}
