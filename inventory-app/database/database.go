package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/n-nourdine/play-with-containers/inventory-app/model"
)

type MovieStream struct {
	db *pgxpool.Pool
}

// Créer un pool de connexions
func NewConn(connString string) (*MovieStream, error) {
	// Créer un pool de connexions
	dbpool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("erreur de connexion à la base de données: %w", err)
	}

	// Vérifier la connexion
	err = dbpool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("impossible de ping la base de données: %w", err)
	}

	return &MovieStream{db: dbpool}, nil
}

// Fermer la connection
func (m *MovieStream) Close() {
	m.db.Close()
}

func (m *MovieStream) InitTable() error {
	_, err := m.db.Exec(context.Background(), "CREATE TABLE movies (id TEXT PRIMARY KEY, title VARCHAR(255) NOT NULL, description TEXT;")
	return err
}

func (m *MovieStream) Add(ctx context.Context, movie model.Movies) error {
	_, err := m.db.Exec(ctx, "INSERT INTO movies (id,title,description) VALUES ($1,$2,$3)", movie.ID, movie.Title, movie.Description)
	return err
}

func (m *MovieStream) GetById(ctx context.Context, id string) (model.Movies, error) {
	var movie model.Movies
	err := m.db.QueryRow(ctx, "SELECT id,title, description FROM movies WHERE id=$1", id).Scan(&movie.ID, &movie.Title, &movie.Description)
	return movie, err
}

func (m *MovieStream) ListeByTitle(ctx context.Context, title string) ([]model.Movies, error) {
	rows, err := m.db.Query(ctx, "SELECT id,title, description FROM movies WHERE title=$1", title)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []model.Movies

	for rows.Next() {
		var movie model.Movies

		err := rows.Scan(&movie.ID, &movie.Title, &movie.Description)
		if err != nil {
			return nil, err
		}

		movies = append(movies, movie)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}

func (m *MovieStream) Liste(ctx context.Context) ([]model.Movies, error) {
	rows, err := m.db.Query(ctx, "SELECT id, title, description FROM movies")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []model.Movies

	for rows.Next() {
		var movie model.Movies

		err := rows.Scan(&movie.ID, &movie.Title, &movie.Description)
		if err != nil {
			return nil, err
		}

		movies = append(movies, movie)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}

func (m *MovieStream) Update(ctx context.Context, id, title, description string) error {
	commandTag, err := m.db.Exec(ctx, "UPDATE movies SET title=$1, description=$2 WHERE id=$3", title, description, id)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("movie with ID %s not found", id)
	}

	return nil
}
func (m *MovieStream) Delete(ctx context.Context, id string) error {
	commandTag, err := m.db.Exec(ctx, "DELETE FROM movies WHERE id=$1", id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("movie with ID %s not found", id)
	}

	return nil
}
