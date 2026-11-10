package data

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var dbg = func() func(format string, as ...any) {
	if os.Getenv("DEBUG") == "" {
		return func(string, ...any) {}
	}
	return func(format string, as ...any) {
		log.Printf(format, as...)
	}
}()

type TrasactionModel struct {
	DB *pgxpool.Pool
}

func (tm *TrasactionModel) GetBookAndChapter(ctx context.Context, bookID, chapterID int64) (Book, Chapter, error) {

	dbg(" in book and chapter")

	tx, err := tm.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Book{}, Chapter{}, err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	bookStmt := `
	SELECT id, title
	FROM books
	WHERE id = $1
	`
	dbg(" get book")
	var book Book
	err = tx.QueryRow(ctx, bookStmt, bookID).Scan(&book.ID, &book.Title)
	if err != nil {
		return Book{}, Chapter{}, err
	}

	chapterStmt := `
	SELECT id, chapter_number, title
	FROM chapters
	WHERE id = $1
	`
	dbg(" get chapter")
	var chapter Chapter
	err = tx.QueryRow(ctx, chapterStmt, chapterID).Scan(&chapter.ID, &chapter.ChapterNumber, &chapter.Title)
	if err != nil {
		return Book{}, Chapter{}, err
	}

	dbg(" commit tx")
	err = tx.Commit(ctx)
	if err != nil {
		return Book{}, Chapter{}, err
	}

	dbg(" returning")
	return book, chapter, nil

}

// call this after chat attempt
func (tm *TrasactionModel) UpdateUserChat(chatEval *ChatEval, chapterNumber int) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
	defer cancel()

	tx, err := tm.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("update user chat: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	chatEvalStmt := `
	INSERT INTO chat_evals
	( user_id, book_id, chapter_id, score, feedback )
	VALUES
	( $1, $2, $3, $4, $5 )
	RETURNING id, created_at
	`
	// insert to chatEvals
	err = tx.QueryRow(ctx, chatEvalStmt, chatEval.UserID, chatEval.BookID, chatEval.ChapterID, chatEval.Score, chatEval.Feedback).Scan(&chatEval.ID, &chatEval.CreatedAt)
	if err != nil {
		return fmt.Errorf("chat eval: %w", err)
	}

	// assume requirement has been fullfilled
	userProgressStmt := `
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

	_, err = tx.Exec(ctx, userProgressStmt, chapterNumber, chatEval.UserID, chatEval.BookID)
	if err != nil {
		return fmt.Errorf("update progress: %w", err)
	}

	// insert to activies progress
	activitiesStmt := `
	UPDATE user_activities
	SET
	score = $1,
	completed_at = $2,
	attempted = true
	WHERE user_id = $3 AND book_id = $4 AND chapter_id = $5 AND activity_type = 'chat'
	`
	_, err = tx.Exec(ctx, activitiesStmt, chatEval.Score, time.Now(), chatEval.UserID, chatEval.BookID, chatEval.ChapterID)
	if err != nil {
		return fmt.Errorf("updateActivity: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("chatEval commit: %w", err)
	}

	return nil

}
