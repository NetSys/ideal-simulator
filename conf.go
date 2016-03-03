package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// read in conf file

type Conf struct {
	Bandwidth float64

	Read          bool
	TraceFileName string

	GenerateOnly bool

	Generate    bool
	Load        float64
	NumFlows    uint
	CDFFileName string
}

func (c Conf) assert_generation() {
	if !c.Generate {
		panic("Conflicting options")
	}
}

func (c Conf) assert_readtrace() {
	if !c.Read {
		panic("Conflicting options")
	}
}

func readConf(fn string) Conf {
	file, ok := os.Open(fn)
	check(ok)
	defer file.Close()

	c := Conf{}
	defer func() {
		if c.Bandwidth == 0 {
			panic("Invalid configuration")
		} else if c.Read == c.Generate {
			panic("Conflicting options")
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		l := strings.Split(scanner.Text(), " ")
		switch {
		case l[0][0] == '#':
			continue
		case l[0] == "Bandwidth":
			b, ok := strconv.ParseFloat(l[1], 64)
			check(ok)
			c.Bandwidth = b

		case l[0] == "Read":
			c.Read = true
			defer func() {
				if c.TraceFileName == "" {
					panic("Invalid configuration")
				}
			}()
		case l[0] == "TraceFile":
			c.TraceFileName = l[1]
			defer c.assert_readtrace()

		case l[0] == "GenerateOnly":
			c.GenerateOnly = true
			fallthrough
		case l[0] == "Generate":
			c.Generate = true
			defer func() {
				if c.CDFFileName == "" || c.Load == 0 || c.NumFlows == 0 {
					panic("Invalid configuration")
				}
			}()
		case l[0] == "Load":
			l, ok := strconv.ParseFloat(l[1], 64)
			check(ok)
			c.Load = l
			defer c.assert_generation()
		case l[0] == "NumFlows":
			n, ok := strconv.ParseUint(l[1], 10, 32)
			check(ok)
			c.NumFlows = uint(n)
			defer c.assert_generation()
		case l[0] == "CDF":
			c.CDFFileName = l[1]
			defer c.assert_generation()
		}
	}

	return c
}
