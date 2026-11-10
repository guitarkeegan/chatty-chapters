package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
)

var dbg = func() func(format string, as ...any) {
	if os.Getenv("DEBUG") == "" {
		return func(string, ...any) {}
	}
	return func(format string, as ...any) {
		log.Printf(format, as...)
	}
}()

const (
	REQUIRED_NUM_OF_RESPONSES = 3
	REQUIRED_NUM_OF_QUESTIONS = 3
	REQUIRED_NUM_OF_ANSWERS   = 3
)

type TextBookMetadata struct {
	Title   string
	Authors []string
	Edition int
	Chapter int
}

type QuizQuestion struct {
	Question string `json:"question"`
}

type ResponseFeedback struct {
	Score    int    `json:"score"`
	Feedback string `json:"feedback"`
}

type ChatConversation struct {
	UserMessage string
	AIMessage   string
}

type Config struct {
	Options []openai.Option
}

type Ai struct {
	LLM *openai.LLM
	DB  *pgvector.Store
}

func New(apikey, env string, pool pgvector.PGXConn) (*Ai, error) {

	embedderOpts := []openai.Option{
		openai.WithToken(os.Getenv("OPENAI")),
		openai.WithModel("gpt-4o-mini"),
		openai.WithEmbeddingModel("text-embedding-3-large"),
	}

	LLMOpts := []openai.Option{
		openai.WithToken(os.Getenv("OPENAI")),
		openai.WithModel("gpt-4o"),
	}

	embedLLM, err := openai.New(embedderOpts...)
	if err != nil {
		return nil, DBInitError("failed to connect with openai embedder model", err)
	}

	LLM, err := openai.New(LLMOpts...)
	if err != nil {
		return nil, DBInitError("failed to connect with openai chat model", err)
	}

	e, err := embeddings.NewEmbedder(embedLLM)
	if err != nil {
		return nil, DBInitError("failed to connect with openai embedder", err)
	}

	// Create a new pgvector store.
	ctx := context.Background()
	db, err := pgvector.New(
		ctx,
		pgvector.WithEmbedder(e),
		pgvector.WithConn(pool),
		pgvector.WithCollectionTableName("documents"),
		pgvector.WithEmbeddingTableName("embedding_table"),
		pgvector.WithCollectionName("my_collection"),
	)
	if err != nil {
		return nil, DBInitError("failed to create new store", err)
	}

	return &Ai{
		LLM: LLM,
		DB:  &db,
	}, nil
}

func (ai *Ai) Chat(conn *websocket.Conn, prompt string, documents []schema.Document, msgHist []llms.MessageContent) (string, error) {
	// check if prompt is empty
	if prompt == "" {
		return "", errors.New("no prompt")
	}

	if len(documents) == 0 {
		return "", errors.New("no documents returned")
	}

	documents = deduplicate(documents)
	docs := stringifyDocs(documents)

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, fmt.Sprintf("You are an academic tutor for a variety of classes. You are currently in a session with a student with a student. Your tone is casual, but appropriate for the tutoring session. Your job is to help the student understand the reading material. The conversation should be a syncratic dialogue. Ask the student follup question if approapriate. Keep your responses as concise as possible. Try to reference the chapter materials provided to answer student questions. Relavent context from the textbook here: %s", docs)),
	}

	// include the most recent chat messages
	switch len(msgHist) {
	case 3:
		// get the first exchange
		content = append(content, msgHist[0:1]...)
	case 5:
		content = append(content, msgHist[len(msgHist)-5:len(msgHist)-1]...)
	default:
		if len(msgHist) > 5 {
			// Get the last 4 items for any length greater than 5
			content = append(content, msgHist[len(msgHist)-5:len(msgHist)-1]...)
		}
	}

	content = append(content, llms.TextParts(llms.ChatMessageTypeHuman, prompt))

	dbg("current chat: ")
	dbg("%+v", content)

	aiRes, err := ai.LLM.GenerateContent(context.Background(), content,
		llms.WithMaxTokens(1024),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			conn.WriteMessage(websocket.TextMessage, chunk)
			return nil
		}))
	if err != nil {
		return "", err
	}

	if aiRes.Choices[0].StopReason == "stop" {
		dbg("StopReason: stop")
		conn.WriteMessage(websocket.TextMessage, []byte("\n\r\n\r"))
	} else {
		conn.WriteMessage(websocket.TextMessage, []byte("There was a problem getting a response. \n\r\n\r"))
	}

	return cleanAIResponse(aiRes.Choices[0].Content), nil

}

