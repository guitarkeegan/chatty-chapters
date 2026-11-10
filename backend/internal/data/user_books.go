package data

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserBook struct {
	UserID int64
	BookID int64
}

type UserBookModel struct {
	DB *pgxpool.Pool
}

func (ubm *UserBookModel) Insert(ctx context.Context, userID int64, bookID int64) error {

	stmt := `
	INSERT INTO user_books (user_id, book_id)
	VALUES ( $1, $2 )
	`

	_, err := ubm.DB.Exec(ctx, stmt, userID, bookID)
	if err != nil {
		return SQLError(fmt.Sprintf("error adding userID: %d, bookID: %d, to user_book", userID, bookID), err)
	}
	return nil

}

// return a slice of Book based on the userID passed in
func (ubm *UserBookModel) GetBooksByUserID(ctx context.Context, userID int64) ([]Book, error) {

	stmt := `
	SELECT b.id, b.title, b.created_at, b.edition, b.year, b.version
	FROM user_books ub
	JOIN books b ON ub.book_id = b.id
	WHERE ub.user_id = $1
	`

	rows, err := ubm.DB.Query(ctx, stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []Book

	for rows.Next() {
		var book Book

		err := rows.Scan(&book.ID, &book.Title, &book.CreatedAt, &book.Edition, &book.Year, &book.Version)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	// Check for errors from iterating over rows. Gotcha.
	if rows.Err() != nil {
		return nil, err
	}

	// send back empty slice if none found

	return books, nil

}
