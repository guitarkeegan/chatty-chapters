package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID         int64     `json:"id"`
	CognitoSub uuid.UUID `json:"cognito_sub"`
	CreatedAt  time.Time `json:"-"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email"`
	Version    int       `json:"version"`
	Role       string    `json:"role"`
}

// was using the embedded user, but had issues
type UserRegistration struct {
	CognitoSub uuid.UUID `json:"cognito_sub"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email"`
	CourseID   int64     `json:"course_id"`
}

type UserResult struct {
	UserID             int64  `json:"user_id"`
	FirstName          string `json:"first_name"`
	LastName           string `json:"last_name"`
	ChatScore          int    `json:"chat_score"`
	QuizScore          int    `json:"quiz_score"`
	TotalActivityScore int    `json:"total_activity_score"`
}

// TODO: these should probably be the same type
type UserChapterResults struct {
	ChatQuestions []string `json:"questions"`
	QuizQuestions []string `json:"quiz_questions"`
	QuizAnswers   []string `json:"quiz_answers"`
	QuizFeedback  []string `json:"quiz_feedback"`
	ChatFeedback  []string `json:"chat_feedback"`
}

type UserActivitiesResult struct {
	ChatQuestions []string `json:"questions"`
	QuizQuestions []string `json:"quiz_questions"`
	QuizAnswers   []string `json:"quiz_answers"`
}

type UserModel struct {
	DB *pgxpool.Pool
}

func NewUser(cognitoSub uuid.UUID, firstName, lastName, email, role string) *User {
	return &User{
		CognitoSub: cognitoSub,
		FirstName:  firstName,
		LastName:   lastName,
		Email:      email,
	}
}

// example
func (um UserModel) Insert(user *User) error {

	query := `
	INSERT INTO users (first_name, last_name, email, cognito_sub)
	VALUES ( $1, $2, $3, $4 )
	RETURNING id, created_at, version`

	args := []any{user.FirstName, user.LastName, user.Email, user.CognitoSub}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return um.DB.QueryRow(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)

}

