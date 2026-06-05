package handler

import (
	"encoding/json"
	"net/http"

	"goshen/backend/internal/model"
	"goshen/backend/pkg/response"

	"github.com/go-chi/chi/v5"
)

// JSON shapes — match ConferencePageForm on the frontend

type conferenceHeroJSON struct {
	HeroImageURL string `json:"heroImageUrl"`
	BadgeText    string `json:"badgeText"`
	Headline     string `json:"headline"`
	SubText      string `json:"subText"`
}

type conferenceTitlesJSON struct {
	ProductGrid string `json:"productGrid"`
	Workspace   string `json:"workspace"`
	Solutions   string `json:"solutions"`
	Contact     string `json:"contact"`
}

type conferenceSolutionJSON struct {
	RoomSize         string   `json:"roomSize"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	KitLabel         string   `json:"kitLabel"`
	KitItems         []string `json:"kitItems"`
	ImageURL         string   `json:"imageUrl"`
	ImageURL2        string   `json:"imageUrl2"`
	Card1Name        string   `json:"card1Name"`
	Card1Category    string   `json:"card1Category"`
	Card1SubCategory string   `json:"card1SubCategory"`
	Card2Name        string   `json:"card2Name"`
	Card2Category    string   `json:"card2Category"`
	Card2SubCategory string   `json:"card2SubCategory"`
	IsHidden         bool     `json:"isHidden"`
}

type conferenceProductJSON struct {
	ID          string `json:"id"`
	ProductID   int64  `json:"productId"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	SubCategory string `json:"subCategory"`
	ImageURL    string `json:"imageUrl"`
	IsHidden    bool   `json:"isHidden"`
}

type conferenceAdminJSON struct {
	Slug                 string                   `json:"slug"`
	Label                string                   `json:"label"`
	IsPublished          bool                     `json:"isPublished"`
	Hero                 conferenceHeroJSON        `json:"hero"`
	Titles               conferenceTitlesJSON      `json:"titles"`
	WorkspaceDescription string                   `json:"workspaceDescription"`
	Solutions            []conferenceSolutionJSON  `json:"solutions"`
	Products             []conferenceProductJSON   `json:"products"`
}

// Public shape — matches ConferencePageData on the frontend

