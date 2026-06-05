package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"goshen/backend/pkg/response"

	"github.com/go-chi/chi/v5"
)

func jsonDecode(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func (h *Handler) ListFeatured(w http.ResponseWriter, r *http.Request) {
	page, limit, offset := parsePage(r)
	items, err := h.featured.ListPaged(r.Context(), limit, offset)
	if err != nil {
		response.InternalError(w)
		return
	}
	response.OK(w, pagedResponse{Data: items, Page: page, Limit: limit, Total: nil})
}

func (h *Handler) GetFeatured(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}
	f, err := h.featured.GetByID(r.Context(), id)
	if err != nil {
		response.NotFound(w, "featured item not found")
		return
	}
	response.OK(w, f)
}

func (h *Handler) CreateFeatured(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		response.BadRequest(w, "invalid multipart form")
		return
	}
	productIDStr := r.FormValue("product_id")
	if productIDStr == "" {
		response.BadRequest(w, "product_id is required")
		return
	}
	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid product_id")
		return
	}
	featuredCategories := parseTags(r.FormValue("featured_categories"))

	f, err := h.featured.Create(r.Context(), productID, featuredCategories)
	if err != nil {
		response.InternalError(w)
		return
	}
	response.Created(w, f)
}

func (h *Handler) UpdateFeatured(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		response.BadRequest(w, "invalid multipart form")
		return
	}
	existing, err := h.featured.GetByID(r.Context(), id)
	if err != nil {
		response.NotFound(w, "featured item not found")
		return
	}

	productID := existing.ProductID
	if v := r.FormValue("product_id"); v != "" {
		if pid, err := strconv.ParseInt(v, 10, 64); err == nil {
			productID = pid
		}
	}

	featuredCategoriesRaw := r.FormValue("featured_categories")
	featuredCategories := existing.FeaturedCategories
	if featuredCategoriesRaw != "" {
		featuredCategories = parseTags(featuredCategoriesRaw)
	}

	f, err := h.featured.Update(r.Context(), id, productID, featuredCategories)
	if err != nil {
		response.NotFound(w, "featured item not found")
		return
	}
	response.OK(w, f)
}

func (h *Handler) DeleteFeatured(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}
	if err := h.featured.Delete(r.Context(), id); err != nil {
		response.InternalError(w)
		return
	}
	response.OK(w, map[string]bool{"deleted": true})
}

// ListHomepageGrid handles GET /api/v1/homepage-grid (public)
func (h *Handler) ListHomepageGrid(w http.ResponseWriter, r *http.Request) {
	items, err := h.homepageGrid.List(r.Context())
	if err != nil {
		response.InternalError(w)
		return
	}
	response.OK(w, items)
}

// ReplaceHomepageGrid handles PUT /api/v1/homepage-grid (protected)
func (h *Handler) ReplaceHomepageGrid(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ProductIDs []int64 `json:"product_ids"`
	}
	if err := jsonDecode(r, &body); err != nil {
		response.BadRequest(w, "invalid body")
		return
	}
	if err := h.homepageGrid.ReplaceAll(r.Context(), body.ProductIDs); err != nil {
		response.InternalError(w)
		return
	}
	response.OK(w, struct{}{})
}

// parseTags splits a comma-separated tag string.
func parseTags(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == '\n'
	})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
