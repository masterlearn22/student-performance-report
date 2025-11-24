package models

type PaginationQuery struct {
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
	Sort   string `query:"sort"`   
	Search string `query:"search"` 
	Status string `query:"status"` 
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