package handler

import (
"errors"
"log/slog"
"net/http"
"strconv"

"goshen/backend/pkg/response"

"github.com/go-chi/chi/v5"
)

const maxUploadSize = 10 << 20 // 10 MB

// parsePage reads ?page= and ?limit= query params with safe defaults and clamping.
// Returns page (1-based), limit, and offset.
func parsePage(r *http.Request) (page, limit, offset int) {
	page = 1
	limit = 20
	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 {
		if l > 100 {
			l = 100
		}
		limit = l
	}
	offset = (page - 1) * limit
	return
}

type pagedResponse struct {
	Data  interface{} `json:"data"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
	Total *int        `json:"total"`
}

func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
page, limit, offset := parsePage(r)
products, err := h.products.ListPaged(r.Context(), limit, offset)
if err != nil {
response.InternalError(w)
return
}
response.OK(w, pagedResponse{Data: products, Page: page, Limit: limit, Total: nil})
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
if err != nil {
response.BadRequest(w, "invalid id")
return
}
p, err := h.products.GetByID(r.Context(), id)
if err != nil {
response.NotFound(w, "product not found")
return
}
response.OK(w, p)
}

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
if err := r.ParseMultipartForm(maxUploadSize); err != nil {
response.BadRequest(w, "invalid multipart form")
return
}
name := r.FormValue("name")
if name == "" {
response.BadRequest(w, "name is required")
return
}
category    := r.FormValue("category")
subCategory := r.FormValue("sub_category")

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

p, err := h.products.Create(r.Context(), name, imageURL, category, subCategory)
if err != nil {
response.InternalError(w)
return
}
response.Created(w, p)
}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
if err != nil {
response.BadRequest(w, "invalid id")
return
}
if err := r.ParseMultipartForm(maxUploadSize); err != nil {
response.BadRequest(w, "invalid multipart form")
return
}

existing, err := h.products.GetByID(r.Context(), id)
if err != nil {
response.NotFound(w, "product not found")
return
}

name := r.FormValue("name")
if name == "" {
name = existing.Name
}
category := r.FormValue("category")
if category == "" {
category = existing.Category
}
subCategory := r.FormValue("sub_category")
if subCategory == "" {
subCategory = existing.SubCategory
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

p, err := h.products.Update(r.Context(), id, name, imageURL, category, subCategory)
if err != nil {
response.NotFound(w, "product not found")
return
}
response.OK(w, p)
}

func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
if err != nil {
response.BadRequest(w, "invalid id")
return
}
existing, err := h.products.GetByID(r.Context(), id)
if err != nil {
response.NotFound(w, "product not found")
return
}
if err := h.products.Delete(r.Context(), id); err != nil {
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
