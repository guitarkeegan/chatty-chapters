package data

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	REQUIRED_NUM_OF_QUIZ_RESPONSES = 3
)

type QuizResponse struct {
	ID              int64     `json:"id"`
	QuizID          int64     `json:"quiz_id"`
	QuizQuestionID  int64     `json:"quiz_question_id"`
	StudentResponse string    `json:"student_response"`
	CreatedAt       time.Time `json:"created_at"`
}

type QuizResponseModel struct {
	DB *pgxpool.Pool
}

func (qqm QuizResponseModel) GetMany(quizID int64) ([]QuizResponse, error) {

	if quizID <= 0 {
		return []QuizResponse{}, fmt.Errorf("QuizResponse: quizID is %d", quizID)
	}

	stmt := `SELECT
	id, quiz_id, quiz_question_id, student_response, created_at
	FROM
	quiz_responses
	WHERE
	quiz_id = $1
	ORDER BY created_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := qqm.DB.Query(ctx, stmt, quizID)
	if err != nil {
		return []QuizResponse{}, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	var responses []QuizResponse
	for rows.Next() {
		var response QuizResponse
		err := rows.Scan(&response.ID, &response.QuizID, &response.QuizQuestionID, &response.StudentResponse, &response.CreatedAt)
		if err != nil {
			return []QuizResponse{}, fmt.Errorf("failed to scan row: %v", err)
		}
		responses = append(responses, response)
	}

	if rows.Err() != nil {
		return []QuizResponse{}, fmt.Errorf("rows error: %v", rows.Err())
	}

	return responses, nil
}

func (qqm QuizResponseModel) Insert(quizResponse *QuizResponse) error {

	stmt := `
	INSERT INTO quiz_responses ( quiz_id, quiz_question_id, student_response )
	VALUES ( $1, $2, $3 )
	RETURNING ( id )
	`

	args := []any{&quizResponse.QuizID, &quizResponse.QuizQuestionID, &quizResponse.StudentResponse}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return qqm.DB.QueryRow(ctx, stmt, args...).Scan(&quizResponse.ID)

}

func (qqm QuizResponseModel) InsertMany(quizResponses []QuizResponse) error {

	if len(quizResponses) != REQUIRED_NUM_OF_QUIZ_RESPONSES {
		return fmt.Errorf("QuizResponseModel InsertMany: not enough responses. Want: %d, Got: %d", REQUIRED_NUM_OF_QUIZ_RESPONSES, len(quizResponses))
	}

	stmt := `
	INSERT INTO quiz_responses ( quiz_id, quiz_question_id, student_response )
	VALUES
	( $1, $2, $3 ),
	( $4, $5, $6 ),
	( $7, $8, $9 )
	`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := qqm.DB.Exec(ctx, stmt, quizResponses[0].QuizID, quizResponses[0].QuizQuestionID, quizResponses[0].StudentResponse, quizResponses[1].QuizID, quizResponses[1].QuizQuestionID, quizResponses[1].StudentResponse, quizResponses[2].QuizID, quizResponses[2].QuizQuestionID, quizResponses[2].StudentResponse)
	if err != nil {
		return fmt.Errorf("QuizResponseModel InsertMany: %w", err)
	}

	return nil

}
