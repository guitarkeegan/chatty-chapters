package canvas

import (
	"fmt"
	"strconv"
	"time"

	"github.com/goccy/go-json"
)

type Assignment struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description,omitempty"` // html format
	DueAt             time.Time `json:"due_at,omitempty"`
	LockAt            time.Time `json:"lock_at,omitempty"`
	PointsPossible    float64   `json:"points_possible,omitempty"`
	CourseID          int64     `json:"course_id,omitempty"`
	AssignmentGroupID int64     `json:"assignment_group_id"`
	SubmissionTypes   []string  `json:"submission_types,omitempty"`
	AllowedAttempts   int64     `json:"allowed_attempts,omitempty"`
}

func (c *Canvas) GetAssignments(params map[string]string) ([]Assignment, error) {

	endpoint := fmt.Sprintf("courses/%d/assignments", c.CourseID)

	res, err := c.get(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	var a []Assignment

	if err := json.Unmarshal(res, &a); err != nil {

		return nil, fmt.Errorf("canvas: Send: unmarshal %w", err)

	}

	return a, nil

}

func (c *Canvas) CreateChatAssignment(chapterNumber int) (*Assignment, error) {

	params := map[string]string{
		"assignment[name]":                fmt.Sprintf("chapter_%d: chat", chapterNumber),
		"assignment[submission_types][]":  "online_text_entry",
		"assignment[notify_of_update]":    "false",
		"assignment[points_possible]":     "1",
		"assignment[grading_type]":        "points",
		"assignment[assignment_group_id]": strconv.Itoa(int(c.AssignmentsGroupID)),
	}

	endpoint := fmt.Sprintf("courses/%d/assignments", c.CourseID)

	res, err := c.post(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	var a Assignment

	if err := json.Unmarshal(res, &a); err != nil {

		return nil, fmt.Errorf("canvas: Send: unmarshal %w", err)

	}

	return &a, nil

}
