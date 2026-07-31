package handlers

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"sangehassan/back/internal/usecase"
)

const maxWorkflowFileBytes = 15 << 20

type WorkflowFileHandler struct {
	service *usecase.OperationsService
	baseDir string
}

type storedWorkflowUpload struct {
	storageKey, originalName, mimeType, path string
	size                                     int64
}

func (h *WorkflowFileHandler) persistUpload(c *gin.Context, policy usecase.WorkflowUploadPolicy) (storedWorkflowUpload, error) {
	var out storedWorkflowUpload
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxWorkflowFileBytes+(1<<20))
	header, err := c.FormFile("file")
	if err != nil || header.Size <= 0 || header.Size > policy.MaxSizeBytes {
		return out, errors.New("invalid file")
	}
	source, err := header.Open()
	if err != nil {
		return out, err
	}
	defer source.Close()
	buffer := make([]byte, 512)
	n, err := io.ReadFull(source, buffer)
	if err != nil && err != io.ErrUnexpectedEOF {
		return out, err
	}
	detected := http.DetectContentType(buffer[:n])
	extensions := map[string]string{"image/png": ".png", "image/jpeg": ".jpg", "application/pdf": ".pdf"}
	extension, known := extensions[detected]
	allowed := false
	for _, value := range policy.AllowedMIMETypes {
		if value == detected {
			allowed = true
			break
		}
	}
	if !known || !allowed {
		return out, errors.New("unsupported file type")
	}
	if err = os.MkdirAll(h.baseDir, 0o750); err != nil {
		return out, err
	}
	out.storageKey = randomToken(20) + extension
	out.path = filepath.Join(h.baseDir, out.storageKey)
	destination, err := os.OpenFile(out.path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return out, err
	}
	written, copyErr := io.Copy(destination, io.MultiReader(bytes.NewReader(buffer[:n]), source))
	closeErr := destination.Close()
	if copyErr != nil || closeErr != nil || written > policy.MaxSizeBytes {
		_ = os.Remove(out.path)
		return out, errors.New("cannot save file")
	}
	out.originalName = filepath.Base(strings.ReplaceAll(header.Filename, "\\", "/"))
	out.mimeType = detected
	out.size = written
	return out, nil
}

func NewWorkflowFileHandler(service *usecase.OperationsService, baseDir string) *WorkflowFileHandler {
	return &WorkflowFileHandler{service: service, baseDir: baseDir}
}

func (h *WorkflowFileHandler) Upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxWorkflowFileBytes+(1<<20))
	fieldID, err := strconv.ParseInt(c.PostForm("field_definition_id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "field_definition_id is required")
		return
	}
	workflowID, customerVisible, policy, err := h.service.PrepareWorkflowFileUpload(c.Request.Context(), actorID(c), c.Param("id"), fieldID)
	if err != nil {
		operationError(c, err)
		return
	}
	header, err := c.FormFile("file")
	if err != nil {
		respondError(c, http.StatusBadRequest, "file is required")
		return
	}
	if header.Size <= 0 || header.Size > policy.MaxSizeBytes {
		respondError(c, http.StatusRequestEntityTooLarge, "file is too large")
		return
	}
	source, err := header.Open()
	if err != nil {
		respondError(c, http.StatusBadRequest, "cannot open file")
		return
	}
	defer source.Close()
	buffer := make([]byte, 512)
	n, err := io.ReadFull(source, buffer)
	if err != nil && err != io.ErrUnexpectedEOF {
		respondError(c, http.StatusBadRequest, "cannot inspect file")
		return
	}
	detected := http.DetectContentType(buffer[:n])
	extensions := map[string]string{"image/png": ".png", "image/jpeg": ".jpg", "application/pdf": ".pdf"}
	extension, known := extensions[detected]
	allowed := false
	for _, allowedMIME := range policy.AllowedMIMETypes {
		if detected == allowedMIME {
			allowed = true
			break
		}
	}
	if !known || !allowed {
		respondError(c, http.StatusBadRequest, "unsupported file type")
		return
	}
	if err = os.MkdirAll(h.baseDir, 0o750); err != nil {
		respondError(c, http.StatusInternalServerError, "cannot prepare private storage")
		return
	}
	storageKey := randomToken(20) + extension
	path := filepath.Join(h.baseDir, storageKey)
	destination, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "cannot save file")
		return
	}
	written, copyErr := io.Copy(destination, io.MultiReader(bytes.NewReader(buffer[:n]), source))
	closeErr := destination.Close()
	if copyErr != nil || closeErr != nil || written > policy.MaxSizeBytes {
		_ = os.Remove(path)
		respondError(c, http.StatusInternalServerError, "cannot save file")
		return
	}
	originalName := filepath.Base(strings.ReplaceAll(header.Filename, "\\", "/"))
	file, err := h.service.RegisterWorkflowFile(c.Request.Context(), actorID(c), workflowID, c.Param("id"), fieldID, storageKey, originalName, detected, written, customerVisible)
	if err != nil {
		_ = os.Remove(path)
		operationError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": file})
}

func (h *WorkflowFileHandler) Download(c *gin.Context) {
	file, err := h.service.WorkflowFileForDownload(c.Request.Context(), actorID(c), c.Param("id"))
	if err != nil {
		operationError(c, err)
		return
	}
	if filepath.Base(file.StorageKey) != file.StorageKey {
		respondError(c, http.StatusNotFound, "file not found")
		return
	}
	path := filepath.Join(h.baseDir, file.StorageKey)
	handle, err := os.Open(path)
	if err != nil {
		respondError(c, http.StatusNotFound, "file not found")
		return
	}
	defer handle.Close()
	headers := map[string]string{
		"Content-Disposition":    mime.FormatMediaType("inline", map[string]string{"filename": file.OriginalFileName}),
		"X-Content-Type-Options": "nosniff",
		"Cache-Control":          "private, no-store",
	}
	c.DataFromReader(http.StatusOK, file.SizeBytes, file.MIMEType, handle, headers)
}

func (h *WorkflowFileHandler) UploadEntity(c *gin.Context) {
	visible, _ := strconv.ParseBool(c.PostForm("customer_visible"))
	entityType, entityID := c.Param("entityType"), c.Param("entityId")
	workflowID, visible, policy, err := h.service.PrepareOperationalFileUpload(c.Request.Context(), actorID(c), entityType, entityID, visible)
	if err != nil {
		operationError(c, err)
		return
	}
	stored, err := h.persistUpload(c, policy)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	file, err := h.service.RegisterOperationalFile(c.Request.Context(), actorID(c), workflowID, entityType, entityID, stored.storageKey, stored.originalName, stored.mimeType, stored.size, visible)
	if err != nil {
		_ = os.Remove(stored.path)
		operationError(c, err)
		return
	}
	respondCreated(c, file)
}
