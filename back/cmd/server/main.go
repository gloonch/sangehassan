package main

import (
	"log"

	httpapi "sangehassan/back/internal/adapters/http"
	"sangehassan/back/internal/adapters/http/handlers"
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
	productTermRepo := postgres.NewProductTermRepository(db)
	blogRepo := postgres.NewBlogRepository(db)
	projectRepo := postgres.NewProjectRepository(db)
	templateRepo := postgres.NewTemplateRepository(db)
	blockRepo := postgres.NewBlockRepository(db)
	contentSectionRepo := postgres.NewContentSectionRepository(db)
	adminRepo := postgres.NewAdminRepository(db)
	userRepo := postgres.NewUserRepository(db)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(db)
	dashboardRepo := postgres.NewDashboardRepository(db)
	listingRepo := postgres.NewListingRepository(db)
	dealRequestRepo := postgres.NewDealRequestRepository(db)

	categoryService := usecase.NewCategoryService(categoryRepo)
	productService := usecase.NewProductService(productRepo)
	productTermService := usecase.NewProductTermService(productTermRepo)
	blogService := usecase.NewBlogService(blogRepo)
	projectService := usecase.NewProjectService(projectRepo)
	templateService := usecase.NewTemplateService(templateRepo)
	blockService := usecase.NewBlockService(blockRepo)
	contentSectionService := usecase.NewContentSectionService(contentSectionRepo)
	authService := usecase.NewAuthService(adminRepo, cfg.JWTSecret, cfg.JWTTTLHours)
	userAuthService := usecase.NewUserAuthService(userRepo, refreshTokenRepo, dealRequestRepo, cfg.JWTSecret, cfg.AccessTokenMinutes, cfg.RefreshTokenDays)
	dashboardService := usecase.NewDashboardService(dashboardRepo)
	listingService := usecase.NewListingService(listingRepo)
	dealRequestService := usecase.NewDealRequestService(dealRequestRepo, listingRepo)

	uploadHandler := handlers.NewUploadHandler("./storage/images")
	router := httpapi.NewRouter(
		cfg,
		categoryService,
		productService,
		productTermService,
		blogService,
		projectService,
		templateService,
		blockService,
		contentSectionService,
		authService,
		userAuthService,
		dashboardService,
		uploadHandler,
		listingService,
		dealRequestService,
	)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
