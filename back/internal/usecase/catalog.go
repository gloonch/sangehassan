package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

type CatalogService struct {
	repo            ports.CatalogRepository
	minimumProducts int
}

type catalogFacetDefinition struct {
	RouteKey string
	Taxonomy string
	LabelEN  string
	LabelFA  string
	LabelAR  string
}

var catalogFacetDefinitions = []catalogFacetDefinition{
	{RouteKey: "color", Taxonomy: "tone", LabelEN: "Color", LabelFA: "رنگ", LabelAR: "اللون"},
	{RouteKey: "application", Taxonomy: "use_case_application", LabelEN: "Application", LabelFA: "کاربرد", LabelAR: "الاستخدام"},
	{RouteKey: "finish", Taxonomy: "finishes", LabelEN: "Finish", LabelFA: "نوع فرآوری", LabelAR: "التشطيب"},
	{RouteKey: "form", Taxonomy: "use_case_form", LabelEN: "Form", LabelFA: "فرم عرضه", LabelAR: "الشكل"},
	{RouteKey: "origin", Taxonomy: "mines", LabelEN: "Origin", LabelFA: "مبدأ یا معدن", LabelAR: "المنشأ"},
	{RouteKey: "pattern", Taxonomy: "pattern", LabelEN: "Pattern", LabelFA: "نوع موج و طرح", LabelAR: "النمط"},
	{RouteKey: "availability", Taxonomy: "availability", LabelEN: "Availability", LabelFA: "وضعیت تأمین", LabelAR: "التوفر"},
}

func NewCatalogService(repo ports.CatalogRepository, minimumProducts int) *CatalogService {
	if minimumProducts < 1 {
		minimumProducts = 2
	}
	return &CatalogService{repo: repo, minimumProducts: minimumProducts}
}

func (s *CatalogService) Hub(ctx context.Context, locale string) (domain.CatalogHub, error) {
	categories, err := s.repo.ListCategories(ctx)
	if err != nil {
		return domain.CatalogHub{}, err
	}
	seo := domain.CatalogSEO{
		Title:       "Natural Stone Categories | SangeHassan",
		Description: "Browse natural stone categories and find the right stone for architectural and building projects.",
		H1:          "Natural stone categories",
		Intro:       "Choose a stone category to view available products, colors, finishes and applications.",
		Canonical:   "/en/products",
		Robots:      "index,follow",
	}
	if locale == "fa" {
		seo = domain.CatalogSEO{
			Title:       "انواع سنگ ساختمانی | تراورتن، گرانیت، مرمریت و سنگ چینی",
			Description: "مشاهده و بررسی دسته‌بندی انواع سنگ ساختمانی و طبیعی شامل تراورتن، گرانیت، مرمریت، چینی کریستال و سایر سنگ‌های پروژه‌ای.",
			H1:          "دسته‌بندی انواع سنگ ساختمانی",
			Intro:       "برای مشاهده محصولات، رنگ‌ها، فرآوری‌ها و کاربردهای هر سنگ، دسته‌بندی موردنظر را انتخاب کنید.",
			Canonical:   "/fa/products",
			Robots:      "index,follow",
		}
	} else if locale == "ar" {
		seo = domain.CatalogSEO{
			Title:       "أنواع الحجر الطبيعي | سانج حسن",
			Description: "تصفح فئات الحجر الطبيعي واختيار الحجر المناسب للمشاريع المعمارية والإنشائية.",
			H1:          "فئات الحجر الطبيعي",
			Intro:       "اختر فئة الحجر لعرض المنتجات والألوان والتشطيبات والاستخدامات المتاحة.",
			Canonical:   "/ar/products",
			Robots:      "index,follow",
		}
	}
	return domain.CatalogHub{Locale: locale, Categories: categories, SEO: seo}, nil
}

