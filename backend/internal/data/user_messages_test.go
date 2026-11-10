package data

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func beforeUserMessages() (*UserMessageModel, error) {

	dns := os.Getenv("DNS_CONN_DEV")

	db, err := openDB(dns)
	if err != nil {
		return nil, err
	}
	return &UserMessageModel{
			DB: db,
		},
		nil
}

func TestGetAll(t *testing.T) {

	var (
		userID    int64 = 1
		bookID    int64 = 1
		chapterID int64 = 1
	)

	err := godotenv.Load()
	if err != nil {
		t.Errorf("env load failed: %q", err)
	}

	umm, err := beforeUserMessages()
	if err != nil {
		t.Fatal(err)
	}

	questions, err := umm.GetAll(context.Background(), userID, bookID, chapterID)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(questions)

	if len(questions) == 0 {
		t.Fatalf("no questions returned")
	}

}
