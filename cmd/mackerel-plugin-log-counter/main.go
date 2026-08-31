package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	mp "github.com/mackerelio/go-mackerel-plugin"
	"github.com/monitoring-forge/flagrun"
)

var version string

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

func (o *Opt) Run(_ []string) (string, int) {
	if len(o.KeyNames) == 0 {
		fmt.Fprint(os.Stderr, "Specify --pattern and --key-name\n")
		return "", flagrun.UNKNOWN
	}
	if len(o.KeyNames) != len(o.Patterns) {
		fmt.Fprint(os.Stderr, "The number of --pattern and --key-name must be the same\n")
		return "", flagrun.UNKNOWN
	}

	patterns := make([]*patternReg, 0)
	for i, k := range o.KeyNames {
		p, err := parseKeyName(o.Patterns[i], k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return "", flagrun.UNKNOWN
		}
		patterns = append(patterns, p)
	}
	o.patternRegs = patterns

	if o.Filter != "" {
		b := []byte(o.Filter)
		o.filterByte = &b
	}
	if o.Ignore != "" {
		b := []byte(o.Ignore)
		o.ignoreByte = &b
	}

	u := LogCounterPlugin{
		opt: o,
	}
	plugin := mp.NewMackerelPlugin(u)
	plugin.Run()
	return "", flagrun.OK

}

func main() {
	os.Exit(flagrun.Go(&Opt{}, flagrun.Version(version)))
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
