package httpapi

import (
	"net/http"
	"time"

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
	teamMemberService *usecase.TeamMemberService,
	userAuthService *usecase.UserAuthService,
	userRepo ports.UserRepository,
	dashboardService *usecase.DashboardService,
	uploadHandler *handlers.UploadHandler,
	listingService *usecase.ListingService,
	dealRequestService *usecase.DealRequestService,
	stoneSampleRequestService *usecase.StoneSampleRequestService,
	contactSubmissionService *usecase.ContactSubmissionService,
	operationsService *usecase.OperationsService,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	corsConfig := cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Idempotency-Key"},
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
	teamMemberHandler := handlers.NewTeamMemberHandler(teamMemberService)
	userAuthHandler := handlers.NewUserAuthHandler(userAuthService, cfg.CookieSecure)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	userMiddleware := middleware.NewUserAuthMiddleware(userAuthService)
	listingHandler := handlers.NewListingHandler(listingService, dealRequestService, userRepo)
	stoneSampleRequestHandler := handlers.NewStoneSampleRequestHandler(stoneSampleRequestService)
	contactSubmissionHandler := handlers.NewContactSubmissionHandler(contactSubmissionService)
	operationsHandler := handlers.NewOperationsHandler(operationsService)
	workflowFileHandler := handlers.NewWorkflowFileHandler(operationsService, cfg.WorkflowFileDir)
	operationsMiddleware := middleware.NewOperationsMiddleware(userAuthService, operationsService)

	api := router.Group("/api")
	api.Use(middleware.OriginGuard(cfg.AllowedOrigins))
	{
		api.GET("/categories", categoryHandler.List)
		api.GET("/products", productHandler.List)
		api.GET("/products/:slug", productHandler.GetBySlug)
		api.GET("/product-terms", productTermHandler.List)
		api.GET("/catalog/categories", catalogHandler.Hub)
		api.GET("/catalog/categories/:categorySlug", catalogHandler.Page)
		api.GET("/catalog/categories/:categorySlug/:facet/:value", catalogHandler.Page)
		api.GET("/catalog/routes", catalogHandler.Routes)
		api.GET("/sample-categories", stoneSampleRequestHandler.SampleCategories)
		api.GET("/sample-products", stoneSampleRequestHandler.SampleProducts)
		api.GET("/ads", listingHandler.List)
		api.GET("/ads/:id", listingHandler.Get)

		adsAuth := api.Group("/ads")
		adsAuth.Use(userMiddleware.RequireUser)
		{
			adsAuth.POST("", listingHandler.Create)
			adsAuth.POST("/product-requests", listingHandler.CreateProductRequest)
			adsAuth.PUT("/:id", listingHandler.Update)
			adsAuth.POST("/:id/requests", listingHandler.CreateDealRequest)
			adsAuth.DELETE("/:id", listingHandler.Delete)
		}
		api.GET("/blogs", blogHandler.ListPublic)
		api.GET("/blogs/:locale/:slug", blogHandler.GetPublicBySlug)
		api.GET("/projects", projectHandler.ListPublic)
		api.GET("/projects/:id", projectHandler.GetByID)
		api.GET("/templates", templateHandler.List)
		api.GET("/blocks", blockHandler.List)
		api.GET("/blocks/:slug", blockHandler.GetBySlug)
		api.GET("/content-sections", contentSectionHandler.ListPublic)
		api.GET("/team-members", teamMemberHandler.ListPublic)
		api.POST("/contact-submissions", contactSubmissionHandler.Create)

		v1 := api.Group("/v1")
		{
			v1.POST("/auth/signup", middleware.RateLimit(10, 15*time.Minute), userAuthHandler.Signup)
			v1.POST("/auth/login", middleware.RateLimit(10, 15*time.Minute), userAuthHandler.Login)
			v1.POST("/auth/internal/login", middleware.RateLimit(10, 15*time.Minute), userAuthHandler.LoginInternal)
			v1.POST("/auth/logout", userMiddleware.RequireUser, userAuthHandler.Logout)
			v1.POST("/auth/refresh", userAuthHandler.Refresh)
			v1.POST("/auth/customer/activate", middleware.RateLimit(6, 15*time.Minute), operationsHandler.Activate)
			v1.GET("/session", userAuthHandler.Session)

			v1.GET("/me", userMiddleware.RequireUser, userAuthHandler.Me)
			v1.PUT("/me", userMiddleware.RequireUser, userAuthHandler.UpdateMe)
			v1.POST("/auth/change-password", userMiddleware.RequireUser, userAuthHandler.ChangePassword)
			v1.GET("/me/requests", userMiddleware.RequireUser, userAuthHandler.Requests)
			v1.GET("/me/listings", userMiddleware.RequireUser, listingHandler.MyListings)
			v1.GET("/sample-request-options", userMiddleware.RequireUser, stoneSampleRequestHandler.Options)
			v1.POST("/sample-requests", userMiddleware.RequireUser, stoneSampleRequestHandler.Create)
			v1.GET("/sample-requests", userMiddleware.RequireUser, stoneSampleRequestHandler.ListMine)
			v1.GET("/sample-requests/:id", userMiddleware.RequireUser, stoneSampleRequestHandler.GetMine)

			v1.GET("/operations/me", operationsMiddleware.RequireUser, operationsHandler.Me)
			v1.GET("/dashboard/action-items", operationsMiddleware.RequirePermission("action_items.view_own"), operationsHandler.ActionItems)
			v1.GET("/dashboard/workflow-templates", operationsMiddleware.RequirePermission("workflow_templates.view"), operationsHandler.Workflows)
			v1.GET("/dashboard/workflow-summary", operationsMiddleware.RequirePermission("dashboard.internal.view"), operationsHandler.WorkflowDashboard)
			v1.GET("/workflow-templates/available", operationsMiddleware.RequirePermission("workflow_templates.view"), operationsHandler.Workflows)
			v1.POST("/workflow-instances", operationsMiddleware.RequirePermission("workflow_instances.start"), operationsHandler.StartWorkflow)
			v1.GET("/workflow-instances/:id/runtime", operationsMiddleware.RequireUser, operationsHandler.WorkflowRuntime)
			v1.POST("/workflow-step-instances/:id/start", operationsMiddleware.RequireInternal, operationsHandler.StartWorkflowStep)
			v1.PUT("/workflow-step-instances/:id/draft", operationsMiddleware.RequireInternal, operationsHandler.DraftWorkflowStep)
			v1.POST("/workflow-step-instances/:id/submit", operationsMiddleware.RequireInternal, operationsHandler.SubmitWorkflowStep)
			v1.POST("/workflow-step-instances/:id/approve", operationsMiddleware.RequireInternal, operationsHandler.ApproveWorkflowStep)
			v1.POST("/workflow-step-instances/:id/reject", operationsMiddleware.RequireInternal, operationsHandler.RejectWorkflowStep)
			v1.POST("/workflow-step-instances/:id/skip", operationsMiddleware.RequireInternal, operationsHandler.SkipWorkflowStep)
			v1.POST("/workflow-step-instances/:id/reopen", operationsMiddleware.RequireInternal, operationsHandler.ReopenWorkflowStep)
			v1.POST("/workflow-step-instances/:id/reassign", operationsMiddleware.RequireInternal, operationsHandler.ReassignWorkflowStep)
			v1.POST("/workflow-step-instances/:id/files", operationsMiddleware.RequirePermission("workflow_files.upload"), workflowFileHandler.Upload)
			v1.GET("/workflow-files/:id", operationsMiddleware.RequireUser, workflowFileHandler.Download)
			v1.POST("/workflow-files/entities/:entityType/:entityId", operationsMiddleware.RequireUser, workflowFileHandler.UploadEntity)
			v1.GET("/workflow-instances/:id/discrepancies", operationsMiddleware.RequireInternal, operationsHandler.WorkflowDiscrepancies)
			v1.POST("/workflow-instances/:id/discrepancies", operationsMiddleware.RequirePermission("workflow_discrepancies.review"), operationsHandler.CreateWorkflowDiscrepancy)
			v1.GET("/workflow-discrepancies/:id", operationsMiddleware.RequireInternal, operationsHandler.WorkflowDiscrepancy)
			v1.POST("/workflow-discrepancies/:id/review", operationsMiddleware.RequirePermission("workflow_discrepancies.review"), operationsHandler.ReviewWorkflowDiscrepancy)
			v1.POST("/workflow-discrepancies/:id/accept", operationsMiddleware.RequirePermission("workflow_discrepancies.resolve"), operationsHandler.AcceptWorkflowDiscrepancy)
			v1.POST("/workflow-discrepancies/:id/require-correction", operationsMiddleware.RequirePermission("workflow_discrepancies.resolve"), operationsHandler.RequireCorrectionForDiscrepancy)
			v1.POST("/workflow-discrepancies/:id/resolve", operationsMiddleware.RequirePermission("workflow_discrepancies.resolve"), operationsHandler.ResolveWorkflowDiscrepancy)
			v1.POST("/action-items/:id/complete", operationsMiddleware.RequirePermission("action_items.complete"), operationsHandler.CompleteActionItem)
			v1.GET("/customers/search", operationsMiddleware.RequirePermission("customers.view"), operationsHandler.SearchCustomer)
			v1.POST("/orders/:id/proformas", operationsMiddleware.RequirePermission("proformas.create"), operationsHandler.CreateProforma)
			v1.POST("/proformas/:id/issue", operationsMiddleware.RequirePermission("proformas.issue"), operationsHandler.IssueProforma)
			v1.GET("/account/orders", operationsMiddleware.RequirePermission("customer_portal.orders.view_own"), operationsHandler.AccountOrders)
			v1.GET("/account/proformas", operationsMiddleware.RequirePermission("customer_portal.proformas.view_own"), operationsHandler.AccountProformas)
			v1.GET("/account/orders/:orderId/progress", operationsMiddleware.RequirePermission("customer_portal.orders.view_own"), operationsHandler.AccountOrderProgress)
			v1.GET("/account/orders/:orderId/shipments", operationsMiddleware.RequirePermission("customer_portal.shipments.view_own"), operationsHandler.AccountShipments)
			v1.POST("/account/orders/:orderId/shipments/:id/confirm-delivery", operationsMiddleware.RequirePermission("customer_portal.shipments.confirm_delivery"), operationsHandler.AccountDeliverShipment)

			v1.GET("/dashboard/phase3-summary", operationsMiddleware.RequirePermission("dashboard.internal.view"), operationsHandler.Phase3Dashboard)
			v1.GET("/orders/:id/items", operationsMiddleware.RequirePermission("orders.view_all"), operationsHandler.OrderItems)
			v1.POST("/orders/:id/items", operationsMiddleware.RequirePermission("orders.update"), operationsHandler.CreateOrderItem)
			v1.PUT("/order-items/:id", operationsMiddleware.RequirePermission("orders.update"), operationsHandler.UpdateOrderItem)
			v1.DELETE("/order-items/:id", operationsMiddleware.RequirePermission("orders.update"), operationsHandler.DeleteOrderItem)
			v1.POST("/order-items/:id/conversions", operationsMiddleware.RequirePermission("orders.update"), operationsHandler.CreateOrderItemConversion)
			v1.GET("/orders/:id/progress", operationsMiddleware.RequirePermission("orders.view_all"), operationsHandler.OrderProgress)

			v1.GET("/batches", operationsMiddleware.RequireAnyPermission("batches.view_assigned", "batches.view_all"), operationsHandler.Batches)
			v1.GET("/batches/:id", operationsMiddleware.RequireAnyPermission("batches.view_assigned", "batches.view_all"), operationsHandler.Batch)
			v1.POST("/orders/:id/batches", operationsMiddleware.RequirePermission("batches.create"), operationsHandler.CreateBatch)
			v1.PUT("/batches/:id", operationsMiddleware.RequirePermission("batches.update"), operationsHandler.UpdateBatch)
			v1.POST("/batches/:id/split", operationsMiddleware.RequirePermission("batches.split"), operationsHandler.SplitBatch)
			v1.POST("/batches/merge", operationsMiddleware.RequirePermission("batches.merge"), operationsHandler.MergeBatches)
			v1.POST("/batches/:id/cancel", operationsMiddleware.RequirePermission("batches.cancel"), operationsHandler.CancelBatch)
			v1.GET("/batches/:id/reservations", operationsMiddleware.RequirePermission("inventory.reservations.view"), operationsHandler.BatchReservations)

			v1.GET("/inventory/locations", operationsMiddleware.RequirePermission("inventory.locations.view"), operationsHandler.Locations)
			v1.POST("/inventory/locations", operationsMiddleware.RequirePermission("inventory.locations.manage"), operationsHandler.CreateLocation)
			v1.PUT("/inventory/locations/:id", operationsMiddleware.RequirePermission("inventory.locations.manage"), operationsHandler.UpdateLocation)
			v1.GET("/inventory/lots", operationsMiddleware.RequirePermission("inventory.lots.view"), operationsHandler.Lots)
			v1.GET("/inventory/lots/:id", operationsMiddleware.RequirePermission("inventory.lots.view"), operationsHandler.Lot)
			v1.PUT("/inventory/lots/:id", operationsMiddleware.RequirePermission("inventory.lots.update_metadata"), operationsHandler.UpdateLot)
			v1.GET("/inventory/lots/:id/traceability", operationsMiddleware.RequirePermission("inventory.lots.view"), operationsHandler.LotTraceability)
			v1.POST("/inventory/receipts", operationsMiddleware.RequirePermission("inventory.lots.create"), operationsHandler.ReceiveInventory)
			v1.POST("/inventory/lots/:id/reservations", operationsMiddleware.RequirePermission("inventory.reservations.create"), operationsHandler.CreateReservation)
			v1.POST("/inventory/reservations/:id/release", operationsMiddleware.RequirePermission("inventory.reservations.release"), operationsHandler.ReleaseReservation)
			v1.POST("/inventory/transfers", operationsMiddleware.RequirePermission("inventory.transfers.create"), operationsHandler.TransferInventory)
			v1.POST("/inventory/adjustments", operationsMiddleware.RequirePermission("inventory.adjustments.create"), operationsHandler.AdjustInventory)
			v1.POST("/inventory/conversions", operationsMiddleware.RequirePermission("inventory.conversions.create"), operationsHandler.ConvertInventory)
			v1.GET("/inventory/movements", operationsMiddleware.RequirePermission("inventory.movements.view"), operationsHandler.InventoryMovements)
			v1.GET("/inventory/summary", operationsMiddleware.RequirePermission("inventory.lots.view"), operationsHandler.InventorySummary)

			v1.GET("/vehicles", operationsMiddleware.RequirePermission("vehicles.view"), operationsHandler.Vehicles)
			v1.POST("/vehicles", operationsMiddleware.RequirePermission("vehicles.manage"), operationsHandler.CreateVehicle)
			v1.PUT("/vehicles/:id", operationsMiddleware.RequirePermission("vehicles.manage"), operationsHandler.UpdateVehicle)
			v1.GET("/shipments", operationsMiddleware.RequireAnyPermission("shipments.view_assigned", "shipments.view_all"), operationsHandler.Shipments)
			v1.GET("/shipments/:id", operationsMiddleware.RequireAnyPermission("shipments.view_assigned", "shipments.view_all"), operationsHandler.Shipment)
			v1.POST("/orders/:id/shipments", operationsMiddleware.RequirePermission("shipments.create"), operationsHandler.CreateShipment)
			v1.PUT("/shipments/:id", operationsMiddleware.RequirePermission("shipments.update"), operationsHandler.UpdateShipment)
			v1.POST("/shipments/:id/items", operationsMiddleware.RequirePermission("shipments.plan"), operationsHandler.AddShipmentItem)
			v1.POST("/shipments/:id/load", operationsMiddleware.RequirePermission("shipments.load"), operationsHandler.LoadShipment)
			v1.POST("/shipments/:id/dispatch", operationsMiddleware.RequirePermission("shipments.dispatch"), operationsHandler.DispatchShipment)
			v1.POST("/shipments/:id/arrive", operationsMiddleware.RequirePermission("shipments.confirm_arrival"), operationsHandler.ArriveShipment)
			v1.POST("/shipments/:id/deliver", operationsMiddleware.RequirePermission("shipments.confirm_delivery"), operationsHandler.DeliverShipment)
			v1.POST("/shipments/:id/cancel", operationsMiddleware.RequirePermission("shipments.cancel"), operationsHandler.CancelShipment)
			v1.GET("/packaging", operationsMiddleware.RequirePermission("packaging.view"), operationsHandler.Packaging)
			v1.POST("/batches/:id/packages", operationsMiddleware.RequirePermission("packaging.create"), operationsHandler.CreatePackaging)
			v1.POST("/packages/:id/status", operationsMiddleware.RequirePermission("packaging.update"), operationsHandler.UpdatePackagingStatus)
			v1.POST("/packages/:id/assign", operationsMiddleware.RequirePermission("packaging.assign_to_shipment"), operationsHandler.AssignPackage)
			v1.POST("/shipments/:id/containers", operationsMiddleware.RequirePermission("containers.manage"), operationsHandler.CreateContainer)
			v1.PUT("/containers/:id", operationsMiddleware.RequirePermission("containers.manage"), operationsHandler.UpdateContainer)
			v1.POST("/containers/:id/items", operationsMiddleware.RequirePermission("containers.manage"), operationsHandler.AddContainerItem)

			v1.GET("/operational-costs", operationsMiddleware.RequirePermission("finance.operational_costs.view"), operationsHandler.OperationalCosts)
			v1.POST("/operational-costs", operationsMiddleware.RequirePermission("finance.operational_costs.record"), operationsHandler.CreateOperationalCost)
			v1.POST("/operational-costs/:id/approve", operationsMiddleware.RequirePermission("finance.operational_costs.approve"), operationsHandler.ApproveOperationalCost)
			v1.POST("/operational-costs/:id/void", operationsMiddleware.RequirePermission("finance.operational_costs.approve"), operationsHandler.VoidOperationalCost)
			v1.GET("/workflow-step-instances/:id/transitions", operationsMiddleware.RequirePermission("workflow_transitions.view"), operationsHandler.RuntimeTransitions)
			v1.POST("/workflow-step-instances/:id/select-transition", operationsMiddleware.RequirePermission("workflow_transitions.select"), operationsHandler.SelectRuntimeTransition)

			opsAdmin := v1.Group("/admin")
			{
				opsAdmin.GET("/users", operationsMiddleware.RequirePermission("users.view"), operationsHandler.Users)
				opsAdmin.POST("/users", operationsMiddleware.RequirePermission("users.create"), operationsHandler.CreateUser)
				opsAdmin.PUT("/users/:id/roles", operationsMiddleware.RequirePermission("users.assign_roles"), operationsHandler.AssignRoles)
				opsAdmin.POST("/users/:id/status", operationsMiddleware.RequirePermission("users.disable"), operationsHandler.SetUserStatus)
				opsAdmin.POST("/users/:id/reset-password", operationsMiddleware.RequirePermission("users.reset_password"), operationsHandler.ResetPassword)
				opsAdmin.POST("/customers/:id/regenerate-activation", operationsMiddleware.RequirePermission("customers.regenerate_activation"), operationsHandler.RegenerateActivation)
				opsAdmin.GET("/roles", operationsMiddleware.RequirePermission("roles.view"), operationsHandler.Roles)
				opsAdmin.POST("/roles", operationsMiddleware.RequirePermission("roles.create"), operationsHandler.CreateRole)
				opsAdmin.GET("/roles/:id", operationsMiddleware.RequirePermission("roles.view"), operationsHandler.Role)
				opsAdmin.PATCH("/roles/:id", operationsMiddleware.RequirePermission("roles.update"), operationsHandler.UpdateRole)
				opsAdmin.PUT("/roles/:id/permissions", operationsMiddleware.RequirePermission("roles.assign_permissions"), operationsHandler.RolePermissions)
				opsAdmin.GET("/permissions", operationsMiddleware.RequirePermission("permissions.view"), operationsHandler.Permissions)
				opsAdmin.GET("/audit", operationsMiddleware.RequirePermission("audit.view"), operationsHandler.Audit)
				opsAdmin.GET("/workflow-step-catalogue", operationsMiddleware.RequirePermission("workflow_templates.manage"), operationsHandler.WorkflowStepCatalogue)

				workflowAdmin := opsAdmin.Group("/workflow-templates")
				workflowAdmin.Use(operationsMiddleware.RequirePermission("workflow_templates.manage"))
				{
					workflowAdmin.GET("", operationsHandler.WorkflowTemplateVersions)
					workflowAdmin.POST("", operationsHandler.CreateWorkflowTemplate)
					workflowAdmin.GET("/:id", operationsHandler.WorkflowTemplateVersion)
					workflowAdmin.PATCH("/:id", operationsHandler.UpdateWorkflowTemplate)
					workflowAdmin.POST("/:id/clone", operationsHandler.CloneWorkflowTemplate)
					workflowAdmin.POST("/:id/publish", operationsMiddleware.RequirePermission("workflow_templates.publish"), operationsHandler.PublishWorkflowTemplate)
					workflowAdmin.POST("/:id/archive", operationsMiddleware.RequirePermission("workflow_templates.archive"), operationsHandler.ArchiveWorkflowTemplate)
					workflowAdmin.POST("/:id/steps", operationsHandler.AddWorkflowStep)
					workflowAdmin.PUT("/:id/steps/reorder", operationsHandler.ReorderWorkflowSteps)
					workflowAdmin.PATCH("/:id/steps/:stepId", operationsHandler.UpdateWorkflowStep)
					workflowAdmin.DELETE("/:id/steps/:stepId", operationsHandler.DeleteWorkflowStep)
					workflowAdmin.POST("/:id/steps/:stepId/duplicate", operationsHandler.DuplicateWorkflowStep)
					workflowAdmin.POST("/:id/steps/:stepId/fields", operationsHandler.AddWorkflowField)
					workflowAdmin.PUT("/:id/steps/:stepId/fields/reorder", operationsHandler.ReorderWorkflowFields)
					workflowAdmin.PATCH("/:id/steps/:stepId/fields/:fieldId", operationsHandler.UpdateWorkflowField)
					workflowAdmin.DELETE("/:id/steps/:stepId/fields/:fieldId", operationsHandler.DeleteWorkflowField)
					workflowAdmin.POST("/:id/steps/:stepId/tasks", operationsHandler.AddWorkflowTask)
					workflowAdmin.PATCH("/:id/steps/:stepId/tasks/:taskId", operationsHandler.UpdateWorkflowTask)
					workflowAdmin.DELETE("/:id/steps/:stepId/tasks/:taskId", operationsHandler.DeleteWorkflowTask)
					workflowAdmin.POST("/:id/tasks", operationsHandler.AddWorkflowLevelTask)
					workflowAdmin.PATCH("/:id/tasks/:taskId", operationsHandler.UpdateWorkflowLevelTask)
					workflowAdmin.DELETE("/:id/tasks/:taskId", operationsHandler.DeleteWorkflowLevelTask)
					workflowAdmin.POST("/:id/handoff-metrics", operationsHandler.AddHandoffMetric)
					workflowAdmin.PATCH("/:id/handoff-metrics/:metricId", operationsHandler.UpdateHandoffMetric)
					workflowAdmin.DELETE("/:id/handoff-metrics/:metricId", operationsHandler.DeleteHandoffMetric)
					workflowAdmin.GET("/:id/transitions", operationsMiddleware.RequirePermission("workflow_transitions.view"), operationsHandler.WorkflowTransitions)
					workflowAdmin.POST("/:id/transitions", operationsMiddleware.RequirePermission("workflow_transitions.manage"), operationsHandler.CreateWorkflowTransition)
					workflowAdmin.PATCH("/:id/transitions/:transitionId", operationsMiddleware.RequirePermission("workflow_transitions.manage"), operationsHandler.UpdateWorkflowTransition)
					workflowAdmin.DELETE("/:id/transitions/:transitionId", operationsMiddleware.RequirePermission("workflow_transitions.manage"), operationsHandler.DeleteWorkflowTransition)
				}
			}
		}

		admin := api.Group("/admin")
		admin.Use(operationsMiddleware.RequirePermission("content.manage"))
		{
			admin.POST("/upload/template", uploadHandler.UploadTemplate)
			admin.POST("/upload/product", uploadHandler.UploadProduct)
			admin.POST("/upload/block", uploadHandler.UploadBlock)
			admin.POST("/upload/content", uploadHandler.UploadContent)
			admin.POST("/upload/team", uploadHandler.UploadTeam)
			admin.POST("/upload/blog", uploadHandler.UploadBlog)
			admin.POST("/upload/project", uploadHandler.UploadProject)

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

			admin.GET("/blogs", blogHandler.ListAdmin)
			admin.GET("/blogs/:id", blogHandler.GetByID)
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

			admin.GET("/team-members", teamMemberHandler.List)
			admin.GET("/team-members/:id", teamMemberHandler.GetByID)
			admin.POST("/team-members", teamMemberHandler.Create)
			admin.PUT("/team-members/:id", teamMemberHandler.Update)
			admin.DELETE("/team-members/:id", teamMemberHandler.Delete)

			admin.GET("/requests", listingHandler.AdminListRequests)
			admin.GET("/requests/:id", listingHandler.AdminGetRequest)
			admin.PUT("/requests/:id/status", listingHandler.AdminUpdateRequestStatus)
			admin.GET("/sample-requests", stoneSampleRequestHandler.AdminList)
			admin.GET("/sample-requests/:id", stoneSampleRequestHandler.AdminGet)
			admin.PUT("/sample-requests/:id/status", stoneSampleRequestHandler.AdminUpdateStatus)
			admin.GET("/ads", listingHandler.AdminListListings)
			admin.DELETE("/ads/:id", listingHandler.AdminDeleteListing)
			admin.GET("/product-requests", listingHandler.AdminListProductRequests)
			admin.GET("/contact-submissions", contactSubmissionHandler.AdminList)
			admin.GET("/contact-submissions/:id", contactSubmissionHandler.AdminGet)
			admin.GET("/users", listingHandler.AdminListUsers)
		}
	}

	return router
}
