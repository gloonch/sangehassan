package domain

type ProductTerm struct {
	ID       int64  `json:"id"`
	Taxonomy string `json:"taxonomy"`
	Key      string `json:"key"`
	LabelEN  string `json:"label_en"`
	LabelFA  string `json:"label_fa"`
	LabelAR  string `json:"label_ar"`
}

