package main

import (
	"testing"
)

func TestValidate_NoKeyNames(t *testing.T) {
	ops := &Opt{
		Patterns: []string{`foo`},
		KeyNames: []string{},
	}
	err := ops.Validate(nil)
	if err == nil {
		t.Error("expected error for missing key names, got nil")
	}
}

func TestValidate_MismatchedPatternsAndKeyNames(t *testing.T) {
	ops := &Opt{
		Patterns: []string{`foo`, `bar`},
		KeyNames: []string{`foo`},
	}
	err := ops.Validate(nil)
	if err == nil {
		t.Error("expected error for mismatched patterns and key names, got nil")
	}
}

func TestValidate_Valid(t *testing.T) {
	ops := &Opt{
		Patterns: []string{`foo`},
		KeyNames: []string{`foo`},
	}
	err := ops.Validate(nil)
	if err != nil {
		t.Errorf("unexpected error for valid input: %v", err)
	}
}
