package data

import (
	"reflect"
	"testing"

	"github.com/shrtyk/greenlight/internal/validator"
)

func TestFilters(t *testing.T) {
	filters := Filters{
		Page:     1,
		PageSize: 20,
		Sort:     "-title",
		SortSafelist: []string{
			"id", "title", "year", "runtime",
			"-id", "-title", "-year", "-runtime",
		},
	}

	v := validator.New()

	filters.Validate(v)
	if !v.Valid() {
		t.Errorf("expected to get valid filters but got: %v", v.Errors)
	}

	gotSortCol := filters.sortColumn()
	if gotSortCol != "title" {
		t.Errorf("got: %s, want: %s", gotSortCol, "title")
	}

	gotSortDir := filters.sortDirection()
	if gotSortDir != "DESC" {
		t.Errorf("got: %s, want: %s", gotSortCol, "DESC")
	}

	totalRecs := 5
	gotMeta := calculateMetadata(totalRecs, filters.Page, filters.PageSize)
	wantMeta := Metadata{
		CurrentPage:  filters.Page,
		PageSize:     filters.PageSize,
		FirstPage:    1,
		LastPage:     1,
		TotalRecords: totalRecs,
	}
	if !reflect.DeepEqual(gotMeta, wantMeta) {
		t.Errorf("got: %+v, want: %+v", gotMeta, wantMeta)
	}
}
