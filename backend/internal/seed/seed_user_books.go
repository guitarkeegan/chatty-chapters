package seed

import (
	"context"
	"fmt"

	"github.com/guitarkeegan/music-class-tutor/internal/data"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SeedUserBooks(db *pgxpool.Pool, userID, bookID int64) error {

	ubm := &data.UserBookModel{DB: db}

	err := ubm.Insert(context.Background(), userID, bookID)
	if err != nil {
		return fmt.Errorf("error inserting to user_books with userID: %d, and bookID: %d: %w", userID, bookID, err)
	}

	return nil
}
