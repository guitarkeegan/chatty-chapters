package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/guitarkeegan/music-class-tutor/internal/data"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/vectorstores"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	RETURN_THREE_DOCS  = 3
	RETURN_TEN_DOCS    = 10
	RETURN_TWENTY_DOCS = 20
)

// send the chat history
func (app *application) chapterChatStartHandler(w http.ResponseWriter, r *http.Request) {

	dbg("CHAPTER_CHAT_START_HANDLER")

	userID, err := app.getUserIDFromSubContext(r)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("userID not in context"))
		return
	}

	dbg("userID: %d", userID)

	bookID, chapterID, err := app.readBookChapterParams(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	dbg("bookID: %d, chapterID: %d", bookID, chapterID)

	convoHistory, err := app.models.AIUserConversation.GetChapterConversation(context.Background(), userID, bookID, chapterID)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	dbg("past convoHistory")

	// get ai evaluation
	eval := data.NewChatEval(userID, bookID, chapterID)
	evals := &[]*data.ChatEval{eval}
	if err = app.models.ChatEval.Get(r.Context(), evals); err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("chatStart: ChatEval: %w", err))
		return
	}

	dbg("evals: %d", len(*evals))
	// loop through and find the one that is completed
	for _, e := range *evals {
		if e.Score != 0 && e.Feedback != "" {
			eval = e
			break
		}
	}

	// find the one with feedback if the score is 0
	if eval.Score == 0 {
		dbg("score is 0")
		for _, e := range *evals {
			if e.Feedback != "" {
				eval = e
				break
			}
		}
	}
	dbg("eval to return: %+v", eval)
	dbg("CHAPTER_CHAT_START_HANDLER_END")
	app.writeJSON(w, http.StatusOK, envelope{"chat": envelope{"history": convoHistory, "eval": eval}}, nil)
}

