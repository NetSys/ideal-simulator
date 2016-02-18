package main

// read in input
// initialize objects

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	args := os.Args[1:]
	fn := args[0]
	bw, err := strconv.ParseFloat(args[1], 64)
	check(err)

	fmt.Println(fn, ":", bw)

	file, err := os.Open(fn)
	check(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	eventQueue := make(EventQueue, 0)
	for scanner.Scan() {
		flow := makeFlow(scanner.Text())
		eventQueue = append(eventQueue, makeArrivalEvent(flow))
	}

	flows := ideal(eventQueue, bw)
	numFlows := len(flows)

	slowdown := 0.0
	for i := 0; i < numFlows; i++ {
		slowdown += calculateFlowSlowdown(flows[i])
	}

	fmt.Println(slowdown / float64(numFlows))
}
