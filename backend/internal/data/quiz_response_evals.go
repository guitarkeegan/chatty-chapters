package data

//CREATE TABLE IF NOT EXISTS quiz_response_evals (
//
//    id BIGSERIAL PRIMARY KEY,
//    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
//    book_id BIGINT REFERENCES books(id) ON DELETE CASCADE,
//    chapter_id BIGINT REFERENCES chapters(id) ON DELETE CASCADE,
//    quiz_id BIGINT REFERENCES quizzes(id) ON DELETE CASCADE,
//    quiz_question_id BIGINT REFERENCES quiz_questions(id) ON DELETE CASCADE,
//    quiz_response_id BIGINT REFERENCES quiz_responses(id) ON DELETE CASCADE,
//    score INTEGER DEFAULT 0,
//    response_feedback TEXT,
//    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
//);

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	REQUIRED_NUM_OF_RESPONSE_EVALS = 3
)

type QuizResponseEval struct {
	ID               int64     `json:"id"`
	UserID           int64     `json:"user_id"`
	BookID           int64     `json:"book_id"`
	ChapterID        int64     `json:"chapter_id"`
	QuizID           int64     `json:"quiz_id"`
	QuizQuestionID   int64     `json:"quiz_question_id"`
	QuizResponseID   int64     `json:"quiz_response_id"`
	Score            int       `json:"score"`
	ResponseFeedback string    `json:"response_feedback"`
	CreatedAt        time.Time `json:"created_at"`
}

type QuizResult struct {
	QuizQuestion     string `json:"quiz_question"`
	StudentResponse  string `json:"quiz_response"`
	Score            int    `json:"score"`
	ResponseFeedback string `json:"response_feedback"`
}

type QuizResponseEvalModel struct {
	DB *pgxpool.Pool
}

func (qrem QuizResponseEvalModel) Insert(quizResponseEval *QuizResponseEval) error {

	stmt := `
	INSERT INTO quiz_response_evals ( user_id, book_id, chapter_id, quiz_id, quiz_question_id, quiz_response_id, score, response_feedback )
	VALUES ( $1, $2, $3, $4, $5, $6, $7, $8 )
	RETURNING ( id )
	`

	args := []any{quizResponseEval.UserID, quizResponseEval.BookID, quizResponseEval.ChapterID, quizResponseEval.QuizID, quizResponseEval.QuizQuestionID, quizResponseEval.QuizResponseID, quizResponseEval.Score, quizResponseEval.ResponseFeedback}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return qrem.DB.QueryRow(ctx, stmt, args...).Scan(&quizResponseEval.ID)

}

func (qrem QuizResponseEvalModel) InsertMany(quizResponseEvals *[]QuizResponseEval) error {

	if len(*quizResponseEvals) < 3 {
		return fmt.Errorf("QuizResponseEval InsertMany: not enough items in params. Want: %d, Got: %d", REQUIRED_NUM_OF_RESPONSE_EVALS, len(*quizResponseEvals))
	}

	stmt := `
	INSERT INTO quiz_response_evals ( user_id, book_id, chapter_id, quiz_id, quiz_question_id, quiz_response_id, score, response_feedback )
	VALUES
	( $1, $2, $3, $4, $5, $6, $7, $8 ),
	( $9, $10, $11, $12, $13, $14, $15, $16 ),
	( $17, $18, $19, $20, $21, $22, $23, $24 )
	`

	args := []any{}
	for _, eval := range *quizResponseEvals {
		args = append(args, eval.UserID, eval.BookID, eval.ChapterID, eval.QuizID, eval.QuizQuestionID, eval.QuizResponseID, eval.Score, eval.ResponseFeedback)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := qrem.DB.Exec(ctx, stmt, args...)
	if err != nil {
		return fmt.Errorf("QuizResponseEval InsertMany: exec: %w", err)
	}

	return nil

}

// get by user, book, chapter
func (qrem QuizResponseEvalModel) GetMany(userID, bookID, chapterID int64) ([]QuizResponseEval, error) {

	stmt := `
	SELECT
	id, quiz_id, quiz_question_id, quiz_response_id, score, response_feedback, created_at
	FROM
	quiz_response_evals
	WHERE user_id = $1 AND book_id = $2 AND chapter_id = $3
	`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	rows, err := qrem.DB.Query(ctx, stmt, userID, bookID, chapterID)
	if err != nil {
		return []QuizResponseEval{}, fmt.Errorf("QuizResponseEval GetMany: rows: %w", err)
	}
	defer rows.Close()

	var quizResponseEvals []QuizResponseEval

	defer rows.Close()

	for rows.Next() {

		var quizResponseEval QuizResponseEval

		err := rows.Scan(&quizResponseEval.ID, &quizResponseEval.QuizID, &quizResponseEval.QuizQuestionID, &quizResponseEval.QuizResponseID, &quizResponseEval.Score, &quizResponseEval.ResponseFeedback, &quizResponseEval.CreatedAt)

		if err != nil {
			return []QuizResponseEval{}, fmt.Errorf("QuizResponseEval row scan: %w", err)
		}

		quizResponseEvals = append(quizResponseEvals, quizResponseEval)

	}

	if rows.Err() != nil {
		return []QuizResponseEval{}, fmt.Errorf("QuizResponseEval after full scan: %w", err)
	}

	return quizResponseEvals, nil

}

func (qrem QuizResponseEvalModel) GetQuizResults(userID, bookID, chapterID int64) ([]*QuizResult, error) {

	stmt := `
	SELECT 
	qq.quiz_question, qr.student_response, qre.score, qre.response_feedback
	FROM quiz_response_evals qre
	JOIN quiz_questions qq ON qre.quiz_question_id = qq.id
	JOIN quiz_responses qr ON qre.quiz_response_id = qr.id
	WHERE qre.user_id = $1 AND qre.book_id = $2 AND qre.chapter_id = $3
	ORDER BY qre.created_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	quizResults := []*QuizResult{}

	rows, err := qrem.DB.Query(ctx, stmt, userID, bookID, chapterID)
	if err != nil {
		return nil, fmt.Errorf("quiz_response_evals GetQuizResults: query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {

		var quizResult QuizResult

		err := rows.Scan(&quizResult.QuizQuestion, &quizResult.StudentResponse, &quizResult.Score, &quizResult.ResponseFeedback)
		if err != nil {
			return nil, fmt.Errorf("quiz_response_evals GetQuizResults rows scan: %w", err)
		}

		quizResults = append(quizResults, &quizResult)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("quiz_response_evals GetQuizResults final scan: %w", rows.Err())
	}

	return quizResults, nil

}
