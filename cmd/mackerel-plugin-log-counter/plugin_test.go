package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogCounterPlugin_FetchMetrics(t *testing.T) {
	tmpdir := t.TempDir()
	logFileName := filepath.Join(tmpdir, "log")
	fh, err := os.Create(logFileName)
	if err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	patterns := []*patternReg{
		{name: "pattern1", reg: regexp.MustCompile(`error`)},
		{name: "pattern2", reg: regexp.MustCompile(`warning`)},
	}
	{
		opt := &Opt{
			Prefix:      "TestLogCounterPlugin_FetchMetrics",
			patternRegs: patterns,
			LogFile:     logFileName,
			PerSec:      true,
		}
		plugin := LogCounterPlugin{opt: opt}

		_, err := plugin.FetchMetrics()
		assert.NoError(t, err)
	}

	for i := range 10 {
		fmt.Fprintf(fh, "warning msg %08d\n", i)
	}
	for i := range 10 {
		fmt.Fprintf(fh, "error msg %08d\n", i)
	}

	time.Sleep(time.Second)

	{
		opt := &Opt{
			Prefix:      "TestLogCounterPlugin_FetchMetrics",
			patternRegs: patterns,
			LogFile:     logFileName,
			PerSec:      true,
		}
		plugin := LogCounterPlugin{opt: opt}

		m, err := plugin.FetchMetrics()
		assert.NoError(t, err)
		assert.Equal(t, m, map[string]float64{
			"pattern1": 10,
			"pattern2": 10,
		}, "match metrics")
	}

	for i := range 5 {
		fmt.Fprintf(fh, "warning msg %08d\n", i)
	}
	for i := range 5 {
		fmt.Fprintf(fh, "error msg %08d\n", i)
	}
	time.Sleep(time.Second)
	{
		opt := &Opt{
			Prefix:      "TestLogCounterPlugin_FetchMetrics",
			patternRegs: patterns,
			LogFile:     logFileName,
			PerSec:      false,
		}
		plugin := LogCounterPlugin{opt: opt}

		m, err := plugin.FetchMetrics()
		assert.NoError(t, err)
		assert.Equal(t, m, map[string]float64{
			"pattern1": 300,
			"pattern2": 300,
		}, "match metrics")
	}

}

func TestLogCounterPlugin_FetchMetrics_Uniq(t *testing.T) {
	tmpdir := t.TempDir()
	logFileName := filepath.Join(tmpdir, "log")
	fh, err := os.Create(logFileName)
	if err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	// uniq=true のパターンを用意
	patterns := []*patternReg{
		{name: "pattern1", reg: regexp.MustCompile(`error`), uniq: true},
		{name: "pattern2", reg: regexp.MustCompile(`error: \w+`), uniq: true},
	}

	// initialize parser to create tracking files
	{
		opt := &Opt{
			Prefix:      "TestLogCounterPlugin_FetchMetrics_Uniq",
			patternRegs: patterns,
			LogFile:     logFileName,
			PerSec:      true,
		}
		plugin := LogCounterPlugin{opt: opt}

		_, err := plugin.FetchMetrics()
		assert.NoError(t, err)
	}

	keys := []string{"aaaa", "bbbb", "cccc", "dddd", "eeee"}
	for _, key := range keys {
		for range 5 {
			fmt.Fprintf(fh, "error: %s something failed\n", key)
		}
	}

	err = fh.Sync()
	require.NoError(t, err, "failed to sync log file")

	// allow followparser to measure a non-zero duration
	time.Sleep(time.Second)

	opt := &Opt{
		Prefix:      "TestLogCounterPlugin_FetchMetrics_Uniq",
		patternRegs: patterns,
		LogFile:     logFileName,
		PerSec:      true,
	}
	plugin := LogCounterPlugin{opt: opt}

	m, err := plugin.FetchMetrics()
	assert.NoError(t, err)
	// uniq=true なので重複行は1件のみカウントされることを期待
	assert.Equal(t, m, map[string]float64{
		"pattern1": 1,
		"pattern2": 5,
	}, "match metrics")
}

func TestLogCounterPlugin_RotateAndArchive(t *testing.T) {
	tmpdir := t.TempDir()
	archiveDir := filepath.Join(tmpdir, "archive")
	if err := os.Mkdir(archiveDir, 0755); err != nil {
		t.Fatalf("failed to create archive dir: %v", err)
	}

	logFileName := filepath.Join(tmpdir, "log")
	fh, err := os.Create(logFileName)
	if err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	patterns := []*patternReg{
		{name: "pattern1", reg: regexp.MustCompile(`error`)},
		{name: "pattern2", reg: regexp.MustCompile(`warning`)},
	}

	// initial run to create tracking files
	{
		opt := &Opt{
			Prefix:        "TestLogCounterPlugin_RotateAndArchive",
			patternRegs:   patterns,
			LogFile:       logFileName,
			LogArchiveDir: archiveDir,
			PerSec:        true,
		}
		plugin := LogCounterPlugin{opt: opt}

		_, err := plugin.FetchMetrics()
		assert.NoError(t, err)
	}

	beforeSleep := time.Now()

	// write some lines to the original log file, then rotate (move to archive)
	for i := range 10 {
		fmt.Fprintf(fh, "warning msg %08d\n", i)
	}
	for i := range 10 {
		fmt.Fprintf(fh, "error msg %08d\n", i)
	}
	err = fh.Sync()
	require.NoError(t, err, "failed to sync log file")

	// rotate: move the file to archive
	if err := fh.Close(); err != nil {
		t.Fatalf("failed to close log file: %v", err)
	}
	archived := filepath.Join(archiveDir, "log.1")
	if err := os.Rename(logFileName, archived); err != nil {
		t.Fatalf("failed to rotate log file: %v", err)
	}

	// create a new log file with the same name and write additional lines
	fh2, err := os.Create(logFileName)
	if err != nil {
		t.Fatalf("failed to create new log file: %v", err)
	}
	for i := range 5 {
		fmt.Fprintf(fh2, "warning msg %08d\n", i)
	}
	for i := range 5 {
		fmt.Fprintf(fh2, "error msg %08d\n", i)
	}
	err = fh2.Sync()
	require.NoError(t, err, "failed to sync new log file")

	// beforeSleepから1秒変わるまでsleep
	time.Sleep(time.Until(beforeSleep.Add(time.Second)))

	{
		opt := &Opt{
			Prefix:        "TestLogCounterPlugin_RotateAndArchive",
			patternRegs:   patterns,
			LogFile:       logFileName,
			LogArchiveDir: archiveDir,
			PerSec:        true,
		}
		plugin := LogCounterPlugin{opt: opt}

		m, err := plugin.FetchMetrics()
		assert.NoError(t, err)
		// expect counts from both archived (10 each) and new file (5 each)
		assert.Equal(t, map[string]float64{
			"pattern1": 15,
			"pattern2": 15,
		}, m, "counts should include archived and new entries")
	}
}
