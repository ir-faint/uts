package repository

import (
	"context"
	"errors"

	"siakad-mini/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id int) (*model.User, error)
	CreateUserTx(ctx context.Context, tx pgx.Tx, user *model.User) error
}

type pgxUserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &pgxUserRepository{db: db}
}

func (r *pgxUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, email, password, role, created_at FROM users WHERE email = $1`
	user := &model.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *pgxUserRepository) FindByID(ctx context.Context, id int) (*model.User, error) {
	query := `SELECT id, email, password, role, created_at FROM users WHERE id = $1`
	user := &model.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *pgxUserRepository) CreateUserTx(ctx context.Context, tx pgx.Tx, user *model.User) error {
	query := `INSERT INTO users (email, password, role) VALUES ($1, $2, $3) RETURNING id, created_at`
	return tx.QueryRow(ctx, query, user.Email, user.Password, user.Role).Scan(&user.ID, &user.CreatedAt)
}
