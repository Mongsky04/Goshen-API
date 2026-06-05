package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"goshen/backend/internal/config"
	"goshen/backend/internal/db"
	"goshen/backend/internal/handler"
	"goshen/backend/internal/middleware"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg := config.Load()

	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("db connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	h := handler.New(pool, cfg)
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CORS(middleware.ParseOrigins(cfg.FrontendOrigin)))

	// Serve locally uploaded files
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.UploadDir))))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`)) //nolint:errcheck
	})

	// Admin auth
	r.With(httprate.LimitByIP(5, time.Minute)).Post("/admin/login", h.Login)

	// Nav publish status (used by frontend to filter navbar links)
	r.Get("/api/v1/nav", h.GetNavStatus)

	// Public page data
	r.Get("/api/v1/conference-pages/{slug}", h.GetConferencePage)
	r.Get("/api/v1/performer-pages/{slug}", h.GetPerformerPage)

	// Public read endpoints (consumed by frontend)
	r.Get("/api/v1/products", h.ListProducts)
	r.Get("/api/v1/products/{id}", h.GetProduct)
	r.Get("/api/v1/featured", h.ListFeatured)
	r.Get("/api/v1/featured/{id}", h.GetFeatured)
	r.Get("/api/v1/articles", h.ListArticles)
	r.Get("/api/v1/articles/{id}", h.GetArticle)
	r.Get("/api/v1/slider", h.ListSlider)
	r.Get("/api/v1/slider/{id}", h.GetSlider)
	r.Get("/api/v1/brands", h.ListBrands)
	r.Get("/api/v1/brands/{id}", h.GetBrand)
	r.Get("/api/v1/customers", h.ListCustomers)
	r.Get("/api/v1/customers/{id}", h.GetCustomer)
	r.Get("/api/v1/support-cards", h.ListSupportCards)
	r.Get("/api/v1/homepage-grid", h.ListHomepageGrid)
	r.Get("/api/v1/banners", h.ListBanners)
	r.Get("/api/v1/media", h.ListMedia)

	// Protected CMS write routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(cfg.JWTSecret))

		r.Get("/admin/me", h.Me)

		r.Post("/api/v1/products", h.CreateProduct)
		r.Put("/api/v1/products/{id}", h.UpdateProduct)
		r.Delete("/api/v1/products/{id}", h.DeleteProduct)

		r.Post("/api/v1/featured", h.CreateFeatured)
		r.Put("/api/v1/featured/{id}", h.UpdateFeatured)
		r.Delete("/api/v1/featured/{id}", h.DeleteFeatured)

		r.Post("/api/v1/articles", h.CreateArticle)
		r.Put("/api/v1/articles/{id}", h.UpdateArticle)
		r.Delete("/api/v1/articles/{id}", h.DeleteArticle)

		r.Post("/api/v1/slider", h.CreateSlider)
		r.Put("/api/v1/slider/{id}", h.UpdateSlider)
		r.Delete("/api/v1/slider/{id}", h.DeleteSlider)

		r.Post("/api/v1/brands", h.CreateBrand)
		r.Put("/api/v1/brands/{id}", h.UpdateBrand)
		r.Delete("/api/v1/brands/{id}", h.DeleteBrand)

		r.Post("/api/v1/customers", h.CreateCustomer)
		r.Put("/api/v1/customers/{id}", h.UpdateCustomer)
		r.Delete("/api/v1/customers/{id}", h.DeleteCustomer)

		// File upload
		r.Post("/api/v1/upload", h.UploadImage)

		// Media library
		r.Post("/api/v1/media", h.UploadMedia)
		r.Delete("/api/v1/media/{id}", h.DeleteMedia)

		// Banners catalog
		r.Post("/api/v1/banners", h.CreateBanner)
		r.Delete("/api/v1/banners/{id}", h.DeleteBanner)

		// Support cards
		r.Post("/api/v1/support-cards", h.CreateSupportCard)
		r.Delete("/api/v1/support-cards/{id}", h.DeleteSupportCard)

		// Homepage grid (replaces product_ids list)
		r.Put("/api/v1/homepage-grid", h.ReplaceHomepageGrid)

		// Conference CMS
		r.Get("/api/v1/admin/conference-pages/{slug}", h.GetAdminConferencePage)
		r.Put("/api/v1/admin/conference-pages/{slug}", h.PutAdminConferencePage)

		// Performer CMS
		r.Get("/api/v1/admin/performer-pages/{slug}", h.GetAdminPerformerPage)
		r.Put("/api/v1/admin/performer-pages/{slug}", h.PutAdminPerformerPage)
	})

	slog.Info("server started", "port", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
