package task

import "testing"

func TestValidStateTransition(t *testing.T) {
	if !ValidStateTransition(Pending, Scheduled) {
		t.Fatal("expected true, got false")
	}
	if ValidStateTransition(Scheduled, Pending) {
		t.Fatal("expected false, got true")
	}
	if !ValidStateTransition(Running, Failed) {
		t.Fatal("expected true, got false")
	}
	if ValidStateTransition(Completed, Scheduled) {
		t.Fatal("expected false, got true")
	}
}
