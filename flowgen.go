package main

import (
	"bufio"
	"container/heap"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

// generate flows

// read input flow trace

type FlowReader struct {
	file string
}

func makeFlowReader(fn string) FlowReader {
	return FlowReader{file: fn}
}

func (fr FlowReader) makeFlows(logger chan LogEvent) EventQueue {
	eventQueue := make(EventQueue, 0)
	file, err := os.Open(fr.file)
	check(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		flow := makeFlow(scanner.Text())
		eventQueue = append(eventQueue, makeArrivalEvent(flow))
	}

	return eventQueue
}

// read flow size CDF file and load for Poisson arrival

type CDF struct {
	values []uint
	distrs []float64
}

func readCDF(fn string) CDF {
	file, err := os.Open(fn)
	check(err)
	defer file.Close()

	vals := make([]uint, 16)
	cdfs := make([]float64, 16)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		sp := strings.Split(scanner.Text(), " ")
		// structure of line: <value> <blah> <cdf>
		v, ok := strconv.ParseUint(sp[0], 10, 32)
		check(ok)
		c, ok := strconv.ParseFloat(sp[2], 64)
		check(ok)

		if v == 0 {
			panic("invalid cdf flow size")
		}

		vals = append(vals, uint(v)*1460)
		cdfs = append(cdfs, c)
	}

	return CDF{values: vals, distrs: cdfs}
}

func (cdf CDF) meanFlowSize() float64 {
	avg := 0.0
	lastCdf := 0.0
	for i := 0; i < len(cdf.values); i++ {
		avg += float64(cdf.values[i]) * (cdf.distrs[i] - lastCdf)
		lastCdf = cdf.distrs[i]
	}
	return avg
}

func (cdf CDF) value() uint {
	rand := rand.Float64()
	for i := 0; i < len(cdf.values); i++ {
		if cdf.distrs[i] >= rand {
			if cdf.values[i] == 0 {
				panic("invalid flow size")
			}
			return cdf.values[i]
		}
	}

	panic("reached end of cdf function without value")
}

type FlowGenerator struct {
	load      float64
	bandwidth float64
	cdf       CDF
	numFlows  uint
}

func makeFlowGenerator(load float64, bw float64, cdfFile string, nf uint) FlowGenerator {
	return FlowGenerator{load: load, bandwidth: bw, cdf: readCDF(cdfFile), numFlows: nf}
}

func makeCreationEvent(f *Flow) *Event {
	return &Event{Time: f.Start, Flow: f, Type: FlowArrival, Cancelled: false}
}

func (fg FlowGenerator) makeFlows(logger chan LogEvent) EventQueue {
	lambda := (fg.bandwidth * 1e9 * fg.load) / (fg.cdf.meanFlowSize() * 1500 * 8)
	lambda /= 143

	creationQueue := make(EventQueue, 0)
	defer func() {
		creationQueue = nil
	}()

	heap.Init(&creationQueue)
	for i := 0; i < NUM_HOSTS; i++ {
		for j := 0; j < NUM_HOSTS; j++ {
			if i == j {
				continue
			}
			f := &Flow{Start: 1e6 + (rand.ExpFloat64()/lambda)*1e6, Size: fg.cdf.value(), Source: uint8(i), Dest: uint8(j), LastTime: 0, FinishEvent: nil}
			heap.Push(&creationQueue, makeCreationEvent(f))
		}
	}

	eventQueue := make(EventQueue, 0)
	for uint(len(eventQueue)) < fg.numFlows {
		ev := heap.Pop(&creationQueue).(*Event)
		logger <- LogEvent{Time: 0, Type: LOG_FLOW_GEN, Flow: ev.Flow}
		eventQueue = append(eventQueue, makeArrivalEvent(ev.Flow))

		nextTime := ev.Time + (rand.ExpFloat64()/lambda)*1e6
		f := &Flow{Start: nextTime, Size: fg.cdf.value(), Source: ev.Flow.Source, Dest: ev.Flow.Dest, LastTime: 0, FinishEvent: nil}
		heap.Push(&creationQueue, makeCreationEvent(f))
	}

	return eventQueue
}
