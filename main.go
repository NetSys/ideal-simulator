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

// wait for all threads to finish
func shutdown(chans ...(chan int)) {
	for i := 0; i < len(chans); i++ {
		<-chans[i]
	}
}

func main() {
	args := os.Args[1:]
	conf := readConf(args[0])

	var eventQueue EventQueue
	var genQuit chan int
	switch {
	case conf.Read:
		fr := makeFlowReader(conf.TraceFileName)
		eventQueue = fr.makeFlows()
	case conf.GenerateOnly:
		fallthrough
	case conf.Generate:
		fg := makeFlowGenerator(conf.Load, conf.Bandwidth, conf.CDFFileName, conf.NumFlows)
		eventQueue, genQuit = fg.makeFlows()
	default:
		panic("Invalid configuration")
	}

	if conf.Generate || conf.Read {
		flows, idQuit := ideal(eventQueue, conf.Bandwidth)
		numFlows := len(flows)

		slowdown := 0.0
		for i := 0; i < numFlows; i++ {
			slowdown += calculateFlowSlowdown(flows[i])
		}

		if conf.Generate {
			shutdown(genQuit, idQuit)
		} else {
			shutdown(idQuit)
		}

		fmt.Println(slowdown / float64(numFlows))
	} else if conf.GenerateOnly {
		shutdown(genQuit)
	} else {
		panic("Unknown mode")
	}
}
