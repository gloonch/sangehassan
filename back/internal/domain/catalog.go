package domain

type CatalogCategory struct {
	Category
	ProductCount int    `json:"product_count"`
	PreviewImage string `json:"preview_image,omitempty"`
}

type CatalogFacetValue struct {
	ID          int64  `json:"id"`
	Key         string `json:"key"`
	LabelEN     string `json:"label_en"`
	LabelFA     string `json:"label_fa"`
	LabelAR     string `json:"label_ar"`
	Count       int    `json:"count"`
	IsIndexable bool   `json:"is_indexable"`
}

type CatalogFacetPage struct {
	ID            int64  `json:"id"`
	CategoryID    int64  `json:"category_id"`
	TermID        int64  `json:"term_id"`
	CategorySlug  string `json:"category_slug,omitempty"`
	Taxonomy      string `json:"taxonomy,omitempty"`
	TermKey       string `json:"term_key,omitempty"`
	TitleEN       string `json:"title_en,omitempty"`
	TitleFA       string `json:"title_fa,omitempty"`
	TitleAR       string `json:"title_ar,omitempty"`
	DescriptionEN string `json:"description_en,omitempty"`
	DescriptionFA string `json:"description_fa,omitempty"`
	DescriptionAR string `json:"description_ar,omitempty"`
	H1EN          string `json:"h1_en,omitempty"`
	H1FA          string `json:"h1_fa,omitempty"`
	H1AR          string `json:"h1_ar,omitempty"`
	IntroEN       string `json:"intro_en,omitempty"`
	IntroFA       string `json:"intro_fa,omitempty"`
	IntroAR       string `json:"intro_ar,omitempty"`
	IsActive      bool   `json:"is_active"`
	IsIndexable   bool   `json:"is_indexable"`
}

type CatalogFacet struct {
	Key      string              `json:"key"`
	Taxonomy string              `json:"taxonomy"`
	LabelEN  string              `json:"label_en"`
	LabelFA  string              `json:"label_fa"`
	LabelAR  string              `json:"label_ar"`
	Values   []CatalogFacetValue `json:"values"`
}

type CatalogSEO struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	H1          string `json:"h1"`
	Intro       string `json:"intro"`
	Canonical   string `json:"canonical"`
	Robots      string `json:"robots"`
}

type CatalogPagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

type CatalogPage struct {
	Locale            string              `json:"locale"`
	Category          CatalogCategory     `json:"category"`
	SEO               CatalogSEO          `json:"seo"`
	Facets            []CatalogFacet      `json:"facets"`
	Selected          map[string][]string `json:"selected_filters"`
	Products          []Product           `json:"products"`
	Pagination        CatalogPagination   `json:"pagination"`
	Indexable         bool                `json:"indexable"`
	SelectedFacet     *CatalogFacetValue  `json:"selected_facet,omitempty"`
	SelectedFacetKey  string              `json:"selected_facet_key,omitempty"`
	RelatedCategories []CatalogCategory   `json:"related_categories,omitempty"`
	RelatedProjects   []ProjectCard       `json:"related_projects,omitempty"`
}

type CatalogHub struct {
	Locale     string            `json:"locale"`
	Categories []CatalogCategory `json:"categories"`
	SEO        CatalogSEO        `json:"seo"`
}

type CatalogRoute struct {
	Path         string `json:"path"`
	Type         string `json:"type"`
	CategorySlug string `json:"category_slug,omitempty"`
	Facet        string `json:"facet,omitempty"`
	Value        string `json:"value,omitempty"`
	ProductCount int    `json:"product_count"`
	Indexable    bool   `json:"indexable"`
}
