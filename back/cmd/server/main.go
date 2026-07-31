package main

import (
	"context"
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
	catalogRepo := postgres.NewCatalogRepository(db)
	blogRepo := postgres.NewBlogRepository(db)
	projectRepo := postgres.NewProjectRepository(db)
	templateRepo := postgres.NewTemplateRepository(db)
	blockRepo := postgres.NewBlockRepository(db)
	contentSectionRepo := postgres.NewContentSectionRepository(db)
	teamMemberRepo := postgres.NewTeamMemberRepository(db)
	userRepo := postgres.NewUserRepository(db)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(db)
	dashboardRepo := postgres.NewDashboardRepository(db)
	listingRepo := postgres.NewListingRepository(db)
	dealRequestRepo := postgres.NewDealRequestRepository(db)
	stoneSampleRequestRepo := postgres.NewStoneSampleRequestRepository(db)
	contactSubmissionRepo := postgres.NewContactSubmissionRepository(db)

	categoryService := usecase.NewCategoryService(categoryRepo)
	productService := usecase.NewProductService(productRepo)
	productTermService := usecase.NewProductTermService(productTermRepo)
	catalogService := usecase.NewCatalogService(catalogRepo, cfg.CatalogMinProducts)
	blogService := usecase.NewBlogService(blogRepo)
	projectService := usecase.NewProjectService(projectRepo)
	templateService := usecase.NewTemplateService(templateRepo)
	blockService := usecase.NewBlockService(blockRepo)
	contentSectionService := usecase.NewContentSectionService(contentSectionRepo)
	teamMemberService := usecase.NewTeamMemberService(teamMemberRepo)
	userAuthService := usecase.NewUserAuthService(userRepo, refreshTokenRepo, dealRequestRepo, cfg.JWTSecret, cfg.AccessTokenMinutes, cfg.RefreshTokenDays)
	dashboardService := usecase.NewDashboardService(dashboardRepo)
	listingService := usecase.NewListingService(listingRepo)
	dealRequestService := usecase.NewDealRequestService(dealRequestRepo, listingRepo)
	stoneSampleRequestService := usecase.NewStoneSampleRequestService(stoneSampleRequestRepo, userRepo)
	contactSubmissionService := usecase.NewContactSubmissionService(contactSubmissionRepo)
	operationsService := usecase.NewOperationsService(db)
	if err := operationsService.BootstrapSuperAdmin(context.Background(), cfg.BootstrapSuperAdminPhone, cfg.BootstrapSuperAdminPassword, cfg.BootstrapSuperAdminFirstName, cfg.BootstrapSuperAdminLastName); err != nil {
		log.Fatalf("super admin bootstrap error: %v", err)
	}

	uploadHandler := handlers.NewUploadHandler(cfg.UploadDir)
	router := httpapi.NewRouter(
		cfg,
		categoryService,
		productService,
		productTermService,
		catalogService,
		blogService,
		projectService,
		templateService,
		blockService,
		contentSectionService,
		teamMemberService,
		userAuthService,
		userRepo,
		dashboardService,
		uploadHandler,
		listingService,
		dealRequestService,
		stoneSampleRequestService,
		contactSubmissionService,
		operationsService,
	)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
