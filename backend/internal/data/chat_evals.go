package data

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//CREATE TABLE IF NOT EXISTS chat_evals (
//    id BIGSERIAL PRIMARY KEY,
//    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
//    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
//    book_id BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
//    chapter_id BIGINT NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
//    score INTEGER NOT NULL DEFAULT 0 CHECK (score BETWEEN 0 AND 3),
//    feedback TEXT NOT NULL DEFAULT ''
//);

type ChatEval struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UserID    int64     `json:"user_id"`
	BookID    int64     `json:"book_id"`
	ChapterID int64     `json:"chapter_id"`
	Score     int       `json:"score"`
	Feedback  string    `json:"feedback"`
}

type ChatEvalModel struct {
	DB *pgxpool.Pool
}

func NewChatEval(userID, bookID, chapterID int64) *ChatEval {
	return &ChatEval{
		UserID:    userID,
		BookID:    bookID,
		ChapterID: chapterID,
	}
}

func (ce *ChatEvalModel) Insert(chatEval *ChatEval) error {

	stmt := `
	INSERT INTO chat_evals
	( user_id, book_id, chapter_id, score, feedback )
	VALUES
	( $1, $2, $3, $4, $5 )
	RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := ce.DB.QueryRow(ctx, stmt, chatEval.UserID, chatEval.BookID, chatEval.ChapterID, chatEval.Score, chatEval.Feedback).Scan(&chatEval.ID, &chatEval.CreatedAt)
	if err != nil {
		return fmt.Errorf("ChatEvalModel insert: %w", err)
	}

	return nil

}

func (ce *ChatEvalModel) Get(ctx context.Context, chatEvals *[]*ChatEval) error {
	if len(*chatEvals) == 0 {
		return fmt.Errorf("chatEvals is empty")
	}

	stmt := `
	SELECT id, created_at, score, feedback
	FROM chat_evals
	WHERE user_id = $1 AND book_id = $2 AND chapter_id = $3
	`

	args := []any{(*chatEvals)[0].UserID, (*chatEvals)[0].BookID, (*chatEvals)[0].ChapterID}

	rows, err := ce.DB.Query(ctx, stmt, args...)
	if err != nil {
		return fmt.Errorf("chatEval Get query: %w", err)
	}
	defer rows.Close()

	foundEvals := []*ChatEval{}

	for rows.Next() {
		chatEval := &ChatEval{
			UserID:    (*chatEvals)[0].UserID,
			BookID:    (*chatEvals)[0].BookID,
			ChapterID: (*chatEvals)[0].ChapterID,
		}
		err = rows.Scan(&chatEval.ID, &chatEval.CreatedAt, &chatEval.Score, &chatEval.Feedback)
		if err != nil {
			return fmt.Errorf("chatEval Get scan: %w", err)
		}
		foundEvals = append(foundEvals, chatEval)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("chatEval Get rows: %w", err)
	}

	if len(foundEvals) == 0 {
		return pgx.ErrNoRows
	}

	// Clear the original slice and append the found evaluations
	*chatEvals = (*chatEvals)[:0]
	*chatEvals = append(*chatEvals, foundEvals...)

	return nil

}
