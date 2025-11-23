package models

type PaginationQuery struct {
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
	Sort   string `query:"sort"`   // created_at_desc, created_at_asc
	Search string `query:"search"` // Optional: Search by title (butuh join mongo, kompleks)
	Status string `query:"status"` // Filter status
}

type PaginationMeta struct {
	CurrentPage int `json:"currentPage"`
	TotalPage   int `json:"totalPage"`
	TotalData   int `json:"totalData"`
	Limit       int `json:"limit"`
}

type PaginatedResponse struct {
	Data []interface{}  `json:"data"`
	Meta PaginationMeta `json:"meta"`
}