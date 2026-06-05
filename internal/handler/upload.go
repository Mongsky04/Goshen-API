package handler

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"

	"goshen/backend/pkg/response"
)

var errUnsupportedFileType = errors.New("unsupported file type")

var allowedImageMIME = map[string]bool{
	"image/jpeg":    true,
	"image/png":     true,
	"image/webp":    true,
	"image/gif":     true,
	"image/svg+xml": true,
}

// UploadImage handles POST /api/v1/upload — accepts multipart "file" field,
// saves to local disk, and returns the public URL.
func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		response.BadRequest(w, "invalid multipart form")
		return
	}
	url, err := h.uploadFormImage(r, "file")
	if err != nil {
		response.InternalError(w)
		return
	}
	if url == "" {
		response.BadRequest(w, "no file provided")
		return
	}
	response.OK(w, map[string]string{"url": url})
}

// uploadFormImage saves the file from the named form field to local disk.
// Returns ("", nil) when the field is absent (no file chosen).
// Returns ("", err) on upload failure.
func (h *Handler) uploadFormImage(r *http.Request, field string) (string, error) {
	file, header, err := r.FormFile(field)
	if err != nil {
		// Field not present — caller decides whether that's an error.
		return "", nil
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(header.Filename))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if !allowedImageMIME[contentType] {
		return "", errUnsupportedFileType
	}

	url, err := h.storage.Upload(file, header.Filename, contentType)
	if err != nil {
		return "", fmt.Errorf("upload image: %w", err)
	}
	return url, nil
}
