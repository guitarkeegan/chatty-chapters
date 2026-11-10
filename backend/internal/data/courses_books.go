package data

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CoursesBooks struct {
	CourseID int64 `json:"course_id"`
	BookID   int64 `json:"book_id"`
}

type CoursesBooksModel struct {
	DB *pgxpool.Pool
}

func NewCoursesBooks(courseID, bookID int64) *CoursesBooks {
	return &CoursesBooks{
		CourseID: courseID,
		BookID:   bookID,
	}
}

func (cb *CoursesBooksModel) GetBooks(courseID int64) ([]Book, error) {

	stmt := `
	SELECT b.id, b.title, b.created_at, b.edition, b.year
	FROM books b
	JOIN courses_books cb ON cb.book_id = b.id
	WHERE cb.course_id = $1;
	`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	rows, err := cb.DB.Query(ctx, stmt, courseID)
	if err != nil {
		return []Book{}, fmt.Errorf("CoursesBooksModel GetBooks: %w", err)
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {

		var book Book
		err := rows.Scan(&book.ID, &book.Title, &book.CreatedAt, &book.Edition, &book.Year)
		if err != nil {
			return []Book{}, fmt.Errorf("CoursesBooksModel GetBooks: scan row: %w", err)
		}

		books = append(books, book)
	}

	if rows.Err() != nil {
		return []Book{}, fmt.Errorf("CoursesBooksModel GetBooks: %w", err)
	}

	return books, nil

}

func (cb *CoursesBooksModel) Insert(coursesBooks *CoursesBooks) error {
	stmt := `
        INSERT INTO courses_books (course_id, book_id)
        VALUES ($1, $2)
	`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := cb.DB.Exec(ctx, stmt, coursesBooks.CourseID, coursesBooks.BookID)
	if err != nil {
		return fmt.Errorf("CoursesBooksModel Insert: %w", err)
	}

	return nil

}
