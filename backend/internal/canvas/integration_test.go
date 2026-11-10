//go:build integration

package canvas

import "testing"

// test will check assignment groups, create if not exist
// check assignments, create if not exists
func TestAssignmentAndGroupCreationFlow(t *testing.T) {

	canvas := beforeTest()

	// check if group exists

	ags, err := canvas.GetAssignmentGroups()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("found assignment groups: %+v\n", ags)

	// check all args for matching groupname
	var matchedAG AssignmentGroup
	for _, ag := range ags {
		if ag.Name == canvas.AssignmentGroupTitle && canvas.AssignmentGroupTitle != "" {
			matchedAG.ID = ag.ID
			matchedAG.Assignments = ag.Assignments
		}
	}

	t.Logf("matchedAG is now: %+v\n", matchedAG)

	// check if matchedAG was found
	if matchedAG.ID != 0 {
		t.Logf("matchedAG found ID: %d...\n", matchedAG.ID)
		// check assignments
		for _, a := range matchedAG.Assignments {
			// somewhat magic string to look for...
			if a.Name == "chapter_1: chat" {
				t.Log("found assignment!\n")
				return
			}
			t.Logf("a: %+v\n", a)
		}
		// assignment was not found
		// create the assignment
		// attempt to create chapter_1
		ch, err := canvas.CreateChatAssignment(1)
		if err != nil {
			t.Fatal("could not create chapter")
		}

		t.Logf("found ch: %+v", ch)
	}

	// group was not found...
	// create group
	t.Log("no group found...\n")
	ag, err := canvas.CreateAssignmentGroup()
	if err != nil {
		t.Fatal(err)
	}

	err = canvas.SetAssignmentGroupID(int64(ag.ID))
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("created assignment group: %+v\n", ag)

	ch, err := canvas.CreateChatAssignment(1)
	if err != nil {
		t.Fatal("could not create chapter")
	}

	t.Logf("created assignment: %+v\n", ch)

}
