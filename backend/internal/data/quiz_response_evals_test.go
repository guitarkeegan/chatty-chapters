//go:build integration

package data

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func beforeQuizEvals() (*QuizResponseEvalModel, error) {

	dns := os.Getenv("DNS_CONN_DEV")

	db, err := openDB(dns)
	if err != nil {
		return nil, err
	}

	return &QuizResponseEvalModel{
		DB: db,
	}, nil
}

func TestGetQuizResults(t *testing.T) {

	err := godotenv.Load()
	if err != nil {
		t.Errorf("env load failed: %q", err)
	}

	qrem, err := beforeQuizEvals()
	if err != nil {
		t.Fatalf("beforeQuizEvals failed: %v", err)
	}

	var (
		testUserID    int64 = 1
		testBookID    int64 = 1
		testChapterID int64 = 1
	)

	quizResults, err := qrem.GetQuizResults(testUserID, testBookID, testChapterID)
	if err != nil {
		t.Fatalf("func call failed: %v", err)
	}

	if len(quizResults) != 3 {
		t.Fatalf("quizResults: Want: %d, Got: %d", 3, len(quizResults))
	}

	for _, res := range quizResults {
		t.Logf("%+v", *res)
	}

}
