package utils

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

// PaginationParams thông số phân trang từ query string
type PaginationParams struct {
	Page  int
	Limit int
}

// ParsePagination đọc và validate tham số phân trang từ request
func ParsePagination(c echo.Context) PaginationParams {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	if page <= 0 {
		page = DefaultPage
	}
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	return PaginationParams{Page: page, Limit: limit}
}

// Offset tính offset cho SQL query
func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.Limit
}
