package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	baseDir string
}

const (
	maxImageUploadBytes = 25 << 20
	maxMediaUploadBytes = 250 << 20
)

func NewUploadHandler(baseDir string) *UploadHandler {
	return &UploadHandler{baseDir: baseDir}
}

func (h *UploadHandler) UploadTemplate(c *gin.Context) {
	h.uploadFile(c, "templates", false)
}

func (h *UploadHandler) UploadProduct(c *gin.Context) {
	h.uploadFile(c, "products", true)
}

func (h *UploadHandler) UploadBlock(c *gin.Context) {
	h.uploadFile(c, "blocks", false)
}

func (h *UploadHandler) UploadContent(c *gin.Context) {
	h.uploadFile(c, "content", false)
}

func (h *UploadHandler) UploadTeam(c *gin.Context) {
	h.uploadFile(c, "team", false)
}

func (h *UploadHandler) UploadBlog(c *gin.Context) {
	h.uploadFile(c, "blogs", false)
}

func (h *UploadHandler) UploadProject(c *gin.Context) {
	h.uploadFile(c, "projects", true)
}

func (h *UploadHandler) UploadListing(c *gin.Context) {
	h.uploadFile(c, "listings", false)
}

func (h *UploadHandler) uploadFile(c *gin.Context, subdir string, allowVideo bool) {
	limit := int64(maxImageUploadBytes)
	if allowVideo {
		limit = maxMediaUploadBytes
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "too large") {
			respondError(c, http.StatusRequestEntityTooLarge, "file is too large")
			return
		}
		respondError(c, http.StatusBadRequest, "file is required")
		return
	}
	if fileHeader.Size > limit {
		respondError(c, http.StatusRequestEntityTooLarge, "file is too large")
		return
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	isVideo := isAllowedVideoExt(ext)
	if !isAllowedImageExt(ext) && !(allowVideo && isVideo) {
		respondError(c, http.StatusBadRequest, "unsupported file type")
		return
	}

	filename := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), randomToken(4), ext)
	destDir := filepath.Join(h.baseDir, subdir)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create directory")
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to open file")
		return
	}
	defer src.Close()

	destPath := filepath.Join(destDir, filename)
	dst, err := os.Create(destPath)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to save file")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to save file")
		return
	}

	publicPath := fmt.Sprintf("/images/%s/%s", subdir, filename)
	payload := gin.H{"file_url": publicPath}
	if isVideo {
		payload["video_url"] = publicPath
	} else {
		payload["image_url"] = publicPath
	}
	respondOK(c, payload)
}

func isAllowedImageExt(ext string) bool {
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp":
		return true
	default:
		return false
	}
}

func isAllowedVideoExt(ext string) bool {
	switch ext {
	case ".mp4", ".webm", ".mov", ".m4v":
		return true
	default:
		return false
	}
}

func randomToken(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "rand"
	}
	return hex.EncodeToString(buf)
}
