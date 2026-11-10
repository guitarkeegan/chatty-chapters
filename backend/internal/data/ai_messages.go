package data

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AIMessage struct {
	ID        int64     `json:"id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type AIMessageModel struct {
	DB *pgxpool.Pool
}

func (aim *AIMessageModel) Insert(ctx context.Context, message string, tx pgx.Tx) (*AIMessage, error) {

	stmt := `
	INSERT INTO ai_messages (message)
	VALUES ( $1 )
	RETURNING id, message, created_at
	`
	aiMessage := &AIMessage{}

	err := tx.QueryRow(ctx, stmt, message).Scan(&aiMessage.ID, &aiMessage.Message, &aiMessage.CreatedAt)
	if err != nil {
		return nil, err
	}

	return aiMessage, nil
}
