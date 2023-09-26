package usage

import (
	"testing"
)

func TestTopFrameworks(t *testing.T) {
	frameworks, err := TopFrameworks()
	if err != nil {
		t.Error(err)
	}
	if len(frameworks) == 0 {
		t.Error("no frameworks found")
	}
}

