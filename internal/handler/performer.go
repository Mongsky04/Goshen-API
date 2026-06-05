package handler

import (
	"encoding/json"
	"net/http"

	"goshen/backend/internal/model"
	"goshen/backend/pkg/response"

	"github.com/go-chi/chi/v5"
)

// JSON shapes — match PerformerPageForm / PerformerPageData on the frontend

type performerCellJSON struct {
	ProductID   int64  `json:"productId"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	SubCategory string `json:"subCategory"`
	ImageURL    string `json:"imageUrl"`
	IsHidden    bool   `json:"isHidden"`
	SortOrder   int    `json:"sortOrder"`
}

type performerVideoJSON struct {
	IsMain       bool   `json:"isMain"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	ThumbnailURL string `json:"thumbnailUrl"`
	VideoURL     string `json:"videoUrl"`
	SortOrder    int    `json:"sortOrder"`
}

type performerAdminJSON struct {
	Slug               string              `json:"slug"`
	Label              string              `json:"label"`
	IsPublished        bool                `json:"isPublished"`
	HeroImageURL       string              `json:"heroImageUrl"`
	ProductGridTitle   string              `json:"productGridTitle"`
	VideosSectionTitle string              `json:"videosSectionTitle"`
	Products           []performerCellJSON `json:"products"`
	MainVideo          performerVideoJSON  `json:"mainVideo"`
	RelatedVideos      []performerVideoJSON `json:"relatedVideos"`
}

// Public shape — products as flat list; frontend builds 2D grid
type performerPublicJSON struct {
	HeroImageURL       string              `json:"heroImageUrl"`
	ProductGridTitle   string              `json:"productGridTitle"`
	VideosSectionTitle string              `json:"videosSectionTitle"`
	Products           []performerCellJSON `json:"products"`
	MainVideo          *performerVideoJSON  `json:"mainVideo"`
	RelatedVideos      []performerVideoJSON `json:"relatedVideos"`
}

var validPerformerSlugs = map[string]bool{
	"musician": true, "vocalist": true, "master-ceremony": true,
}

func emptyPerformerAdminJSON(slug string) performerAdminJSON {
	return performerAdminJSON{
		Slug:               slug,
		Label:              slug,
		IsPublished:        false,
		HeroImageURL:       "",
		ProductGridTitle:   "Professional Wireless Systems",
		VideosSectionTitle: "Related Videos",
		Products:           []performerCellJSON{},
		MainVideo:          performerVideoJSON{IsMain: true},
		RelatedVideos:      []performerVideoJSON{},
	}
}

func toCellJSON(c model.PerformerCellData) performerCellJSON {
	return performerCellJSON{
		ProductID:   c.ProductID,
		Name:        c.Name,
		Category:    c.Category,
		SubCategory: c.SubCategory,
		ImageURL:    c.ImageURL,
		IsHidden:    c.IsHidden,
		SortOrder:   c.SortOrder,
	}
}

func toPerformerVideoJSON(v model.PerformerVideoData) performerVideoJSON {
	return performerVideoJSON{
		IsMain: v.IsMain, Title: v.Title, Subtitle: v.Subtitle,
		ThumbnailURL: v.ThumbnailURL, VideoURL: v.VideoURL, SortOrder: v.SortOrder,
	}
}

// GetAdminPerformerPage handles GET /api/v1/admin/performer-pages/{slug}
func (h *Handler) GetAdminPerformerPage(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if !validPerformerSlugs[slug] {
		response.BadRequest(w, "invalid slug")
		return
	}
	data, err := h.performer.Get(r.Context(), slug)
	if err != nil {
		response.InternalError(w)
		return
	}
	if data == nil {
		response.OK(w, emptyPerformerAdminJSON(slug))
		return
	}

	cells := make([]performerCellJSON, len(data.Products))
	for i, c := range data.Products {
		cells[i] = toCellJSON(c)
	}
	related := make([]performerVideoJSON, len(data.RelatedVideos))
	for i, v := range data.RelatedVideos {
		related[i] = toPerformerVideoJSON(v)
	}
	response.OK(w, performerAdminJSON{
		Slug:               data.Slug,
		Label:              data.Label,
		IsPublished:        data.IsPublished,
		HeroImageURL:       data.HeroImageURL,
		ProductGridTitle:   data.ProductGridTitle,
		VideosSectionTitle: data.VideosSectionTitle,
		Products:           cells,
		MainVideo:          toPerformerVideoJSON(data.MainVideo),
		RelatedVideos:      related,
	})
}

// PutAdminPerformerPage handles PUT /api/v1/admin/performer-pages/{slug}
func (h *Handler) PutAdminPerformerPage(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if !validPerformerSlugs[slug] {
		response.BadRequest(w, "invalid slug")
		return
	}

	var req performerAdminJSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	products := make([]model.PerformerCellData, len(req.Products))
	for i, c := range req.Products {
		products[i] = model.PerformerCellData{
			ProductID: c.ProductID,
			IsHidden:  c.IsHidden,
			SortOrder: i,
		}
	}

	related := make([]model.PerformerVideoData, len(req.RelatedVideos))
	for i, v := range req.RelatedVideos {
		related[i] = model.PerformerVideoData{
			IsMain: v.IsMain, Title: v.Title, Subtitle: v.Subtitle,
			ThumbnailURL: v.ThumbnailURL, VideoURL: v.VideoURL, SortOrder: v.SortOrder,
		}
	}

	data := &model.PerformerAdminData{
		Slug:               slug,
		Label:              req.Label,
		IsPublished:        req.IsPublished,
		HeroImageURL:       req.HeroImageURL,
		ProductGridTitle:   req.ProductGridTitle,
		VideosSectionTitle: req.VideosSectionTitle,
		Products:           products,
		MainVideo: model.PerformerVideoData{
			IsMain: true, Title: req.MainVideo.Title, Subtitle: req.MainVideo.Subtitle,
			ThumbnailURL: req.MainVideo.ThumbnailURL, VideoURL: req.MainVideo.VideoURL,
		},
		RelatedVideos: related,
	}

	if err := h.performer.Save(r.Context(), data); err != nil {
		response.InternalError(w)
		return
	}
	response.OK(w, struct{}{})
}

// GetPerformerPage handles GET /api/v1/performer-pages/{slug} (public)
func (h *Handler) GetPerformerPage(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if !validPerformerSlugs[slug] {
		response.BadRequest(w, "invalid slug")
		return
	}
	data, err := h.performer.Get(r.Context(), slug)
	if err != nil {
		response.InternalError(w)
		return
	}
	if data == nil {
		response.OK(w, performerPublicJSON{
			Products:      []performerCellJSON{},
			RelatedVideos: []performerVideoJSON{},
		})
		return
	}

	cells := make([]performerCellJSON, 0, len(data.Products))
	for _, c := range data.Products {
		if !c.IsHidden {
			cells = append(cells, toCellJSON(c))
		}
	}

	related := make([]performerVideoJSON, len(data.RelatedVideos))
	for i, v := range data.RelatedVideos {
		related[i] = toPerformerVideoJSON(v)
	}

	var mainVideo *performerVideoJSON
	if data.MainVideo.Title != "" || data.MainVideo.ThumbnailURL != "" {
		mv := toPerformerVideoJSON(data.MainVideo)
		mainVideo = &mv
	}

	response.OK(w, performerPublicJSON{
		HeroImageURL:       data.HeroImageURL,
		ProductGridTitle:   data.ProductGridTitle,
		VideosSectionTitle: data.VideosSectionTitle,
		Products:           cells,
		MainVideo:          mainVideo,
		RelatedVideos:      related,
	})
}
