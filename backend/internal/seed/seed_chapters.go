package seed

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/guitarkeegan/music-class-tutor/internal/data"
)

func SeedChapters(db *pgxpool.Pool) error {
	bookTitle := "Alice in Wonderland"

	var aliceChapters = []data.Chapter{
		{ChapterNumber: 1, Title: "Down the Rabbit-Hole"},
		{ChapterNumber: 2, Title: "The Pool of Tears"},
		{ChapterNumber: 3, Title: "A Caucus-Race and a Long Tale"},
		{ChapterNumber: 4, Title: "The Rabbit Sends in a Little Bill"},
		{ChapterNumber: 5, Title: "Advice from a Caterpillar"},
		{ChapterNumber: 6, Title: "Pig and Pepper"},
		{ChapterNumber: 7, Title: "A Mad Tea-Party"},
		{ChapterNumber: 8, Title: "The Queen’s Croquet-Ground"},
		{ChapterNumber: 9, Title: "The Mock Turtle’s Story"},
		{ChapterNumber: 10, Title: "The Lobster Quadrille"},
		{ChapterNumber: 11, Title: "Who Stole the Tarts?"},
		{ChapterNumber: 12, Title: "Alice’s Evidence"},
	}

	bm := &data.BookModel{
		DB: db,
	}

	ctx := context.Background()
	tx, err := bm.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx for chapters failed: %w", err)
	}
	defer tx.Rollback(ctx)

	bookID, err := bm.GetBookByTitle(ctx, bookTitle)
	if err != nil {
		return fmt.Errorf("call to get book by title failed: %w", err)
	}

	cm := &data.ChapterModel{
		DB: db,
	}

	if err := cm.InsertChapters(ctx, bookID, aliceChapters); err != nil {
		return fmt.Errorf("insert chapters for alice book failed: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction failed: %w", err)
	}

	return nil
}
