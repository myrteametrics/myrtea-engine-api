package model

// SearchOptions contains standard listing / search options
type SearchOptions struct {
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
	SortBy []SortOption `json:"sortBy"`
}

// SortOption contains a sort clause (field + ordering)
type SortOption struct {
	Field string          `json:"field"`
	Order SortOptionOrder `json:"order"`
}

// PaginatedResource wraps a list of resources with additionnal informations
type PaginatedResource struct {
	Total int         `json:"total"`
	Items interface{} `json:"items"` // Usually a slice of struct
}
