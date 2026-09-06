package main

import (
	"fmt"
	"net/url"

	mp "github.com/mackerelio/go-mackerel-plugin"
	"github.com/mackerelio/golib/pluginutil"
	"github.com/monitoring-forge/followparser"
)

func (opt *Opt) GraphDefinition() map[string]mp.Graphs {
	metrics := make([]mp.Metrics, 0, len(opt.patternRegs))
	for _, pr := range opt.patternRegs {
		metrics = append(metrics, mp.Metrics{
			Name:    pr.name,
			Label:   pr.name,
			Diff:    false,
			Stacked: false,
		})
	}
	comment := "(per minute)"
	if opt.PerSec {
		comment = "(per second)"
	}
	return map[string]mp.Graphs{
		"": {
			Label:   fmt.Sprintf("LogCounter %s %s", opt.Prefix, comment),
			Unit:    mp.UnitFloat,
			Metrics: metrics,
		},
	}
}

func (opt *Opt) FetchMetrics() (map[string]float64, error) {
	p := NewParser(
		opt.patternRegs,
		ParserPerSec(opt.PerSec),
		ParserFilter(opt.filterByte),
		ParserIgnore(opt.ignoreByte),
	)
	fp := &followparser.Parser{
		WorkDir:  pluginutil.PluginWorkDir(),
		Callback: p,
		Silent:   !opt.Verbose,
	}
	if opt.LogArchiveDir != "" {
		fp.ArchiveDir = opt.LogArchiveDir
	}
	_, err := fp.Parse(
		fmt.Sprintf("%s-mp-log-counter-%s", opt.Prefix, url.PathEscape(opt.LogFile)),
		opt.LogFile,
	)
	if err != nil {
		return nil, err
	}
	return p.GetResult(), nil
}

func (opt *Opt) MetricKeyPrefix() string {
	return opt.Prefix
}
