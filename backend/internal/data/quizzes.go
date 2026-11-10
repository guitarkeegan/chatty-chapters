package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//CREATE TABLE IF NOT EXISTS quizzes (
//
//    id BIGSERIAL PRIMARY KEY,
//    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
//    book_id BIGINT REFERENCES books(id) ON DELETE CASCADE,
//    chapter_id BIGINT REFERENCES chapters(id) ON DELETE CASCADE,
//    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
//    attempted BOOLEAN DEFAULT false,
//    score INTEGER DEFAULT 0
//);

type Quiz struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	BookID    int64     `json:"book_id"`
	ChapterID int64     `json:"chapter_id"`
	CreatedAt time.Time `json:"created_at"`
	Attempted bool      `json:"attempted"`
	Score     int       `json:"score"`
}

type QuizModel struct {
	DB *pgxpool.Pool
}

func NewQuiz(userID, bookID, chapterID int64) *Quiz {
	return &Quiz{
		UserID:    userID,
		BookID:    bookID,
		ChapterID: chapterID,
	}
}

func (qm QuizModel) Insert(quiz *Quiz) error {

	stmt := `
	INSERT INTO quizzes ( user_id, book_id, chapter_id )
	VALUES ( $1, $2, $3 )
	RETURNING ( id )
	`

	args := []any{quiz.UserID, quiz.BookID, quiz.ChapterID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return qm.DB.QueryRow(ctx, stmt, args...).Scan(&quiz.ID)

}

func (qm QuizModel) Get(quiz *Quiz) error {

	stmt := `
	SELECT id, attempted, created_at, score
	FROM quizzes
	WHERE user_id = $1 AND book_id = $2 AND chapter_id = $3
	`

	args := []any{quiz.UserID, quiz.BookID, quiz.ChapterID}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	err := qm.DB.QueryRow(ctx, stmt, args...).Scan(&quiz.ID, &quiz.Attempted, &quiz.CreatedAt, &quiz.Score)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return ErrRecordNotFound
		default:
			return fmt.Errorf("quiz get: %w", err)
		}
	}

	return nil

}

func (qm QuizModel) Delete(quizID int64) error {

	if quizID <= 0 {
		return fmt.Errorf("QuizModel: Delete: quizID is %d", quizID)
	}

	stmt := `
	DELETE FROM quizzes WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := qm.DB.Exec(ctx, stmt, quizID)
	if err != nil {
		return fmt.Errorf("QuizModel: Delete: %w", err)
	}

	return nil

}
