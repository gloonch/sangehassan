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

func NewUploadHandler(baseDir string) *UploadHandler {
	return &UploadHandler{baseDir: baseDir}
}

func (h *UploadHandler) UploadTemplate(c *gin.Context) {
	h.uploadFile(c, "templates")
}

func (h *UploadHandler) UploadProduct(c *gin.Context) {
	h.uploadFile(c, "products")
}

func (h *UploadHandler) uploadFile(c *gin.Context, subdir string) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		respondError(c, http.StatusBadRequest, "file is required")
		return
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !isAllowedExt(ext) {
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
	respondOK(c, gin.H{"image_url": publicPath})
}

func isAllowedExt(ext string) bool {
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp":
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
