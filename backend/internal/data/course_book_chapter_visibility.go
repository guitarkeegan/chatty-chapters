package data

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CourseBookChapter struct {
	BookID    int64 `json:"book_id"`
	ChapterID int64 `json:"chapter_id"`
	CourseID  int64 `json:"course_id"`
	Visible   bool  `json:"visible"`
}

type CourseBookChapterModel struct {
	DB *pgxpool.Pool
}

func NewCourseBookChapter(bookID, chapterID, courseID int64) *CourseBookChapter {
	return &CourseBookChapter{
		BookID:    bookID,
		ChapterID: chapterID,
		CourseID:  courseID,
		Visible:   true,
	}
}

func (m *CourseBookChapterModel) Insert(ctx context.Context, ucc *CourseBookChapter) error {
	query := `
        INSERT INTO course_book_chapter_visibility (book_id, chapter_id, course_id, visible)
        VALUES ($1, $2, $3, $4)
    `

	_, err := m.DB.Exec(ctx, query, ucc.BookID, ucc.ChapterID, ucc.CourseID, ucc.Visible)
	if err != nil {
		return fmt.Errorf("failed to insert CourseBookChapter: %w", err)
	}
	return nil
}

func (m *CourseBookChapterModel) InsertMany(ctx context.Context, uccs []CourseBookChapter) error {
	batch := &pgx.Batch{}
	for _, ucc := range uccs {
		batch.Queue(`
            INSERT INTO course_book_chapter_visibility (book_id, chapter_id, course_id, visible)
            VALUES ($1, $2, $3, $4)
        `, ucc.BookID, ucc.ChapterID, ucc.CourseID, ucc.Visible)
	}

	br := m.DB.SendBatch(ctx, batch)
	err := br.Close()
	if err != nil {
		return fmt.Errorf("failed to insert multiple CourseBookChapters: %w", err)
	}

	return nil
}

func (m *CourseBookChapterModel) Get(ctx context.Context, bookID, chapterID, courseID int64) (CourseBookChapter, error) {
	query := `
        SELECT book_id, chapter_id, course_id, visible
        FROM course_book_chapter_visibility
        WHERE book_id = $1 AND chapter_id = $2 AND course_id = $3
    `

	var ucc CourseBookChapter
	err := m.DB.QueryRow(ctx, query, bookID, chapterID, courseID).Scan(
		&ucc.BookID,
		&ucc.ChapterID,
		&ucc.CourseID,
		&ucc.Visible,
	)

	if err != nil {
		return CourseBookChapter{}, fmt.Errorf("failed to get CourseBookChapter: %w", err)
	}

	return ucc, nil
}

func (m *CourseBookChapterModel) GetMany(ctx context.Context, courseID int64, bookID int64) ([]CourseBookChapter, error) {
	query := `
        SELECT book_id, chapter_id, course_id, visible
        FROM course_book_chapter_visibility
        WHERE course_id = $1 AND book_id = $2
    `

	rows, err := m.DB.Query(ctx, query, courseID, bookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get CourseBookChapters: %w", err)
	}
	defer rows.Close()

	var uccs []CourseBookChapter
	for rows.Next() {
		var ucc CourseBookChapter
		if err := rows.Scan(&ucc.BookID, &ucc.ChapterID, &ucc.CourseID, &ucc.Visible); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		uccs = append(uccs, ucc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during row iteration: %w", err)
	}

	return uccs, nil
}
