package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/guitarkeegan/music-class-tutor/internal/data"
	"github.com/tmc/langchaingo/vectorstores"
)

var ErrRecordNotFound = errors.New("record not found")

const (
	MINIMUM_CHAT_REQUIREMENT         = 3
	MINIMUM_QUIZ_RESPONSE_REQUIRMENT = 3
	NUM_OF_RETURNED_DOCS             = 3
	MINIMUM_EVAL_REQUIREMENT         = 3
	NUM_OF_DOCS_ON_EVAL              = 5
	MINIMUM_QUIZ_QUESTIONS           = 3
)

func (app *application) quizTakeHandler(w http.ResponseWriter, r *http.Request) {

	dbg("quizTakeHandler")
	userID, err := app.getUserIDFromSubContext(r)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("get user id: %w", err))
		return
	}

	dbg("readBookChapterParams")
	bookID, chapterID, err := app.readBookChapterParams(r)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("url is malformed"))
		return
	}

	// check if quiz has been taken
	evals, err := app.models.QuizResponseEval.GetMany(userID, bookID, chapterID)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("quizTakeHandler: QuizRepsonseEval GetMany: %w", err))
		return
	}

	if len(evals) == MINIMUM_EVAL_REQUIREMENT {

		quizResults, err := app.models.QuizResponseEval.GetQuizResults(userID, bookID, chapterID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		err = app.writeJSON(w, http.StatusOK, envelope{
			"quiz":   quizResults,
			"status": "completed",
		}, nil)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		return
	}

	chatEval := data.ChatEval{
		UserID:    userID,
		BookID:    bookID,
		ChapterID: chapterID,
	}

	chatEvals := &[]*data.ChatEval{&chatEval}

	dbg("get chat eval")
	// have they completed the chat req?
	// if they did pass, they'll have a score and feedback
	err = app.models.ChatEval.Get(context.Background(), chatEvals)
	if err != nil {

		app.serverErrorResponse(w, r, fmt.Errorf("ChatEval.Get: %w", err))
		// 404
		return
	}

	for _, eval := range *chatEvals {
		// check one of the evals is passing
		if eval.Score == 1 && eval.Feedback != "" {
			chatEval = *eval
			break
		}
	}
	// user has either not chatted, or not chatted enough just yet
	if chatEval.Score == 0 && chatEval.Feedback == "" {
		err = app.writeJSON(w, http.StatusOK, envelope{
			"status":  "incomplete",
			"message": "User must first finish chat requirement: 3 quality questions or interactions",
		}, nil)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		return
	}
	// must have completed the chat req

	var quiz data.Quiz
	dbg("populate quiz object")
	quiz.UserID = userID
	quiz.BookID = bookID
	quiz.ChapterID = chapterID
	dbg("%+v", quiz)
	// check if quiz is already done.
	dbg("get quiz")
	err = app.models.Quiz.Get(&quiz)
	if err != nil {
		// return an error if anything other than record not found
		if !errors.Is(err, data.ErrRecordNotFound) {
			app.serverErrorResponse(w, r, fmt.Errorf("unexpected err on quiz scan: %w", err))
			return
		}
	}

	// check if was attempted yet, then just send the quiz if it already exists
	dbg("check quizID")
	dbg("quizID: %d", quiz.ID)
	if quiz.ID != 0 {
		dbg("quiz exists")
		quizQuestionsFromDB, err := app.models.QuizQuestion.GetMany(quiz.ID)
		if err != nil {
			app.serverErrorResponse(w, r, fmt.Errorf("getMany in handler: %w", err))
			return
		}
		// what if the quiz was taken, but not graded?
		// at this point we know that there are not enough evals
		// so we can check if they have responses for this quiz
		// get responses
		dbg("check num of questions returned")
		if len(quizQuestionsFromDB) < MINIMUM_QUIZ_QUESTIONS {
			// quiz exists.. but no questions have been created so...
			// delete the quiz, and create a new one..
			app.logger.Info("recreate quiz", "number of questions found", len(quizQuestionsFromDB))
			if err = app.models.Quiz.Delete(quiz.ID); err != nil {
				app.serverErrorResponse(w, r, fmt.Errorf("delete: %w", err))
				return
			}

			quizQuestionsFromDB, err := app.createQuiz(userID, bookID, chapterID, &quiz)
			if err != nil {
				app.serverErrorResponse(w, r, fmt.Errorf("createQuiz: %w", err))
				return
			}

			app.serverErrorResponse(w, r, fmt.Errorf("quizQuestionsFromDB length: %d", len(quizQuestionsFromDB)))
			return
		}

		err = app.writeJSON(w, http.StatusOK, envelope{
			"quiz": quizQuestionsFromDB,
		}, nil)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		return
	}

	quizQuestionsFromDB, err := app.createQuiz(userID, bookID, chapterID, &quiz)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("createQuiz: %w", err))
		return
	}

	dbg("writeJSON")
	err = app.writeJSON(w, http.StatusOK, envelope{
		"quiz": quizQuestionsFromDB,
	}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) quizGenerateHandler(w http.ResponseWriter, r *http.Request) {

	// sub from context
	userID, err := app.getUserIDFromSubContext(r)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	bookID, chapterID, err := app.readBookChapterParams(r)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("url is malformed"))
		return
	}

	// check that chat session exist
	convoHistory, err := app.models.AIUserConversation.GetConversationHistory(r.Context(), userID, bookID, chapterID)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("convoHistory: %q", err))
		return
	}

	// dereference to get length?
	if convoHistory != nil && len(*convoHistory) < MINIMUM_CHAT_REQUIREMENT {
		app.serverErrorResponse(w, r, fmt.Errorf("no chat history"))
		return
	}

}

