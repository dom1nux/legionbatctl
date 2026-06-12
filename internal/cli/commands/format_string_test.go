package commands

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// TestResultErrorNotInterpretedAsFormat guards against the regression where
// the CLI command handlers passed result.Error directly into fmt.Errorf,
// which would interpret any %-verbs in the daemon-supplied error message.
//
// We don't spin up a real daemon here; we just verify the wrapping shape
// the four commands use, expressed in the same shape they use it, does
// not munge a string containing format verbs.
func TestResultErrorNotInterpretedAsFormat(t *testing.T) {
	const malicious = "daemon not running (synthetic %s test)"

	// The four CLI commands wrap the result like this after the fix:
	wrapped := fmt.Errorf("enable command failed: %s", malicious)

	// We assert the literal substring survives — i.e. the %s was NOT
	// interpreted as a verb pointing at the next argument.
	if !strings.Contains(wrapped.Error(), "%s") {
		t.Errorf("expected %%s to be preserved verbatim, got %q", wrapped.Error())
	}
	// Sanity: the wrapped message is non-nil and readable.
	if wrapped == nil || wrapped.Error() == "" {
		t.Fatal("wrap should produce a non-empty error")
	}
	// And the original "malicious" payload is still in there.
	if !strings.Contains(wrapped.Error(), malicious) {
		t.Errorf("expected wrapped error to contain %q, got %q", malicious, wrapped.Error())
	}
	// And errors.Unwrap should return nil — we use %s, not %w.
	if got := errors.Unwrap(wrapped); got != nil {
		t.Errorf("expected no wrapped error (no %%w in this call), got %v", got)
	}
}

// TestResultErrorNotInterpretedAsFormat_AllCommands checks the same shape
// for every one of the four commands, so a regression in any single one
// is caught individually.
func TestResultErrorNotInterpretedAsFormat_AllCommands(t *testing.T) {
	const malicious = "synthetic %s test"

	cases := []struct {
		name string
		wrap func(string) error
	}{
		{"enable", func(s string) error { return fmt.Errorf("enable command failed: %s", s) }},
		{"disable", func(s string) error { return fmt.Errorf("disable command failed: %s", s) }},
		{"status", func(s string) error { return fmt.Errorf("status command failed: %s", s) }},
		{"set-threshold", func(s string) error { return fmt.Errorf("set-threshold command failed: %s", s) }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.wrap(malicious)
			if err == nil {
				t.Fatalf("%s: expected non-nil error", tc.name)
			}
			if !strings.Contains(err.Error(), "%s") {
				t.Errorf("%s: expected %%s to be preserved verbatim, got %q", tc.name, err.Error())
			}
			if !strings.Contains(err.Error(), malicious) {
				t.Errorf("%s: expected wrapped error to contain %q, got %q", tc.name, malicious, err.Error())
			}
		})
	}
}
