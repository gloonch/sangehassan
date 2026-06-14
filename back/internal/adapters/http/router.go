package httpapi

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"sangehassan/back/internal/adapters/http/handlers"
	"sangehassan/back/internal/adapters/http/middleware"
	"sangehassan/back/internal/config"
	"sangehassan/back/internal/ports"
	"sangehassan/back/internal/usecase"
)

func NewRouter(
	cfg config.Config,
	categoryService *usecase.CategoryService,
	productService *usecase.ProductService,
	productTermService *usecase.ProductTermService,
	catalogService *usecase.CatalogService,
	blogService *usecase.BlogService,
	projectService *usecase.ProjectService,
	templateService *usecase.TemplateService,
	blockService *usecase.BlockService,
	contentSectionService *usecase.ContentSectionService,
	adminAuthService *usecase.AuthService,
	userAuthService *usecase.UserAuthService,
	userRepo ports.UserRepository,
	dashboardService *usecase.DashboardService,
	uploadHandler *handlers.UploadHandler,
	listingService *usecase.ListingService,
	dealRequestService *usecase.DealRequestService,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	corsConfig := cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}
	if len(corsConfig.AllowOrigins) == 0 {
		corsConfig.AllowOrigins = []string{"http://localhost:5173", "http://localhost:5174"}
	}
	router.Use(cors.New(corsConfig))

	router.Static("/images", cfg.UploadDir)

	imageHandler := handlers.NewImageHandler(cfg.UploadDir, "")
	router.GET("/protected-images/*filepath", imageHandler.Serve)
	router.HEAD("/protected-images/*filepath", imageHandler.Serve)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	categoryHandler := handlers.NewCategoryHandler(categoryService)
	productHandler := handlers.NewProductHandler(productService)
	productTermHandler := handlers.NewProductTermHandler(productTermService)
	catalogHandler := handlers.NewCatalogHandler(catalogService)
	blogHandler := handlers.NewBlogHandler(blogService)
	projectHandler := handlers.NewProjectHandler(projectService)
	templateHandler := handlers.NewTemplateHandler(templateService)
	blockHandler := handlers.NewBlockHandler(blockService)
	contentSectionHandler := handlers.NewContentSectionHandler(contentSectionService)
	adminAuthHandler := handlers.NewAuthHandler(adminAuthService, cfg.CookieSecure, cfg.JWTTTLHours)
	userAuthHandler := handlers.NewUserAuthHandler(userAuthService, cfg.CookieSecure)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	authMiddleware := middleware.NewAuthMiddleware(adminAuthService)
	userMiddleware := middleware.NewUserAuthMiddleware(userAuthService)
	listingHandler := handlers.NewListingHandler(listingService, dealRequestService, userRepo)

	api := router.Group("/api")
	{
		api.GET("/categories", categoryHandler.List)
		api.GET("/products", productHandler.List)
		api.GET("/products/:slug", productHandler.GetBySlug)
		api.GET("/product-terms", productTermHandler.List)
		api.GET("/catalog/categories", catalogHandler.Hub)
		api.GET("/catalog/categories/:categorySlug", catalogHandler.Page)
		api.GET("/catalog/categories/:categorySlug/:facet/:value", catalogHandler.Page)
		api.GET("/catalog/routes", catalogHandler.Routes)
		api.GET("/ads", listingHandler.List)
		api.GET("/ads/:id", listingHandler.Get)

		adsAuth := api.Group("/ads")
		adsAuth.Use(userMiddleware.RequireUser)
		{
			adsAuth.POST("", listingHandler.Create)
			adsAuth.POST("/upload-image", uploadHandler.UploadListing)
			adsAuth.PUT("/:id", listingHandler.Update)
			adsAuth.POST("/:id/requests", listingHandler.CreateDealRequest)
			adsAuth.DELETE("/:id", listingHandler.Delete)
		}
		api.GET("/blogs", blogHandler.List)
		api.GET("/projects", projectHandler.ListPublic)
		api.GET("/projects/:id", projectHandler.GetByID)
		api.GET("/templates", templateHandler.List)
		api.GET("/blocks", blockHandler.List)
		api.GET("/blocks/:slug", blockHandler.GetBySlug)
		api.GET("/content-sections", contentSectionHandler.ListPublic)

		api.POST("/admin/login", adminAuthHandler.Login)
		api.POST("/admin/logout", adminAuthHandler.Logout)

		v1 := api.Group("/v1")
		{
			v1.POST("/auth/signup", userAuthHandler.Signup)
			v1.POST("/auth/login", userAuthHandler.Login)
			v1.POST("/auth/logout", userMiddleware.RequireUser, userAuthHandler.Logout)
			v1.POST("/auth/refresh", userAuthHandler.Refresh)

			v1.GET("/me", userMiddleware.RequireUser, userAuthHandler.Me)
			v1.PUT("/me", userMiddleware.RequireUser, userAuthHandler.UpdateMe)
			v1.GET("/me/requests", userMiddleware.RequireUser, userAuthHandler.Requests)
			v1.GET("/me/listings", userMiddleware.RequireUser, listingHandler.MyListings)
		}

		admin := api.Group("/admin")
		admin.Use(authMiddleware.RequireAdmin)
		{
			admin.POST("/upload/template", uploadHandler.UploadTemplate)
			admin.POST("/upload/product", uploadHandler.UploadProduct)
			admin.POST("/upload/block", uploadHandler.UploadBlock)
			admin.POST("/upload/content", uploadHandler.UploadContent)
			admin.POST("/upload/blog", uploadHandler.UploadBlog)
			admin.POST("/upload/project", uploadHandler.UploadProject)

			admin.GET("/session", adminAuthHandler.Session)
			admin.GET("/dashboard", dashboardHandler.Stats)
			admin.GET("/protected-images/settings", imageHandler.AdminSettings)
			admin.PUT("/protected-images/settings", imageHandler.AdminUpdateSettings)

			admin.GET("/product-terms", productTermHandler.List)
			admin.POST("/product-terms", productTermHandler.Upsert)
			admin.DELETE("/product-terms/:id", productTermHandler.Delete)
			admin.GET("/catalog-facet-pages", catalogHandler.AdminListFacetPages)
			admin.POST("/catalog-facet-pages", catalogHandler.AdminUpsertFacetPage)
			admin.DELETE("/catalog-facet-pages/:id", catalogHandler.AdminDeleteFacetPage)

			admin.GET("/categories", categoryHandler.List)
			admin.POST("/categories", categoryHandler.Create)
			admin.PUT("/categories/:id", categoryHandler.Update)
			admin.DELETE("/categories/:id", categoryHandler.Delete)

			admin.GET("/products", productHandler.List)
			admin.GET("/products/:id", productHandler.GetByID)
			admin.POST("/products", productHandler.Create)
			admin.PUT("/products/:id", productHandler.Update)
			admin.DELETE("/products/:id", productHandler.Delete)

			admin.GET("/blocks", blockHandler.List)
			admin.GET("/blocks/:id", blockHandler.GetByID)
			admin.POST("/blocks", blockHandler.Create)
			admin.PUT("/blocks/:id", blockHandler.Update)
			admin.DELETE("/blocks/:id", blockHandler.Delete)

			admin.GET("/blogs", blogHandler.List)
			admin.POST("/blogs", blogHandler.Create)
			admin.PUT("/blogs/:id", blogHandler.Update)
			admin.DELETE("/blogs/:id", blogHandler.Delete)

			admin.GET("/projects", projectHandler.List)
			admin.GET("/projects/:id", projectHandler.GetByID)
			admin.POST("/projects", projectHandler.Create)
			admin.PUT("/projects/:id", projectHandler.Update)
			admin.DELETE("/projects/:id", projectHandler.Delete)

			admin.GET("/templates", templateHandler.List)
			admin.POST("/templates", templateHandler.Create)
			admin.PUT("/templates/:id", templateHandler.Update)
			admin.DELETE("/templates/:id", templateHandler.Delete)

			admin.GET("/content-sections", contentSectionHandler.List)
			admin.GET("/content-sections/:id", contentSectionHandler.GetByID)
			admin.POST("/content-sections", contentSectionHandler.Create)
			admin.PUT("/content-sections/:id", contentSectionHandler.Update)
			admin.DELETE("/content-sections/:id", contentSectionHandler.Delete)

			admin.GET("/requests", listingHandler.AdminListRequests)
			admin.GET("/requests/:id", listingHandler.AdminGetRequest)
			admin.PUT("/requests/:id/status", listingHandler.AdminUpdateRequestStatus)
			admin.GET("/ads", listingHandler.AdminListListings)
			admin.DELETE("/ads/:id", listingHandler.AdminDeleteListing)
			admin.GET("/users", listingHandler.AdminListUsers)
		}
	}

	return router
}
