package routes

import (
	"go_market_otomasyon/controllers"
	"go_market_otomasyon/middleware"
	"net/http"
)

func RegisterRoutes() {
	// Statik dosya servisi
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Ana sayfa yönlendirmesi
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	// Auth işlemleri
	http.HandleFunc("/login", controllers.ShowLoginPage)
	http.HandleFunc("/login/submit", controllers.LoginHandler)
	http.HandleFunc("/register", controllers.RegisterHandler)
	http.HandleFunc("/logout", controllers.LogoutHandler)

	// Dashboard rotaları
	http.Handle("/admin/dashboard", middleware.JWTMiddleware(middleware.RoleBasedMiddleware(controllers.ShowAdminDashboard, "superuser")))
	http.Handle("/dashboard", middleware.JWTMiddleware(http.HandlerFunc(controllers.ShowDashboard)))

	// User CRUD rotaları (Sadece Superuser erişebilir)
	http.Handle("/user/crud", middleware.JWTMiddleware(middleware.RoleBasedMiddleware(controllers.ShowUserCrud, "superuser")))
	http.Handle("/user/crud/add", middleware.JWTMiddleware(middleware.RoleBasedMiddleware(controllers.ShowUserCrudAdd, "superuser")))
	http.Handle("/user/crud/delete", middleware.JWTMiddleware(middleware.RoleBasedMiddleware(controllers.ShowUserCrudDelete, "superuser")))
	http.Handle("/user/crud/list", middleware.JWTMiddleware(middleware.RoleBasedMiddleware(controllers.ShowUserCrudList, "superuser")))
	http.Handle("/user/crud/update", middleware.JWTMiddleware(middleware.RoleBasedMiddleware(controllers.ShowUserCrudUpdate, "superuser")))

	// Product CRUD rotaları (Hem yeni hem eski URL şemaları desteklenir)
	http.Handle("/product/crud", middleware.JWTMiddleware(http.HandlerFunc(controllers.ShowProductCrud)))
	http.Handle("/product_crud", middleware.JWTMiddleware(http.HandlerFunc(controllers.ShowProductCrud)))
	http.Handle("/product/crud/add", middleware.JWTMiddleware(middleware.RoleBasedMiddleware(controllers.ShowProductCrudAdd, "superuser")))
	http.Handle("/product/crud/delete", middleware.JWTMiddleware(middleware.RoleBasedMiddleware(controllers.ShowProductCrudDelete, "superuser")))
	http.Handle("/product/crud/list", middleware.JWTMiddleware(http.HandlerFunc(controllers.ShowProductCrudList)))
	http.Handle("/product_crud_list", middleware.JWTMiddleware(http.HandlerFunc(controllers.ShowProductCrudList)))
	http.Handle("/products", middleware.JWTMiddleware(http.HandlerFunc(controllers.ShowProductCrudList)))
	http.Handle("/product/crud/update", middleware.JWTMiddleware(middleware.RoleBasedMiddleware(controllers.ShowProductCrudUpdate, "superuser")))
	http.Handle("/product_update", middleware.JWTMiddleware(middleware.RoleBasedMiddleware(controllers.UpdateProduct, "superuser")))
	http.Handle("/product/order", middleware.JWTMiddleware(http.HandlerFunc(controllers.OrderProduct)))

	// Info CRUD rotaları (Hem yeni hem eski URL şemaları desteklenir)
	http.Handle("/info/crud", middleware.JWTMiddleware(http.HandlerFunc(controllers.ShowInfoCrud)))
	http.Handle("/info_crud", middleware.JWTMiddleware(http.HandlerFunc(controllers.ShowInfoCrud)))
	http.Handle("/info/crud/add", middleware.JWTMiddleware(middleware.RoleBasedMiddleware(controllers.ShowInfoCrudAdd, "superuser")))
	http.Handle("/info/crud/delete", middleware.JWTMiddleware(middleware.RoleBasedMiddleware(controllers.ShowInfoCrudDelete, "superuser")))
	http.Handle("/info/crud/list", middleware.JWTMiddleware(http.HandlerFunc(controllers.ShowInfoCrudList)))
	http.Handle("/info_crud_list", middleware.JWTMiddleware(http.HandlerFunc(controllers.ShowInfoCrudList)))
	http.Handle("/info", middleware.JWTMiddleware(http.HandlerFunc(controllers.ShowInfoCrudList)))
	http.Handle("/info/crud/update", middleware.JWTMiddleware(middleware.RoleBasedMiddleware(controllers.ShowInfoCrudUpdate, "superuser")))
	http.Handle("/info_update", middleware.JWTMiddleware(middleware.RoleBasedMiddleware(controllers.ShowInfoCrudUpdate, "superuser")))

	// Merkezi CRUD form aksiyonu
	http.Handle("/perform_crud", middleware.JWTMiddleware(http.HandlerFunc(controllers.PerformCrudOperation)))
}
