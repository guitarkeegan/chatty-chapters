package seed

import (
	"context"
	"fmt"
	"log"

	"github.com/guitarkeegan/music-class-tutor/internal/data"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedUsersBooksChaptersRelations makes 2 demo users, and for each seed book
// inserts the book, its chapters, and a course with every chapter visible,
// then enrolls both users in every course.
func SeedUsersBooksChaptersRelations(db *pgxpool.Pool) error {
	ctx := context.Background()

	if err := SeedUsers(db); err != nil {
		return fmt.Errorf("seedUsers failed: %w", err)
	}

	bm := &data.BookModel{DB: db}
	cm := &data.ChapterModel{DB: db}
	uam := &data.UserActivityModel{DB: db}
	um := &data.UserModel{DB: db}
	upm := &data.UserProgressModel{DB: db}
	crm := &data.CourseModel{DB: db}
	ucm := &data.UserCourseModel{DB: db}
	cbc := &data.CourseBookChapterModel{DB: db}
	cbm := &data.CoursesBooksModel{DB: db}

	var userIDs []int64
	for _, email := range []string{demoStudentEmail, demoTeacherEmail} {
		userID, err := um.GetUserByEmail(ctx, email)
		if err != nil {
			return fmt.Errorf("error getting user by email: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	for _, sb := range seedBooks() {
		book := sb.book
		if err := bm.Insert(&book); err != nil {
			return fmt.Errorf("couldn't insert book %q: %w", book.Title, err)
		}

		if err := cm.InsertChapters(ctx, book.ID, sb.chapters); err != nil {
			return fmt.Errorf("insert chapters for %q failed: %w", book.Title, err)
		}

		chapters, err := cm.GetAllChaptersByBookID(ctx, book.ID)
		if err != nil {
			return fmt.Errorf("could not get chapters for %q: %w", book.Title, err)
		}

		var chapterIDs []int64
		for _, ch := range chapters {
			chapterIDs = append(chapterIDs, ch.ID)
		}

		course := data.NewCourse(book.Title, "fall", 2024)
		if err = crm.Insert(course); err != nil {
			return seedError("could not seed course", err.Error())
		}

		if err = cbm.Insert(data.NewCoursesBooks(course.ID, book.ID)); err != nil {
			return seedError("could not seed courses_books", err.Error())
		}

		// make every chapter visible in the course
		var visible []data.CourseBookChapter
		for _, ch := range chapters {
			visible = append(visible, *data.NewCourseBookChapter(book.ID, ch.ID, course.ID))
		}
		if err = cbc.InsertMany(ctx, visible); err != nil {
			return seedError("course book chapter insert many failed", err.Error())
		}

		for _, userID := range userIDs {
			if err = uam.CreateActivitiesForUser(ctx, userID, book.ID, chapterIDs); err != nil {
				return seedError(fmt.Sprintf("could not create activities for user: %d", userID), err.Error())
			}

			if err = SeedUserBooks(db, userID, book.ID); err != nil {
				return seedError(fmt.Sprintf("could not seed user_books for user: %d", userID), err.Error())
			}

			if err = upm.Insert(&data.UserProgress{UserID: userID, BookID: book.ID}); err != nil {
				return seedError(fmt.Sprintf("could not seed user_progress for user: %d", userID), err.Error())
			}

			userCourse, err := data.NewUserCourse(userID, course.ID)
			if err != nil {
				return seedError("could not seed users_courses", err.Error())
			}
			if err = ucm.Insert(userCourse); err != nil {
				return seedError("could not seed users_courses", err.Error())
			}
		}
	}

	log.Printf("Seed of all users, books, chapters, and user_activities was successful! ✨\n")

	return nil
}

// ClearAllUserProgress deletes all chats, quizzes, and evals, and resets progress
func ClearAllUserProgress(ctx context.Context, db *pgxpool.Pool) error {
	stmt := `
	DELETE FROM user_activities;
	DELETE FROM user_messages;
	UPDATE user_progress
	SET
		chapter_chats_completed = '{}',
		chat_score_attempted = 0,
		chats_achieved = 0,
		quizzes_score_attempted = 0,
		quizzes_score_achieved = 0,
		quiz_chapters_completed = '{}',
		total_score_attempted = 0,
		total_score_achieved = 0;
	DELETE FROM ai_messages;
	DELETE FROM ai_user_conversations;
	DELETE FROM chat_evals;
	DELETE FROM quizzes;
	DELETE FROM quiz_questions;
	DELETE FROM quiz_responses;
	DELETE FROM quiz_response_evals;
	`

	_, err := db.Exec(ctx, stmt)
	if err != nil {
		return fmt.Errorf("unable to clear all user data: %w", err)
	}

	log.Printf("all user data had been cleared! 😬")
	return nil
}
