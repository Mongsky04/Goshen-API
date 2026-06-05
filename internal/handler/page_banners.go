package handler

import (
	"encoding/json"
	"net/http"

	"goshen/backend/pkg/response"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetPageBanners(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	banners, err := h.pageBanners.ListBySlug(r.Context(), slug)
	if err != nil {
		response.InternalError(w)
		return
	}
	response.OK(w, banners)
}

func (h *Handler) ReplacePageBanners(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	var body struct {
		BannerIDs []int64 `json:"banner_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid JSON")
		return
	}
	if body.BannerIDs == nil {
		body.BannerIDs = []int64{}
	}
	if err := h.pageBanners.Replace(r.Context(), slug, body.BannerIDs); err != nil {
		response.InternalError(w)
		return
	}
	response.OK(w, map[string]bool{"updated": true})
}
