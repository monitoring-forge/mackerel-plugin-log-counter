package main

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_Parse(t *testing.T) {
	patterns := []*patternReg{
		{name: "pattern1", reg: regexp.MustCompile(`error`)},
		{name: "pattern2", reg: regexp.MustCompile(`warning`)},
	}

	tests := []struct {
		input    []byte
		expected map[string]float64
	}{
		{
			input:    []byte("error occurred"),
			expected: map[string]float64{"pattern1": 1, "pattern2": 0},
		},
		{
			input:    []byte("warning issued"),
			expected: map[string]float64{"pattern1": 0, "pattern2": 1},
		},
		{
			input:    []byte("no match here"),
			expected: map[string]float64{"pattern1": 0, "pattern2": 0},
		},
	}

	for _, test := range tests {
		parser := NewParser(patterns)
		err := parser.Parse(test.input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		for key, value := range test.expected {
			if parser.mapCounter[key] != value {
				t.Errorf("expected %v for key %v, got %v", value, key, parser.mapCounter[key])
			}
		}
	}
}

func TestParser_FilterIgnore(t *testing.T) {
	filter := []byte("filter")
	ignore := []byte("ignore")
	patterns := []*patternReg{
		{name: "pattern1", reg: regexp.MustCompile(`error`)},
	}

	tests := []struct {
		input    []byte
		expected map[string]float64
	}{
		{
			input:    []byte("filter error occurred"),
			expected: map[string]float64{"pattern1": 1},
		},
		{
			input:    []byte("ignore error occurred"),
			expected: map[string]float64{"pattern1": 0},
		},
		{
			input:    []byte("no match here"),
			expected: map[string]float64{"pattern1": 0},
		},
	}

	for _, test := range tests {
		parser := NewParser(patterns, ParserFilter(&filter), ParserIgnore(&ignore))

		err := parser.Parse(test.input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		for key, value := range test.expected {
			if parser.mapCounter[key] != value {
				t.Errorf("expected %v for key %v, got %v", value, key, parser.mapCounter[key])
			}
		}
	}
}

func TestParser_GetResult(t *testing.T) {
	patterns := []*patternReg{
		{name: "pattern1", reg: regexp.MustCompile(`error`)},
	}
	parser := NewParser(patterns, ParserPerSec(true))

	err := parser.Parse([]byte("error occurred"))
	require.NoError(t, err, "Parse should not return an error")
	parser.Finish(10) // 10 seconds

	result := parser.GetResult()
	expected := map[string]float64{"pattern1": 0.1}

	for key, value := range expected {
		assert.Equal(t, value, result[key], "expected %v for key %v, got %v", value, key, result[key])
	}
}
