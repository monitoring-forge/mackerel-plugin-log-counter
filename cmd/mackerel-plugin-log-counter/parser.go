package main

import (
	"bytes"
)

type Parser struct {
	opt         *Opt
	mapCounter  map[string]float64
	uniqCounter map[string]map[string]struct{}
	duration    float64
}

func NewParser(opt *Opt) *Parser {
	m := map[string]float64{}
	uq := map[string]map[string]struct{}{}
	for _, pr := range opt.patternRegs {
		m[pr.name] = float64(0)
		if pr.uniq {
			uq[pr.name] = map[string]struct{}{}
		}
	}
	return &Parser{
		opt:         opt,
		mapCounter:  m,
		uniqCounter: uq,
	}
}

func (p *Parser) Parse(b []byte) error {
	if p.opt.filterByte != nil && !bytes.Contains(b, *p.opt.filterByte) {
		return nil
	}
	if p.opt.ignoreByte != nil && bytes.Contains(b, *p.opt.ignoreByte) {
		return nil
	}
	for _, pr := range p.opt.patternRegs {
		if pr.uniq {
			f := pr.reg.Find(b)
			if len(f) > 0 {
				p.uniqCounter[pr.name][string(f)] = struct{}{}
			}
		} else {
			if pr.reg.Match(b) {
				p.mapCounter[pr.name]++
			}
		}
	}
	return nil
}

func (p *Parser) Finish(duration float64) {
	p.duration = duration
}

func (p *Parser) GetResult() map[string]float64 {
	m := map[string]float64{}
	if p.duration == 0 {
		// first running
		return m
	}
	for _, pr := range p.opt.patternRegs {
		if pr.uniq {
			m[pr.name] = float64(len(p.uniqCounter[pr.name]))
		} else {
			m[pr.name] = p.mapCounter[pr.name]
		}
		if p.opt.PerSec {
			m[pr.name] = m[pr.name] / p.duration
		} else {
			m[pr.name] = (m[pr.name] / p.duration) * 60
		}
	}
	return m
}
