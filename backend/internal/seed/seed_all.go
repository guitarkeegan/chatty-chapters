package seed

import (
	"context"
	"fmt"
	"log"

	"github.com/guitarkeegan/music-class-tutor/internal/data"
	"github.com/jackc/pgx/v5/pgxpool"
)

// make 2 demo users, the alice book and its chapters, an alice course,
// and enroll both users in it
func SeedUsersBooksChaptersRelations(db *pgxpool.Pool) error {
	fmt.Println("SeedUsers")
	err := SeedUsers(db)
	if err != nil {
		return fmt.Errorf("seedUsers failed: %w", err)
	}
	fmt.Println("SeedBooks")
	err = SeedBooks(db)
	if err != nil {
		return fmt.Errorf("seedBooks failed: %w", err)
	}
	err = SeedChapters(db)
	if err != nil {
		return fmt.Errorf("seedChapters failed: %w", err)
	}

	ctx := context.Background()
	cm := &data.ChapterModel{DB: db}
	bm := &data.BookModel{DB: db}
	uam := &data.UserActivityModel{DB: db}
	um := &data.UserModel{DB: db}
	upm := &data.UserProgressModel{DB: db}
	crm := &data.CourseModel{DB: db}
	ucm := &data.UserCourseModel{DB: db}
	cbc := &data.CourseBookChapterModel{DB: db}
	cbm := &data.CoursesBooksModel{DB: db}

	aliceID, err := bm.GetBookByTitle(ctx, "Alice in Wonderland")
	if err != nil {
		return fmt.Errorf("could not get alice id: %w", err)
	}

	aliceChapters, err := cm.GetAllChaptersByBookID(ctx, aliceID)
	if err != nil {
		return fmt.Errorf("could not get alice chapters: %w", err)
	}

	var aliceIDs []int64
	for _, chapter := range aliceChapters {
		aliceIDs = append(aliceIDs, chapter.ID)
	}

	// create the Alice course
	aliceCourse := data.NewCourse("Alice in Wonderland", "fall", 2024)
	if err = crm.Insert(aliceCourse); err != nil {
		return seedError("could not seed course", err.Error())
	}

	if err = cbm.Insert(data.NewCoursesBooks(aliceCourse.ID, aliceID)); err != nil {
		return seedError("could not seed courses_books", err.Error())
	}

	// make every alice chapter visible in the course
	var aliceCourseBookChapters []data.CourseBookChapter
	for _, ch := range aliceChapters {
		aliceCourseBookChapters = append(aliceCourseBookChapters, *data.NewCourseBookChapter(aliceID, ch.ID, aliceCourse.ID))
	}
	if err = cbc.InsertMany(ctx, aliceCourseBookChapters); err != nil {
		return seedError("course book chapter insert many failed", err.Error())
	}

	// enroll both demo users in the course
	for _, email := range []string{demoStudentEmail, demoTeacherEmail} {
		userID, err := um.GetUserByEmail(ctx, email)
		if err != nil {
			return fmt.Errorf("error getting user by email: %w", err)
		}

		if err = uam.CreateActivitiesForUser(ctx, userID, aliceID, aliceIDs); err != nil {
			return seedError(fmt.Sprintf("could not create activities for user: %d", userID), err.Error())
		}

		if err = SeedUserBooks(db, userID, aliceID); err != nil {
			return seedError(fmt.Sprintf("could not seed user_books for user: %d", userID), err.Error())
		}

		if err = upm.Insert(&data.UserProgress{UserID: userID, BookID: aliceID}); err != nil {
			return seedError(fmt.Sprintf("could not seed user_progress for user: %d", userID), err.Error())
		}

		userCourse, err := data.NewUserCourse(userID, aliceCourse.ID)
		if err != nil {
			return seedError("could not seed users_courses", err.Error())
		}
		if err = ucm.Insert(userCourse); err != nil {
			return seedError("could not seed users_courses", err.Error())
		}
	}

	log.Printf("Seed of all users, books, chapters, and user_activities was successful! ✨\n")

	return nil
}

// update this function
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
