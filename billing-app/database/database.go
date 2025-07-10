package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderStore struct {
	db *pgxpool.Pool
}

type Order struct {
	ID            string `json:"id"`
	UserID        string `json:"user_id"`
	NumberOfItems string `json:"number_of_items"`
	TotalAmount   string `json:"total_amount"`
}

func NewConn() (*OrderStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	dbpool, err := pgxpool.New(ctx,
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
			os.Getenv("BILLING_DB_USER"),
			os.Getenv("BILLING_DB_PASSWORD"),
			os.Getenv("BILLING_DB_HOST"),
			os.Getenv("BILLING_DB_PORT"),
			os.Getenv("BILLING_DB_NAME"),
		))

	if err != nil {
		return nil, fmt.Errorf("erreur de connexion à la base de données: %w", err)
	}

	err = dbpool.Ping(ctx)
	if err != nil {
		fmt.Printf("user: %v; db: %v, password: %v \n", os.Getenv("BILLING_DB_USER"), os.Getenv("BILLING_DB_NAME"), os.Getenv("BILLING_DB_PASSWORD"))
		return nil, fmt.Errorf("impossible de ping la base de données: %w", err)
	}

	return &OrderStore{db: dbpool}, nil
}

func (o *OrderStore) Close() {
	o.db.Close()
}

func (o *OrderStore) CreateOrder(ctx context.Context, order Order) error {
	tx, err := o.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("erreur lors du début de la transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := o.db.Exec(ctx, "INSERT INTO orders (id, user_id, number_of_items, total_amount) VALUES ($1, $2, $3, $4)", order.ID, order.UserID, order.NumberOfItems, order.TotalAmount); err != nil {

		return fmt.Errorf("%v erreur lors de l'insertion de la commande: %w", order, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("erreur lors du commit de la transaction: %w", err)
	}

	return nil
}

func (o *OrderStore) GetAllOrders(ctx context.Context) ([]Order, error) {
	rows, err := o.db.Query(ctx, `SELECT id, user_id, number_of_items, total_amount FROM orders`)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("sql no row")
		}
		return nil, fmt.Errorf("erreur lors de la récupération des commandes: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		err := rows.Scan(&order.ID, &order.UserID, &order.NumberOfItems, &order.TotalAmount)
		if err != nil {
			return nil, fmt.Errorf("erreur lors du scan de la commande: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors de l'itération des lignes: %w", err)
	}

	return orders, nil
}
