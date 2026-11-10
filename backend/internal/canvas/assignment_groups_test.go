package canvas

import "testing"

func TestListAssignmentGroups(t *testing.T) {

	canvas := beforeTest()

	ag, err := canvas.GetAssignmentGroups()
	if err != nil {

		t.Fatal(err)
	}

	t.Logf("ag: %+v", ag)

}
