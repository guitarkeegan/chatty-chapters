package canvas

import (
	"testing"
)

func TestGetAssignments(t *testing.T) {

	canvas := beforeTest()

	a, err := canvas.GetAssignments(map[string]string{})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("a: %+v", a)

}
