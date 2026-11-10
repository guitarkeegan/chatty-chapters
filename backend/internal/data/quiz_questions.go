package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	NumberOfRequiredQuizQuestions = 3
)

//CREATE TABLE IF NOT EXISTS quiz_questions (
//
//    id BIGSERIAL PRIMARY KEY,
//    quiz_id BIGINT REFERENCES quizzes(id) ON DELETE CASCADE,
//    quiz_question TEXT,
//    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
//);

type QuizQuestion struct {
	ID           int64     `json:"id"`
	QuizID       int64     `json:"quiz_id"`
	QuizQuestion string    `json:"quiz_question"`
	CreatedAt    time.Time `json:"created_at"`
}

type QuizQuestionModel struct {
	DB *pgxpool.Pool
}

func (qqm QuizQuestionModel) GetMany(quizID int64) ([]QuizQuestion, error) {

	if quizID < 1 {
		return []QuizQuestion{}, ErrRecordNotFound
	}

	stmt := `
	SELECT id, quiz_id, quiz_question, created_at
	FROM quiz_questions
	WHERE quiz_id = $1
	ORDER BY
	created_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	rows, err := qqm.DB.Query(ctx, stmt, quizID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []QuizQuestion{}, nil
		}
	}

	defer rows.Close()

	var questions []QuizQuestion
	for rows.Next() {
		var question QuizQuestion

		err := rows.Scan(&question.ID, &question.QuizID, &question.QuizQuestion, &question.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("QuizQuestion scan row: %w", err)
		}

		questions = append(questions, question)
	}

	if rows.Err() != nil {
		return []QuizQuestion{}, fmt.Errorf("QuizQuestion after scan all: %w", rows.Err())
	}

	return questions, nil

}

func (qqm QuizQuestionModel) Insert(quizQuestion *QuizQuestion) error {

	stmt := `
	INSERT INTO quiz_questions ( quiz_id, quiz_question )
	VALUES ( $1, $2 )
	RETURNING ( id )
	`

	args := []any{&quizQuestion.QuizID, &quizQuestion.QuizQuestion}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return qqm.DB.QueryRow(ctx, stmt, args...).Scan(&quizQuestion.ID)

}

// planning for 3 questions
func (qqm QuizQuestionModel) InsertMany(quizQuestions []QuizQuestion) error {
	dbg(" QuizQuestionsModel start")
	dbg(" params: %+v", quizQuestions)

	stmt := `
	INSERT INTO quiz_questions (quiz_id, quiz_question)
	VALUES 
	($1, $2),
	($1, $3),
	($1, $4)
	`

	if len(quizQuestions) < NumberOfRequiredQuizQuestions {
		return fmt.Errorf("QuizQuestionModel insert many: less than 3 quiz questions")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	dbg(" quizID: %d", quizQuestions[0].QuizID)
	args := []any{quizQuestions[0].QuizID, quizQuestions[0].QuizQuestion, quizQuestions[1].QuizQuestion, quizQuestions[2].QuizQuestion}

	_, err := qqm.DB.Exec(ctx, stmt, args...)
	if err != nil {
		return fmt.Errorf("insert many: %w", err)
	}

	return nil
}
