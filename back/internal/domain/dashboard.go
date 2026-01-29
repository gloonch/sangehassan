package domain

type DashboardStats struct {
	ProductsCount   int64           `json:"products_count"`
	CategoriesCount int64           `json:"categories_count"`
	TemplatesCount  int64           `json:"templates_count"`
	BlogsCount      int64           `json:"blogs_count"`
	CategoryUsage   []CategoryUsage `json:"category_usage"`
	Templates       []Template      `json:"templates"`
}

type CategoryUsage struct {
	ID           int64  `json:"id"`
	TitleEN      string `json:"title_en"`
	TitleFA      string `json:"title_fa"`
	TitleAR      string `json:"title_ar"`
	Slug         string `json:"slug"`
	ProductCount int64  `json:"product_count"`
}
