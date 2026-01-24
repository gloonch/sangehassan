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

func NewRouter(cfg config.Config, categoryService *usecase.CategoryService, productService *usecase.ProductService, blogService *usecase.BlogService, authService *usecase.AuthService) *gin.Engine {
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

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	categoryHandler := handlers.NewCategoryHandler(categoryService)
	productHandler := handlers.NewProductHandler(productService)
	blogHandler := handlers.NewBlogHandler(blogService)
	authHandler := handlers.NewAuthHandler(authService, cfg.CookieSecure, cfg.JWTTTLHours)
	authMiddleware := middleware.NewAuthMiddleware(authService)

	api := router.Group("/api")
	{
		api.GET("/categories", categoryHandler.List)
		api.GET("/products", productHandler.List)
		api.GET("/blogs", blogHandler.List)

		api.POST("/admin/login", authHandler.Login)
		api.POST("/admin/logout", authHandler.Logout)

		admin := api.Group("/admin")
		admin.Use(authMiddleware.RequireAdmin)
		{
			admin.GET("/session", authHandler.Session)

			admin.GET("/categories", categoryHandler.List)
			admin.POST("/categories", categoryHandler.Create)
			admin.PUT("/categories/:id", categoryHandler.Update)
			admin.DELETE("/categories/:id", categoryHandler.Delete)

			admin.GET("/products", productHandler.List)
			admin.POST("/products", productHandler.Create)
			admin.PUT("/products/:id", productHandler.Update)
			admin.DELETE("/products/:id", productHandler.Delete)

			admin.GET("/blogs", blogHandler.List)
			admin.POST("/blogs", blogHandler.Create)
			admin.PUT("/blogs/:id", blogHandler.Update)
			admin.DELETE("/blogs/:id", blogHandler.Delete)
		}
	}

	return router
}
