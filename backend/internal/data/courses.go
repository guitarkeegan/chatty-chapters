package data

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Course struct {
	ID       int64     `json:"id"`
	Title    string    `json:"title"`
	StartAt  time.Time `json:"start_at"`
	EndAt    time.Time `json:"end_at"`
	Semester string    `json:"semester"`
	Year     int       `json:"year"`
}

type CourseModel struct {
	DB *pgxpool.Pool
}

func NewCourse(title, semester string, year int) *Course {
	return &Course{
		Title:    title,
		Semester: semester,
		Year:     year,
	}
}

func (cm *CourseModel) Insert(course *Course) error {

	stmt := `
	INSERT INTO courses (title, semester, year)
	VALUES ( $1, $2, $3 )
	RETURNING id
	`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	err := cm.DB.QueryRow(ctx, stmt, course.Title, course.Semester, course.Year).Scan(&course.ID)
	if err != nil {
		return fmt.Errorf("CourseModel Insert: %w", err)
	}

	return nil

}
