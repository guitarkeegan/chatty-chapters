package data

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"golang.org/x/net/context"
)

func beforeConversation() (*AIUserConversationModel, error) {

	dns := os.Getenv("DNS_CONN_DEV")

	db, err := openDB(dns)
	if err != nil {
		return nil, err
	}
	return &AIUserConversationModel{
			DB: db,
		},
		nil
}

func TestGetChapterConversation(t *testing.T) {

	var (
		userID    int64 = 1
		bookID    int64 = 1
		chapterID int64 = 1
	)

	err := godotenv.Load()
	if err != nil {
		t.Errorf("env load failed: %q", err)
	}

	ucm, err := beforeConversation()
	if err != nil {
		t.Fatal(err)
	}

	res, err := ucm.GetChapterConversation(context.Background(), userID, bookID, chapterID)
	if err != nil {
		t.Fatal(err)
	}

	if res == nil {
		t.Fatal("returned nil")
	}

	t.Logf("%+v", res)

}
