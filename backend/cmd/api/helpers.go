package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/google/uuid"
	"github.com/guitarkeegan/music-class-tutor/internal/ai"
	"github.com/guitarkeegan/music-class-tutor/internal/data"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type envelope map[string]any

const NEW_PASSWORD_REQUIRED = "NEW_PASSWORD_REQUIRED"

// change marshal indent to marshal for a slight performance boost
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// noline prefix "" and tab \t
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readBookIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("book_id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

// func (ap *application) readStudentIDParam(r *http.Request) (int64, error) {
// 	params := httprouter.ParamsFromContext(r.Context())

// 	id, err := strconv.ParseInt(params.ByName("student_id"), 10, 64)
// 	if err != nil || id < 1 {
// 		return 0, errors.New("invalid id parameter")
// 	}

// 	return id, nil
// }

func (app *application) readBookChapterParams(r *http.Request) (int64, int64, error) {

	dbg("inside readBookChapterParams")
	params := httprouter.ParamsFromContext(r.Context())

	bookID, err := strconv.ParseInt(params.ByName("book_id"), 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid book id")
	}
	if bookID <= 0 {
		return 0, 0, fmt.Errorf("invalid book id")
	}
	chapterID, err := strconv.ParseInt(params.ByName("chapter_id"), 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid chapter id")
	}
	if chapterID <= 0 {
		return 0, 0, fmt.Errorf("invalid chapter id")
	}
	dbg("leaving readBookChapterParams")
	return bookID, chapterID, nil

}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	dbg("reading json")
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dbg("after max bytes")
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	dbg("decoder created")
	err := dec.Decode(dst)
	if err != nil {
		dbg("in error after decode")
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// send the Identity Token and refresh token back to the client
func (app *application) sendIDAndRefreshToClient(w http.ResponseWriter, status int, data envelope, authResult *cognitoidentityprovider.AuthenticationResultType) {

	dbg("inside send id and refresh to client..")
	if authResult == nil || authResult.IdToken == nil || authResult.RefreshToken == nil {
		http.Error(w, "Authentication result is missing required tokens", http.StatusInternalServerError)
		return
	}
	// ID tokens typically expire after 1 hour, refresh tokens after 30 days
	app.setAuthCookie(w, idCookieName, *authResult.IdToken, time.Now().Add(1*time.Hour))
	app.setAuthCookie(w, refreshCookieName, *authResult.RefreshToken, time.Now().Add(30*24*time.Hour))
	app.writeJSON(w, status, data, nil)
}

func (app *application) validateContextSub(r *http.Request) (string, error) {

	userSub, ok := r.Context().Value(userCognitoSubKey).(string)
	if !ok {
		return "", fmt.Errorf("validateContextSub: ")
	}
	return userSub, nil

}

func (app *application) getRole(r *http.Request) (string, error) {

	userSub, err := app.validateContextSub(r)
	if err != nil {
		return "", fmt.Errorf("getRole: invalid: %w", err)
	}
	userSubUUID, err := uuid.Parse(userSub)
	if err != nil {
		return "", fmt.Errorf("getRole: parsing uuid error: %w", err)

	}

	// check role
	role, err := app.models.User.CheckRole(r.Context(), userSubUUID)
	if err != nil {
		return "", fmt.Errorf("checkRole: %w", err)
	}

	return role, nil
}

func (app *application) getUserIDFromSubContext(r *http.Request) (int64, error) {

	userSub, err := app.validateContextSub(r)
	if err != nil {
		return 0, fmt.Errorf("getUserSubID: invalid: %q", err)

	}
	userSubUUID, err := uuid.Parse(userSub)
	if err != nil {
		return 0, fmt.Errorf("getUserSubID: parsing uuid error: %q", err)

	}
	userID, err := app.models.User.GetUserByCognitoSub(r.Context(), userSubUUID)
	if err != nil {
		return 0, fmt.Errorf("getUserSubID: user not found")

	}

	return userID, nil
}

func titleCase(s string) string {

	caser := cases.Title(language.English)

	return caser.String(s)

}

// ai generated validator...
func validateRegistrationInput(firstName, lastName, email string) error {
	// Check if FirstName is empty
	if firstName == "" {
		return fmt.Errorf("first name is required")
	}

	// Check if LastName is empty
	if lastName == "" {
		return fmt.Errorf("last name is required")
	}

	// Check if Email is empty
	if email == "" {
		return fmt.Errorf("email is required")
	}

	// Validate Email format using a simple regex pattern
	emailRegex := `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`
	matched, err := regexp.MatchString(emailRegex, email)
	if err != nil || !matched {
		return fmt.Errorf("invalid email format")
	}

	return nil

}

func (app *application) createQuiz(userID, bookID, chapterID int64, quiz *data.Quiz) ([]data.QuizQuestion, error) {

	dbg("get chapter conversation")
	// quiz has not been taken, create the quiz
	conversations, err := app.models.AIUserConversation.GetChapterConversation(context.Background(), userID, bookID, chapterID)
	if err != nil {
		return nil, fmt.Errorf("createQuiz: %w", err)
	}

	var conversationsFormatted []ai.ChatConversation
	// convert the conversations to the correct format
	dbg("format conversations")
	for _, c := range conversations {
		var formatted ai.ChatConversation
		formatted.UserMessage = c.UserMessage
		formatted.AIMessage = c.AIMessage
		conversationsFormatted = append(conversationsFormatted, formatted)
	}
	dbg("formatted conversations: %+v", conversationsFormatted)

	// need book title...
	dbg("get book")
	bookData, err := app.models.Book.Get(context.Background(), bookID)
	if err != nil {
		return nil, fmt.Errorf("createQuiz: %w", err)
	}

	// need chapter number...
	dbg("get chapter")
	chapterData, err := app.models.Chapter.Get(context.Background(), chapterID)
	if err != nil {
		return nil, fmt.Errorf("createQuiz: %w", err)

	}

	// create a quiz
	dbg("insert quiz")
	err = app.models.Quiz.Insert(quiz)
	if err != nil {
		return nil, fmt.Errorf("createQuiz: %w", err)
	}

	dbg("generate quiz questions")
	quizQuestions, err := app.ai.GenerateQuizQuestions(conversationsFormatted, bookData.Title, chapterData.ChapterNumber, NUM_OF_RETURNED_DOCS)
	if err != nil {
		return nil, fmt.Errorf("createQuiz: %w", err)
	}
	dbg("quizQuestions: %+v", quizQuestions)
	// if you've made it this far, populate the quiz object

	// store quiz questions
	var storedQuestions []data.QuizQuestion
	// format quiz questions to store
	dbg("format quiz questions")
	for _, qq := range quizQuestions {
		var q data.QuizQuestion
		q.QuizID = quiz.ID
		q.QuizQuestion = qq.Question
		storedQuestions = append(storedQuestions, q)
	}

	// store the new questions
	dbg("storedQuestions length: %d", len(storedQuestions))
	dbg("insert many quiz questions")
	err = app.models.QuizQuestion.InsertMany(storedQuestions)
	if err != nil {
		return nil, fmt.Errorf("createQuiz: %w", err)
	}

	// retrieve the questions that were just stored
	dbg("get many quiz questions")
	quizQuestionsFromDB, err := app.models.QuizQuestion.GetMany(quiz.ID)
	if err != nil {
		return nil, fmt.Errorf("createQuiz: %w", err)
	}

	return quizQuestionsFromDB, nil

}
