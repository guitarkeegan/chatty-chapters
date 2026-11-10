package canvas

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var dbg = func() func(format string, as ...any) {
	if os.Getenv("DEBUG") == "" {
		return func(string, ...any) {}
	}
	return func(format string, as ...any) {
		log.Printf(format, as...)
	}
}()

var (
	ErrEmptyStrings = errors.New("error empty strings")
	ErrNoCommaStr   = "expected comma after object element"
	baseProtocol    = "https://"
	baseEndpoint    = ".instructure.com/api/v1/"
)

type Canvas struct {
	Client               *http.Client
	APIKey               string
	School               School
	CourseID             int64
	AssignmentsGroupID   int64
	AssignmentGroupTitle string
	baseProtocol         string
	baseEndpoint         string
}

type School string

type User struct {
	Name             string `json:"name"`
	ShortName        string `json:"short_name"`
	TermsOfUse       bool   `json:"terms_of_use"`
	SkipRegistration bool   `json:"skip_registration"`
	Unique           string `json:"unique_id"`
	SendConfirmation bool   `json:"send_confirmation"`
}

type Enrollment struct {
	UserID          string `json:"user_id"`
	Type            string `json:"type"`
	EnrollmentState string `json:"enrollment_state"`
	CourseSectionID string `json:"course_section_id"`
}

func New(apiKey, groupTitle string, school School, courseID int64) (*Canvas, error) {

	if apiKey == "" || school == "" || courseID <= 0 {
		return nil, ErrEmptyStrings
	}

	client := &http.Client{
		Timeout: time.Second * 5,
	}

	return &Canvas{
		Client:               client,
		APIKey:               apiKey,
		School:               school,
		CourseID:             courseID,
		AssignmentGroupTitle: groupTitle,
		baseProtocol:         baseProtocol,
		baseEndpoint:         baseEndpoint,
	}, nil
}

func (c *Canvas) SetSchool(school School) error {

	if school == "" {
		return ErrEmptyStrings
	}
	c.School = school

	return nil

}

// make API key type?
func (c *Canvas) SetApiKey(apiKey string) error {

	if apiKey == "" {
		return ErrEmptyStrings
	}
	c.APIKey = apiKey

	return nil

}

func (c *Canvas) SetAssignmentGroupID(ag int64) error {

	if ag <= 0 {
		return fmt.Errorf("canvas: SetAssignmentGroupID: param must be greater than 0: ag: %d", ag)
	}
	c.AssignmentsGroupID = ag

	return nil

}

// endpoint can be "" in this case
func (c *Canvas) get(endpoint string, params map[string]string) ([]byte, error) {

	reqURL := fmt.Sprintf("%s%s%s%s", c.baseProtocol, c.School, c.baseEndpoint, endpoint)
	parsedURL, err := url.Parse(reqURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	q := parsedURL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	parsedURL.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("canvas get: %w", err)
	}

	// Add the Authorization header to the request
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Canvas get: client do: %w", err)
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Canvas ReadAll: %w", err)
	}

	return bytes, nil

}

func (c *Canvas) post(endpoint string, params map[string]string) ([]byte, error) {

	reqURL := fmt.Sprintf("%s%s%s%s", c.baseProtocol, c.School, c.baseEndpoint, endpoint)
	parsedURL, err := url.Parse(reqURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	q := parsedURL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	parsedURL.RawQuery = q.Encode()

	dbg("parsedURL", parsedURL.String())

	req, err := http.NewRequest("POST", parsedURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("Canvas post: %w", err)
	}

	// Add the Authorization header to the request
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	// Add Content-Type header if necessary
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Canvas post: client do: %w", err)
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Canvas ReadAll: %w", err)
	}

	return bytes, nil

}

// submit the assignment for the student
// give score and feedback
// find a place to show instructor chat sessions feedback
// maybe through another assignment group?
