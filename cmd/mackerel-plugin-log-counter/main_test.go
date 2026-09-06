package main

import (
	"regexp"
	"testing"
)

func TestParseKeyName_Normal(t *testing.T) {
	p, err := parseKeyName(`foo.*`, "foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.name != "foo" {
		t.Errorf("expected name 'foo', got '%s'", p.name)
	}
	if p.uniq {
		t.Errorf("expected uniq false, got true")
	}
	if p.reg.String() != regexp.MustCompile(`foo.*`).String() {
		t.Errorf("unexpected regexp: %v", p.reg)
	}
}

func TestParseKeyName_Uniq(t *testing.T) {
	p, err := parseKeyName(`bar`, "bar|uniq")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.name != "bar" {
		t.Errorf("expected name 'bar', got '%s'", p.name)
	}
	if !p.uniq {
		t.Errorf("expected uniq true, got false")
	}
}

func TestParseKeyName_InvalidKeyName(t *testing.T) {
	_, err := parseKeyName(`baz`, "baz|uniq|extra")
	if err == nil {
		t.Error("expected error for invalid keyName format, got nil")
	}
}

func TestParseKeyName_InvalidRegexp(t *testing.T) {
	_, err := parseKeyName(`*invalid`, "foo")
	if err == nil {
		t.Error("expected error for invalid regexp, got nil")
	}
}

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
