package models

import (
	"errors"
	"strings"

)
const (
	defaultPage    = 1
	defaultPerPage = 5
)


type Response struct{
	Data any `json:"data"`
	Error any `json:"error"`
}
type ErrorResponse struct{
	Message string `json:"message"`
	Code int `json:"code"`
}


type PaginationParams struct {
	Cursor    string   `query:"cursor"`
	PerPage uint   `query:"per_page"`
	Search  string `query:"search"`
}

func (i *PaginationParams) Validate() error {
	i.Search = strings.Trim(i.Search, " ")

	if i.PerPage == 0 {
		i.PerPage = defaultPerPage
	}
	if i.Search != "" && len(i.Search) < 5 {
		return errors.New("bad request")
	}
	return nil
}


