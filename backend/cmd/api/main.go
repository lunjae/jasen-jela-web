package main

import (
	"context"
	"errors"
	firebase "firebase.google.com/go/v4"
	"github.com/lunjae/jasen-jela-web/backend/internal/config"
	"github.com/lunjae/jasen-jela-web/backend/internal/handlers"
	"github.com/lunjae/jasen-jela-web/backend/internal/middleware"
	"github.com/lunjae/jasen-jela-web/backend/internal/repository"
	"github.com/lunjae/jasen-jela-web/backend/internal/services"
	"google.golang.org/api/option"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, e := config.Load()
	if e != nil {
		slog.Error("invalid configuration", "error", e)
		os.Exit(1)
	}
	ctx := context.Background()
	var store repository.Store
	var verifier middleware.TokenVerifier = services.RejectVerifier{}
	var closeFn func() error = func() error { return nil }
	uploader := services.CloudinaryUploader{CloudName: cfg.CloudinaryCloudName, APIKey: cfg.CloudinaryAPIKey, APISecret: cfg.CloudinaryAPISecret}
	if cfg.UseMemoryStore {
		store = repository.NewMemory()
		slog.Warn("using in-memory development store")
	} else {
		var options []option.ClientOption
		if credentials := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); credentials != "" {
			options = append(options, option.WithCredentialsFile(credentials))
		}
		app, e := firebase.NewApp(ctx, &firebase.Config{ProjectID: cfg.FirebaseProjectID}, options...)
		if e != nil {
			fatal(e)
		}
		fc, e := app.Firestore(ctx)
		if e != nil {
			fatal(e)
		}
		auth, e := app.Auth(ctx)
		if e != nil {
			fatal(e)
		}
		store = repository.NewFirestore(fc)
		verifier = services.FirebaseVerifier{Client: auth}
		closeFn = fc.Close
	}
	defer closeFn()
	h := handlers.New(store, uploader)
	mux := routes(h, verifier, store, cfg.UseMemoryStore)
	server := &http.Server{Addr: ":" + cfg.Port, Handler: middleware.Chain(mux, middleware.RequestLog, middleware.Recover, middleware.CORS(cfg.FrontendOrigin)), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: cfg.ReadTimeout, WriteTimeout: cfg.WriteTimeout, IdleTimeout: cfg.IdleTimeout}
	go func() {
		slog.Info("server listening", "port", cfg.Port)
		if e := server.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
			fatal(e)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if e := server.Shutdown(shutdown); e != nil {
		slog.Error("shutdown failed", "error", e)
	}
}
func routes(h *handlers.Handler, v middleware.TokenVerifier, s repository.Store, dev bool) *http.ServeMux {
	m := http.NewServeMux()
	m.HandleFunc("GET /api/health", h.Health)
	m.HandleFunc("GET /api/products", h.Products)
	m.HandleFunc("GET /api/products/{slug}", h.Product)
	m.HandleFunc("GET /api/categories", h.Categories)
	m.HandleFunc("POST /api/inquiries", h.CreateInquiry)
	admin := http.NewServeMux()
	admin.HandleFunc("GET /api/admin/products", h.AdminProducts)
	admin.HandleFunc("POST /api/admin/products", h.SaveProduct)
	admin.HandleFunc("GET /api/admin/products/{id}", h.AdminProduct)
	admin.HandleFunc("PUT /api/admin/products/{id}", h.SaveProduct)
	admin.HandleFunc("DELETE /api/admin/products/{id}", h.DeleteProduct)
	admin.HandleFunc("POST /api/admin/products/{id}/images", h.UploadImage)
	admin.HandleFunc("DELETE /api/admin/products/{id}/images", h.DeleteImage)
	admin.HandleFunc("PATCH /api/admin/products/{id}/images/primary", h.SetPrimaryImage)
	admin.HandleFunc("PATCH /api/admin/products/{id}/images/order", h.ReorderImages)
	admin.HandleFunc("GET /api/admin/categories", h.AdminCategories)
	admin.HandleFunc("POST /api/admin/categories", h.SaveCategory)
	admin.HandleFunc("PUT /api/admin/categories/{id}", h.SaveCategory)
	admin.HandleFunc("DELETE /api/admin/categories/{id}", h.DeleteCategory)
	admin.HandleFunc("GET /api/admin/inquiries", h.Inquiries)
	admin.HandleFunc("GET /api/admin/inquiries/{id}", h.Inquiry)
	admin.HandleFunc("PATCH /api/admin/inquiries/{id}/status", h.InquiryStatus)
	m.Handle("/api/admin/", middleware.Chain(admin, middleware.Authenticate(v, dev), middleware.RequireAdmin(s)))
	return m
}
func fatal(e error) { slog.Error("fatal error", "error", e); os.Exit(1) }
