package handlers

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/color"
	stddraw "image/draw"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/image/draw"
	"golang.org/x/image/webp"
)

const defaultWatermarkOpacity = 0.18
const protectedThumbnailMaxSide = 220

//go:embed assets/sangehassan-watermark.png
var embeddedWatermark []byte

type ImageHandler struct {
	baseDir       string
	cacheDir      string
	settingsPath  string
	watermarkPath string
	opacity       float64

	watermarkOnce sync.Once
	watermark     image.Image
	watermarkErr  error
	settingsMu    sync.RWMutex
	settings      protectedImageSettings
	cacheLocks    sync.Map
}

type protectedImageSettings struct {
	WatermarkEnabled bool `json:"watermark_enabled"`
}

type protectedImageSettingsPayload struct {
	WatermarkEnabled *bool `json:"watermark_enabled"`
}

func NewImageHandler(baseDir, watermarkPath string) *ImageHandler {
	cleanBaseDir := filepath.Clean(baseDir)
	if strings.TrimSpace(watermarkPath) == "" {
		watermarkPath = filepath.Join(cleanBaseDir, "watermark", "sangehassan-watermark.png")
	}
	handler := &ImageHandler{
		baseDir:       cleanBaseDir,
		cacheDir:      filepath.Join(cleanBaseDir, ".protected-cache"),
		settingsPath:  filepath.Join(cleanBaseDir, ".protected-image-settings.json"),
		watermarkPath: filepath.Clean(watermarkPath),
		opacity:       defaultWatermarkOpacity,
		settings: protectedImageSettings{
			WatermarkEnabled: true,
		},
	}
	handler.loadSettings()
	return handler
}

func (h *ImageHandler) Serve(c *gin.Context) {
	relativePath, ok := cleanImageRequestPath(c.Param("filepath"))
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}
	if !isProtectedProductImagePath(relativePath) {
		c.Status(http.StatusNotFound)
		return
	}

	fullPath := filepath.Join(h.baseDir, relativePath)
	if !isInsideBaseDir(h.baseDir, fullPath) {
		c.Status(http.StatusNotFound)
		return
	}

	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		c.Status(http.StatusNotFound)
		return
	}

	ext := strings.ToLower(filepath.Ext(fullPath))
	variant := protectedImageVariant(c.Query("variant"))
	watermarkEnabled := h.WatermarkEnabled()
	setProtectedImageHeaders(c.Writer.Header())
	c.Header("Last-Modified", info.ModTime().UTC().Format(http.TimeFormat))

	if !isWatermarkableImageExt(ext) || shouldServeOriginal(relativePath) {
		c.Status(http.StatusNotFound)
		return
	}

	if c.Request.Method == http.MethodHead {
		c.Header("Content-Type", protectedContentType(ext, variant, watermarkEnabled))
		c.Header("ETag", h.protectedImageETag(relativePath, ext, variant, info, watermarkEnabled))
		c.Status(http.StatusOK)
		return
	}

	payload, contentType, etag, err := h.cachedProtectedImage(fullPath, relativePath, ext, variant, info, watermarkEnabled)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Header("ETag", etag)
	if etagMatches(c.GetHeader("If-None-Match"), etag) {
		c.Status(http.StatusNotModified)
		return
	}
	c.Header("Content-Length", strconv.Itoa(len(payload)))
	c.Data(http.StatusOK, contentType, payload)
}

func (h *ImageHandler) AdminSettings(c *gin.Context) {
	respondOK(c, h.Settings())
}

func (h *ImageHandler) AdminUpdateSettings(c *gin.Context) {
	var payload protectedImageSettingsPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "invalid payload")
		return
	}
	if payload.WatermarkEnabled == nil {
		respondError(c, http.StatusBadRequest, "watermark_enabled is required")
		return
	}

	settings := protectedImageSettings{
		WatermarkEnabled: *payload.WatermarkEnabled,
	}
	if err := h.UpdateSettings(settings); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to save settings")
		return
	}

	respondOK(c, settings)
}

