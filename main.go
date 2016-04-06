package main

// read in input
// initialize objects

import (
	"fmt"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	args := os.Args[1:]
	conf := readConf(args[0])

	var logger = make(chan LogEvent, 1000)
	loggingDone := make(chan bool)
	go log(logger, loggingDone)

	var eventQueue EventQueue
	switch {
	case conf.Read:
		fr := makeFlowReader(conf.TraceFileName)
		eventQueue = fr.makeFlows(logger)
	case conf.GenerateOnly:
		fg := makeFlowGenerator(conf.Load, conf.Bandwidth, conf.CDFFileName, conf.NumFlows)
		eventQueue = fg.makeFlows(logger)
		close(logger)
		<-loggingDone
		return
	case conf.Generate:
		fg := makeFlowGenerator(conf.Load, conf.Bandwidth, conf.CDFFileName, conf.NumFlows)
		eventQueue = fg.makeFlows(logger)
	default:
		panic("Invalid configuration")
	}

	flows := ideal(conf.Bandwidth, logger, eventQueue)

	close(logger)
	<-loggingDone

	numFlows := len(flows)

	slowdown := 0.0
	for i := 0; i < numFlows; i++ {
		slowdown += calculateFlowSlowdown(flows[i])
	}
	fmt.Println(slowdown / float64(numFlows))
}
