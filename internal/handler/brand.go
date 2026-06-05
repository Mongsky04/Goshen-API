package handler

import (
"errors"
"log/slog"
"net/http"
"strconv"

"goshen/backend/pkg/response"

"github.com/go-chi/chi/v5"
)

func (h *Handler) ListBrands(w http.ResponseWriter, r *http.Request) {
items, err := h.brands.List(r.Context())
if err != nil {
response.InternalError(w)
return
}
response.OK(w, items)
}

func (h *Handler) GetBrand(w http.ResponseWriter, r *http.Request) {
id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
if err != nil {
response.BadRequest(w, "invalid id")
return
}
b, err := h.brands.GetByID(r.Context(), id)
if err != nil {
response.NotFound(w, "brand not found")
return
}
response.OK(w, b)
}

func (h *Handler) CreateBrand(w http.ResponseWriter, r *http.Request) {
if err := r.ParseMultipartForm(maxUploadSize); err != nil {
response.BadRequest(w, "invalid multipart form")
return
}
name := r.FormValue("name")
if name == "" {
response.BadRequest(w, "name is required")
return
}

imageURL, err := h.uploadFormImage(r, "image")
if errors.Is(err, errUnsupportedFileType) {
response.BadRequest(w, "unsupported file type: only JPEG, PNG, WebP, GIF, SVG allowed")
return
}
if err != nil {
response.InternalError(w)
return
}

b, err := h.brands.Create(r.Context(), name, imageURL)
if err != nil {
response.InternalError(w)
return
}
response.Created(w, b)
}

func (h *Handler) UpdateBrand(w http.ResponseWriter, r *http.Request) {
id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
if err != nil {
response.BadRequest(w, "invalid id")
return
}
if err := r.ParseMultipartForm(maxUploadSize); err != nil {
response.BadRequest(w, "invalid multipart form")
return
}

existing, err := h.brands.GetByID(r.Context(), id)
if err != nil {
response.NotFound(w, "brand not found")
return
}

name := r.FormValue("name")
if name == "" {
name = existing.Name
}

imageURL := existing.ImageURL
if _, _, err := r.FormFile("image"); err == nil {
newURL, uploadErr := h.uploadFormImage(r, "image")
if errors.Is(uploadErr, errUnsupportedFileType) {
response.BadRequest(w, "unsupported file type: only JPEG, PNG, WebP, GIF, SVG allowed")
return
}
if uploadErr != nil {
response.InternalError(w)
return
}
go func() {
if err := h.storage.Delete(existing.ImageURL); err != nil {
slog.Error("storage: failed to delete image", "err", err)
}
}()
imageURL = newURL
}

b, err := h.brands.Update(r.Context(), id, name, imageURL)
if err != nil {
response.NotFound(w, "brand not found")
return
}
response.OK(w, b)
}

func (h *Handler) DeleteBrand(w http.ResponseWriter, r *http.Request) {
id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
if err != nil {
response.BadRequest(w, "invalid id")
return
}
existing, err := h.brands.GetByID(r.Context(), id)
if err != nil {
response.NotFound(w, "brand not found")
return
}
if err := h.brands.Delete(r.Context(), id); err != nil {
response.InternalError(w)
return
}
go func() {
if err := h.storage.Delete(existing.ImageURL); err != nil {
slog.Error("storage: failed to delete image", "err", err)
}
}()
response.OK(w, map[string]bool{"deleted": true})
}
