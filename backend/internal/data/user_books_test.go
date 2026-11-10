package data

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var (
	demoStudentEmail = "student@example.com"
)

func before() (*UserModel, *UserBookModel, error) {

	dns := os.Getenv("DNS_CONN_DEV")

	db, err := openDB(dns)
	if err != nil {
		return nil, nil, err
	}

	return &UserModel{
			DB: db,
		},
		&UserBookModel{
			DB: db,
		}, nil
}

func TestGetBooksByUserID(t *testing.T) {

	err := godotenv.Load()
	if err != nil {
		t.Errorf("env load failed: %q", err)
	}

	var want = 4

	ctx := context.Background()

	// get user id
	um, ubm, err := before()
	if err != nil {
		t.Errorf("before failed: %q", err.Error())
	}
	userID, err := um.GetUserByEmail(ctx, demoStudentEmail)
	if err != nil {
		t.Errorf("no user with that email: %q\n", err.Error())
	}
	// make call
	books, err := ubm.GetBooksByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("getBooksByUserID: %v", err)
	}
	// expect length of array to be for test
	got := len(books)
	if got != want {
		t.Errorf("GOT: %d, WANT: %d\n", got, want)
	}

}

func openDB(dsn string) (*pgxpool.Pool, error) {

	// TODO: add config for DB
	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	err = db.Ping(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	// Return the sql.DB connection pool.
	return db, nil
}
