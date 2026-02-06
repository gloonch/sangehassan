package httpapi

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"sangehassan/back/internal/adapters/http/handlers"
	"sangehassan/back/internal/adapters/http/middleware"
	"sangehassan/back/internal/config"
	"sangehassan/back/internal/usecase"
)

func NewRouter(
	cfg config.Config,
	categoryService *usecase.CategoryService,
	productService *usecase.ProductService,
	blogService *usecase.BlogService,
	templateService *usecase.TemplateService,
	blockService *usecase.BlockService,
	contentSectionService *usecase.ContentSectionService,
	adminAuthService *usecase.AuthService,
	userAuthService *usecase.UserAuthService,
	dashboardService *usecase.DashboardService,
	uploadHandler *handlers.UploadHandler,
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

	router.Static("/images", "./storage/images")

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	categoryHandler := handlers.NewCategoryHandler(categoryService)
	productHandler := handlers.NewProductHandler(productService)
	blogHandler := handlers.NewBlogHandler(blogService)
	templateHandler := handlers.NewTemplateHandler(templateService)
	blockHandler := handlers.NewBlockHandler(blockService)
	contentSectionHandler := handlers.NewContentSectionHandler(contentSectionService)
	adminAuthHandler := handlers.NewAuthHandler(adminAuthService, cfg.CookieSecure, cfg.JWTTTLHours)
	userAuthHandler := handlers.NewUserAuthHandler(userAuthService, cfg.CookieSecure)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	authMiddleware := middleware.NewAuthMiddleware(adminAuthService)
	userMiddleware := middleware.NewUserAuthMiddleware(userAuthService)

	api := router.Group("/api")
	{
		api.GET("/categories", categoryHandler.List)
		api.GET("/products", productHandler.List)
		api.GET("/products/:slug", productHandler.GetBySlug)
		api.GET("/blogs", blogHandler.List)
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
		}

		api.POST("/admin/upload/template", uploadHandler.UploadTemplate)
		api.POST("/admin/upload/product", uploadHandler.UploadProduct)
		api.POST("/admin/upload/block", uploadHandler.UploadBlock)
		api.POST("/admin/upload/content", uploadHandler.UploadContent)

		admin := api.Group("/admin")
		admin.Use(authMiddleware.RequireAdmin)
		{
			admin.GET("/session", adminAuthHandler.Session)
			admin.GET("/dashboard", dashboardHandler.Stats)

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

			admin.GET("/templates", templateHandler.List)
			admin.POST("/templates", templateHandler.Create)
			admin.PUT("/templates/:id", templateHandler.Update)
			admin.DELETE("/templates/:id", templateHandler.Delete)

			admin.GET("/content-sections", contentSectionHandler.List)
			admin.GET("/content-sections/:id", contentSectionHandler.GetByID)
			admin.POST("/content-sections", contentSectionHandler.Create)
			admin.PUT("/content-sections/:id", contentSectionHandler.Update)
			admin.DELETE("/content-sections/:id", contentSectionHandler.Delete)
		}
	}

	return router
}
