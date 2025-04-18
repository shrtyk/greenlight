package data

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgtype"
	"github.com/shortykevich/greenlight/internal/validator"
)

type Movie struct {
	ID        int64     `json:"id,omitempty"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title,omitempty"`
	Year      int32     `json:"year,omitempty"`
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version,omitempty"`
}

type MovieRepository interface {
	MovieGetter
	MovieMutator
}

type MovieMutator interface {
	Insert(movie *Movie) error
	Delete(id int64) error
	Update(movie *Movie) error
}

type MovieGetter interface {
	GetByID(id int64) (*Movie, error)
	GetAll() ([]Movie, error)
}

type MovieModel struct {
	DB *sql.DB
}

func (m MovieModel) Insert(movie *Movie) error {
	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`
	args := []any{movie.Title, movie.Year, movie.Runtime, movie.Genres}

	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) GetByID(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	movie := new(Movie)
	query := `
		SELECT id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE id = $1`

	var genres pgtype.TextArray
	err := m.DB.QueryRow(query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		&genres,
		&movie.Version,
	)

	movie.Genres = make([]string, 0, len(genres.Elements))
	for _, genre := range genres.Elements {
		movie.Genres = append(movie.Genres, genre.String)
	}

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return movie, nil
}

func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM movies
		WHERE id = $1`

	res, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m MovieModel) Update(movie *Movie) error {
	query := `
		UPDATE movies
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5 and version = $6
		RETURNING version`

	args := []any{movie.Title, movie.Year, movie.Runtime, movie.Genres, movie.ID, movie.Version}

	if err := m.DB.QueryRow(query, args...).Scan(&movie.Version); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m MovieModel) GetAll() ([]Movie, error) {
	query := `
		SELECT title, year, runtime, genres
		FROM movies`

	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("%w: %v", ErrCloseRows, cerr)
		}
	}()

	movies := []Movie{}
	var movie Movie
	for rows.Next() {
		err := rows.Scan(&movie.Title, &movie.Year, &movie.Runtime, &movie.Genres)
		if err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return movies, nil
}

type MovieInMemRepo struct {
	// mu        sync.Mutex
	idCounter int64
	movies    map[int64]*Movie
}

func (m *MovieInMemRepo) Insert(movie *Movie) error {
	movie.ID = m.idCounter
	movie.Version++

	m.movies[m.idCounter] = movie
	m.idCounter++
	return nil
}

func (m *MovieInMemRepo) GetByID(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	movie, ok := m.movies[id]
	if !ok {
		return nil, ErrRecordNotFound
	}
	return movie, nil
}

func (m *MovieInMemRepo) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	if _, ok := m.movies[id]; !ok {
		return ErrRecordNotFound
	}

	delete(m.movies, id)
	return nil
}
func (m *MovieInMemRepo) Update(movie *Movie) error {
	id := movie.ID
	if _, ok := m.movies[id]; !ok {
		return ErrRecordNotFound
	}

	movie.Version = m.movies[id].Version + 1
	m.movies[id] = movie
	return nil
}

func (m *MovieInMemRepo) GetAll() ([]Movie, error) {
	moviesList := make([]Movie, 0, len(m.movies))
	for _, v := range m.movies {
		moviesList = append(moviesList, *v)
	}
	sort.Slice(moviesList, func(i, j int) bool {
		return moviesList[i].ID < moviesList[j].ID
	})
	return moviesList, nil
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")

	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}