type conferencePublicSolutionJSON struct {
	RoomSize         string   `json:"roomSize"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	KitLabel         string   `json:"kitLabel"`
	KitItems         []string `json:"kitItems"`
	ImageURL         string   `json:"imageUrl"`
	ImageURL2        string   `json:"imageUrl2"`
	Card1Name        string   `json:"card1Name"`
	Card1Category    string   `json:"card1Category"`
	Card1SubCategory string   `json:"card1SubCategory"`
	Card2Name        string   `json:"card2Name"`
	Card2Category    string   `json:"card2Category"`
	Card2SubCategory string   `json:"card2SubCategory"`
}

type conferencePublicProductJSON struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	SubCategory string `json:"subCategory"`
	ImageURL    string `json:"imageUrl"`
}

type conferencePublicJSON struct {
	HeroImageURL         string                         `json:"heroImageUrl"`
	ProductGridTitle     string                         `json:"productGridTitle"`
	WorkspaceTitle       string                         `json:"workspaceTitle"`
	WorkspaceDescription string                         `json:"workspaceDescription"`
	SolutionsTitle       string                         `json:"solutionsTitle"`
	Solutions            []conferencePublicSolutionJSON  `json:"solutions"`
	Products             []conferencePublicProductJSON   `json:"products"`
}

var validConferenceSlugs = map[string]bool{
	"enterprise": true, "government": true,
	"higher-education": true, "hospitality": true,
}

var defaultConferenceTitles = conferenceTitlesJSON{
	ProductGrid: "Recommended Products",
	Workspace:   "Workspace Solutions",
	Solutions:   "Room Solutions",
	Contact:     "Contact Enterprise",
}

func toConferenceAdminJSON(data *model.ConferenceAdminData) conferenceAdminJSON {
	solutions := make([]conferenceSolutionJSON, len(data.Solutions))
	for i, s := range data.Solutions {
		kitItems := s.KitItems
		if kitItems == nil {
			kitItems = []string{}
		}
		solutions[i] = conferenceSolutionJSON{
			RoomSize:         s.RoomSize,
			Title:            s.Title,
			Description:      s.Description,
			KitLabel:         s.KitLabel,
			KitItems:         kitItems,
			ImageURL:         s.ImageURL,
			ImageURL2:        s.ImageURL2,
			Card1Name:        s.Card1Name,
			Card1Category:    s.Card1Category,
			Card1SubCategory: s.Card1SubCategory,
			Card2Name:        s.Card2Name,
			Card2Category:    s.Card2Category,
			Card2SubCategory: s.Card2SubCategory,
			IsHidden:         s.IsHidden,
		}
	}
	products := make([]conferenceProductJSON, len(data.Products))
	for i, p := range data.Products {
		products[i] = conferenceProductJSON{
			ID: p.ID, ProductID: p.ProductID, Name: p.Name,
			Category: p.Category, SubCategory: p.SubCategory, ImageURL: p.ImageURL,
			IsHidden: p.IsHidden,
		}
	}
	return conferenceAdminJSON{
		Slug:        data.Slug,
		Label:       data.Label,
		IsPublished: data.IsPublished,
		Hero: conferenceHeroJSON{
			HeroImageURL: data.Hero.HeroImageURL,
			BadgeText:    data.Hero.BadgeText,
			Headline:     data.Hero.Headline,
			SubText:      data.Hero.SubText,
		},
		Titles: conferenceTitlesJSON{
			ProductGrid: data.Titles.ProductGrid,
			Workspace:   data.Titles.Workspace,
			Solutions:   data.Titles.Solutions,
			Contact:     data.Titles.Contact,
		},
		WorkspaceDescription: data.WorkspaceDescription,
		Solutions:            solutions,
		Products:             products,
	}
}

func emptyConferenceAdminJSON(slug string) conferenceAdminJSON {
	return conferenceAdminJSON{
		Slug:        slug,
		Label:       slug,
		IsPublished: false,
		Hero:        conferenceHeroJSON{},
		Titles:      defaultConferenceTitles,
		Solutions:   []conferenceSolutionJSON{},
		Products:    []conferenceProductJSON{},
	}
}

// GetAdminConferencePage handles GET /api/v1/admin/conference-pages/{slug}
func (h *Handler) GetAdminConferencePage(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if !validConferenceSlugs[slug] {
		response.BadRequest(w, "invalid slug")
		return
	}
	data, err := h.conference.Get(r.Context(), slug)
	if err != nil {
		response.InternalError(w)
		return
	}
	if data == nil {
		response.OK(w, emptyConferenceAdminJSON(slug))
		return
	}
	response.OK(w, toConferenceAdminJSON(data))
}

// PutAdminConferencePage handles PUT /api/v1/admin/conference-pages/{slug}
func (h *Handler) PutAdminConferencePage(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if !validConferenceSlugs[slug] {
		response.BadRequest(w, "invalid slug")
		return
	}

	var req conferenceAdminJSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	solutions := make([]model.ConferenceSolutionData, len(req.Solutions))
	for i, s := range req.Solutions {
		solutions[i] = model.ConferenceSolutionData{
			RoomSize:         s.RoomSize,
			Title:            s.Title,
			Description:      s.Description,
			KitLabel:         s.KitLabel,
			KitItems:         s.KitItems,
			ImageURL:         s.ImageURL,
			ImageURL2:        s.ImageURL2,
			Card1Name:        s.Card1Name,
			Card1Category:    s.Card1Category,
			Card1SubCategory: s.Card1SubCategory,
			Card2Name:        s.Card2Name,
			Card2Category:    s.Card2Category,
			Card2SubCategory: s.Card2SubCategory,
			IsHidden:         s.IsHidden,
		}
	}
	products := make([]model.ConferenceProductData, len(req.Products))
	for i, p := range req.Products {
		products[i] = model.ConferenceProductData{
			ProductID: p.ProductID,
			IsHidden:  p.IsHidden,
		}
	}

	data := &model.ConferenceAdminData{
		Slug:        slug,
		Label:       req.Label,
		IsPublished: req.IsPublished,
		Hero: model.ConferenceHeroData{
			HeroImageURL: req.Hero.HeroImageURL,
			BadgeText:    req.Hero.BadgeText,
			Headline:     req.Hero.Headline,
			SubText:      req.Hero.SubText,
		},
		Titles: model.ConferenceTitlesData{
			ProductGrid: req.Titles.ProductGrid,
			Workspace:   req.Titles.Workspace,
			Solutions:   req.Titles.Solutions,
			Contact:     req.Titles.Contact,
		},
		WorkspaceDescription: req.WorkspaceDescription,
		Solutions:            solutions,
		Products:             products,
	}

	if err := h.conference.Save(r.Context(), data); err != nil {
		response.InternalError(w)
		return
	}
	response.OK(w, struct{}{})
}

// GetConferencePage handles GET /api/v1/conference-pages/{slug} (public)
func (h *Handler) GetConferencePage(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if !validConferenceSlugs[slug] {
		response.BadRequest(w, "invalid slug")
		return
	}
	data, err := h.conference.Get(r.Context(), slug)
	if err != nil {
		response.InternalError(w)
		return
	}
	if data == nil || !data.IsPublished {
		response.OK(w, conferencePublicJSON{
			Solutions: []conferencePublicSolutionJSON{},
			Products:  []conferencePublicProductJSON{},
		})
		return
	}

	solutions := make([]conferencePublicSolutionJSON, 0, len(data.Solutions))
	for _, s := range data.Solutions {
		if s.IsHidden {
			continue
		}
		kitItems := s.KitItems
		if kitItems == nil {
			kitItems = []string{}
		}
		solutions = append(solutions, conferencePublicSolutionJSON{
			RoomSize:         s.RoomSize,
			Title:            s.Title,
			Description:      s.Description,
			KitLabel:         s.KitLabel,
			KitItems:         kitItems,
			ImageURL:         s.ImageURL,
			ImageURL2:        s.ImageURL2,
			Card1Name:        s.Card1Name,
			Card1Category:    s.Card1Category,
			Card1SubCategory: s.Card1SubCategory,
			Card2Name:        s.Card2Name,
			Card2Category:    s.Card2Category,
			Card2SubCategory: s.Card2SubCategory,
		})
	}
	products := make([]conferencePublicProductJSON, 0, len(data.Products))
	for _, p := range data.Products {
		if p.IsHidden {
			continue
		}
		products = append(products, conferencePublicProductJSON{
			ID: p.ID, Name: p.Name,
			Category: p.Category, SubCategory: p.SubCategory, ImageURL: p.ImageURL,
		})
	}

	response.OK(w, conferencePublicJSON{
		HeroImageURL:         data.Hero.HeroImageURL,
		ProductGridTitle:     data.Titles.ProductGrid,
		WorkspaceTitle:       data.Titles.Workspace,
		WorkspaceDescription: data.WorkspaceDescription,
		SolutionsTitle:       data.Titles.Solutions,
		Solutions:            solutions,
		Products:             products,
	})
}
