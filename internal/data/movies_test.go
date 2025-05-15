package data_test

import (
	"testing"

	"github.com/shrtyk/greenlight/internal/data"
	"github.com/shrtyk/greenlight/internal/testutils/assertions"
)

func TestMovies(t *testing.T) {
	models := data.NewMockModels()
	movies := models.Movies

	moana := &data.Movie{
		Title:   "Moana",
		Year:    2016,
		Runtime: 107,
		Genres:  data.Genres{"animation", "adventure"},
	}
	err := movies.Insert(moana)
	assertions.AssertNoError(t, err)

	gotMov, err := movies.GetByID(1)
	assertions.AssertNoError(t, err)
	assertions.AssertMovies(t, *gotMov, *moana)

	err = movies.Insert(moana)
	assertions.AssertExpectedError(t, err)

	blackPanther := &data.Movie{
		Title:   "Black Panther",
		Year:    2018,
		Runtime: 134,
		Genres:  data.Genres{"action", "adventure"},
	}
	_ = movies.Insert(blackPanther)

	deadPool := &data.Movie{
		Title:   "Deadpool",
		Year:    2016,
		Runtime: 108,
		Genres:  data.Genres{"action", "comedy"},
	}
	_ = movies.Insert(deadPool)

	err = movies.Delete(3)
	assertions.AssertNoError(t, err)

	_, err = movies.GetByID(3)
	assertions.AssertNotFoundError(t, err)

	moana.Title = "moana"
	err = movies.Update(moana)
	assertions.AssertNoError(t, err)

	gotMov, _ = movies.GetByID(1)
	if gotMov.Title == moana.Title && gotMov.Version != 2 {
		t.Errorf("expected title: '%s', but got old one: '%s'\nexpected version: %d, got: %d", moana.Title, gotMov.Title, 2, gotMov.Version)
	}

	filters := data.Filters{
		Page:     1,
		PageSize: 20,
		Sort:     "id",
		SortSafelist: []string{
			"id", "title", "year", "runtime",
			"-id", "-title", "-year", "-runtime",
		},
	}

	gotMovs, gotMeta, err := movies.GetAll("", data.Genres{}, filters)
	assertions.AssertNoError(t, err)
	assertions.AssertMovieLists(t, gotMovs, []*data.Movie{moana, blackPanther})
	assertions.AssertMoviesMetadata(t, gotMeta, data.Metadata{
		CurrentPage:  1,
		PageSize:     20,
		FirstPage:    1,
		LastPage:     1,
		TotalRecords: 2,
	})

	filters.PageSize = 1
	filters.Sort = "-id"

	gotMovs, gotMeta, _ = movies.GetAll("", data.Genres{}, filters)
	assertions.AssertMovieLists(t, gotMovs, []*data.Movie{blackPanther})
	assertions.AssertMoviesMetadata(t, gotMeta, data.Metadata{
		CurrentPage:  1,
		PageSize:     1,
		FirstPage:    1,
		LastPage:     2,
		TotalRecords: 2,
	})
	gotMovs, _, _ = movies.GetAll("mo", data.Genres{}, filters)
	assertions.AssertMovieLists(t, gotMovs, []*data.Movie{moana})

	filters.PageSize = 5

	gotMovs, gotMeta, _ = movies.GetAll("", data.Genres{"adventure"}, filters)
	assertions.AssertMovieLists(t, gotMovs, []*data.Movie{blackPanther, moana})
	assertions.AssertMoviesMetadata(t, gotMeta, data.Metadata{
		CurrentPage:  1,
		PageSize:     5,
		FirstPage:    1,
		LastPage:     1,
		TotalRecords: 2,
	})
}
