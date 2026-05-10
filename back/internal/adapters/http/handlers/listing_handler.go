package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
	"sangehassan/back/internal/usecase"
)

type ListingHandler struct {
	listings *usecase.ListingService
	deals    *usecase.DealRequestService
	users    ports.UserRepository
}

func NewListingHandler(listingService *usecase.ListingService, dealService *usecase.DealRequestService, userRepo ports.UserRepository) *ListingHandler {
	return &ListingHandler{
		listings: listingService,
		deals:    dealService,
		users:    userRepo,
	}
}

type listingPayload struct {
	Title       string         `json:"title"`
	StoneType   string         `json:"stone_type"`
	Form        string         `json:"form"`
	Tonnage     *float64       `json:"tonnage"`
	Province    string         `json:"province"`
	City        string         `json:"city"`
	PriceAmount *float64       `json:"price_amount"`
	PriceUnit   string         `json:"price_unit"`
	Description string         `json:"description"`
	ExtraProps  map[string]any `json:"extra_props"`
	Images      []string       `json:"images"`
	Status      string         `json:"status"`
}

func (h *ListingHandler) List(c *gin.Context) {
	filter := buildListingFilter(c)
	items, err := h.listings.List(c.Request.Context(), filter)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load listings")
		return
	}
	for i := range items {
		items[i].CreatedBy = nil // mask owner id
	}
	respondOK(c, items)
}

// MyListings returns listings created by the authenticated user.
func (h *ListingHandler) MyListings(c *gin.Context) {
	userID, _ := c.Get("user_id")
	idStr, _ := userID.(string)
	if idStr == "" {
		respondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	filter := buildListingFilter(c)
	filter.OwnerID = &idStr
	items, err := h.listings.List(c.Request.Context(), filter)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load listings")
		return
	}
	respondOK(c, items)
}

func (h *ListingHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid listing id")
		return
	}
	item, err := h.listings.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "listing not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to load listing")
		return
	}
	item.CreatedBy = nil // mask owner id
	respondOK(c, item)
}

func (h *ListingHandler) Create(c *gin.Context) {
	var payload listingPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}

	listing := buildListingFromPayload(payload)
	if userID, ok := c.Get("user_id"); ok {
		if idStr, ok := userID.(string); ok && idStr != "" {
			listing.CreatedBy = &idStr
		}
	}
	created, err := h.listings.Create(c.Request.Context(), listing)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create listing")
		return
	}
	respondCreated(c, created)
}

func (h *ListingHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid listing id")
		return
	}
	current, err := h.listings.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "listing not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to load listing")
		return
	}
	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)
	if current.CreatedBy != nil && userIDStr != "" && *current.CreatedBy != userIDStr {
		respondError(c, http.StatusForbidden, "not allowed to update this listing")
		return
	}
	var payload listingPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}

	listing := buildListingFromPayload(payload)
	listing.ID = id
	if userIDStr != "" {
		listing.CreatedBy = &userIDStr
	}
	updated, err := h.listings.Update(c.Request.Context(), listing)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to update listing")
		return
	}
	respondOK(c, updated)
}

type dealRequestPayload struct {
	RequestType string `json:"request_type"`
	BuyerNote   string `json:"buyer_note"`
}

func (h *ListingHandler) CreateDealRequest(c *gin.Context) {
	listingID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid listing id")
		return
	}
	var payload dealRequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}
	reqType := domain.DealRequestType(strings.ToUpper(payload.RequestType))
	if reqType != domain.DealRequestTypeInspection && reqType != domain.DealRequestTypePurchase && reqType != domain.DealRequestTypeBoth {
		respondError(c, http.StatusBadRequest, "invalid request_type")
		return
	}

	req := domain.DealRequest{
		ListingID:   listingID,
		RequestType: reqType,
		BuyerNote:   payload.BuyerNote,
	}
	if userID, ok := c.Get("user_id"); ok {
		if idStr, ok := userID.(string); ok && idStr != "" {
			req.BuyerID = &idStr
		}
	}

	created, err := h.deals.Create(c.Request.Context(), req)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create request")
		return
	}
	respondCreated(c, created)
}

func (h *ListingHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid listing id")
		return
	}
	current, err := h.listings.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "listing not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to load listing")
		return
	}
	userID, _ := c.Get("user_id")
	userIDStr, _ := userID.(string)
	if current.CreatedBy != nil && userIDStr != "" && *current.CreatedBy != userIDStr {
		respondError(c, http.StatusForbidden, "not allowed to delete this listing")
		return
	}
	if err := h.listings.DeleteOwned(c.Request.Context(), id, &userIDStr); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to delete listing")
		return
	}
	c.Status(http.StatusNoContent)
}

// Admin endpoints

func (h *ListingHandler) AdminListRequests(c *gin.Context) {
	filter := ports.DealRequestFilter{
		Limit:  parseIntDefault(c.Query("limit"), 50),
		Offset: parseIntDefault(c.Query("offset"), 0),
	}
	if statusParam := strings.TrimSpace(c.Query("status")); statusParam != "" {
		for _, s := range strings.Split(statusParam, ",") {
			filter.Status = append(filter.Status, domain.DealRequestStatus(strings.ToUpper(strings.TrimSpace(s))))
		}
	}
	items, err := h.deals.ListAdmin(c.Request.Context(), filter)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load requests")
		return
	}
	respondOK(c, items)
}

