# 🐹 Market & Stok Otomasyonu - Go Web Frontend (BFF)

Bu proje, **Go** dili ile geliştirilmiş; kullanıcı kimlik doğrulamasını (JWT/RBAC), Server-Side Rendering (SSR) arayüz yönetimini yürüten ve arka plandaki **Rust REST API (`:3030`)** servisine bağlanan **BFF (Backend-for-Frontend)** web uygulamasıdır.

---

## 🚀 Hızlı Başlangıç

### 1. Gereksinimler
* Go 1.20+
* MongoDB (`localhost:27017`)
* Rust Backend API (`localhost:3030`)

### 2. Ortam Değişkenleri
Dizinde `.env` dosyası hazır bulunmaktadır:
```env
JWT_SECRET_KEY=your_secret_key
SUPER_KEY=super123
MONGODB_URI=mongodb://localhost:27017
RUST_API_URL=http://localhost:3030
PORT=8080
```

### 3. Çalıştırma
```bash
go run main.go
```
*Sunucu `http://localhost:8080` adresinde başlayacaktır.*

---

## 🔑 Özellikler
* **Rol Tabanlı Yetkilendirme (RBAC):** `superuser` (Yönetici) ve `normal` kullanıcı ayrımı.
* **Super Key ile Admin Kaydı:** Kayıt ekranında (`/register`) gizli anahtarı (`super123`) giren kullanıcılar otomatik olarak Superuser (Admin) yetkisine sahip olur.
* **Rust API Entegrasyonu:** Ürün, kullanıcı ve şube CRUD işlemleri `http://localhost:3030` adresindeki Warp REST servisi üzerinden JWT Bearer token ile gerçekleştirilir.
* **Modern UI:** Responsive kartlar, modern tablolar, renkli rozetler ve temiz form tasarımları.
