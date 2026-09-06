package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

type ParserOption func(*Parser)

type patternReg struct {
	reg  *regexp.Regexp
	name string
	uniq bool
}

type Parser struct {
	PerSec      bool
	patternRegs []*patternReg
	filterByte  *[]byte
	ignoreByte  *[]byte
	mapCounter  map[string]float64
	uniqCounter map[string]map[string]struct{}
	duration    float64
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

func NewParser(patternRegs []*patternReg, opts ...ParserOption) *Parser {
	m := map[string]float64{}
	uq := map[string]map[string]struct{}{}
	for _, pr := range patternRegs {
		m[pr.name] = float64(0)
		if pr.uniq {
			uq[pr.name] = map[string]struct{}{}
		}
	}
	p := &Parser{
		patternRegs: patternRegs,
		mapCounter:  m,
		uniqCounter: uq,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func ParserPerSec(perSec bool) ParserOption {
	return func(p *Parser) {
		p.PerSec = perSec
	}
}

func ParserFilter(filter *[]byte) ParserOption {
	return func(p *Parser) {
		p.filterByte = filter
	}
}

func ParserIgnore(ignore *[]byte) ParserOption {
	return func(p *Parser) {
		p.ignoreByte = ignore
	}
}

func (p *Parser) Parse(b []byte) error {
	if p.filterByte != nil && !bytes.Contains(b, *p.filterByte) {
		return nil
	}
	if p.ignoreByte != nil && bytes.Contains(b, *p.ignoreByte) {
		return nil
	}
	for _, pr := range p.patternRegs {
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
	for _, pr := range p.patternRegs {
		if pr.uniq {
			m[pr.name] = float64(len(p.uniqCounter[pr.name]))
		} else {
			m[pr.name] = p.mapCounter[pr.name]
		}
		if p.PerSec {
			m[pr.name] = m[pr.name] / p.duration
		} else {
			m[pr.name] = (m[pr.name] / p.duration) * 60
		}
	}
	return m
}
