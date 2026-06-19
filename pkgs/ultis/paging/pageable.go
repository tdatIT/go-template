package pagable

import (
	"math"
	"strconv"
)

const (
	defaultSize = 15
	maxSize     = 100
	defaultPage = 1
	ASCENDING   = "asc"
	DESCENDING  = "desc"
)

type ListQuery struct {
	Page          int     `query:"page"`
	Size          int     `query:"size"`
	Search        *string `query:"search"`
	SortField     *string `query:"sort_field"`
	SortDirection *string `query:"sort_direction"`
	DateFrom      *string `query:"date_from"`
	DateTo        *string `query:"date_to"`
}

type ListResponse struct {
	Items   any  `json:"items"`
	Total   int  `json:"total"`
	Page    int  `json:"page"`
	Size    int  `json:"size"`
	HasMore bool `json:"hasMore"`
}

// SetSize Set page size
func (q *ListQuery) SetSize(sizeQuery string) error {
	if sizeQuery == "" {
		q.Size = defaultSize
		return nil
	}

	n, err := strconv.ParseUint(sizeQuery, 10, 32)
	if err != nil {
		return err
	}

	q.Size = int(n)
	if q.Size > maxSize {
		q.Size = maxSize
	}

	return nil
}

// SetPage Set page number
func (q *ListQuery) SetPage(pageQuery string) error {
	if pageQuery == "" {
		q.Page = defaultPage
		return nil
	}
	n, err := strconv.ParseUint(pageQuery, 10, 32)
	if err != nil {
		return err
	}
	q.Page = int(n)

	return nil
}

// GetOffset Get offset
func (q *ListQuery) GetOffset() int {
	return (q.GetPage() - 1) * q.GetSize()
}

func (q *ListQuery) GetPage() int {
	if q.Page == 0 {
		return defaultPage
	}
	return q.Page
}

func (q *ListQuery) GetSize() int {
	if q.Size == 0 {
		return defaultSize
	}
	return q.Size
}

func (q *ListQuery) GetTotalPages(totalCount int) int {
	d := float64(totalCount) / float64(q.GetSize())
	return int(math.Ceil(d))
}

func (q *ListQuery) GetHasMore(total int) bool {
	return q.GetPage() < q.GetTotalPages(total)
}
