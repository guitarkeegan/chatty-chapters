package canvas

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func beforeTest() *Canvas {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error on env load: %q", err)
	}

	apiKey := os.Getenv("CANVAS_API")
	// the subdomain of your school's canvas instance, e.g. <subdomain>.instructure.com
	subdomain := School(os.Getenv("CANVAS_SUBDOMAIN"))

	courseID, err := strconv.ParseInt(os.Getenv("CANVAS_COURSE_ID"), 10, 64)
	if err != nil {
		log.Fatalf("CANVAS_COURSE_ID must be set to a number: %q", err)
	}

	canvas, err := New(apiKey, "music-tutor-app", subdomain, courseID)
	if err != nil {
		log.Fatal(err)
	}

	return canvas

}

//func TestListAssignments(t *testing.T) {
//
//	canvas := beforeTest()
//
//	ag := AssignmentGroup{
//		Name: "music-tutor-quiz-assignments",
//	}
//
//	err := canvas.Post("courses/:course_id/assignment_groups", map[string]string{"name": ag.Name}, &ag)
//	if err != nil {
//		t.Error(err)
//	}
//
//	var a []Assignment
//
//	err = canvas.Get(fmt.Sprintf("courses/:course_id/assignment_groups/%d/assignments", ag.ID), map[string]string{}, &a)
//	if err != nil {
//		t.Error(err)
//	}
//
//	t.Logf("ag: %+v", ag)
//	t.Logf("a: %+v", a)
//
//}
//
//func TestCreateAssignment(t *testing.T) {
//
//	canvas := beforeTest()
//
//	ag := AssignmentGroup{
//		Name: "music-tutor-quiz-assignments",
//	}
//
//	err := canvas.Post("courses/:course_id/assignment_groups", map[string]string{"name": ag.Name}, &ag)
//	if err != nil {
//		t.Error(err)
//	}
//
//	var a Assignment
//
//	err = canvas.Post(fmt.Sprintf("courses/:course_id/assignments"), map[string]string{"assignment[name]": "test-1", "assignment[submission_types][]": "online_text_entry", "assignment[points_possible]": "1", "assignment[allowed_attempts]": "3", "assignment[assignment_group_id]": strconv.Itoa(int(ag.ID)), "assignment[description]": "<h1>my description</h1>", "assignment[published]": "true"}, &a)
//
//	var newAssignment Assignment
//
//	err = canvas.Get(fmt.Sprintf("courses/:course_id/assignment_groups/%d/assignments", ag.ID), map[string]string{}, &newAssignment)
//	if err != nil {
//		t.Error(err)
//	}
//
//	t.Logf("ag: %+v", ag)
//	t.Logf("a: %+v", a)
//	t.Logf("newAssignment: %+v", newAssignment)
//
//}
//
//func TestCreateUser(t *testing.T) {
//
//	canvas := beforeTest()
//
//	var user User
//
//	err := canvas.Post("accounts/:accountID/users", map[string]string{
//		"user[name]":                   "Nageek Milgna",
//		"user[short_name]":             "Nageek",
//		"user[terms_of_use]":           "false",
//		"user[skip_registration]":      "true",
//		"pseudonym[send_confirmation]": "false",
//	}, &user)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	t.Logf("user: %+v", user)
//
//}
//
//func TestEnrollUser(t *testing.T) {
//
//	canvas := beforeTest()
//
//	var e Enrollment
//
//	err := canvas.Post("courses/:course_id/enrollments", map[string]string{
//		"enrollment[user_id]": "auserid83kjasd8f23ikjfs",
//		"enrollment[type]":    "StudentEnrollment",
//	}, e)
//
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	t.Logf("result: %+v", e)
//
//}
//
//func TestSubmitAssignment(t *testing.T) {
//
//	client := beforeTest()
//
//	var ag []AssignmentGroup
//
//	err := client.Get("courses/:course_id/assignment_groups", map[string]string{}, &ag)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	t.Log(ag)
//
//}