func (s *CatalogService) Page(
	ctx context.Context,
	locale, categorySlug, routeFacet, routeValue string,
	routeFilters map[string][]string,
	limit, offset int,
) (domain.CatalogPage, error) {
	category, err := s.repo.GetCategory(ctx, categorySlug)
	if err != nil {
		return domain.CatalogPage{}, err
	}

	filters := make(map[string][]string)
	for routeKey, values := range routeFilters {
		if routeKey == "__search" {
			filters[routeKey] = uniqueStrings(values)
			continue
		}
		definition, ok := findCatalogFacet(routeKey)
		if !ok || len(values) == 0 {
			continue
		}
		filters[definition.Taxonomy] = uniqueStrings(values)
	}

	var selectedFacet *domain.CatalogFacetValue
	var facetPage *domain.CatalogFacetPage
	selectedFacetKey := ""
	if routeFacet != "" || routeValue != "" {
		definition, ok := findCatalogFacet(routeFacet)
		if !ok || strings.TrimSpace(routeValue) == "" {
			return domain.CatalogPage{}, sql.ErrNoRows
		}
		value, valueErr := s.repo.GetFacetValue(ctx, category.ID, definition.Taxonomy, routeValue)
		if valueErr != nil {
			return domain.CatalogPage{}, valueErr
		}
		selectedFacet = &value
		override, overrideErr := s.repo.GetFacetPage(ctx, category.ID, value.ID)
		if overrideErr != nil && overrideErr != sql.ErrNoRows {
			return domain.CatalogPage{}, overrideErr
		}
		if overrideErr == nil {
			if !override.IsActive {
				return domain.CatalogPage{}, sql.ErrNoRows
			}
			facetPage = &override
		}
		selectedFacetKey = definition.RouteKey
		filters[definition.Taxonomy] = []string{routeValue}
	}

	products, total, err := s.repo.ListProducts(ctx, category.ID, filters, limit, offset)
	if err != nil {
		return domain.CatalogPage{}, err
	}

	facets := make([]domain.CatalogFacet, 0, len(catalogFacetDefinitions))
	for _, definition := range catalogFacetDefinitions {
		values, valueErr := s.repo.ListFacetValues(ctx, category.ID, definition.Taxonomy, filters)
		if valueErr != nil {
			return domain.CatalogPage{}, valueErr
		}
		if len(values) == 0 {
			continue
		}
		facets = append(facets, domain.CatalogFacet{
			Key: definition.RouteKey, Taxonomy: definition.Taxonomy,
			LabelEN: definition.LabelEN, LabelFA: definition.LabelFA, LabelAR: definition.LabelAR,
			Values: values,
		})
	}

	selected := make(map[string][]string)
	for _, definition := range catalogFacetDefinitions {
		if values := filters[definition.Taxonomy]; len(values) > 0 {
			selected[definition.RouteKey] = values
		}
	}

	indexable := category.IsIndexable && len(routeFilters) == 0
	if selectedFacet != nil {
		indexable = indexable && selectedFacet.IsIndexable && selectedFacet.Count >= s.minimumProducts
		if facetPage != nil {
			indexable = indexable && facetPage.IsIndexable
		}
	}
	seo := buildCatalogSEO(locale, category, routeFacet, selectedFacet, indexable)
	if facetPage != nil {
		seo = applyFacetPageSEO(seo, *facetPage, locale)
	}
	relatedCategories, err := s.repo.ListCategories(ctx)
	if err != nil {
		return domain.CatalogPage{}, err
	}
	filteredRelated := make([]domain.CatalogCategory, 0, len(relatedCategories))
	for _, related := range relatedCategories {
		if related.ID == category.ID {
			continue
		}
		filteredRelated = append(filteredRelated, related)
	}
	relatedProjects, err := s.repo.ListRelatedProjects(ctx, category.ID, 6)
	if err != nil {
		return domain.CatalogPage{}, err
	}

	return domain.CatalogPage{
		Locale:   locale,
		Category: category, SEO: seo, Facets: facets, Selected: selected,
		Products: products, Pagination: domain.CatalogPagination{Limit: limit, Offset: offset, Total: total},
		Indexable: indexable, SelectedFacet: selectedFacet, SelectedFacetKey: selectedFacetKey,
		RelatedCategories: filteredRelated, RelatedProjects: relatedProjects,
	}, nil
}

func (s *CatalogService) Routes(ctx context.Context) ([]domain.CatalogRoute, error) {
	return s.repo.ListIndexableRoutes(ctx, s.minimumProducts)
}

func (s *CatalogService) ListFacetPages(ctx context.Context) ([]domain.CatalogFacetPage, error) {
	return s.repo.ListFacetPages(ctx)
}

func (s *CatalogService) UpsertFacetPage(ctx context.Context, page domain.CatalogFacetPage) (domain.CatalogFacetPage, error) {
	if page.CategoryID < 1 || page.TermID < 1 {
		return domain.CatalogFacetPage{}, fmt.Errorf("category_id and term_id are required")
	}
	return s.repo.UpsertFacetPage(ctx, page)
}

func (s *CatalogService) DeleteFacetPage(ctx context.Context, id int64) error {
	return s.repo.DeleteFacetPage(ctx, id)
}

