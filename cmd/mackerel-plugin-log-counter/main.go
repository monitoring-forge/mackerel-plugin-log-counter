package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/jessevdk/go-flags"
	mp "github.com/mackerelio/go-mackerel-plugin"
)

// version by Makefile
var version string
var commit string

const (
	OK = iota
	WARNING
	CRITICAL
	UNKNOWN
)

type patternReg struct {
	reg  *regexp.Regexp
	name string
	uniq bool
}

type Opt struct {
	Version       bool     `short:"v" long:"version" description:"Show version"`
	Filter        string   `long:"filter" description:"filter string used before check pattern."`
	Ignore        string   `long:"ignore" description:"ignore string used before check pattern."`
	Patterns      []string `short:"p" long:"pattern" required:"true" description:"Regexp pattern to search for."`
	KeyNames      []string `short:"k" long:"key-name" required:"true" description:"Key name for pattern. if key has '|uniq' suffix, this plugin count unique matches."`
	Prefix        string   `long:"prefix" required:"true" description:"Metric key prefix"`
	LogFile       string   `long:"log-file" default:"/var/log/messages" description:"Path to log file" required:"true"`
	LogArchiveDir string   `long:"log-archive-dir" default:"" description:"Path to log archive directory"`
	PerSec        bool     `long:"per-second" description:"calculate per-seconds count. default per minute count"`
	Verbose       bool     `long:"verbose" description:"display informational logs"`
	patternRegs   []*patternReg
	filterByte    *[]byte
	ignoreByte    *[]byte
}

func main() {
	os.Exit(_main())
}

func _main() int {
	opt := &Opt{}
	psr := flags.NewParser(opt, flags.HelpFlag|flags.PassDoubleDash)
	_, err := psr.Parse()
	if opt.Version {
		if commit == "" {
			commit = "dev"
		}
		fmt.Printf(
			"%s-%s\n%s/%s, %s, %s\n",
			filepath.Base(os.Args[0]),
			version,
			runtime.GOOS,
			runtime.GOARCH,
			runtime.Version(),
			commit)
		return OK
	} else if flags.WroteHelp(err) {
		fmt.Fprintf(os.Stdout, "%v\n", err)
		return OK
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return UNKNOWN
	}

	if len(opt.KeyNames) == 0 {
		fmt.Fprint(os.Stderr, "Specify --pattern and --key-name\n")
		return UNKNOWN
	}
	if len(opt.KeyNames) != len(opt.Patterns) {
		fmt.Fprint(os.Stderr, "The number of --pattern and --key-name must be the same\n")
		return UNKNOWN
	}

	patterns := make([]*patternReg, 0)
	for i, k := range opt.KeyNames {
		p, err := parseKeyName(opt.Patterns[i], k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return UNKNOWN
		}
		patterns = append(patterns, p)
	}
	opt.patternRegs = patterns

	if opt.Filter != "" {
		b := []byte(opt.Filter)
		opt.filterByte = &b
	}
	if opt.Ignore != "" {
		b := []byte(opt.Ignore)
		opt.ignoreByte = &b
	}

	u := LogCounterPlugin{
		opt: opt,
	}
	plugin := mp.NewMackerelPlugin(u)
	plugin.Run()
	return OK
}

func parseKeyName(pattern, keyName string) (*patternReg, error) {
	reg, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("pattern '%s' compile error. %w", pattern, err)
	}

	uniq := false

	fields := strings.FieldsFunc(keyName, func(r rune) bool {
		return r == '|'
	})
	if len(fields) == 2 && fields[1] == "uniq" {
		uniq = true
	} else if len(fields) >= 2 {
		return nil, fmt.Errorf("key name '%s' format error. must be <name> or <name>|uniq", keyName)
	}

	return &patternReg{
		reg:  reg,
		name: fields[0],
		uniq: uniq,
	}, nil
}
