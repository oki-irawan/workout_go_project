package store

import (
	"database/sql"
	"github.com/oki-irawan/fem_project/internal/tokens"
	"time"
)

type PostgresTokenStore struct {
	db *sql.DB
}

func NewPostgresTokenStore(db *sql.DB) *PostgresTokenStore {
	return &PostgresTokenStore{
		db: db,
	}
}

type TokenStore interface {
	Insert(token *tokens.Token) error
	CreateNewToken(userID int64, ttl time.Duration, scope string) (*tokens.Token, error)
	DeleteAllTokensForUser(userID int64, scope string) error
}

func (p *PostgresTokenStore) CreateNewToken(userID int64, ttl time.Duration, scope string) (*tokens.Token, error) {
	token, err := tokens.GenerateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = p.Insert(token)
	return token, nil
}

func (p *PostgresTokenStore) Insert(token *tokens.Token) error {
	query := `
		INSERT INTO tokens (hash, user_id, expiry, scope)
		VALUES ($1, $2, $3, $4)
	`
	_, err := p.db.Exec(query, token.Hash, token.UserID, token.Expiry, token.Scope)
	return err
}

func (p *PostgresTokenStore) DeleteAllTokensForUser(userID int64, scope string) error {
	query := `DELETE FROM tokens WHERE scope = $1 AND user_id = $2`

	_, err := p.db.Exec(query, scope, userID)
	return err
}
