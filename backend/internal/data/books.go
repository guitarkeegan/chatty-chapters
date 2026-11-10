package data

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Book struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	Edition   int       `json:"edition"`
	Year      int       `json:"year"`
	Version   int       `json:"version"`
}

type BookModel struct {
	DB *pgxpool.Pool
}

func (bm *BookModel) Insert(book *Book) error {
	query := `
	INSERT INTO books (title, edition, year)
	VALUES ( $1, $2, $3 )
	RETURNING id, created_at, version`

	args := []any{book.Title, book.Edition, book.Year}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return bm.DB.QueryRow(ctx, query, args...).Scan(&book.ID, &book.CreatedAt, &book.Version)
}

func (bm *BookModel) Get(ctx context.Context, bookID int64) (Book, error) {

	stmt := `
	SELECT id, title
	FROM books
	WHERE id = $1
	`

	var book Book

	err := bm.DB.QueryRow(ctx, stmt, bookID).Scan(&book.ID, &book.Title)
	if err != nil {
		// should i make a new Book instead?
		return book, err
	}

	return book, nil

}

// returns the book id
func (bm *BookModel) GetBookByTitle(ctx context.Context, title string) (int64, error) {
	query := `
	SELECT id 
	FROM books
	WHERE title = $1
	`

	var id int64
	err := bm.DB.QueryRow(ctx, query, title).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, ErrRecordNotFound
		}
		return 0, err
	}

	return id, nil
}
