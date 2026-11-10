package canvas

import (
	"fmt"

	"github.com/goccy/go-json"
)

type AssignmentGroup struct {
	ID              int               `json:"id"`
	Name            string            `json:"name"`
	Position        int               `json:"position"`
	GroupWeight     float64           `json:"group_weight"`
	SISSourceID     *string           `json:"sis_source_id,omitempty"`
	IntegrationData map[string]string `json:"integration_data,omitempty"`
	Assignments     []Assignment      `json:"assignments,omitempty"`
	Rules           json.RawMessage   `json:"rules,omitempty"`
}

func (c *Canvas) GetAssignmentGroup() (*AssignmentGroup, error) {

	params := map[string]string{}

	endpoint := fmt.Sprintf("courses/%d/assignment_groups/%d", c.CourseID, c.AssignmentsGroupID)

	res, err := c.get(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	var ag AssignmentGroup

	if err := json.Unmarshal(res, &ag); err != nil {

		return nil, fmt.Errorf("canvas: Send: unmarshal %w", err)

	}

	return &ag, nil

}

func (c *Canvas) GetAssignmentGroups() ([]AssignmentGroup, error) {

	params := map[string]string{
		"include[]": "assignments",
	}

	endpoint := fmt.Sprintf("courses/%d/assignment_groups", c.CourseID)

	res, err := c.get(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	var ag []AssignmentGroup

	if err := json.Unmarshal(res, &ag); err != nil {

		return nil, fmt.Errorf("canvas: Send: unmarshal %w", err)

	}

	return ag, nil

}

func (c *Canvas) CreateAssignmentGroup() (*AssignmentGroup, error) {

	params := map[string]string{
		"name": c.AssignmentGroupTitle,
	}

	endpoint := fmt.Sprintf("courses/%d/assignment_groups", c.CourseID)

	res, err := c.post(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("post: %w", err)
	}

	var ag AssignmentGroup

	if err := json.Unmarshal(res, &ag); err != nil {

		return nil, fmt.Errorf("canvas: Send: unmarshal %w", err)

	}

	return &ag, nil

}
