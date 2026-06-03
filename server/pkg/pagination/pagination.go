package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type PaginationParams struct {
	Page   int
	Limit  int
	Offset int
}

// ParseQuery reads page and limit from query params.
// Defaults: page=1, limit=10. Max limit=100.
func ParseQuery(c *gin.Context) PaginationParams {
	page := 1
	limit := 10

	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit
	return PaginationParams{
		Page:   page,
		Limit:  limit,
		Offset: offset,
	}
}
