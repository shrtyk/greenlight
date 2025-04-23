package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/jackc/pgtype"
	"github.com/shortykevich/greenlight/internal/validator"
)

type MovieRepository interface {
	MovieReader
	MovieWriter
}

type MovieWriter interface {
	Insert(movie *Movie) error
	Delete(id int64) error
	Update(movie *Movie) error
}

type MovieReader interface {
	GetByID(id int64) (*Movie, error)
	GetAll(title string, genres Genres, filters Filters) ([]*Movie, Metadata, error)
}

type Movie struct {
	ID        int64     `json:"id,omitempty"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title,omitempty"`
	Year      int32     `json:"year,omitempty"`
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    Genres    `json:"genres,omitempty"`
	Version   int32     `json:"version,omitempty"`
}

type Genres []string

func (g *Genres) Scan(src any) error {
	var arr pgtype.TextArray
	if err := arr.Scan(src); err != nil {
		return err
	}

	genres := make([]string, len(arr.Elements))
	for i, el := range arr.Elements {
		genres[i] = el.String
	}

	*g = genres
	return nil
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		&movie.Genres,
		&movie.Version,
	)
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := m.DB.ExecContext(ctx, query, id)
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m MovieModel) GetAll(title string, genres Genres, filters Filters) ([]*Movie, Metadata, error) {
	// #nosec G201 -- filters validated in handler
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (genres @> $2 OR $2 = '{}')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	args := []any{title, genres, filters.limit(), filters.offset()}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer func() {
		cerr := rows.Close()
		if cerr != nil {
			if err != nil {
				err = fmt.Errorf("%w: %v", err, cerr)
			} else {
				err = fmt.Errorf("%w: %v", ErrCloseRows, cerr)
			}
		}
	}()

	movies := []*Movie{}
	totalRecords := 0
	for rows.Next() {
		var movie Movie

		err = rows.Scan(
			&totalRecords,
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			&movie.Genres,
			&movie.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		movies = append(movies, &movie)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return movies, metadata, nil
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(int(movie.Year) <= time.Now().Year(), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")

	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

type MovieInMemRepo struct {
	mu        sync.RWMutex
	idCounter int64
	movies    map[int64]*Movie
}

func (m *MovieInMemRepo) Insert(movie *Movie) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	movie.ID = m.idCounter
	movie.Version++

	m.movies[m.idCounter] = movie
	m.idCounter++

	return nil
}

func (m *MovieInMemRepo) GetByID(id int64) (*Movie, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

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
	m.mu.Lock()
	defer m.mu.Unlock()

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
	m.mu.Lock()
	defer m.mu.Unlock()

	id := movie.ID
	if _, ok := m.movies[id]; !ok {
		return ErrRecordNotFound
	}

	movie.Version = m.movies[id].Version + 1
	m.movies[id] = movie
	return nil
}

func (m *MovieInMemRepo) GetAll(_ string, _ Genres, _ Filters) ([]*Movie, Metadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	moviesList := make([]*Movie, 0, len(m.movies))
	for _, v := range m.movies {
		movie := v
		moviesList = append(moviesList, movie)
	}
	sort.Slice(moviesList, func(i, j int) bool {
		return moviesList[i].ID < moviesList[j].ID
	})
	return moviesList, Metadata{}, nil
}
