package cmdrunner

import "testing"

func TestExecSuccess(t *testing.T) {
	out, err := Exec("echo", []string{"hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "hello\n" {
		t.Errorf("got %q", out)
	}
}

func TestExecFailure(t *testing.T) {
	if _, err := Exec("this-command-should-not-exist-xyz", nil); err == nil {
		t.Fatal("expected error for missing command")
	}
}