func (h *ListingHandler) AdminListListings(c *gin.Context) {
	filter := buildListingFilter(c)
	filter.Limit = parseIntDefault(c.Query("limit"), 50)
	filter.Offset = parseIntDefault(c.Query("offset"), 0)
	if strings.TrimSpace(c.Query("status")) == "" {
		filter.Status = nil
	}

	items, err := h.listings.ListAdmin(c.Request.Context(), filter)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load ads")
		return
	}

	for i := range items {
		if items[i].CreatedBy == nil || strings.TrimSpace(*items[i].CreatedBy) == "" {
			continue
		}
		user, err := h.users.GetByID(c.Request.Context(), *items[i].CreatedBy)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			respondError(c, http.StatusInternalServerError, "failed to load ad author")
			return
		}
		info := user.SafeInfo()
		items[i].Author = &info
	}

	respondOK(c, items)
}

func (h *ListingHandler) AdminListUsers(c *gin.Context) {
	limit := parseIntDefault(c.Query("limit"), 50)
	offset := parseIntDefault(c.Query("offset"), 0)

	items, err := h.users.List(c.Request.Context(), limit, offset)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load users")
		return
	}
	total, err := h.users.Count(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to load users count")
		return
	}

	users := make([]domain.UserInfo, 0, len(items))
	for _, u := range items {
		users = append(users, u.SafeInfo())
	}

	respondOK(c, gin.H{
		"total":  total,
		"limit":  limit,
		"offset": offset,
		"items":  users,
	})
}

func (h *ListingHandler) AdminGetRequest(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid request id")
		return
	}
	item, err := h.deals.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "request not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to load request")
		return
	}
	respondOK(c, item)
}

type requestStatusPayload struct {
	Status    string `json:"status"`
	MeetingAt string `json:"meeting_at"` // RFC3339 optional
	AdminNote string `json:"admin_note"`
}

func (h *ListingHandler) AdminUpdateRequestStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid request id")
		return
	}
	var payload requestStatusPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}

	status := domain.DealRequestStatus(strings.ToUpper(payload.Status))
	if status == "" {
		respondError(c, http.StatusBadRequest, "status is required")
		return
	}

	var meeting *time.Time
	if strings.TrimSpace(payload.MeetingAt) != "" {
		t, err := time.Parse(time.RFC3339, payload.MeetingAt)
		if err != nil {
			respondError(c, http.StatusBadRequest, "invalid meeting_at, must be RFC3339")
			return
		}
		meeting = &t
	}

	var adminNote *string
	if strings.TrimSpace(payload.AdminNote) != "" {
		n := payload.AdminNote
		adminNote = &n
	}

	if err := h.deals.UpdateStatus(c.Request.Context(), id, status, meeting, adminNote, nil); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to update status")
		return
	}
	c.Status(http.StatusNoContent)
}

// helpers

func buildListingFilter(c *gin.Context) ports.ListingFilter {
	filter := ports.ListingFilter{
		StoneType: c.Query("stone_type"),
		Form:      c.Query("form"),
		Province:  c.Query("province"),
		City:      c.Query("city"),
		Query:     c.Query("q"),
		Sort:      c.Query("sort"),
		Limit:     parseIntDefault(c.Query("limit"), 20),
		Offset:    parseIntDefault(c.Query("offset"), 0),
	}
	if v := parseFloatPointer(c.Query("min_tonnage")); v != nil {
		filter.MinTonnage = v
	}
	if v := parseFloatPointer(c.Query("max_tonnage")); v != nil {
		filter.MaxTonnage = v
	}
	if v := parseFloatPointer(c.Query("min_price")); v != nil {
		filter.MinPrice = v
	}
	if v := parseFloatPointer(c.Query("max_price")); v != nil {
		filter.MaxPrice = v
	}
	if statusParam := strings.TrimSpace(c.Query("status")); statusParam != "" {
		filter.Status = strings.Split(statusParam, ",")
	}
	return filter
}

func buildListingFromPayload(p listingPayload) domain.Listing {
	listing := domain.Listing{
		Title:       p.Title,
		StoneType:   p.StoneType,
		Form:        p.Form,
		Tonnage:     p.Tonnage,
		Province:    p.Province,
		City:        p.City,
		PriceAmount: p.PriceAmount,
		PriceUnit:   p.PriceUnit,
		Description: p.Description,
		ExtraProps:  p.ExtraProps,
		Status:      p.Status,
	}
	for i, url := range p.Images {
		if strings.TrimSpace(url) == "" {
			continue
		}
		listing.Images = append(listing.Images, domain.ListingImage{
			ImageURL: url,
			Position: i,
		})
	}
	return listing
}

func parseFloatPointer(raw string) *float64 {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil
	}
	return &val
}

func parseIntDefault(raw string, def int) int {
	if strings.TrimSpace(raw) == "" {
		return def
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return val
}
