package main

import (
	"context"
	"fmt"
	"net/http"
)

func (app *application) teacherDashHandler(w http.ResponseWriter, r *http.Request) {

	userID, err := app.getUserIDFromSubContext(r)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// get user courses, assuming they can't also be a student
	courseInfo, err := app.models.UserCourse.GetAll(userID)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("instructorDashHandler: %w", err))
		return
	}

	if len(courseInfo) == 0 {
		app.badRequestResponse(w, r, fmt.Errorf("no courses found"))
		return
	}

	if err = app.writeJSON(w, http.StatusOK, envelope{"courses": courseInfo}, nil); err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("instructorDashHandler: %w", err))
	}

}

func (app *application) teacherCourseBookHandler(w http.ResponseWriter, r *http.Request) {

	// context here?
	ctx := context.Background()

	bookID, err := app.readBookIDParam(r)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	chapters, err := app.models.Chapter.GetAllChaptersByBookID(ctx, bookID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	env := envelope{"chapters": chapters}

	app.writeJSON(w, http.StatusOK, env, nil)
}

func (app *application) teacherChapterResultsHandler(w http.ResponseWriter, r *http.Request) {

	var dst struct {
		CourseID  int64 `json:"course_id"`
		BookID    int64 `json:"book_id"`
		ChapterID int64 `json:"chapter_id"`
	}

	if err := app.readJSON(w, r, &dst); err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("must include: course_id, book_id, and chapter_id"))
		return
	}

	results, err := app.models.User.GetChapterResult(dst.CourseID, dst.BookID, dst.ChapterID)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("teacherChapterResults: GetChapterResult: %w", err))
		return
	}

	env := envelope{"results": results}

	app.writeJSON(w, http.StatusOK, env, nil)

}

func (app *application) teacherChapterResultsStudentHandler(w http.ResponseWriter, r *http.Request) {

	var queryData struct {
		UserID    int64 `json:"user_id"`
		BookID    int64 `json:"book_id"`
		ChapterID int64 `json:"chapter_id"`
	}

	if err := app.readJSON(w, r, &queryData); err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("must include: user_id, book_id, and chapter_id"))
		return
	}

	userChapterResults, err := app.models.User.GetChapterResultsStudent(queryData.BookID, queryData.ChapterID, queryData.UserID)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("teacherChapterResultsStudent: GetChapterResultsStudent: %w", err))
		return
	}

	env := envelope{"student_results": userChapterResults}

	if err := app.writeJSON(w, http.StatusOK, env, nil); err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("teacherChapterResultsStudent: writeJSON: %w", err))
		return
	}

}