func (h *ImageHandler) Settings() protectedImageSettings {
	h.settingsMu.RLock()
	defer h.settingsMu.RUnlock()
	return h.settings
}

func (h *ImageHandler) WatermarkEnabled() bool {
	return h.Settings().WatermarkEnabled
}

func (h *ImageHandler) UpdateSettings(next protectedImageSettings) error {
	h.settingsMu.Lock()
	defer h.settingsMu.Unlock()

	h.settings = next
	return h.writeSettingsLocked()
}

func (h *ImageHandler) loadSettings() {
	payload, err := os.ReadFile(h.settingsPath)
	if err != nil {
		return
	}

	var settings protectedImageSettings
	if err := json.Unmarshal(payload, &settings); err != nil {
		return
	}

	h.settingsMu.Lock()
	h.settings = settings
	h.settingsMu.Unlock()
}

func (h *ImageHandler) writeSettingsLocked() error {
	payload, err := json.MarshalIndent(h.settings, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(h.settingsPath), 0755); err != nil {
		return err
	}
	tmpFile, err := os.CreateTemp(filepath.Dir(h.settingsPath), ".protected-image-settings-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	if _, err := tmpFile.Write(payload); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, h.settingsPath)
}

func isProtectedProductImagePath(relativePath string) bool {
	firstSegment := strings.Split(filepath.ToSlash(relativePath), "/")[0]
	return firstSegment == "products"
}

func cleanImageRequestPath(value string) (string, bool) {
	value = strings.TrimPrefix(value, "/")
	if value == "" || strings.Contains(value, "\x00") {
		return "", false
	}
	cleaned := filepath.Clean(value)
	if cleaned == "." || strings.HasPrefix(cleaned, "..") || filepath.IsAbs(cleaned) {
		return "", false
	}
	return cleaned, true
}

func isInsideBaseDir(baseDir, fullPath string) bool {
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return false
	}
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(absBase, absPath)
	return err == nil && rel != "." && !strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel)
}

func setProtectedImageHeaders(headers http.Header) {
	headers.Set("Content-Disposition", "inline")
	headers.Set("Cache-Control", "private, no-cache, max-age=0")
	headers.Set("Cross-Origin-Resource-Policy", "same-origin")
	headers.Set("X-Content-Type-Options", "nosniff")
	headers.Set("Referrer-Policy", "no-referrer")
}

func isWatermarkableImageExt(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	default:
		return false
	}
}

func shouldServeOriginal(relativePath string) bool {
	firstSegment := strings.Split(filepath.ToSlash(relativePath), "/")[0]
	switch firstSegment {
	case "watermark":
		return true
	default:
		return false
	}
}

func contentTypeForExt(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mov":
		return "video/quicktime"
	default:
		return "application/octet-stream"
	}
}

func protectedImageVariant(value string) string {
	if value == "thumb" {
		return "thumb"
	}
	return "full"
}

func protectedContentType(ext, variant string, watermarkEnabled bool) string {
	if variant != "thumb" && !watermarkEnabled {
		return contentTypeForExt(ext)
	}
	if ext == ".png" && variant != "thumb" {
		return "image/png"
	}
	return "image/jpeg"
}

func protectedCacheExt(ext, variant string, watermarkEnabled bool) string {
	if protectedContentType(ext, variant, watermarkEnabled) == "image/png" {
		return ".png"
	}
	return ".jpg"
}

