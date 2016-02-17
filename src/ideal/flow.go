package main

// define data types

import (
	"fmt"
	"strconv"
	"strings"
)

type SortedFlows []*Flow

func (sf SortedFlows) Len() int {
	return len(sf)
}

func (sf SortedFlows) Less(i, j int) bool {
	return (sf[i].TimeRemaining < sf[j].TimeRemaining)
}

func (sf SortedFlows) Swap(i, j int) {
	tmp := sf[i]
	sf[i] = sf[j]
	sf[j] = tmp
}

type Flow struct {
	Start         float64
	Size          uint32
	Source        uint8
	Dest          uint8
	End           float64
	TimeRemaining float64
	OracleFct     float64
	LastTime      float64
	FinishEvent   *Event
	PropDelay     float64
	FinishSending bool
	Finish        bool
}

func flowToString(f *Flow) {
	fmt.Printf("%d %d %d %f %f %f\n", f.Source, f.Dest, f.Size, f.Start, f.OracleFct, f.End)
}

func calculateFlowSlowdown(f *Flow) float64 {
	if f.End < f.Start {
		panic("flow has negative fct")
	}

	fct := f.End - f.Start
	slowdown := fct / f.OracleFct
	switch {
	case slowdown >= 1:
		return slowdown
	case slowdown < 0.999:
		flowToString(f)
		fmt.Println(slowdown)
		panic("flow has fct better than oracle")
	default:
		return 1.000000
	}
}

func makeFlow(l string) *Flow {
	sp := strings.Split(l, " ")

	size, err := strconv.ParseUint(sp[1], 10, 32)
	check(err)
	src, err := strconv.ParseUint(sp[2], 10, 8)
	check(err)
	dst, err := strconv.ParseUint(sp[3], 10, 8)
	check(err)
	time, err := strconv.ParseFloat(sp[4], 64)
	check(err)

	return &Flow{Start: time, Size: uint32(size), Source: uint8(src), Dest: uint8(dst), LastTime: 0, FinishEvent: nil}
}

type EventType int

const (
	FlowArrival EventType = iota
	FlowSourceFree
	FlowDestFree
)

type Event struct {
	Time      float64
	Flow      *Flow
	Type      EventType
	Cancelled bool
}

func makeArrivalEvent(f *Flow) *Event {
	return &Event{Time: f.Start, Flow: f, Type: FlowArrival, Cancelled: false}
}

func makeCompletionEvent(t float64, f *Flow, ty EventType) *Event {
	return &Event{Time: t, Flow: f, Type: ty, Cancelled: false}
}

type EventQueue []*Event

func (e EventQueue) Len() int {
	return len(e)
}

func (e EventQueue) Less(i, j int) bool {
	return (e[i].Time < e[j].Time)
}

func (e EventQueue) Swap(i, j int) {
	tmp := e[i]
	e[i] = e[j]
	e[j] = tmp
}

func (e *EventQueue) Push(x interface{}) {
	ev := x.(*Event)
	*e = append(*e, ev)
}

func (e *EventQueue) Pop() interface{} {
	old := *e
	n := len(old)
	ev := old[n-1]
	*e = old[0 : n-1]
	return ev
}
