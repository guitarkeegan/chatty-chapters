package data

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserMessage struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	BookID    int64     `json:"book_id"`
	ChapterID int64     `json:"chapter_id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type UserMessageModel struct {
	DB *pgxpool.Pool
}

func (um *UserMessageModel) Insert(ctx context.Context, userID, bookID, chapterID int64, message string, tx pgx.Tx) (*UserMessage, error) {

	stmt := `
	INSERT INTO user_messages (user_id, book_id, chapter_id, message)
	VALUES ( $1, $2, $3, $4 )
	RETURNING id, user_id, message, created_at
	`

	userMessage := &UserMessage{}

	err := tx.QueryRow(ctx, stmt, userID, bookID, chapterID, message).Scan(&userMessage.ID, &userMessage.UserID, &userMessage.Message, &userMessage.CreatedAt)
	if err != nil {
		return nil, err
	}

	return userMessage, nil
}

// return all user questions for question analysis
func (umm *UserMessageModel) GetAll(ctx context.Context, userID, bookID, chapterID int64) ([]string, error) {

	stmt := `
	SELECT message
	FROM user_messages
	WHERE user_id = $1 AND book_id = $2 AND chapter_id = $3
	ORDER BY created_at ASC
	`

	rows, err := umm.DB.Query(ctx, stmt, userID, bookID, chapterID)
	if err != nil {
		return nil, fmt.Errorf("GetAll: %w", err)
	}
	defer rows.Close()

	var messages []string
	for rows.Next() {

		var message string
		err := rows.Scan(&message)
		if err != nil {
			return nil, fmt.Errorf("err scanning row: %w", err)
		}

		// this should be trimmed already now
		messages = append(messages, message)
	}

	return messages, nil

}