func (h *ImageHandler) cachedProtectedImage(path, relativePath, ext, variant string, info os.FileInfo, watermarkEnabled bool) ([]byte, string, string, error) {
	cacheKey := h.protectedImageCacheKey(relativePath, ext, variant, info, watermarkEnabled)
	etag := `"` + cacheKey + `"`
	contentType := protectedContentType(ext, variant, watermarkEnabled)

	if variant != "thumb" && !watermarkEnabled {
		payload, err := os.ReadFile(path)
		if err != nil {
			return nil, "", "", err
		}
		return payload, contentType, etag, nil
	}

	cachePath := filepath.Join(h.cacheDir, cacheKey+protectedCacheExt(ext, variant, watermarkEnabled))

	if payload, err := os.ReadFile(cachePath); err == nil {
		return payload, contentType, etag, nil
	}

	lockValue, _ := h.cacheLocks.LoadOrStore(cacheKey, &sync.Mutex{})
	lock := lockValue.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()

	if payload, err := os.ReadFile(cachePath); err == nil {
		return payload, contentType, etag, nil
	}

	payload, contentType, err := h.protectedImage(path, ext, variant, watermarkEnabled)
	if err != nil {
		return nil, "", "", err
	}
	if err := os.MkdirAll(h.cacheDir, 0755); err != nil {
		return payload, contentType, etag, nil
	}

	tmpFile, err := os.CreateTemp(h.cacheDir, cacheKey+"-*")
	if err != nil {
		return payload, contentType, etag, nil
	}
	tmpPath := tmpFile.Name()
	if _, writeErr := tmpFile.Write(payload); writeErr != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return payload, contentType, etag, nil
	}
	if closeErr := tmpFile.Close(); closeErr != nil {
		os.Remove(tmpPath)
		return payload, contentType, etag, nil
	}
	if err := os.Rename(tmpPath, cachePath); err != nil {
		os.Remove(tmpPath)
	}

	return payload, contentType, etag, nil
}

func (h *ImageHandler) protectedImageETag(relativePath, ext, variant string, info os.FileInfo, watermarkEnabled bool) string {
	return `"` + h.protectedImageCacheKey(relativePath, ext, variant, info, watermarkEnabled) + `"`
}

func (h *ImageHandler) protectedImageCacheKey(relativePath, ext, variant string, info os.FileInfo, watermarkEnabled bool) string {
	parts := []string{
		filepath.ToSlash(relativePath),
		ext,
		variant,
		strconv.FormatBool(watermarkEnabled),
		strconv.FormatInt(info.Size(), 10),
		strconv.FormatInt(info.ModTime().UnixNano(), 10),
		strconv.FormatFloat(h.opacity, 'f', 4, 64),
		strconv.Itoa(protectedThumbnailMaxSide),
		h.watermarkSignature(),
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:])
}

func (h *ImageHandler) watermarkSignature() string {
	if info, err := os.Stat(h.watermarkPath); err == nil && !info.IsDir() {
		parts := []string{
			filepath.ToSlash(h.watermarkPath),
			strconv.FormatInt(info.Size(), 10),
			strconv.FormatInt(info.ModTime().UnixNano(), 10),
		}
		sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
		return hex.EncodeToString(sum[:])
	}

	sum := sha256.Sum256(embeddedWatermark)
	return hex.EncodeToString(sum[:])
}

func etagMatches(headerValue, etag string) bool {
	if headerValue == "" {
		return false
	}
	for _, value := range strings.Split(headerValue, ",") {
		value = strings.TrimSpace(value)
		if value == "*" || value == etag {
			return true
		}
	}
	return false
}