func (app *application) quizSubmitAnswersHandler(w http.ResponseWriter, r *http.Request) {

	dbg("start quizSubmitAnswersHandler")

	userID, err := app.getUserIDFromSubContext(r)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("get user id: %w", err))
		return
	}

	dbg("readBookChapterParams")

	bookID, chapterID, err := app.readBookChapterParams(r)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("url is malformed"))
		return
	}

	// TODO:
	// check if quiz eval exists
	dbg("quizResEvals")
	quizResEvals, err := app.models.QuizResponseEval.GetMany(userID, bookID, chapterID)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("quizSubmitAnswers: quiz res evals: GetMany: %w", err))
		return
	}
	// if it does, return early
	dbg("quizResEvals.length: %d", len(quizResEvals))
	if len(quizResEvals) == MINIMUM_EVAL_REQUIREMENT {
		app.writeJSON(w, http.StatusOK, envelope{"quiz": map[string]string{
			"status":  "completed",
			"message": "Quiz has already been submitted",
		}}, nil)
	}
	// else, create the quiz evail

	var quizResponses []data.QuizResponse

	dbg("readJSON")
	err = app.readJSON(w, r, &quizResponses)
	if err != nil {
		app.logger.Error("readJSON", "error", err)
		app.badRequestResponse(w, r, fmt.Errorf("malformed JSON"))
		return
	}
	// store responses
	dbg("quizResponse InsertMany")
	err = app.models.QuizResponse.InsertMany(quizResponses)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("insertMany in handler: %w", err))
		return
	}

	// grade the quiz

	dbg("START GRADE QUIZ")

	// get quizID
	var quiz data.Quiz
	quiz.UserID = userID
	quiz.BookID = bookID
	quiz.ChapterID = chapterID
	err = app.models.Quiz.Get(&quiz)
	if err != nil {
		app.logger.Error(err.Error())
		return
	}
	// get the questions
	questions, err := app.models.QuizQuestion.GetMany(quiz.ID)
	if err != nil {
		app.logger.Error(err.Error())
		return
	}
	// get the responses
	answers, err := app.models.QuizResponse.GetMany(quiz.ID)
	if err != nil {
		app.logger.Error(err.Error())
		return
	}

	// format questions and answers
	var questionsFmt []string
	for _, q := range questions {
		questionsFmt = append(questionsFmt, q.QuizQuestion)
	}
	var answersFmt []string
	for _, a := range answers {
		answersFmt = append(answersFmt, a.StudentResponse)
	}
	// get the docs
	// get book info

	dbg("quistionsFmt: %+v", questionsFmt)
	dbg("answersFmt: %+v", answersFmt)

	book, chapter, err := app.models.Transaction.GetBookAndChapter(context.Background(), bookID, chapterID)
	if err != nil {
		app.logger.Error(err.Error())
	}

	filter := map[string]any{"title": titleCase(book.Title), "chapter": strconv.Itoa(chapter.ChapterNumber)}

	// get textbook context
	// testing 5 docs returned
	dbg("similaritySearch")
	docs, err := app.ai.DB.SimilaritySearch(context.Background(), strings.Join(questionsFmt, " "), NUM_OF_DOCS_ON_EVAL, vectorstores.WithFilters(filter))
	if err != nil {
		app.logger.Error(err.Error())
		return
	}
	dbg("docs returned: ", len(docs))
	dbg("responseFeedback")
	// assess the quiz responses
	// what to do if this fails???
	responseFeedback, err := app.ai.AssessAnswers(questionsFmt, answersFmt, docs)
	if err != nil {
		app.logger.Error(err.Error())
		return
	}
	// check if 3 were returned
	if len(responseFeedback) != 3 {
		dbg("responseFeedback: length: %d", responseFeedback)
		app.logger.Error("less than 3 assesments were returned")
		return
	}

	var totalScore int
	for _, res := range responseFeedback {
		totalScore += res.Score
	}
	dbg("quizResEvals")
	// create the quiz evals
	var qre []data.QuizResponseEval
	for i, e := range responseFeedback {
		newQRE := data.QuizResponseEval{
			UserID:           userID,
			BookID:           bookID,
			ChapterID:        chapterID,
			QuizID:           quiz.ID,
			QuizQuestionID:   questions[i].ID,
			QuizResponseID:   answers[i].ID,
			Score:            e.Score,
			ResponseFeedback: e.Feedback,
		}

		qre = append(qre, newQRE)
	}

	dbg("QuizResponseEval.InsertMany")
	err = app.models.QuizResponseEval.InsertMany(&qre)
	if err != nil {
		app.logger.Error(err.Error())
		return
	}

	// update the chapter activities
	ua := data.UserActivity{
		UserID:       userID,
		BookID:       bookID,
		ChapterID:    chapterID,
		ActivityType: "quiz",
		Score:        totalScore,
	}

	dbg("UserActivity.Update")
	err = app.models.UserActivity.Update(&ua)
	if err != nil {
		app.logger.Error(err.Error())
		return
	}
	// update the student progress
	dbg("UserProgress.QupdateQuiz")
	err = app.models.UserProgress.UpdateQuiz(userID, bookID, totalScore, chapter.ChapterNumber)
	if err != nil {
		app.logger.Error(err.Error())
		return
	}

	dbg("END GRADE QUIZ")

	err = app.writeJSON(w, http.StatusCreated, envelope{"quiz": map[string]string{
		"status": "success",
	}}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}
