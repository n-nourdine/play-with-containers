package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MovieStream struct {
	db *pgxpool.Pool
}

type Movies struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func NewConn() (*MovieStream, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	dbpool, err := pgxpool.New(ctx,
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
			os.Getenv("INVENTORY_DB_USER"),
			os.Getenv("INVENTORY_DB_PASSWORD"),
			os.Getenv("INVENTORY_DB_HOST"),
			os.Getenv("INVENTORY_DB_PORT"),
			os.Getenv("INVENTORY_DB_NAME"),
		))

	if err != nil {
		return nil, fmt.Errorf("erreur de connexion à la base de données: %w", err)
	}

	err = dbpool.Ping(ctx)
	if err != nil {
		fmt.Printf("user: %v; db: %v, password: %v \n", os.Getenv("INVENTORY_DB_USER"), os.Getenv("INVENTORY_DB_NAME"), os.Getenv("INVENTORY_DB_PASSWORD"))
		return nil, fmt.Errorf("impossible de ping la base de données: %w", err)
	}

	return &MovieStream{db: dbpool}, nil
}

func (m *MovieStream) Close() {
	m.db.Close()
}

func (m *MovieStream) Add(ctx context.Context, movie Movies) error {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("erreur lors du début de la transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := m.db.Exec(ctx, "INSERT INTO movies (id,title,description) VALUES ($1,$2,$3)", movie.ID, movie.Title, movie.Description); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("erreur lors du commit de la transaction: %w", err)
	}

	return nil
}

func (m *MovieStream) GetById(ctx context.Context, id string) (Movies, error) {
	var movie Movies
	err := m.db.QueryRow(ctx, "SELECT id,title, description FROM movies WHERE id=$1", id).Scan(&movie.ID, &movie.Title, &movie.Description)
	return movie, err
}

func (m *MovieStream) ListeByTitle(ctx context.Context, title string) ([]Movies, error) {
	rows, err := m.db.Query(ctx, "SELECT id,title, description FROM movies WHERE title=$1", title)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []Movies

	for rows.Next() {
		var movie Movies

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

func (m *MovieStream) Liste(ctx context.Context) ([]Movies, error) {
	rows, err := m.db.Query(ctx, "SELECT id, title, description FROM movies")
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des films: %w", err)
	}
	defer rows.Close()

	var movies []Movies

	for rows.Next() {
		var movie Movies

		err := rows.Scan(&movie.ID, &movie.Title, &movie.Description)
		if err != nil {
			return nil, fmt.Errorf("erreur lors du scan : %w", err)
		}

		movies = append(movies, movie)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors de l'itération des lignes: %w", err)
	}

	return movies, nil
}

func (m *MovieStream) Update(ctx context.Context, movie Movies) error {

	commandTag, err := m.db.Exec(ctx, "UPDATE movies SET title=$1, description=$2 WHERE id=$3", movie.Title, movie.Description, movie.ID)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (m *MovieStream) Delete(ctx context.Context, id string) error {
	commandTag, err := m.db.Exec(ctx, "DELETE FROM movies WHERE id=$1", id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (m *MovieStream) DeleteAll(ctx context.Context) error {
	commandTag, err := m.db.Exec(ctx, "DELETE FROM movies")
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
