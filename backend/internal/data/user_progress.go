package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//CREATE TABLE IF NOT EXISTS user_progress (
//    id BIGSERIAL PRIMARY KEY,
//    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
//    book_id BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
//    chapter_chats_completed INTEGER[] DEFAULT '{}',
//    chat_score_attempted INTEGER NOT NULL DEFAULT 0,
//    chats_achieved INTEGER NOT NULL DEFAULT 0,
//    quizzes_score_attempted INTEGER NOT NULL DEFAULT 0,
//    quizzes_score_achieved INTEGER NOT NULL DEFAULT 0,
//    quiz_chapters_completed INTEGER[] DEFAULT '{}',
//    total_score_attempted INTEGER NOT NULL DEFAULT 0,
//    total_score_achieved INTEGER NOT NULL DEFAULT 0
//);

type UserProgress struct {
	ID                    int64 `json:"id"`
	UserID                int64 `json:"user_id"`
	BookID                int64 `json:"book_id"`
	ChapterChatsCompleted []int `json:"chapter_chats_completed"`
	ChatScoreAttempted    int   `json:"chat_score_attempted"`
	ChatsAchieved         int   `json:"chats_achieved"`
	QuizzesScoreAttempted int   `json:"quizzes_score_attempted"`
	QuizzesScoreAchieved  int   `json:"quizzes_score_achieved"`
	QuizChaptersCompleted []int `json:"quiz_chapters_completed"`
	TotalScoreAttempted   int   `json:"total_score_attempted"`
	TotalScoreAchieved    int   `json:"total_score_achieved"`
}

type UserProgressModel struct {
	DB *pgxpool.Pool
}

// Get retrieves a UserProgress record from the database based on UserID and BookID
func (upm *UserProgressModel) Get(ctx context.Context, userID, bookID int64) (*UserProgress, error) {
	query := `
		SELECT id, user_id, book_id, chapter_chats_completed, chat_score_attempted,
			chats_achieved, quizzes_score_attempted, quizzes_score_achieved,
			quiz_chapters_completed, total_score_attempted, total_score_achieved
		FROM user_progress
		WHERE user_id = $1 AND book_id = $2
	`

	var up UserProgress

	row := upm.DB.QueryRow(ctx, query, userID, bookID)
	err := row.Scan(
		&up.ID,
		&up.UserID,
		&up.BookID,
		&up.ChapterChatsCompleted,
		&up.ChatScoreAttempted,
		&up.ChatsAchieved,
		&up.QuizzesScoreAttempted,
		&up.QuizzesScoreAchieved,
		&up.QuizChaptersCompleted,
		&up.TotalScoreAttempted,
		&up.TotalScoreAchieved,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.New("user progress not found")
		}
		return nil, err
	}

	return &up, nil
}

// GetAll retrieves all UserProgress records associated with the given UserID
func (upm *UserProgressModel) GetAll(ctx context.Context, userID int64) ([]UserProgress, error) {
	query := `
		SELECT id, user_id, book_id, chapter_chats_completed, chat_score_attempted,
			chats_achieved, quizzes_score_attempted, quizzes_score_achieved,
			quiz_chapters_completed, total_score_attempted, total_score_achieved
		FROM user_progress
		WHERE user_id = $1
	`

	rows, err := upm.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("GetAll: rows: %w", err)
	}
	defer rows.Close()

	var progresses []UserProgress

	for rows.Next() {
		var up UserProgress
		err := rows.Scan(
			&up.ID,
			&up.UserID,
			&up.BookID,
			&up.ChapterChatsCompleted,
			&up.ChatScoreAttempted,
			&up.ChatsAchieved,
			&up.QuizzesScoreAttempted,
			&up.QuizzesScoreAchieved,
			&up.QuizChaptersCompleted,
			&up.TotalScoreAttempted,
			&up.TotalScoreAchieved,
		)
		if err != nil {
			return nil, fmt.Errorf("GetAll: rows.Next: %w", err)
		}
		progresses = append(progresses, up)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("GetAll: rows.Err: %w", err)
	}

	return progresses, nil
}

// TODO: when to create this for the user
func (upm UserProgressModel) Insert(userProgress *UserProgress) error {

	stmt := `
	INSERT INTO user_progress ( user_id, book_id )
	VALUES ( $1, $2 )
	RETURNING ( id )
	`

	args := []any{userProgress.UserID, userProgress.BookID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return upm.DB.QueryRow(ctx, stmt, args...).Scan(&userProgress.ID)

}

func (upm *UserProgressModel) Update(userProgress *UserProgress, chapterNumber int) error {

	stmt := `
	UPDATE user_progress
	SET
	  chapter_chats_completed = array_append(chapter_chats_completed, $1),
	  chat_score_attempted = chat_score_attempted + 1,
	  chats_achieved = chats_achieved + 1,
	  total_score_attempted = total_score_attempted + 1,
	  total_score_achieved = total_score_achieved + 1
	WHERE 
	  user_id = $2 AND book_id = $3;
	`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := upm.DB.Exec(ctx, stmt, chapterNumber, userProgress.UserID, userProgress.BookID)
	if err != nil {
		return fmt.Errorf("userProgress update: %w", err)
	}

	return nil

}

func (upm *UserProgressModel) UpdateQuiz(userID, bookID int64, quizScore, chapterNumber int) error {

	stmt := `
	UPDATE user_progress
	SET
  	quizzes_score_attempted = quizzes_score_attempted + 3,
  	quizzes_score_achieved = quizzes_score_achieved + $1,
	quiz_chapters_completed = array_append(quiz_chapters_completed, $2),
	total_score_attempted = total_score_attempted + 3,
	total_score_achieved = total_score_achieved + $1
	WHERE 
	user_id = $3 AND book_id = $4
	`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := upm.DB.Exec(ctx, stmt, quizScore, chapterNumber, userID, bookID)
	if err != nil {
		return fmt.Errorf("userProgress updateQuiz: %w", err)
	}

	return nil

}
