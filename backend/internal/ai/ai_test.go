package ai

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/vectorstores"
)

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
}

func TestSimularitySearch(t *testing.T) {

	// setup db and connect
	err := godotenv.Load()
	if err != nil {
		t.Errorf("error on env load: %q", err.Error())
		return
	}

	cfg := config{
		port: 5432,
		env:  "development",
		db: struct{ dsn string }{
			dsn: os.Getenv("DNS_CONN_DEV"),
		},
	}

	// create connection pool
	db, err := openDB(cfg)
	if err != nil {
		t.Error(err.Error())
		return
	}
	defer db.Close()

	ai, err := New(os.Getenv("OPENAI"), "development", db)
	if err != nil {
		t.Error(err.Error())
		return
	}
	// end setup

	// unable to search by author / slice, but can do title, edition, and chapter
	filters := []map[string]any{
		{"chapter": "2"},
		{"title": "Enjoyment Of Music"},
		{"title": "History Of Jazz", "chapter": "1"},
	}

	var count int

	for _, filter := range filters {

		docs, err := ai.DB.SimilaritySearch(context.Background(), "music", 3, vectorstores.WithFilters(filter))

		if err != nil {
			t.Errorf("Error during similarity search with filter %v: %s", filter, err.Error())
			return
		}

		if len(docs) == 0 || len(docs) != 3 {
			t.Errorf("Search did not return correct amount of docs. WANT: %d, GOT: %d", 3, len(docs))
			return
		}

		for _, doc := range docs {
			fmt.Printf("returned values: %+v\n", doc.Metadata)
			for key, value := range filter {
				fmt.Printf("key: %s value: %s", key, value)
				count++
			}
		}

	}

	fmt.Println(count)

}

func TestAssessStudentQuestionQuality(t *testing.T) {

	err := godotenv.Load()
	if err != nil {
		t.Errorf("error on env load: %q", err.Error())
		return
	}

	cfg := config{
		port: 5432,
		env:  "development",
		db: struct{ dsn string }{
			dsn: os.Getenv("DNS_CONN_DEV"),
		},
	}
	// create connection pool
	db, err := openDB(cfg)
	if err != nil {
		t.Error(err.Error())
		return
	}
	defer db.Close()

	ai, err := New(os.Getenv("OPENAI"), "development", db)
	if err != nil {
		t.Error(err.Error())
		return
	}
	// end setup
	questions := []string{
		"What is jazz?",
		"Why does knowing about jazz matter today?",
		"I like to listen to a lot of country music. Is there any relationship to jazz?",
		"What are some of the instruments mentioned in this chapter?",
		"What are some of the materials that make up the old drums that they were talking about?",
		"Where does jazz come from?",
		"Where does jazz come from?",
		"Were there any instruments that were derived from traditional African ones?",
		"What are some other things from this chapter that are important that I keep in mind, while continuing through the course?",
	}

	filter := map[string]any{"title": "History Of Jazz", "chapter": "1"}

	// get textbook context
	// testing 5 docs returned
	docs, err := ai.DB.SimilaritySearch(context.Background(), strings.Join(questions, " "), 5, vectorstores.WithFilters(filter))
	if err != nil {
		t.Fatal(err)
	}

	res, err := ai.AssessStudentQuestionQuality(questions, docs)
	if err != nil {
		t.Fatal(err)
	}

	if res.Score > 1 || res.Score < 0 || res.Feedback == "" {
		t.Fatal("score and/or feedback unacceptable")
	}

	t.Log("structured output: ", res)

}

