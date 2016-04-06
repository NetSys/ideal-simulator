package main

import (
	"fmt"
)

type LogEventType int

const (
	LOG_FLOW_ARRIVAL LogEventType = iota
	LOG_FLOW_GEN
	LOG_FLOW_FINISHED
	OTHER
)

type LogEvent struct {
	Time float64
	Type LogEventType
	Flow *Flow
	msg  string
}

func log(lgs chan LogEvent, done chan bool) {
	backlog := uint(0)
	for l := range lgs {
		f := l.Flow
		t := l.Time
		switch l.Type {
		case LOG_FLOW_GEN:
			fmt.Printf("generated: %d %d %d %f\n", f.Size, f.Source, f.Dest, f.Start)
		case LOG_FLOW_ARRIVAL:
			backlog += f.Size
			fmt.Printf("backlog %.6f %d : starting %d %d %d\n", t, backlog, f.Source, f.Dest, f.Size)
		case LOG_FLOW_FINISHED:
			backlog -= f.Size
			fmt.Print(flowToString(f))
		}
	}
	done <- false
}
