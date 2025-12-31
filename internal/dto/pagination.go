package dto

type PaginationParams struct {
	Take int `json:"take" query:"take" validate:"min=1,max=100"`
	Skip int `json:"skip" query:"skip" validate:"min=0"`
}

func (p *PaginationParams) GetOffset() int {
	if p.Skip < 0 {
		return 0
	}
	return p.Skip
}

func (p *PaginationParams) GetLimit() int {
	if p.Take < 1 {
		return 10
	}
	if p.Take > 100 {
		return 100
	}
	return p.Take
}

type PaginatedResponse struct {
	Data       interface{}    `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

type PaginationMeta struct {
	Take  int   `json:"take"`
	Skip  int   `json:"skip"`
	Total int64 `json:"total"`
}
