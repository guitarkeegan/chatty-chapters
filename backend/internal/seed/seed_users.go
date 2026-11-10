package seed

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/guitarkeegan/chatty-chapters/backend/internal/data"
)

const (
	demoStudentEmail = "student@example.com"
	demoTeacherEmail = "teacher@example.com"
)

// insert 2 demo users into the db. The cognito subs are random, so to log in
// as one of these users, update cognito_sub to match a user in your own pool.
func SeedUsers(db *pgxpool.Pool) error {
	um := &data.UserModel{
		DB: db,
	}

	student := &data.User{
		FirstName:  "Demo",
		LastName:   "Student",
		Email:      demoStudentEmail,
		CognitoSub: uuid.New(),
	}

	teacher := &data.User{
		FirstName:  "Demo",
		LastName:   "Teacher",
		Email:      demoTeacherEmail,
		CognitoSub: uuid.New(),
	}

	if err := um.Insert(student); err != nil {
		return err
	}
	if err := um.Insert(teacher); err != nil {
		return err
	}

	return nil
}
