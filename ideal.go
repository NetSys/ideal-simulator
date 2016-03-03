package main

// run ideal event loop until completion.

import (
	"container/heap"
	"container/list"
	"fmt"
	"sort"
)

const (
	NUM_HOSTS                    = 144
	IN_RACK_PROPAGATION_DELAY    = 0.440 // microseconds
	INTER_RACK_PROPAGATION_DELAY = 2.040 // microseconds
	HOSTS_IN_RACK                = 16
)

func removeFromActiveFlows(ls *list.List, f *Flow) {
	for e := ls.Front(); e != nil; e = e.Next() {
		if e.Value.(*Flow) == f {
			ls.Remove(e)
			break
		}
	}
}

func getOracleFCT(flow *Flow, bw float64) (float64, float64) {
	var td, pd float64
	if same_rack := flow.Source/HOSTS_IN_RACK == flow.Dest/HOSTS_IN_RACK; same_rack {
		pd = IN_RACK_PROPAGATION_DELAY
	} else {
		pd = INTER_RACK_PROPAGATION_DELAY
	}

	td = (float64(flow.Size) / (bw * 1e9 / 8)) * 1e6
	return pd, td
}

// input: linked list of flows
// output: sorted slice of flows, number of flows
func getSortedFlows(actives *list.List) SortedFlows {
	sortedFlows := make(SortedFlows, actives.Len())

	i := 0
	for e := actives.Front(); e != nil; e = e.Next() {
		sortedFlows[i] = e.Value.(*Flow)
		i++
	}

	sort.Sort(sortedFlows)
	return sortedFlows
}

func trackBacklog(times chan float64, bytes chan int, quit chan int) {
	backlog := 0
	for b := range bytes {
		t := <-times
		backlog += b
		fmt.Printf("backlog %6.3f %d %d\n", t, b, backlog)
	}
	quit <- 0
}

// input: eventQueue of FlowArrival events, topology bandwidth (to determine oracle FCT)
// output: slice of pointers to completed Flows
func ideal(eventQueue EventQueue, bandwidth float64) ([]*Flow, chan int) {
	heap.Init(&eventQueue)

	activeFlows := list.New()
	completedFlows := make([]*Flow, 0)
	var srcPorts [NUM_HOSTS]*Flow
	var dstPorts [NUM_HOSTS]*Flow
	var currentTime float64

	timesC := make(chan float64)
	backlogC := make(chan int)
	quit := make(chan int)
	defer close(backlogC)
	go trackBacklog(timesC, backlogC, quit)

	for len(eventQueue) > 0 {
		ev := heap.Pop(&eventQueue).(*Event)
		if ev.Cancelled {
			continue
		}

		if ev.Time < currentTime {
			panic("going backwards!")
		}

		currentTime = ev.Time
		flow := ev.Flow

		switch ev.Type {
		case FlowArrival:
			backlogC <- int(flow.Size)
			timesC <- currentTime
			prop_delay, trans_delay := getOracleFCT(flow, bandwidth)
			flow.TimeRemaining = trans_delay
			flow.OracleFct = prop_delay + trans_delay
			flow.PropDelay = prop_delay
			activeFlows.PushBack(flow)
		case FlowSourceFree:
			removeFromActiveFlows(activeFlows, flow)
			flow.FinishSending = true
			flow.FinishEvent = makeCompletionEvent(currentTime+flow.PropDelay, flow, FlowDestFree)
			heap.Push(&eventQueue, flow.FinishEvent)
		case FlowDestFree:
			backlogC <- (-1 * int(flow.Size))
			timesC <- currentTime
			if !flow.FinishSending {
				panic("finish without finishSending")
			}

			flow.End = currentTime
			flow.Finish = true
			completedFlows = append(completedFlows, flow)
		}

		for i := 0; i < len(srcPorts); i++ {
			if srcPorts[i] != nil {
				inProgressFlow := srcPorts[i]

				if inProgressFlow.LastTime == 0 {
					panic("flow in progress without LastTime set")
				}

				if inProgressFlow.FinishEvent == nil {
					panic("flow in progress without FinishEvent set")
				}

				inProgressFlow.TimeRemaining -= (currentTime - inProgressFlow.LastTime)
				inProgressFlow.LastTime = 0

				if !inProgressFlow.FinishSending {
					inProgressFlow.FinishEvent.Cancelled = true
					inProgressFlow.FinishEvent = nil
				}
			}
			srcPorts[i] = nil
			dstPorts[i] = nil
		}

		sortedActiveFlows := getSortedFlows(activeFlows)
		numActiveFlows := len(sortedActiveFlows)

		for i := 0; i < numActiveFlows; i++ {
			f := sortedActiveFlows[i]
			src := f.Source
			dst := f.Dest

			if f.FinishSending {
				if f.Finish {
					panic("finished flow in actives")
				}

				if srcPorts[src] != nil || dstPorts[dst] != nil {
					panic("ports taken on still sending flow")
				}

				dstPorts[dst] = f
				continue
			}

			if srcPorts[src] == nil && dstPorts[dst] == nil {
				//this flow gets scheduled.
				f.LastTime = currentTime
				srcPorts[src] = f
				dstPorts[dst] = f

				if f.FinishEvent != nil {
					panic("flow being scheduled, finish event non-nil")
				}

				f.FinishEvent = makeCompletionEvent(currentTime+f.TimeRemaining, f, FlowSourceFree)
				heap.Push(&eventQueue, f.FinishEvent)
			}
		}
	}

	return completedFlows, quit
}
