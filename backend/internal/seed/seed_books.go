package seed

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/guitarkeegan/music-class-tutor/internal/data"
)

func SeedBooks(db *pgxpool.Pool) error {

	alice := &data.Book{
		Title:   "Alice in Wonderland",
		Edition: 3,
		Year:    1865,
	}

	bm := &data.BookModel{
		DB: db,
	}

	if err := bm.Insert(alice); err != nil {
		return fmt.Errorf("couldn't insert book %+v: %w", alice, err)
	}

	return nil
}
