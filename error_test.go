package fetch

import (
	"errors"
	"fmt"
	"testing"
)

func TestError_Unwrap(t *testing.T) {
	err := nonHttpErr("my message: ", errors.New("my error"))
	if err.Error() != "my message: my error" {
		t.Errorf("wrong error format")
	}
	if errors.Unwrap(err).Error() != "my error" {
		t.Errorf("wrong inner error")
	}
}

func TestError_Format(t *testing.T) {
	err := nonHttpErr("my message: ", errors.New("my error"))
	if fmt.Sprintf("%s", err) != "my message: my error" {
		t.Errorf("error failed, got: %s", err)
	}

	if fmt.Sprintf("%v", err) != "my message: my error" {
		t.Errorf("error failed, got: %s", err)
	}
}
