package handler

import (
	"encoding/json"
	"net/http"

	"goshen/backend/pkg/response"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListSupportCards(w http.ResponseWriter, r *http.Request) {
	items, err := h.supportCards.List(r.Context())
	if err != nil {
		response.InternalError(w)
		return
	}
	response.OK(w, items)
}

func (h *Handler) CreateSupportCard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		CtaLabel    string `json:"cta_label"`
		CtaHref     string `json:"cta_href"`
		SortOrder   int    `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.Title == "" {
		response.BadRequest(w, "title is required")
		return
	}
	card, err := h.supportCards.Create(r.Context(), req.Title, req.Description, req.CtaLabel, req.CtaHref, req.SortOrder)
	if err != nil {
		response.InternalError(w)
		return
	}
	response.Created(w, card)
}

func (h *Handler) DeleteSupportCard(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "id is required")
		return
	}
	if err := h.supportCards.Delete(r.Context(), id); err != nil {
		response.NotFound(w, "support card not found")
		return
	}
	response.OK(w, map[string]bool{"deleted": true})
}
