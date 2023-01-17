package data

import (
	"easylist/internal/validator"
	"math"
	"strings"
)

type Filters struct {
	Page         int
	Size         int
	Sort         string
	SortSafelist []string
	Includes     []string
}

type Metadata struct {
	CurrentPage  int
	PageSize     int
	FirstPage    int
	LastPage     int
	TotalRecords int
	NextPage     int
	PrevPage     int
}

func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}
	var lastPage = int(math.Ceil(float64(totalRecords) / float64(pageSize)))
	var nextPage = page + 1
	if nextPage > lastPage {
		nextPage = 0
	}
	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     lastPage,
		TotalRecords: totalRecords,
		NextPage:     nextPage,
		PrevPage:     page - 1,
	}
}

func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter: " + f.Sort)
}

func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

func (f Filters) limit() int {
	return f.Size
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.Size
}

func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page[number]", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page[number]", "must be maximum 10 mln")
	v.Check(f.Size > 0, "page[size]", "must be greater than zero")
	v.Check(f.Size <= 200, "page[size]", "must be a maximum 200")
	v.Check(validator.In(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}
