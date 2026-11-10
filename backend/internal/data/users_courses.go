package data

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserCourse struct {
	UserID   int64 `json:"user_id"`
	CourseID int64 `json:"course_id"`
}

type UserCourseInfo struct {
	CourseID    int64  `json:"course_id"`
	CourseTitle string `json:"course_title"`
	Semester    string `json:"semester"`
	Year        int    `json:"year"`
	BookTitle   string `json:"book_title"`
	BookID      int64  `json:"book_id"`
}

type UserCourseModel struct {
	DB *pgxpool.Pool
}

func NewUserCourse(userID, courseID int64) (*UserCourse, error) {

	return &UserCourse{
		UserID:   userID,
		CourseID: courseID,
	}, nil

}

func (cb *UserCourseModel) Insert(userCourse *UserCourse) error {
	stmt := `
        INSERT INTO users_courses (user_id, course_id)
        VALUES ($1, $2)
	`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := cb.DB.Exec(ctx, stmt, userCourse.UserID, userCourse.CourseID)
	if err != nil {
		return fmt.Errorf("UserCourseModel Insert: %w", err)
	}

	return nil

}

func (cb *UserCourseModel) GetAll(userID int64) ([]UserCourseInfo, error) {

	stmt := `
	SELECT c.id, c.title as course_title, c.semester, c.year, b.title as book_title, b.id as book_id
	FROM users_courses uc
	JOIN courses c ON c.id = uc.course_id
	JOIN courses_books cb ON cb.course_id = uc.course_id
	JOIN books b ON cb.book_id = b.id
	WHERE uc.user_id = $1
	`

	rows, err := cb.DB.Query(context.Background(), stmt, userID)
	if err != nil {
		return nil, fmt.Errorf("UserCourseModel: GetAll: %w", err)
	}
	defer rows.Close()

	var courses []UserCourseInfo
	for rows.Next() {
		var course UserCourseInfo

		err := rows.Scan(&course.CourseID, &course.CourseTitle, &course.Semester, &course.Year, &course.BookTitle, &course.BookID)
		if err != nil {
			return nil, fmt.Errorf("UserCourseModel: GetAll: %w", err)
		}

		courses = append(courses, course)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("UserCourseModel: GetAll: %w", rows.Err())
	}

	return courses, nil
}
