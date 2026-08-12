package httpapi

import (
	"log/slog"
	"net/http"
	"strings"
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
	logger := slog.Default()
	router.Use(middleware.RequestContext(), middleware.SecurityHeaders(strings.EqualFold(cfg.AppEnv, "production")), middleware.StructuredLogger(logger), middleware.StructuredRecovery(logger))

	corsConfig := cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Idempotency-Key"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
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
	router.GET("/ready", func(c *gin.Context) {
		if err := operationsService.Ready(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "requestId": usecase.RequestIDFromContext(c.Request.Context())})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
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
	operationsHandler.ConfigureBuildInfo(cfg.AppVersion, cfg.GitCommit, cfg.BuildTime, cfg.AppEnv)
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
		api.GET("/ads/live-feed", listingHandler.LiveFeed)
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
			v1.GET("/version", operationsHandler.Version)
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
			v1.GET("/operations/features", operationsMiddleware.RequireInternal, operationsHandler.FeatureFlags)
			v1.GET("/dashboard/home", operationsMiddleware.RequirePermission("dashboard.internal.view"), operationsHandler.DashboardHome)
			v1.GET("/search", middleware.RateLimit(60, time.Minute), operationsMiddleware.RequireInternal, operationsHandler.GlobalSearch)
			v1.GET("/saved-views", operationsMiddleware.RequireInternal, operationsHandler.SavedViews)
			v1.POST("/saved-views", operationsMiddleware.RequireInternal, operationsHandler.CreateSavedView)
			v1.PUT("/saved-views/:id", operationsMiddleware.RequireInternal, operationsHandler.UpdateSavedView)
			v1.DELETE("/saved-views/:id", operationsMiddleware.RequireInternal, operationsHandler.DeleteSavedView)
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
			v1.GET("/orders/:id/commercial-terms", operationsMiddleware.RequirePermission("finance.commercial_terms.view"), operationsHandler.CommercialTerms)
			v1.PUT("/orders/:id/commercial-terms", operationsMiddleware.RequirePermission("finance.commercial_terms.manage"), operationsHandler.SaveCommercialTerms)
			v1.GET("/orders/:id/payment-schedule", operationsMiddleware.RequirePermission("finance.payment_schedule.view"), operationsHandler.PaymentSchedule)
			v1.PUT("/orders/:id/payment-schedule", operationsMiddleware.RequirePermission("finance.payment_schedule.manage"), operationsHandler.SavePaymentSchedule)
			v1.GET("/orders/:id/payments", operationsMiddleware.RequireAnyPermission("finance.payments.view", "finance.customer_payments.view"), operationsHandler.Payments)
			v1.POST("/orders/:id/payments", operationsMiddleware.RequireAnyPermission("finance.payments.record", "finance.customer_payments.record"), operationsHandler.RecordPayment)
			v1.GET("/orders/:id/costs", operationsMiddleware.RequireAnyPermission("finance.costs.view", "finance.costs.view_all", "finance.costs.view_assigned"), operationsHandler.OrderCosts)
			v1.GET("/orders/:id/financial-summary", operationsMiddleware.RequirePermission("finance.commercial_terms.view"), operationsHandler.FinancialSummary)
			v1.POST("/orders/:id/confirm", operationsMiddleware.RequirePermission("orders.confirm"), operationsHandler.ConfirmOrder)
			v1.POST("/payments/:id/confirm", operationsMiddleware.RequireAnyPermission("finance.payments.confirm", "finance.customer_payments.confirm"), operationsHandler.ConfirmPayment)
			v1.GET("/payments/:id/allocations", operationsMiddleware.RequireAnyPermission("finance.payments.view", "finance.customer_payments.view"), operationsHandler.PaymentAllocations)
			v1.POST("/payments/:id/reject", operationsMiddleware.RequireAnyPermission("finance.payments.confirm", "finance.customer_payments.reject"), operationsHandler.RejectPayment)
			v1.POST("/payments/:id/refund", operationsMiddleware.RequireAnyPermission("finance.payments.refund", "finance.customer_payments.refund"), operationsHandler.RefundPayment)
			v1.POST("/costs", operationsMiddleware.RequirePermission("finance.costs.record"), operationsHandler.CreateCost)
			v1.POST("/costs/:id/submit", operationsMiddleware.RequirePermission("finance.costs.record"), operationsHandler.SubmitCost)
			v1.POST("/costs/:id/approve", operationsMiddleware.RequirePermission("finance.costs.approve"), operationsHandler.ApproveCost)
			v1.POST("/costs/:id/reject", operationsMiddleware.RequirePermission("finance.costs.approve"), operationsHandler.RejectCost)
			v1.POST("/costs/:id/mark-paid", operationsMiddleware.RequireAnyPermission("finance.costs.pay", "finance.costs.mark_paid"), operationsHandler.MarkCostPaid)
			v1.POST("/costs/:id/cancel", operationsMiddleware.RequirePermission("finance.costs.approve"), operationsHandler.CancelCost)

			v1.GET("/notifications", operationsMiddleware.RequirePermission("notifications.view_own"), operationsHandler.Notifications)
			v1.POST("/notifications/read-all", operationsMiddleware.RequirePermission("notifications.view_own"), operationsHandler.ReadAllNotifications)
			v1.POST("/notifications/:id/read", operationsMiddleware.RequirePermission("notifications.view_own"), operationsHandler.ReadNotification)
			v1.GET("/notifications/preferences", operationsMiddleware.RequirePermission("notifications.preferences.manage"), operationsHandler.NotificationPreferences)
			v1.PUT("/notifications/preferences", operationsMiddleware.RequirePermission("notifications.preferences.manage"), operationsHandler.SaveNotificationPreferences)

			v1.GET("/orders/:id/documents", operationsMiddleware.RequireAnyPermission("documents.view", "documents.view_all", "documents.view_assigned"), operationsHandler.Documents)
			v1.POST("/orders/:id/documents/generate", operationsMiddleware.RequirePermission("documents.generate"), operationsHandler.GenerateDocument)
			v1.POST("/orders/:id/documents/upload", operationsMiddleware.RequireAnyPermission("documents.upload", "documents.create"), workflowFileHandler.UploadDocument)
			v1.POST("/documents/:id/issue", operationsMiddleware.RequirePermission("documents.issue"), operationsHandler.IssueDocument)
			v1.POST("/documents/:id/cancel", operationsMiddleware.RequirePermission("documents.cancel"), operationsHandler.CancelDocument)
			v1.GET("/documents/:id/download", operationsMiddleware.RequireAnyPermission("documents.download", "documents.download_internal"), operationsHandler.DownloadDocument)

			v1.GET("/customers/:id/contact-logs", operationsMiddleware.RequirePermission("customers.contacts.view"), operationsHandler.CustomerContactLogs)
			v1.POST("/customers/:id/contact-logs", operationsMiddleware.RequirePermission("customers.contacts.record"), operationsHandler.CreateCustomerContactLog)
			v1.GET("/orders/:id/contact-logs", operationsMiddleware.RequirePermission("customers.contacts.view"), operationsHandler.OrderContactLogs)
			v1.POST("/orders/:id/contact-logs", operationsMiddleware.RequirePermission("customers.contacts.record"), operationsHandler.CreateOrderContactLog)
			v1.GET("/account/orders", operationsMiddleware.RequirePermission("customer_portal.orders.view_own"), operationsMiddleware.RequireFeature("customer_portal_enabled"), operationsHandler.AccountOrders)
			v1.GET("/account/proformas", operationsMiddleware.RequirePermission("customer_portal.proformas.view_own"), operationsMiddleware.RequireFeature("customer_portal_enabled"), operationsHandler.AccountProformas)
			v1.GET("/account/orders/:id/progress", operationsMiddleware.RequirePermission("customer_portal.orders.view_own"), operationsMiddleware.RequireFeature("customer_portal_enabled"), operationsHandler.AccountOrderProgress)
			v1.GET("/account/orders/:id/shipments", operationsMiddleware.RequirePermission("customer_portal.shipments.view_own"), operationsMiddleware.RequireFeature("customer_portal_enabled"), operationsHandler.AccountShipments)
			v1.POST("/account/orders/:id/shipments/:shipmentId/confirm-delivery", operationsMiddleware.RequirePermission("customer_portal.shipments.confirm_delivery"), operationsMiddleware.RequireFeature("customer_portal_enabled"), operationsHandler.AccountDeliverShipment)
			v1.GET("/account/orders/:id/financial-summary", operationsMiddleware.RequirePermission("customer_portal.financial_summary.view_own"), operationsMiddleware.RequireFeature("customer_portal_enabled"), operationsHandler.AccountFinancialSummary)
			v1.GET("/account/orders/:id/payment-schedule", operationsMiddleware.RequirePermission("customer_portal.payments.view_own"), operationsMiddleware.RequireFeature("customer_portal_enabled"), operationsHandler.AccountPaymentSchedule)
			v1.GET("/account/orders/:id/payments", operationsMiddleware.RequirePermission("customer_portal.payments.view_own"), operationsMiddleware.RequireFeature("customer_portal_enabled"), operationsHandler.AccountPayments)
			v1.GET("/account/orders/:id/documents", operationsMiddleware.RequirePermission("customer_portal.documents.view_own"), operationsMiddleware.RequireFeature("customer_portal_enabled"), operationsHandler.AccountDocuments)
			v1.GET("/account/documents/:id/download", operationsMiddleware.RequirePermission("customer_portal.documents.view_own"), operationsMiddleware.RequireFeature("customer_portal_enabled"), operationsHandler.DownloadAccountDocument)

			v1.GET("/dashboard/operations-summary", operationsMiddleware.RequirePermission("dashboard.internal.view"), operationsHandler.OperationsDashboardSummary)
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
			v1.POST("/inventory/locations", operationsMiddleware.RequirePermission("inventory.locations.manage"), operationsMiddleware.RequireFeature("inventory_module_enabled"), operationsHandler.CreateLocation)
			v1.PUT("/inventory/locations/:id", operationsMiddleware.RequirePermission("inventory.locations.manage"), operationsMiddleware.RequireFeature("inventory_module_enabled"), operationsHandler.UpdateLocation)
			v1.GET("/inventory/lots", operationsMiddleware.RequirePermission("inventory.lots.view"), operationsHandler.Lots)
			v1.GET("/inventory/lots/:id", operationsMiddleware.RequirePermission("inventory.lots.view"), operationsHandler.Lot)
			v1.PUT("/inventory/lots/:id", operationsMiddleware.RequirePermission("inventory.lots.update_metadata"), operationsMiddleware.RequireFeature("inventory_module_enabled"), operationsHandler.UpdateLot)
			v1.GET("/inventory/lots/:id/traceability", operationsMiddleware.RequirePermission("inventory.lots.view"), operationsHandler.LotTraceability)
			v1.POST("/inventory/receipts", operationsMiddleware.RequirePermission("inventory.lots.create"), operationsMiddleware.RequireFeature("inventory_module_enabled"), operationsHandler.ReceiveInventory)
			v1.POST("/inventory/lots/:id/reservations", operationsMiddleware.RequirePermission("inventory.reservations.create"), operationsMiddleware.RequireFeature("inventory_module_enabled"), operationsHandler.CreateReservation)
			v1.POST("/inventory/reservations/:id/release", operationsMiddleware.RequirePermission("inventory.reservations.release"), operationsMiddleware.RequireFeature("inventory_module_enabled"), operationsHandler.ReleaseReservation)
			v1.POST("/inventory/transfers", operationsMiddleware.RequirePermission("inventory.transfers.create"), operationsMiddleware.RequireFeature("inventory_module_enabled"), operationsHandler.TransferInventory)
			v1.POST("/inventory/adjustments", operationsMiddleware.RequirePermission("inventory.adjustments.create"), operationsMiddleware.RequireFeature("inventory_module_enabled"), operationsHandler.AdjustInventory)
			v1.POST("/inventory/conversions", operationsMiddleware.RequirePermission("inventory.conversions.create"), operationsMiddleware.RequireFeature("inventory_module_enabled"), operationsHandler.ConvertInventory)
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

			v1.GET("/suppliers", operationsMiddleware.RequirePermission("suppliers.view"), operationsHandler.Suppliers)
			v1.POST("/suppliers", operationsMiddleware.RequirePermission("suppliers.create"), operationsMiddleware.RequireFeature("supplier_module_enabled"), operationsHandler.CreateSupplier)
			v1.GET("/suppliers/:id", operationsMiddleware.RequirePermission("suppliers.view"), operationsHandler.Supplier)
			v1.PATCH("/suppliers/:id", operationsMiddleware.RequirePermission("suppliers.update"), operationsMiddleware.RequireFeature("supplier_module_enabled"), operationsHandler.UpdateSupplier)
			v1.POST("/suppliers/:id/disable", operationsMiddleware.RequirePermission("suppliers.disable"), operationsMiddleware.RequireFeature("supplier_module_enabled"), operationsHandler.DisableSupplier)

			v1.GET("/purchases", operationsMiddleware.RequireAnyPermission("purchases.view_assigned", "purchases.view_all"), operationsHandler.Purchases)
			v1.POST("/orders/:id/purchases", operationsMiddleware.RequirePermission("purchases.create"), operationsMiddleware.RequireFeature("supplier_module_enabled"), operationsHandler.CreatePurchase)
			v1.GET("/purchases/:id", operationsMiddleware.RequireAnyPermission("purchases.view_assigned", "purchases.view_all"), operationsHandler.Purchase)
			v1.PATCH("/purchases/:id", operationsMiddleware.RequirePermission("purchases.update"), operationsMiddleware.RequireFeature("supplier_module_enabled"), operationsHandler.UpdatePurchase)
			v1.POST("/purchases/:id/confirm", operationsMiddleware.RequirePermission("purchases.confirm"), operationsMiddleware.RequireFeature("supplier_module_enabled"), operationsHandler.ConfirmPurchase)
			v1.POST("/purchases/:id/receive", operationsMiddleware.RequirePermission("purchases.receive"), operationsMiddleware.RequireFeature("supplier_module_enabled"), operationsHandler.ReceivePurchase)
			v1.POST("/purchases/:id/cancel", operationsMiddleware.RequirePermission("purchases.cancel"), operationsMiddleware.RequireFeature("supplier_module_enabled"), operationsHandler.CancelPurchase)

			v1.GET("/quality-inspections", operationsMiddleware.RequireAnyPermission("quality.view_assigned", "quality.view_all"), operationsHandler.QualityInspections)
			v1.POST("/quality-inspections", operationsMiddleware.RequirePermission("quality.inspect"), operationsHandler.CreateQualityInspection)
			v1.GET("/quality-inspections/:id", operationsMiddleware.RequireAnyPermission("quality.view_assigned", "quality.view_all"), operationsHandler.QualityInspection)
			v1.POST("/quality-inspections/:id/pass", operationsMiddleware.RequirePermission("quality.inspect"), operationsHandler.PassQualityInspection)
			v1.POST("/quality-inspections/:id/fail", operationsMiddleware.RequirePermission("quality.inspect"), operationsHandler.FailQualityInspection)
			v1.POST("/quality-inspections/:id/request-rework", operationsMiddleware.RequirePermission("quality.inspect"), operationsHandler.RequestQualityRework)
			v1.POST("/quality-inspections/:id/reject", operationsMiddleware.RequirePermission("quality.reject"), operationsHandler.RejectQualityInspection)
			v1.POST("/quality-inspections/:id/override", operationsMiddleware.RequireAnyPermission("quality.accept", "quality.override"), operationsHandler.OverrideQualityInspection)

			v1.GET("/installations", operationsMiddleware.RequireAnyPermission("installation.view_assigned", "installation.view_all"), operationsHandler.Installations)
			v1.POST("/orders/:id/installations", operationsMiddleware.RequirePermission("installation.create"), operationsMiddleware.RequireFeature("installation_module_enabled"), operationsHandler.CreateInstallation)
			v1.GET("/installations/:id", operationsMiddleware.RequireAnyPermission("installation.view_assigned", "installation.view_all"), operationsHandler.Installation)
			v1.PATCH("/installations/:id", operationsMiddleware.RequirePermission("installation.update"), operationsMiddleware.RequireFeature("installation_module_enabled"), operationsHandler.UpdateInstallation)
			v1.PUT("/installations/:id/members", operationsMiddleware.RequirePermission("installation.update"), operationsMiddleware.RequireFeature("installation_module_enabled"), operationsHandler.ReplaceInstallationMembers)
			v1.POST("/installations/:id/start", operationsMiddleware.RequirePermission("installation.start"), operationsMiddleware.RequireFeature("installation_module_enabled"), operationsHandler.StartInstallation)
			v1.POST("/installations/:id/pause", operationsMiddleware.RequirePermission("installation.start"), operationsMiddleware.RequireFeature("installation_module_enabled"), operationsHandler.PauseInstallation)
			v1.POST("/installations/:id/updates", operationsMiddleware.RequirePermission("installation.progress"), operationsMiddleware.RequireFeature("installation_module_enabled"), operationsHandler.AddInstallationUpdate)
			v1.POST("/installations/:id/issues", operationsMiddleware.RequirePermission("installation.progress"), operationsMiddleware.RequireFeature("installation_module_enabled"), operationsHandler.AddInstallationIssue)
			v1.POST("/installations/:id/issues/:issueId/resolve", operationsMiddleware.RequirePermission("installation.update"), operationsMiddleware.RequireFeature("installation_module_enabled"), operationsHandler.ResolveInstallationIssue)
			v1.POST("/installations/:id/materials", operationsMiddleware.RequirePermission("installation.progress"), operationsMiddleware.RequireFeature("installation_module_enabled"), operationsHandler.AddInstallationMaterial)
			v1.POST("/installations/:id/complete", operationsMiddleware.RequirePermission("installation.complete"), operationsMiddleware.RequireFeature("installation_module_enabled"), operationsHandler.CompleteInstallation)
			v1.POST("/installations/:id/cancel", operationsMiddleware.RequirePermission("installation.cancel"), operationsMiddleware.RequireFeature("installation_module_enabled"), operationsHandler.CancelInstallation)

			v1.POST("/orders/:id/customer-acceptances", operationsMiddleware.RequirePermission("customer_acceptance.record"), operationsHandler.RecordCustomerAcceptance)
			v1.GET("/orders/:id/closure-readiness", operationsMiddleware.RequireAnyPermission("orders.close", "orders.close_with_warnings"), operationsHandler.OrderClosureReadiness)
			v1.POST("/orders/:id/close", operationsMiddleware.RequireAnyPermission("orders.close", "orders.close_with_warnings"), operationsHandler.CloseOrder)
			v1.GET("/account/orders/:id/installation", operationsMiddleware.RequirePermission("customer_portal.installation.view_own"), operationsMiddleware.RequireFeature("customer_portal_enabled"), operationsHandler.AccountInstallation)
			v1.POST("/account/orders/:id/acceptance", operationsMiddleware.RequirePermission("customer_portal.acceptance.confirm_own"), operationsMiddleware.RequireFeature("customer_portal_enabled"), operationsHandler.AccountCustomerAcceptance)

			opsAdmin := v1.Group("/admin")
			{
				opsAdmin.GET("/settings", operationsMiddleware.RequirePermission("settings.view"), operationsHandler.Settings)
				opsAdmin.PUT("/settings", operationsMiddleware.RequirePermission("settings.manage"), operationsHandler.UpdateSetting)
				opsAdmin.GET("/system-info", operationsMiddleware.RequirePermission("settings.view"), operationsHandler.SystemInfo)
				opsAdmin.GET("/diagnostics/workflows/:id", operationsMiddleware.RequirePermission("diagnostics.view"), operationsHandler.WorkflowDiagnostics)
				opsAdmin.POST("/diagnostics/workflows/:id/repair", operationsMiddleware.RequirePermission("diagnostics.repair"), operationsHandler.RepairWorkflow)
				opsAdmin.POST("/tools/orders/:id/estimated-delivery", operationsMiddleware.RequirePermission("admin_tools.order_repair"), operationsHandler.CorrectEstimatedDelivery)
				opsAdmin.POST("/tools/orders/:id/recalculate-progress", operationsMiddleware.RequirePermission("admin_tools.order_repair"), operationsHandler.RecalculateProgress)
				opsAdmin.POST("/tools/orders/:id/reconcile-payment", operationsMiddleware.RequirePermission("admin_tools.order_repair"), operationsHandler.ReconcilePayment)
				opsAdmin.POST("/tools/action-items/:id/resolve", operationsMiddleware.RequirePermission("admin_tools.workflow_repair"), operationsHandler.ResolveStuckAction)
				opsAdmin.POST("/users/:id/revoke-sessions", operationsMiddleware.RequirePermission("admin_tools.sessions.revoke"), operationsHandler.RevokeSessions)
				opsAdmin.GET("/exports/:kind", operationsMiddleware.RequireInternal, operationsHandler.ExportCSV)
				opsAdmin.GET("/exchange-rates", operationsMiddleware.RequireAnyPermission("finance.exchange_rates.view", "finance.exchange_rates.manage"), operationsHandler.ExchangeRates)
				opsAdmin.POST("/exchange-rates", operationsMiddleware.RequirePermission("finance.exchange_rates.manage"), operationsHandler.SaveExchangeRate)
				opsAdmin.GET("/notification-templates", operationsMiddleware.RequirePermission("notifications.templates.manage"), operationsHandler.NotificationTemplates)
				opsAdmin.PUT("/notification-templates/:id", operationsMiddleware.RequirePermission("notifications.templates.manage"), operationsHandler.UpdateNotificationTemplate)
				opsAdmin.GET("/notification-deliveries", operationsMiddleware.RequireAnyPermission("notifications.delivery.view", "notifications.deliveries.retry", "notifications.retry"), operationsHandler.NotificationDeliveries)
				opsAdmin.POST("/notification-deliveries/:id/retry", operationsMiddleware.RequireAnyPermission("notifications.deliveries.retry", "notifications.retry"), operationsHandler.RetryNotification)
				opsAdmin.GET("/document-templates", operationsMiddleware.RequireAnyPermission("document_templates.manage", "documents.templates.manage"), operationsHandler.DocumentTemplates)
				opsAdmin.PUT("/document-templates/:id", operationsMiddleware.RequireAnyPermission("document_templates.manage", "documents.templates.manage"), operationsHandler.UpdateDocumentTemplate)
				opsAdmin.GET("/reports/overview", operationsMiddleware.RequirePermission("reports.overview.view"), operationsHandler.ReportOverview)
				opsAdmin.GET("/reports/receivables", operationsMiddleware.RequirePermission("reports.receivables.view"), operationsHandler.ReportReceivables)
				opsAdmin.GET("/reports/costs", operationsMiddleware.RequirePermission("reports.costs.view"), operationsHandler.ReportCosts)
				opsAdmin.GET("/reports/profitability", operationsMiddleware.RequirePermission("reports.profitability.view"), operationsHandler.ReportProfitability)
				opsAdmin.GET("/reports/operations", operationsMiddleware.RequirePermission("reports.operations.view"), operationsHandler.ReportOperations)
				opsAdmin.GET("/reports/sales", operationsMiddleware.RequirePermission("reports.sales.view"), operationsHandler.ReportSales)
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
				opsAdmin.GET("/audit", operationsMiddleware.RequirePermission("audit.view"), operationsHandler.AuditPage)
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
					workflowAdmin.GET("/:id/document-requirements", operationsMiddleware.RequirePermission("workflow_document_requirements.manage"), operationsHandler.DocumentRequirements)
					workflowAdmin.POST("/:id/document-requirements", operationsMiddleware.RequirePermission("workflow_document_requirements.manage"), operationsHandler.SaveDocumentRequirement)
					workflowAdmin.DELETE("/:id/document-requirements/:requirementId", operationsMiddleware.RequirePermission("workflow_document_requirements.manage"), operationsHandler.DeleteDocumentRequirement)
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
