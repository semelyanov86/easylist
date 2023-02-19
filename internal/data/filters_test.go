package data

import (
	"reflect"
	"testing"
)

func TestFilters_SortColumn(t *testing.T) {
	tests := []struct {
		name         string
		filters      Filters
		sortSafelist []string
		expected     string
		expectPanic  bool
	}{
		{
			name: "safe sort value",
			filters: Filters{
				Sort:         "id",
				SortSafelist: []string{"id"},
			},
			sortSafelist: []string{"id"},
			expected:     "id",
		},
		{
			name: "unsafe sort value",
			filters: Filters{
				Sort:         "name",
				SortSafelist: []string{"id"},
			},
			sortSafelist: []string{"id"},
			expectPanic:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tt.expectPanic {
					t.Errorf("Filters.SortColumn() panicked unexpectedly: %v", r)
				}
			}()

			filters := tt.filters
			filters.SortSafelist = tt.sortSafelist
			actual := filters.sortColumn()
			if actual != tt.expected {
				t.Errorf("Filters.SortColumn() = %v; want %v", actual, tt.expected)
			}
		})
	}
}

func TestFilters_SortDirection(t *testing.T) {
	tests := []struct {
		name     string
		filters  Filters
		expected string
	}{
		{
			name: "sort in ascending order",
			filters: Filters{
				Sort: "id",
			},
			expected: "ASC",
		},
		{
			name: "sort in descending order",
			filters: Filters{
				Sort: "-id",
			},
			expected: "DESC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := tt.filters
			actual := filters.sortDirection()
			if actual != tt.expected {
				t.Errorf("Filters.SortDirection() = %v; want %v", actual, tt.expected)
			}
		})
	}
}

func TestFilters_Limit(t *testing.T) {
	filters := Filters{
		Size: 100,
	}
	actual := filters.limit()
	expected := 100
	if actual != expected {
		t.Errorf("Filters.Limit() = %v; want %v", actual, expected)
	}
}

func TestFilters_Offset(t *testing.T) {
	filters := Filters{
		Page: 1,
		Size: 100,
	}
	actual := filters.offset()
	expected := 0
	if actual != expected {
		t.Errorf("Filters.Offset() = %v; want %v", actual, expected)
	}
}

func TestCalculateMetadata(t *testing.T) {
	tests := []struct {
		totalRecords int
		page         int
		pageSize     int
		parent       int64
		parentName   string
		expected     Metadata
	}{
		{totalRecords: 0, page: 1, pageSize: 10, parent: 0, parentName: "", expected: Metadata{}},
		{totalRecords: 10, page: 1, pageSize: 10, parent: 0, parentName: "", expected: Metadata{
			CurrentPage:  1,
			PageSize:     10,
			FirstPage:    1,
			LastPage:     1,
			TotalRecords: 10,
			NextPage:     0,
			PrevPage:     0,
			ParentId:     0,
			ParentName:   "",
		}},
		{totalRecords: 20, page: 2, pageSize: 10, parent: 0, parentName: "", expected: Metadata{
			CurrentPage:  2,
			PageSize:     10,
			FirstPage:    1,
			LastPage:     2,
			TotalRecords: 20,
			NextPage:     0,
			PrevPage:     1,
			ParentId:     0,
			ParentName:   "",
		}},
		{totalRecords: 30, page: 2, pageSize: 10, parent: 123, parentName: "Test", expected: Metadata{
			CurrentPage:  2,
			PageSize:     10,
			FirstPage:    1,
			LastPage:     3,
			TotalRecords: 30,
			NextPage:     3,
			PrevPage:     1,
			ParentId:     123,
			ParentName:   "Test",
		}},
	}

	for _, test := range tests {
		result := calculateMetadata(test.totalRecords, test.page, test.pageSize, test.parent, test.parentName)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Test failed - expected: %+v, got: %+v", test.expected, result)
		}
	}
}
