package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"server/internal/config"
	"server/internal/handlers"
	"server/internal/lib"
	"server/internal/middleware"
	"server/internal/repositories"
	"server/internal/routes"
	"server/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// load env config
	cfg := config.Load()

	// init libraries
	db, cacheClient, minioClient, mailer, oauth, ai := lib.Init(cfg)

	if cfg.App.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// setup Gin engine and middleware
	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())
	engine.Use(middleware.CORS(cfg))
	engine.Use(middleware.RateLimiter(cfg, cacheClient, cfg.App.GlobalRateLimit))
	engine.Use(middleware.ErrorHandler())


	// inject dependencies and initialize routes
	repo:= repositories.InitRepository(db)
	services := services.InitServices(repo, cfg, cacheClient, mailer, oauth, minioClient, db, ai)
	handlers := handlers.InitHandlers(services, cfg)

	routes.InitRoutes(engine, cfg, handlers)
	
	// server configuration
	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// start server in a separate goroutine
	go func() {
		log.Printf("✅ Server running on %s (mode: %s)", cfg.App.ServerURL, cfg.App.Mode)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("❌ Server error: %v", err)
		}
	}()

		
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⏳ Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Forced shutdown: %v", err)
	}

	log.Println("✅ Server exited cleanly")
}