// TODO: add validation for token limits
func (app *application) chapterChatContinueHandler(w http.ResponseWriter, r *http.Request) {

	dbg("START_CHAPTER_CONTINUE_HANDLER")

	dbg("get userid")
	userID, err := app.getUserIDFromSubContext(r)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("userID not in context"))
		return
	}

	dbg("get bookid")
	// get param
	bookID, chapterID, err := app.readBookChapterParams(r)
	dbg("bookid: %d, chapterid: %d", bookID, chapterID)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	dbg("got params")

	dbg("get bookdata")
	bookData, chapterData, err := app.models.Transaction.GetBookAndChapter(context.Background(), bookID, chapterID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// fml
	titleCaser := cases.Title(language.English)
	bookData.Title = titleCaser.String(bookData.Title)

	dbg("received user info")
	dbg(fmt.Sprintf("userID: %d, bookID: %d, bookData: %v", userID, bookID, bookData))

	// configure
	dbg("upgrader")
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow connections from any origin for testing purposes. Adjust as necessary for production.
		},
	}

	dbg("upgrader.Upgrade")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		app.logger.Error("error upgrading to ws connection", "error", err.Error())
	}
	defer conn.Close()

	dbg("upgraded connection")
	// way to include this with expvar?
	atomic.AddInt32(&wsConnectionCount, 1)
	app.logger.Info("New connection established", "Current connection count", atomic.LoadInt32(&wsConnectionCount))

	defer func() {
		// Decrement the connection count when the connection is closed.
		atomic.AddInt32(&wsConnectionCount, -1)
		app.logger.Info("Connection closed", "Current connection count", atomic.LoadInt32(&wsConnectionCount))
	}()
	// save the previous 2 exchanges
	var msgHistory []llms.MessageContent

	for {
		_, message, err := conn.ReadMessage()
		// TODO: validate text input
		if err != nil {
			app.logger.Error("Read error", "error", err.Error())
			break
		}

		// send to LLM
		if len(message) > 0 {

			dbg("recieved msg")
			strMsg := strings.TrimSpace(string(message))
			// has to be the string of the chapter
			msgHistory = append(msgHistory, llms.TextParts(llms.ChatMessageTypeHuman, strMsg))

			filter := map[string]any{"title": bookData.Title, "chapter": strconv.Itoa(chapterData.ChapterNumber)}

			dbg(fmt.Sprintf("filter: %v", filter))

			// get textbook context
			docs, err := app.ai.DB.SimilaritySearch(context.Background(), strMsg, RETURN_THREE_DOCS, vectorstores.WithFilters(filter))
			if err != nil {
				app.logger.Error("similarity search", "error", err)
				conn.WriteMessage(websocket.TextMessage, []byte("internal server error"))
				break
			}
			dbg("docs returned: %d", len(docs))

			dbg("sending chat")
			chatRes, err := app.ai.Chat(conn, strMsg, docs, msgHistory)
			if err != nil {
				app.logger.Error("error calling chat", "msg", err)
				conn.WriteMessage(websocket.TextMessage, []byte("Chat Error"))
				break
			}

			msgHistory = append(msgHistory, llms.TextParts(llms.ChatMessageTypeAI, chatRes))

			dbg("calling InsertAll")
			dbRes, err := data.InsertAll(&app.models.UserMessage, &app.models.AIMessage, &app.models.AIUserConversation, userID, bookID, chapterID, strMsg, chatRes)
			if err != nil {
				app.logger.Error("error saving to db", "error from db", err)
				conn.WriteMessage(websocket.TextMessage, []byte("internal server error"))
				break
			}

			dbg("saved to db")
			dbg("id: %d", dbRes.ID)

		}

	}

	go func() {
		dbg("chat evel update start")
		// Run a deferred function which uses recover() to catch any panic, and log an
		// error message instead of terminating the application.
		defer func() {
			if err := recover(); err != nil {
				dbg("chat eval update recover")
				app.logger.Error(fmt.Sprintf("%v", err))
			}
		}()

		// check if the chat as already be achieved
		userActivity := &data.UserActivity{
			UserID:       userID,
			BookID:       bookID,
			ChapterID:    chapterID,
			ActivityType: "chat",
		}
		err := app.models.UserActivity.GetActivityByUserBookChapter(context.Background(), userActivity)
		if err != nil {
			app.logger.Error(err.Error())
			return
		}

		if userActivity.Score != 0 {
			// already credited
			app.logger.Info("chat requirement already completed")
			return
		}
		// get the conversation questions
		// new ctx for now
		questions, err := app.models.UserMessage.GetAll(context.Background(), userID, bookID, chapterID)
		if err != nil {
			app.logger.Error(err.Error())
		}
		if len(questions) < 3 {
			app.logger.Info("fewer than 3 questions", "want", 3, "got", len(questions))
			return
		}

		filter := map[string]any{"title": bookData.Title, "chapter": strconv.Itoa(chapterData.ChapterNumber)}

		docs, err := app.ai.DB.SimilaritySearch(context.Background(), strings.Join(questions, " "), 5, vectorstores.WithFilters(filter))
		if err != nil {
			app.logger.Error(err.Error())
			return
		}

		questionFeedback, err := app.ai.AssessStudentQuestionQuality(questions, docs)
		if err != nil {
			app.logger.Error(err.Error())
			return
		}

		if questionFeedback.Score == 0 {
			app.logger.Info("chat score not achieved", "reason", questionFeedback.Feedback)
		}

		chatEval := &data.ChatEval{
			UserID:    userID,
			BookID:    bookID,
			ChapterID: chapterID,
			Score:     questionFeedback.Score,
			Feedback:  questionFeedback.Feedback,
		}
		// create or store results in DB
		err = app.models.Transaction.UpdateUserChat(chatEval, chapterData.ChapterNumber)
		if err != nil {
			app.logger.Error(err.Error())
			return
		}

		dbg("end of chat eval update")

	}()

	dbg("connection closed")
	dbg("CHAT_HANDLER_CONT_END")
}
