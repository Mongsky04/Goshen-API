package handler

import (
	"encoding/json"
	"net/http"
)

type navStatusJSON struct {
	Hidden []string `json:"hidden"`
}

func (h *Handler) GetNavStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	conferenceSlugs, err := h.conference.ListUnpublishedSlugs(ctx)
	if err != nil {
		conferenceSlugs = []string{}
	}

	performerSlugs, err := h.performer.ListUnpublishedSlugs(ctx)
	if err != nil {
		performerSlugs = []string{}
	}

	hidden := append(conferenceSlugs, performerSlugs...)
	if hidden == nil {
		hidden = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=60")
	json.NewEncoder(w).Encode(navStatusJSON{Hidden: hidden}) //nolint:errcheck
}
