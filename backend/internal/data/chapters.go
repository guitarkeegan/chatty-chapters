package data

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Chapter struct {
	ID            int64  `json:"id"`
	ChapterNumber int    `json:"chapter_number"`
	Title         string `json:"title"`
}

type ChapterModel struct {
	DB *pgxpool.Pool
}

func (cm *ChapterModel) InsertChapters(ctx context.Context, bookID int64, chapters []Chapter) error {
	// Start a transaction
	tx, err := cm.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Prepare the insert statement
	stmt := `
        INSERT INTO chapters (book_id, chapter_number, title)
        VALUES ($1, $2, $3)
    `

	// Insert each chapter
	for _, chapter := range chapters {
		_, err := tx.Exec(ctx, stmt, bookID, chapter.ChapterNumber, chapter.Title)
		if err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (cm *ChapterModel) Get(ctx context.Context, chapterID int64) (Chapter, error) {

	stmt := `
	SELECT id, title, chapter_number
	FROM chapters
	WHERE id = $1
	`

	var chapter Chapter

	err := cm.DB.QueryRow(ctx, stmt, chapterID).Scan(&chapter.ID, &chapter.Title, &chapter.ChapterNumber)
	if err != nil {
		return chapter, err
	}

	return chapter, nil

}

// Use this when create the user_activities
func (cm *ChapterModel) GetAllChaptersByBookID(ctx context.Context, bookID int64) ([]Chapter, error) {
	// write statement
	query := `
	SELECT id, title, chapter_number
	FROM chapters
	WHERE book_id = $1
	`
	// args in params
	// execute query with context
	rows, err := cm.DB.Query(ctx, query, bookID)
	if err != nil {
		return nil, fmt.Errorf("query for all chapters failed: %w", err)
	}
	defer rows.Close()

	var chapters []Chapter
	for rows.Next() {
		var chapter Chapter
		err := rows.Scan(&chapter.ID, &chapter.Title, &chapter.ChapterNumber)
		if err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}
		chapters = append(chapters, chapter)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	return chapters, nil
}

func (cm *ChapterModel) GetAllVisibleChapters(ctx context.Context, userID, bookID int64) ([]Chapter, error) {

	query := `
	SELECT ch.id, ch.title, ch.chapter_number
	FROM course_book_chapter_visibility cbcv
	JOIN chapters ch ON ch.id = cbcv.chapter_id
	JOIN users_courses uc ON uc.course_id = cbcv.course_id
	JOIN user_books ub ON ub.book_id = cbcv.book_id AND ub.user_id = uc.user_id
	WHERE uc.user_id = $1 AND cbcv.book_id = $2
	AND cbcv.visible = TRUE
	ORDER BY ch.chapter_number
	`
	// args in params
	// execute query with context
	rows, err := cm.DB.Query(ctx, query, userID, bookID)
	if err != nil {
		return nil, fmt.Errorf("query for all chapters failed: %w", err)
	}
	defer rows.Close()

	var chapters []Chapter
	for rows.Next() {
		var chapter Chapter
		err := rows.Scan(&chapter.ID, &chapter.Title, &chapter.ChapterNumber)
		if err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}
		chapters = append(chapters, chapter)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	return chapters, nil
}
