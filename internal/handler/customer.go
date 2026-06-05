package handler

import (
"errors"
"log/slog"
"net/http"
"strconv"

"goshen/backend/pkg/response"

"github.com/go-chi/chi/v5"
)

func (h *Handler) ListCustomers(w http.ResponseWriter, r *http.Request) {
items, err := h.customers.List(r.Context())
if err != nil {
response.InternalError(w)
return
}
response.OK(w, items)
}

func (h *Handler) GetCustomer(w http.ResponseWriter, r *http.Request) {
id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
if err != nil {
response.BadRequest(w, "invalid id")
return
}
c, err := h.customers.GetByID(r.Context(), id)
if err != nil {
response.NotFound(w, "customer not found")
return
}
response.OK(w, c)
}

func (h *Handler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
if err := r.ParseMultipartForm(maxUploadSize); err != nil {
response.BadRequest(w, "invalid multipart form")
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
if imageURL == "" {
imageURL = r.FormValue("image_url")
}
if imageURL == "" {
response.BadRequest(w, "image is required")
return
}

altText := r.FormValue("alt_text")

c, err := h.customers.Create(r.Context(), imageURL, altText)
if err != nil {
response.InternalError(w)
return
}
response.Created(w, c)
}

func (h *Handler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
if err != nil {
response.BadRequest(w, "invalid id")
return
}
if err := r.ParseMultipartForm(maxUploadSize); err != nil {
response.BadRequest(w, "invalid multipart form")
return
}

existing, err := h.customers.GetByID(r.Context(), id)
if err != nil {
response.NotFound(w, "customer not found")
return
}

altText := r.FormValue("alt_text")
if altText == "" {
altText = existing.AltText
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

c, err := h.customers.Update(r.Context(), id, imageURL, altText)
if err != nil {
response.NotFound(w, "customer not found")
return
}
response.OK(w, c)
}

func (h *Handler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
if err != nil {
response.BadRequest(w, "invalid id")
return
}
existing, err := h.customers.GetByID(r.Context(), id)
if err != nil {
response.NotFound(w, "customer not found")
return
}
if err := h.customers.Delete(r.Context(), id); err != nil {
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
