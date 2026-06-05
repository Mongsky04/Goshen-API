package model

import "time"

type Product struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	ImageURL    string    `json:"image_url"`
	Category    string    `json:"category"`
	SubCategory string    `json:"sub_category"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Featured struct {
	ID                 int64     `json:"id"`
	ProductID          int64     `json:"product_id"`
	Name               string    `json:"name"`
	ImageURL           string    `json:"image_url"`
	Category           string    `json:"category"`
	SubCategory        string    `json:"sub_category"`
	FeaturedCategories []string  `json:"featured_categories"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type HomepageGridProduct struct {
	ID        int64  `json:"id"`
	ProductID int64  `json:"product_id"`
	Name      string `json:"name"`
	ImageURL  string `json:"image_url"`
	Category  string `json:"category"`
	SubCategory string `json:"sub_category"`
}

type Banner struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	ImageURL  string    `json:"image_url"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Article struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Slider struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	ImageURL  string    `json:"image_url"`
	OrderNum  int       `json:"order_num"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Brand struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	ImageURL  string    `json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Customer struct {
	ID        int64     `json:"id"`
	ImageURL  string    `json:"image_url"`
	AltText   string    `json:"alt_text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SupportCard struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CtaLabel    string `json:"cta_label"`
	CtaHref     string `json:"cta_href"`
	SortOrder   int    `json:"sort_order"`
}

type Admin struct {
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
}

// --- Conference ---

type ConferenceHeroData struct {
	HeroImageURL string
	BadgeText    string
	Headline     string
	SubText      string
}

type ConferenceTitlesData struct {
	ProductGrid string
	Workspace   string
	Solutions   string
	Contact     string
}

type ConferenceSolutionData struct {
	ID              string
	RoomSize        string
	Title           string
	Description     string
	KitLabel        string
	KitItems        []string
	ImageURL        string
	ImageURL2       string
	Card1Name       string
	Card1Category   string
	Card1SubCategory string
	Card2Name       string
	Card2Category   string
	Card2SubCategory string
	IsHidden        bool
	SortOrder       int
}

type ConferenceProductData struct {
	ID          string
	ProductID   int64
	Name        string
	Category    string
	SubCategory string
	ImageURL    string
	IsHidden    bool
}

type ConferenceAdminData struct {
	ID                   string
	Slug                 string
	Label                string
	IsPublished          bool
	Hero                 ConferenceHeroData
	Titles               ConferenceTitlesData
	WorkspaceDescription string
	Solutions            []ConferenceSolutionData
	Products             []ConferenceProductData
}

// --- Performer ---

type PerformerCellData struct {
	ProductID   int64
	Name        string
	Category    string
	SubCategory string
	ImageURL    string
	IsHidden    bool
	SortOrder   int
}

type PerformerVideoData struct {
	IsMain       bool
	Title        string
	Subtitle     string
	ThumbnailURL string
	VideoURL     string
	SortOrder    int
}

type PerformerAdminData struct {
	ID                 string
	Slug               string
	Label              string
	IsPublished        bool
	HeroImageURL       string
	ProductGridTitle   string
	VideosSectionTitle string
	Products           []PerformerCellData
	MainVideo          PerformerVideoData
	RelatedVideos      []PerformerVideoData
}

type MediaAsset struct {
	ID        int64     `json:"id"`
	Filename  string    `json:"filename"`
	URL       string    `json:"url"`
	Size      int64     `json:"size"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
}
