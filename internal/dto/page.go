package dto

import "encoding/json"

type Page[T any] struct {
	Data       []T `json:"data"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
}

func (p *Page[T]) UnmarshalJSON(data []byte) error {
	var decoded struct {
		Data       []T `json:"data"`
		Total      int `json:"total"`
		Page       int `json:"page"`
		PerPage    int `json:"per_page"`
		TotalPages int `json:"total_pages"`
		Meta       *struct {
			Total      int `json:"total"`
			Page       int `json:"page"`
			PerPage    int `json:"per_page"`
			TotalPages int `json:"total_pages"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	p.Data = decoded.Data
	p.Total = decoded.Total
	p.Page = decoded.Page
	p.PerPage = decoded.PerPage
	p.TotalPages = decoded.TotalPages
	if decoded.Meta != nil {
		p.Total = decoded.Meta.Total
		p.Page = decoded.Meta.Page
		p.PerPage = decoded.Meta.PerPage
		p.TotalPages = decoded.Meta.TotalPages
	}
	return nil
}