func findCatalogFacet(routeKey string) (catalogFacetDefinition, bool) {
	for _, definition := range catalogFacetDefinitions {
		if definition.RouteKey == routeKey {
			return definition, true
		}
	}
	return catalogFacetDefinition{}, false
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func buildCatalogSEO(locale string, category domain.CatalogCategory, routeFacet string, value *domain.CatalogFacetValue, indexable bool) domain.CatalogSEO {
	title := category.TitleEN
	intro := category.IntroEN
	seoTitle := category.SEOTitleEN
	description := category.SEODescriptionEN
	canonical := "/en/products/" + category.Slug
	if locale == "fa" {
		title = category.TitleFA
		intro = category.IntroFA
		seoTitle = category.SEOTitleFA
		description = category.SEODescriptionFA
		canonical = "/fa/products/" + category.Slug
	} else if locale == "ar" {
		title = category.TitleAR
		intro = category.IntroAR
		seoTitle = category.SEOTitleAR
		description = category.SEODescriptionAR
		canonical = "/ar/products/" + category.Slug
	}

	h1 := strings.TrimSpace(title)
	if locale == "fa" && !strings.HasPrefix(h1, "سنگ ") {
		h1 = "سنگ " + h1
	} else if locale == "ar" && !strings.HasPrefix(h1, "حجر ") {
		h1 = "حجر " + h1
	}
	if value != nil {
		label := value.LabelEN
		if locale == "fa" {
			label = value.LabelFA
		}
		if locale == "ar" {
			label = value.LabelAR
		}
		h1 = facetHeading(locale, routeFacet, h1, label)
		canonical += "/" + routeFacet + "/" + value.Key
		if locale == "fa" {
			seoTitle = h1 + " | سنگ حسن"
		} else if locale == "ar" {
			seoTitle = h1 + " | سانج حسن"
		} else {
			seoTitle = h1 + " | SangeHassan"
		}
		if locale == "fa" {
			description = fmt.Sprintf("مشاهده و مقایسه انواع %s، بررسی مشخصات، تصاویر و انتخاب سنگ مناسب برای پروژه‌های ساختمانی.", h1)
		} else if locale == "ar" {
			description = fmt.Sprintf("تصفح منتجات %s ومقارنة المواصفات والصور والخيارات المتاحة للمشاريع المعمارية.", h1)
		} else {
			description = fmt.Sprintf("Browse %s products, specifications and available options for architectural projects.", h1)
		}
		intro = description
	}
	if seoTitle == "" {
		if locale == "fa" {
			seoTitle = h1 + " | انواع، کاربرد و خرید"
		} else if locale == "ar" {
			seoTitle = h1 + " | الأنواع والاستخدامات"
		} else {
			seoTitle = h1 + " | SangeHassan"
		}
	}
	if description == "" {
		if locale == "fa" {
			description = fmt.Sprintf("مشاهده انواع %s، رنگ‌ها، فرآوری‌ها، کاربردها و محصولات موجود برای پروژه‌های ساختمانی.", h1)
		} else if locale == "ar" {
			description = fmt.Sprintf("تصفح أنواع %s والألوان والتشطيبات والاستخدامات المتاحة للمشاريع الإنشائية.", h1)
		} else {
			description = fmt.Sprintf("Browse %s products, colors, finishes and applications for building projects.", h1)
		}
	}
	if intro == "" {
		if locale == "fa" {
			intro = description
		} else {
			intro = description
		}
	}
	robots := "index,follow"
	if !indexable {
		robots = "noindex,follow"
	}
	return domain.CatalogSEO{Title: seoTitle, Description: description, H1: h1, Intro: intro, Canonical: canonical, Robots: robots}
}

func facetHeading(locale, facet, category, value string) string {
	if locale == "en" {
		return strings.TrimSpace(category + " " + value)
	}
	if locale == "ar" {
		switch facet {
		case "application":
			return fmt.Sprintf("%s مناسب لـ %s", category, value)
		case "finish":
			return fmt.Sprintf("%s بتشطيب %s", category, value)
		case "origin":
			return fmt.Sprintf("%s من %s", category, value)
		case "pattern":
			return fmt.Sprintf("%s بنمط %s", category, value)
		default:
			return strings.TrimSpace(category + " " + value)
		}
	}
	switch facet {
	case "application":
		return fmt.Sprintf("%s مناسب %s", category, value)
	case "finish":
		return fmt.Sprintf("%s با فرآوری %s", category, value)
	case "origin":
		return fmt.Sprintf("%s معدن %s", category, value)
	case "pattern":
		return fmt.Sprintf("%s با طرح %s", category, value)
	default:
		return strings.TrimSpace(category + " " + value)
	}
}

func applyFacetPageSEO(seo domain.CatalogSEO, page domain.CatalogFacetPage, locale string) domain.CatalogSEO {
	title, description, h1, intro := page.TitleEN, page.DescriptionEN, page.H1EN, page.IntroEN
	if locale == "fa" {
		title, description, h1, intro = page.TitleFA, page.DescriptionFA, page.H1FA, page.IntroFA
	} else if locale == "ar" {
		title, description, h1, intro = page.TitleAR, page.DescriptionAR, page.H1AR, page.IntroAR
	}
	if strings.TrimSpace(title) != "" {
		seo.Title = title
	}
	if strings.TrimSpace(description) != "" {
		seo.Description = description
	}
	if strings.TrimSpace(h1) != "" {
		seo.H1 = h1
	}
	if strings.TrimSpace(intro) != "" {
		seo.Intro = intro
	}
	return seo
}
