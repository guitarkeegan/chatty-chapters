package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {

	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.Handler(http.MethodGet, "/v1/user/logged-in", app.authenticate(http.HandlerFunc(app.loggedIn)))
	router.Handler(http.MethodGet, "/v1/user/logout", app.authenticate(http.HandlerFunc(app.logout)))
	router.Handler(http.MethodGet, "/v1/user/progress/all", app.authenticate(http.HandlerFunc(app.userProgressHandler)))

	router.Handler(http.MethodGet, "/v1/teacher/dashboard", app.authenticate(app.authenticateTeacher(http.HandlerFunc(app.teacherDashHandler))))
	router.Handler(http.MethodGet, "/v1/teacher/course/book/:book_id", app.authenticate(app.authenticateTeacher(http.HandlerFunc(app.teacherCourseBookHandler))))
	router.Handler(http.MethodPost, "/v1/teacher/course/chapter/results", app.authenticate(app.authenticateTeacher(http.HandlerFunc(app.teacherChapterResultsHandler))))
	router.Handler(http.MethodPost, "/v1/teacher/course/chapter/results/student", app.authenticate(app.authenticateTeacher(http.HandlerFunc(app.teacherChapterResultsStudentHandler))))

	// TODO: this is unfortunate. either include courseID in progress OR change the route name to book/:book_id...
	router.Handler(http.MethodGet, "/v1/user/progress/book/:book_id", app.authenticate(http.HandlerFunc(app.userCourseProgressHandler)))
	// TODO protect this admin route
	router.Handler(http.MethodPost, "/v1/register/users", http.HandlerFunc(app.registerUsersAdminHandler))

	router.Handler(http.MethodGet, "/v1/books", app.authenticate(http.HandlerFunc(app.listStudentBooksHandler)))

	router.Handler(http.MethodGet, "/v1/chapters/book/:book_id", app.authenticate(http.HandlerFunc(app.listChapters)))

	router.Handler(http.MethodGet, "/v1/chapters/chat/start/book/:book_id/chapter/:chapter_id", app.authenticate(http.HandlerFunc(app.chapterChatStartHandler)))
	router.Handler(http.MethodGet, "/v1/chapters/chat/ws/book/:book_id/chapter/:chapter_id", app.authenticate(http.HandlerFunc(app.chapterChatContinueHandler)))

	router.Handler(http.MethodGet, "/v1/activities/book/:book_id/chapter/:chapter_id", app.authenticate(http.HandlerFunc(app.listActivities)))

	router.Handler(http.MethodGet, "/v1/quiz/take/book/:book_id/chapter/:chapter_id", app.authenticate(http.HandlerFunc(app.quizTakeHandler)))
	router.Handler(http.MethodGet, "/v1/quiz/generate/book/:book_id/chapter/:chapter_id", app.authenticate(http.HandlerFunc(app.quizGenerateHandler)))
	router.Handler(http.MethodPost, "/v1/quiz/submit-answers/book/:book_id/chapter/:chapter_id", app.authenticate(http.HandlerFunc(app.quizSubmitAnswersHandler)))

	router.HandlerFunc(http.MethodPost, "/v1/login", app.loginHandler)
	router.HandlerFunc(http.MethodPost, "/v1/sign-up", app.signupHandler)
	router.HandlerFunc(http.MethodPost, "/v1/new-password", app.newPasswordHandler)
	router.HandlerFunc(http.MethodPost, "/v1/forgot-password", app.forgotPasswordHandler)
	router.HandlerFunc(http.MethodPost, "/v1/reset-password", app.resetPasswordHandler)

	// protect this route
	router.Handler(http.MethodGet, "/v1/metrics", expvar.Handler())

	return app.recoverPanic(app.enableCoors(app.logRequest(router)))

}
