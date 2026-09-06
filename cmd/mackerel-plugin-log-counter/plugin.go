package main

import (
	"fmt"
	"net/url"

	mp "github.com/mackerelio/go-mackerel-plugin"
	"github.com/mackerelio/golib/pluginutil"
	"github.com/monitoring-forge/followparser"
)

type LogCounterPlugin struct {
	Prefix        string
	LogFile       string
	LogArchiveDir string
	PerSec        bool
	Verbose       bool
	patternRegs   []*patternReg
	filterByte    *[]byte
	ignoreByte    *[]byte
}

func (u LogCounterPlugin) GraphDefinition() map[string]mp.Graphs {
	metrics := make([]mp.Metrics, 0, len(u.patternRegs))
	for _, pr := range u.patternRegs {
		metrics = append(metrics, mp.Metrics{
			Name:    pr.name,
			Label:   pr.name,
			Diff:    false,
			Stacked: false,
		})
	}
	comment := "(per minute)"
	if u.PerSec {
		comment = "(per second)"
	}
	return map[string]mp.Graphs{
		"": {
			Label:   fmt.Sprintf("LogCounter %s %s", u.Prefix, comment),
			Unit:    mp.UnitFloat,
			Metrics: metrics,
		},
	}
}

func (u LogCounterPlugin) FetchMetrics() (map[string]float64, error) {
	p := NewParser(
		u.patternRegs,
		ParserPerSec(u.PerSec),
		ParserFilter(u.filterByte),
		ParserIgnore(u.ignoreByte),
	)
	fp := &followparser.Parser{
		WorkDir:  pluginutil.PluginWorkDir(),
		Callback: p,
		Silent:   !u.Verbose,
	}
	if u.LogArchiveDir != "" {
		fp.ArchiveDir = u.LogArchiveDir
	}
	_, err := fp.Parse(
		fmt.Sprintf("%s-mp-log-counter-%s", u.Prefix, url.PathEscape(u.LogFile)),
		u.LogFile,
	)
	if err != nil {
		return nil, err
	}
	return p.GetResult(), nil
}

func (u LogCounterPlugin) MetricKeyPrefix() string {
	return u.Prefix
}