func (h *ImageHandler) protectedImage(path, ext, variant string, watermarkEnabled bool) ([]byte, string, error) {
	src, err := decodeImage(path, ext)
	if err != nil {
		return nil, "", err
	}
	isThumbnail := variant == "thumb"
	if isThumbnail {
		src = resizeImageToMaxSide(src, protectedThumbnailMaxSide)
	}

	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	stddraw.Draw(dst, bounds, src, bounds.Min, stddraw.Src)

	if watermarkEnabled {
		watermark, err := h.loadWatermark()
		if err != nil {
			return nil, "", err
		}

		scaledWatermark := scaleWatermark(watermark, bounds.Dx(), bounds.Dy())
		wmBounds := scaledWatermark.Bounds()
		wmX := bounds.Min.X + (bounds.Dx()-wmBounds.Dx())/2
		wmY := bounds.Min.Y + (bounds.Dy()-wmBounds.Dy())/2
		target := image.Rect(wmX, wmY, wmX+wmBounds.Dx(), wmY+wmBounds.Dy())
		stddraw.Draw(dst, target, scaledWatermark, wmBounds.Min, stddraw.Over)
	}

	var output bytes.Buffer
	if ext == ".png" && !isThumbnail {
		if err := png.Encode(&output, dst); err != nil {
			return nil, "", err
		}
		return output.Bytes(), "image/png", nil
	}

	quality := 86
	if isThumbnail {
		quality = 72
	}
	if err := jpeg.Encode(&output, dst, &jpeg.Options{Quality: quality}); err != nil {
		return nil, "", err
	}
	return output.Bytes(), "image/jpeg", nil
}

func decodeImage(path, ext string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	switch ext {
	case ".jpg", ".jpeg":
		return jpeg.Decode(file)
	case ".png":
		return png.Decode(file)
	case ".webp":
		return webp.Decode(file)
	default:
		img, _, err := image.Decode(file)
		return img, err
	}
}

func (h *ImageHandler) loadWatermark() (image.Image, error) {
	h.watermarkOnce.Do(func() {
		file, err := os.Open(h.watermarkPath)
		if err != nil {
			img, _, decodeErr := image.Decode(bytes.NewReader(embeddedWatermark))
			if decodeErr != nil {
				h.watermarkErr = err
				return
			}
			h.watermark = normalizeWatermark(img, h.opacity)
			return
		}
		defer file.Close()

		img, _, err := image.Decode(file)
		if err != nil {
			h.watermarkErr = err
			return
		}
		h.watermark = normalizeWatermark(img, h.opacity)
	})

	return h.watermark, h.watermarkErr
}

func normalizeWatermark(src image.Image, opacity float64) *image.NRGBA {
	bounds := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	alphaScale := clampFloat(opacity, 0.02, 0.35)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := src.At(x, y).RGBA()
			if a == 0 || (r > 0xf200 && g > 0xf200 && b > 0xf200) {
				continue
			}

			alpha := uint8(float64(a>>8) * alphaScale)
			if alpha == 0 {
				continue
			}
			dst.SetNRGBA(x-bounds.Min.X, y-bounds.Min.Y, color.NRGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: alpha,
			})
		}
	}

	return dst
}

func scaleWatermark(src image.Image, targetWidth, targetHeight int) *image.NRGBA {
	srcBounds := src.Bounds()
	if srcBounds.Dx() == 0 || srcBounds.Dy() == 0 || targetWidth <= 0 || targetHeight <= 0 {
		return image.NewNRGBA(image.Rect(0, 0, 1, 1))
	}

	maxWidth := int(float64(targetWidth) * 0.72)
	maxHeight := int(float64(targetHeight) * 0.72)
	if maxWidth < 1 {
		maxWidth = 1
	}
	if maxHeight < 1 {
		maxHeight = 1
	}

	width := maxWidth
	height := int(float64(width) * float64(srcBounds.Dy()) / float64(srcBounds.Dx()))
	if height > maxHeight {
		height = maxHeight
		width = int(float64(height) * float64(srcBounds.Dx()) / float64(srcBounds.Dy()))
	}
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	dst := image.NewNRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, srcBounds, draw.Over, nil)
	return dst
}

func resizeImageToMaxSide(src image.Image, maxSide int) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 || maxSide <= 0 || (width <= maxSide && height <= maxSide) {
		return src
	}

	newWidth := maxSide
	newHeight := int(float64(newWidth) * float64(height) / float64(width))
	if height > width {
		newHeight = maxSide
		newWidth = int(float64(newHeight) * float64(width) / float64(height))
	}
	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)
	return dst
}

func clampFloat(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