func (um UserModel) Get(user *User) error {

	stmt := `
	SELECT id, first_name, last_name, email, role
	FROM users
	WHERE cognito_sub = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := um.DB.QueryRow(ctx, stmt, user.CognitoSub).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("UserModel Get: %w", err)
	}

	return nil

}

// get the userID by email
func (um UserModel) GetUserByEmail(ctx context.Context, email string) (int64, error) {

	query := `
	SELECT id
	FROM users
	WHERE email = $1
	`

	var id int64
	// arg in param
	err := um.DB.QueryRow(ctx, query, email).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("no user found with email %s: %w", email, err)
		}
		return 0, fmt.Errorf("error while scanning row: %w", err)
	}
	return id, nil

}

// TODO: combine these two functions?
func (um UserModel) GetUserByCognitoSub(ctx context.Context, sub uuid.UUID) (int64, error) {

	query := `
	SELECT id
	FROM users
	WHERE cognito_sub = $1
	`

	var id int64
	err := um.DB.QueryRow(ctx, query, sub).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("no user found with cognito_sub %s: %w", sub, err)
		}
		return 0, fmt.Errorf("error while scanning row: %w", err)
	}
	return id, nil

}

func (um UserModel) CheckRole(ctx context.Context, sub uuid.UUID) (string, error) {

	query := `
	SELECT role
	FROM users
	WHERE cognito_sub = $1
	`

	var role string

	if err := um.DB.QueryRow(ctx, query, sub).Scan(&role); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("no user found with cognito_sub %s: %w", sub, err)
		}
		return "", fmt.Errorf("error while scanning row: %w", err)
	}

	return role, nil

}

// Use this to register any number of studetns via the userRegistration endpoint.
func (um UserModel) InsertMany(userRegistrations *[]UserRegistration) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
	defer cancel()

	tx, err := um.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	insertUserQuery := `
        INSERT INTO users (cognito_sub, first_name, last_name, email)
        VALUES ($1, $2, $3, $4)
        RETURNING id
	`
	insertCourseBookQuery := `
        INSERT INTO users_courses (user_id, course_id)
        VALUES ($1, $2)
	`

	getBooksQuery := `
	SELECT b.id, b.title, b.created_at, b.edition, b.year
	FROM books b
	JOIN courses_books cb ON cb.book_id = b.id
	WHERE cb.course_id = $1;
	`
	// get user books by users_courses

	for _, reg := range *userRegistrations {
		var userID int64
		err = tx.QueryRow(ctx, insertUserQuery, reg.CognitoSub, reg.FirstName, reg.LastName, reg.Email).Scan(&userID)
		if err != nil {
			return fmt.Errorf("register insert user: %v", err)
		}

		_, err = tx.Exec(ctx, insertCourseBookQuery, userID, reg.CourseID)
		if err != nil {
			return fmt.Errorf("register insert into courses_books: %w", err)
		}

		rows, err := tx.Query(ctx, getBooksQuery, reg.CourseID)

		if err != nil {
			return fmt.Errorf("register getBooks: %w", err)
		}

		var books []Book
		for rows.Next() {

			var book Book
			err := rows.Scan(&book.ID, &book.Title, &book.CreatedAt, &book.Edition, &book.Year)
			if err != nil {
				return fmt.Errorf("register scan book row: %w", err)
			}

			books = append(books, book)
		}

		if rows.Err() != nil {
			return fmt.Errorf("register scan book: %w", err)
		}

		// insert users_books
		// TODO: do this better
		for _, book := range books {

			userBooksStmt := `
			INSERT INTO user_books (user_id, book_id)
			VALUES ( $1, $2 )
			`

			_, err := tx.Exec(ctx, userBooksStmt, userID, book.ID)
			if err != nil {
				return SQLError(fmt.Sprintf("error adding userID: %d, bookID: %d, to user_book", userID, book.ID), err)
			}

			// create user progress
			userProgStmt := `
			INSERT INTO user_progress ( user_id, book_id )
			VALUES ( $1, $2 )
			RETURNING ( id )
			`
			_, err = tx.Exec(ctx, userProgStmt, userID, book.ID)
			if err != nil {
				return fmt.Errorf("registration insert userProgres: %w", err)
			}
		}

	}

	return nil

}

func (um UserModel) InsertPublic(userRegistration *UserRegistration) error {

	// Limit public to Life of Chopin
	// and Alice in Wonderland
	var publicCourseID = 3

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
	defer cancel()

	tx, err := um.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	insertUserQuery := `
        INSERT INTO users (cognito_sub, first_name, last_name, email)
        VALUES ($1, $2, $3, $4)
        RETURNING id
	`
	insertUsersCoursesQuery := `
        INSERT INTO users_courses (user_id, course_id)
        VALUES ($1, $2)
	`

	getBooksQuery := `
	SELECT b.id, b.title, b.created_at, b.edition, b.year
	FROM books b
	JOIN courses_books cb ON cb.book_id = b.id
	WHERE cb.course_id = $1;
	`
	// get user books by users_courses

	var userID int64
	err = tx.QueryRow(ctx, insertUserQuery, userRegistration.CognitoSub, userRegistration.FirstName, userRegistration.LastName, userRegistration.Email).Scan(&userID)
	if err != nil {
		return fmt.Errorf("register insert user: %v", err)
	}

	_, err = tx.Exec(ctx, insertUsersCoursesQuery, userID, publicCourseID)
	if err != nil {
		return fmt.Errorf("register insert into courses_books: %w", err)
	}

	rows, err := tx.Query(ctx, getBooksQuery, publicCourseID)

	if err != nil {
		return fmt.Errorf("register getBooks: %w", err)
	}

	var books []Book
	for rows.Next() {

		var book Book
		err := rows.Scan(&book.ID, &book.Title, &book.CreatedAt, &book.Edition, &book.Year)
		if err != nil {
			return fmt.Errorf("register scan book row: %w", err)
		}

		books = append(books, book)
	}

	if rows.Err() != nil {
		return fmt.Errorf("register scan book: %w", err)
	}

	// insert users_books
	// TODO: do this better
	for _, book := range books {

		userBooksStmt := `
			INSERT INTO user_books (user_id, book_id)
			VALUES ( $1, $2 )
			`

		_, err := tx.Exec(ctx, userBooksStmt, userID, book.ID)
		if err != nil {
			return SQLError(fmt.Sprintf("error adding userID: %d, bookID: %d, to user_book", userID, book.ID), err)
		}

		// create user progress
		userProgStmt := `
			INSERT INTO user_progress ( user_id, book_id )
			VALUES ( $1, $2 )
			RETURNING ( id )
			`
		_, err = tx.Exec(ctx, userProgStmt, userID, book.ID)
		if err != nil {
			return fmt.Errorf("registration insert userProgres: %w", err)
		}
	}

	return nil

}

func (um UserModel) GetChapterResult(courseID, bookID, chapterID int64) ([]UserResult, error) {

	stmt := `
	WITH quiz_scores AS (
	    SELECT
		qre.user_id,
		SUM(qre.score) AS total_quiz_score
	    FROM
		quiz_response_evals qre
	    JOIN
		chapters ch ON qre.chapter_id = ch.id
	    WHERE
		ch.id = $3
	    GROUP BY
		qre.user_id
	),
	chat_scores AS (
	    SELECT
		ce.user_id,
		ce.score AS chat_score
	    FROM
		chat_evals ce
	    JOIN
		books b ON ce.book_id = b.id
	    JOIN
		chapters ch ON ce.chapter_id = ch.id
	    WHERE
		ch.id = $3 AND b.id = $2
	)
	SELECT
	    u.id,
	    u.first_name,
	    u.last_name,
	    cs.chat_score,
	    qs.total_quiz_score,
	    cs.chat_score + qs.total_quiz_score AS total_activity_score
	FROM
	    users u
	JOIN
	    users_courses uc ON uc.user_id = u.id AND uc.course_id = $1
	JOIN
	    chat_scores cs ON cs.user_id = u.id
	JOIN
	    quiz_scores qs ON qs.user_id = u.id
	ORDER BY
	    u.last_name,
	    u.first_name;
	`

	ctx := context.Background()

	rows, err := um.DB.Query(ctx, stmt, courseID, bookID, chapterID)
	if err != nil {
		return []UserResult{}, fmt.Errorf("getChapterResults: rows: %w", err)
	}

	defer rows.Close()

	results := []UserResult{}

	for rows.Next() {

		result := UserResult{}

		if err := rows.Scan(&result.UserID, &result.FirstName, &result.LastName, &result.ChatScore, &result.QuizScore, &result.TotalActivityScore); err != nil {
			return results, fmt.Errorf("getChapterResults: rows.Next(): %w", err)
		}

		results = append(results, result)

	}

	if rows.Err() != nil {
		return results, fmt.Errorf("getChapterResults: rows.Err(): %w", rows.Err())
	}

	return results, nil

}

func (u UserModel) GetChapterResultsStudent(bookID, chapterID, userID int64) (UserChapterResults, error) {
	query := `
	WITH chat_data AS (
		SELECT um.message AS chat_question, ce.score AS chat_score, ce.feedback AS chat_feedback
		FROM user_messages um
		LEFT JOIN chat_evals ce ON ce.user_id = um.user_id AND ce.chapter_id = um.chapter_id
		WHERE um.user_id = $1 AND um.book_id = $2 AND um.chapter_id = $3
	),
	quiz_data AS (
		SELECT qq.quiz_question, qr.student_response, qre.score AS quiz_score, qre.response_feedback
		FROM quizzes q
		JOIN quiz_questions qq ON q.id = qq.quiz_id
		JOIN quiz_responses qr ON qq.id = qr.quiz_question_id
		JOIN quiz_response_evals qre ON qr.id = qre.quiz_response_id
		WHERE q.user_id = $1 AND q.book_id = $2 AND q.chapter_id = $3
	)
	SELECT 
		COALESCE(array_agg(DISTINCT cd.chat_question), ARRAY[]::TEXT[]) AS chat_questions,
		COALESCE(array_agg(DISTINCT qd.quiz_question), ARRAY[]::TEXT[]) AS quiz_questions,
		COALESCE(array_agg(DISTINCT qd.student_response), ARRAY[]::TEXT[]) AS quiz_answers,
		COALESCE(array_agg(DISTINCT qd.response_feedback), ARRAY[]::TEXT[]) AS quiz_feedback,
		COALESCE(array_agg(DISTINCT cd.chat_feedback), ARRAY[]::TEXT[]) AS chat_feedback
	FROM chat_data cd
	FULL OUTER JOIN quiz_data qd ON true;
	`

	var result UserChapterResults
	err := u.DB.QueryRow(context.Background(), query, userID, bookID, chapterID).Scan(
		&result.ChatQuestions,
		&result.QuizQuestions,
		&result.QuizAnswers,
		&result.QuizFeedback,
		&result.ChatFeedback,
	)

	if err != nil {
		return UserChapterResults{}, fmt.Errorf("error fetching chapter results for student: %w", err)
	}

	return result, nil
}