func TestGenerateQuizQuestions(t *testing.T) {

	err := godotenv.Load()
	if err != nil {
		t.Errorf("error on env load: %q", err.Error())
		return
	}

	cfg := config{
		port: 5432,
		env:  "development",
		db: struct{ dsn string }{
			dsn: os.Getenv("DNS_CONN_DEV"),
		},
	}
	// create connection pool
	db, err := openDB(cfg)
	if err != nil {
		t.Error(err.Error())
		return
	}
	defer db.Close()

	ai, err := New(os.Getenv("OPENAI"), "development", db)
	if err != nil {
		t.Error(err.Error())
		return
	}
	conversation := []ChatConversation{
		{
			UserMessage: "i am a user",
			AIMessage:   "hey there user",
		},
		{
			UserMessage: "how can i use this app?",
			AIMessage:   "This app is designed to assist you with various tasks. You can ask questions, get information, or even explore topics like the history of jazz!",
		},
		{
			UserMessage: "what kind of information can you provide?",
			AIMessage:   "I can provide information on a wide range of topics, including music history, especially jazz. For instance, I can tell you about the origins and early influences of jazz music.",
		},
		{
			UserMessage: "can you tell me about the origins of jazz?",
			AIMessage:   "Jazz originated in the late 19th and early 20th centuries in New Orleans, Louisiana. It's a rich blend of African, Caribbean, and European musical traditions, characterized by improvisation, syncopation, and a unique rhythm.",
		},
		{
			UserMessage: "interesting! what were some early influences on jazz?",
			AIMessage:   "Early jazz was influenced by blues, ragtime, brass band music, and spirituals. These genres provided the foundation for the distinctive sound and improvisational nature of jazz.",
		},
		{
			UserMessage: "who were some key figures in the early days of jazz?",
			AIMessage:   "Key figures in the early days of jazz include Buddy Bolden, who is often considered the first jazz musician, and Jelly Roll Morton, who claimed to have invented jazz. These pioneers helped shape the genre's early development.",
		},
		{
			UserMessage: "thanks for the information!",
			AIMessage:   "You're welcome! If you have any more questions about jazz or anything else, feel free to ask.",
		},
	}

	questions, err := ai.GenerateQuizQuestions(conversation, "History Of Jazz", 1, 5)
	if err != nil {
		t.Fatal(err)
	}

	if len(questions) < 3 {
		t.Fatalf("not enough questions returned. Want: 3, Got: %d", len(questions))
	}

}

func TestAssessAnswers(t *testing.T) {

	err := godotenv.Load()
	if err != nil {
		t.Errorf("error on env load: %q", err.Error())
		return
	}

	cfg := config{
		port: 5432,
		env:  "development",
		db: struct{ dsn string }{
			dsn: os.Getenv("DNS_CONN_DEV"),
		},
	}
	// create connection pool
	db, err := openDB(cfg)
	if err != nil {
		t.Error(err.Error())
		return
	}
	defer db.Close()

	ai, err := New(os.Getenv("OPENAI"), "development", db)
	if err != nil {
		t.Error(err.Error())
		return
	}

	/*******BEGIN TEST ********/
	questions := []string{
		"What was the name of the area where the so-called slave dances that took place in New Orleans?",
		"Why does the author describe Congo Square as a nexus where opposites collided?",
		"Who was the musician that coined the phrase, 'if you can't manage to put tinges of Spanish in your tunes, you will never be able to get the right seasoning.'?",
	}

	answers := []string{
		"The area of New Orleans in question was called Congo Square.",
		"The author was making a comment on the differences between how music is performed and participated in within a euro-centric context, verses and African one.",
		"The musician who coined the phrase was pianist, Jelly Roll Morton",
	}

	filter := map[string]any{"title": "History Of Jazz", "chapter": "1"}

	// get textbook context
	// testing 5 docs returned
	docs, err := ai.DB.SimilaritySearch(context.Background(), strings.Join(questions, " "), 5, vectorstores.WithFilters(filter))
	if err != nil {
		t.Fatal(err)
	}

	responseFeedback, err := ai.AssessAnswers(questions, answers, docs)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Response Feedback: %+v", responseFeedback)

}

func openDB(cfg config) (*pgxpool.Pool, error) {

	// TODO: add config for DB
	db, err := pgxpool.New(context.Background(), cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.Ping(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	// Return the sql.DB connection pool.
	return db, nil
}
