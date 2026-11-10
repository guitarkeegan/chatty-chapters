package data

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserActivity struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	BookID       int64     `json:"book_id"`
	ChapterID    int64     `json:"chapter_id"`
	ActivityType string    `json:"activity_type"`
	Score        int       `json:"score"`
	CompletedAt  time.Time `json:"completed_at"`
	Attempted    bool      `json:"attempted"`
}

type UserActivityModel struct {
	DB *pgxpool.Pool
}

// will create a new rows for user and all chapters
// problem, this needs to be called for each activity type, so user 3 for each user for each chapter
func (uam *UserActivityModel) CreateActivitiesForUser(ctx context.Context, userID, bookID int64, chapterIDs []int64) error {
	// Start a transaction
	tx, err := uam.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Prepare the insert statement
	stmt := `
        INSERT INTO user_activities (user_id, book_id, chapter_id, activity_type, score, completed_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

	// Insert each activity
	for _, chapterID := range chapterIDs {
		at := []string{"chat", "quiz", "listening"}
		for _, activity := range at {
			_, err := tx.Exec(ctx, stmt, userID, bookID, chapterID, activity, 0, time.Now())
			if err != nil {
				return fmt.Errorf("failed to execute statement: %w", err)
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (uam *UserActivityModel) GetActivityByUserBookChapter(ctx context.Context, userActivity *UserActivity) error {

	stmt := `
	SELECT id, score, completed_at, attempted
	FROM user_activities
	WHERE user_id = $1 AND book_id = $2 AND chapter_id = $3 AND activity_type = $4
	`

	err := uam.DB.QueryRow(ctx, stmt, userActivity.UserID, userActivity.BookID, userActivity.ChapterID, userActivity.ActivityType).Scan(&userActivity.ID, &userActivity.Score, &userActivity.CompletedAt, &userActivity.Attempted)

	if err != nil {
		return fmt.Errorf("couldn't get activity: %w", err)
	}

	return nil

}

func (uam *UserActivityModel) GetActivitiesByUserBookChapter(ctx context.Context, userID, bookID, chapterID int64) ([]*UserActivity, error) {

	order := "CASE activity_type " +
		"WHEN 'chat' THEN 1 " +
		"WHEN 'quiz' THEN 2 " +
		"WHEN 'listening' THEN 3 " +
		"END"

	stmt := `
	SELECT id, user_id, book_id, chapter_id, activity_type, score, completed_at, attempted
	FROM user_activities
	WHERE user_id = $1 AND book_id = $2 AND chapter_id = $3
	ORDER BY ` + order

	rows, err := uam.DB.Query(ctx, stmt, userID, bookID, chapterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	activities := []*UserActivity{}

	for rows.Next() {
		var activity UserActivity

		err := rows.Scan(&activity.ID, &activity.UserID, &activity.BookID, &activity.ChapterID, &activity.ActivityType, &activity.Score, &activity.CompletedAt, &activity.Attempted)
		if err != nil {
			return nil, err
		}

		activities = append(activities, &activity)

		fmt.Printf("Scanned Activity: %+v\n", activity) // Debug print
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return activities, nil

}

func (uam *UserActivityModel) Update(userActivity *UserActivity) error {

	var activityType string

	switch userActivity.ActivityType {
	case "chat", "listening", "quiz":
		activityType = userActivity.ActivityType
	default:
		return fmt.Errorf("activity type must be one of 'chat', 'quiz', or 'listening'")
	}

	stmt := `
	UPDATE user_activities
	SET
	score = $1, completed_at = $2, attempted = true
	WHERE
	user_id = $3 AND book_id = $4 AND chapter_id = $5 AND activity_type = $6
	`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := uam.DB.Exec(ctx, stmt, userActivity.Score, time.Now(), userActivity.UserID, userActivity.BookID, userActivity.ChapterID, activityType)
	if err != nil {
		return fmt.Errorf("failed to execute update: %w", err)
	}

	return nil

}
