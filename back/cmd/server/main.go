package main

import (
	"log"

	httpapi "sangehassan/back/internal/adapters/http"
	"sangehassan/back/internal/adapters/persistence/postgres"
	"sangehassan/back/internal/config"
	"sangehassan/back/internal/usecase"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	db, err := postgres.NewDB(cfg)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer db.Close()

	categoryRepo := postgres.NewCategoryRepository(db)
	productRepo := postgres.NewProductRepository(db)
	blogRepo := postgres.NewBlogRepository(db)
	adminRepo := postgres.NewAdminRepository(db)

	categoryService := usecase.NewCategoryService(categoryRepo)
	productService := usecase.NewProductService(productRepo)
	blogService := usecase.NewBlogService(blogRepo)
	authService := usecase.NewAuthService(adminRepo, cfg.JWTSecret, cfg.JWTTTLHours)

	router := httpapi.NewRouter(cfg, categoryService, productService, blogService, authService)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