// should i also generate the answers?
func (ai *Ai) GenerateQuizQuestions(chatConversation []ChatConversation, bookTitle string, chapterNumber, docsToReturn int) ([]QuizQuestion, error) {

	dbg(" generating quiz questions")
	dbg(" params: %+v, bookTitle: %s, chapterNumber: %d, docsToReturn: %d", chatConversation, bookTitle, chapterNumber, docsToReturn)
	// format for prompt
	var formattedConvo string
	for _, chatConvo := range chatConversation {
		formattedConvo += fmt.Sprintf("user: %s ai: %s ", chatConvo.UserMessage, chatConvo.AIMessage)
	}
	dbg(" formatted convo")
	dbg(" formattedConvo: %+v", formattedConvo)
	// get relavent docs?
	filter := map[string]any{"title": titleCase(bookTitle), "chapter": strconv.Itoa(chapterNumber)}

	dbg(" filter: %+v", filter)

	ctx := context.Background()
	docs, err := ai.DB.SimilaritySearch(ctx, formattedConvo, docsToReturn, vectorstores.WithFilters(filter))

	dbg(" docs returned: %d", len(docs))
	documents := deduplicate(docs)
	strDocuments := stringifyDocs(documents)

	if err != nil {
		return nil, fmt.Errorf("gen quiz questions: %w", err)
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("no docs returned. Want: %d, Got: %d", docsToReturn, len(docs))
	}
	// write system prompt
	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, fmt.Sprintf("Please use the user and ai conversation, along with the textbook exerpts, to create 3 quiz questions. The questions can be somewhat open ended, and should promote critical thinking. Conversation: %s; Texbook Exerpts: %s", formattedConvo, strDocuments)),
	}

	aiRes, err := ai.LLM.GenerateContent(context.Background(), content,
		llms.WithMaxTokens(1024),
		llms.WithTools(quizQuestionTool),
		llms.WithToolChoice("generateQuizQuestions"),
	)
	if err != nil {
		return nil, fmt.Errorf("gen content: %w", err)
	}

	var args struct {
		Q1 string `json:"q1"`
		Q2 string `json:"q2"`
		Q3 string `json:"q3"`
	}

	err = json.Unmarshal([]byte(aiRes.Choices[0].ToolCalls[0].FunctionCall.Arguments), &args)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling: %w", err)
	}

	dbg(" generateQuizQuestions")
	output, err := generateQuizQuestions(args.Q1, args.Q2, args.Q3)
	if err != nil {
		return nil, fmt.Errorf("generate quiz questions: %w", err)
	}
	dbg(" output length: %d", len(output))

	if len(output) < 3 {
		return nil, fmt.Errorf("not enough quiz questions: Want: %d, Got: %d", 3, len(output))
	}

	dbg(" leaving generateQuizQuestions")

	return output, nil

}

// asses the quiz answers
func (ai *Ai) AssessAnswers(questions, answers []string, documents []schema.Document) ([]ResponseFeedback, error) {

	if len(questions) < 3 || len(answers) < 3 || len(questions) != len(answers) {
		return []ResponseFeedback{}, fmt.Errorf("questions length Want: %d, Got: %d, answer length Want: %d, Got: %d", REQUIRED_NUM_OF_QUESTIONS, len(questions), REQUIRED_NUM_OF_ANSWERS, len(answers))
	}

	if len(documents) == 0 {
		return []ResponseFeedback{}, fmt.Errorf("AssessAnswers: length of documents is 0")
	}

	documents = deduplicate(documents)
	docs := stringifyDocs(documents)

	// format for better prompt
	var fmtQuestions string
	for i, q := range questions {
		fmtQuestions += fmt.Sprintf("%d. %s", i+1, q)
	}

	var fmtAnswers string
	for i, q := range answers {
		fmtAnswers += fmt.Sprintf("%d. %s", i+1, q)
	}

	content := []llms.MessageContent{
		// TODO: it is dangerous to pass in the user responses, as they might divert the prompt
		llms.TextParts(llms.ChatMessageTypeSystem, fmt.Sprintf("Here is a quiz that contained 3 questions: %s. Here are the student answers for each question: %s. Assess how well the student answered the questions based on the reading material, and their critical thinking skills. Textbook exerpts: %s", fmtQuestions, fmtAnswers, docs)),
	}

	aiRes, err := ai.LLM.GenerateContent(context.Background(), content,
		llms.WithMaxTokens(1024),
		llms.WithTools(quizFeedbackTool),
		llms.WithToolChoice("quizFeedbackTool"),
	)
	if err != nil {
		return []ResponseFeedback{}, fmt.Errorf("generateQuizFeedback: %w", err)
	}

	var args QuizFeedbackRequest

	err = json.Unmarshal([]byte(aiRes.Choices[0].ToolCalls[0].FunctionCall.Arguments), &args)
	if err != nil {
		return []ResponseFeedback{}, fmt.Errorf("failed to unmarshal tool call: %w", err)
	}

	output, err := generateQuizFeedback(args)
	if err != nil {
		return []ResponseFeedback{}, fmt.Errorf("generateQuizFeedback: %w", err)
	}

	return output, nil

}

func (ai *Ai) AssessStudentQuestionQuality(questions []string, documents []schema.Document) (QuestionFeedback, error) {

	documents = deduplicate(documents)
	docs := stringifyDocs(documents)

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, fmt.Sprintf("You are a teaching assistant, and you are assessing questions that a student has asked during a meeting with another teaching assistant. Your job is to assess the quality of the questions, and generate a score based on the student asked questions that were related to the reading material in the textbook. Here are the questions asked by the student: '%s'. Here are some chapter exerpts to give you an idea of what the chapter covered: '%s'. Please output a score of 1 if the student full filled the requirement, or 0 if they did not. followed by and explanation for the score.", questions, docs)),
	}

	aiRes, err := ai.LLM.GenerateContent(context.Background(), content,
		llms.WithMaxTokens(1024),
		llms.WithTools(questionQualityTool),
		llms.WithToolChoice("generateScoreAndFeedback"),
	)
	if err != nil {
		return QuestionFeedback{}, fmt.Errorf("GenerateContent: %w", err)
	}

	var args struct {
		Score    int    `json:"score"`
		Feedback string `json:"feedback"`
	}

	err = json.Unmarshal([]byte(aiRes.Choices[0].ToolCalls[0].FunctionCall.Arguments), &args)
	if err != nil {
		return QuestionFeedback{}, fmt.Errorf("failed to unmarshal tool call: %w", err)
	}

	output, err := generateScoreAndFeedback(args.Score, args.Feedback)
	if err != nil {
		return QuestionFeedback{}, fmt.Errorf("generateScoreAndFeedback: %w", err)
	}

	return output, nil

}
