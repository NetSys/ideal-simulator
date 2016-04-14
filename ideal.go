package main

// run ideal event loop until completion.

import (
	"container/heap"
	"container/list"
	"sort"
)

const (
	NUM_HOSTS         = 144
	PROPAGATION_DELAY = 0.4 // microseconds
	HOSTS_IN_RACK     = 16
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
	pd := PROPAGATION_DELAY * 4
	td := (float64(flow.Size) / (bw * 1e9 / 8)) * 1e6 // microseconds
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

// input: eventQueue of FlowArrival events, topology bandwidth (to determine oracle FCT)
// output: slice of pointers to completed Flows
func ideal(bandwidth float64, logger chan LogEvent, eventQueue EventQueue) []*Flow {
	heap.Init(&eventQueue)

	activeFlows := list.New()
	completedFlows := make([]*Flow, 0)
	var srcPorts [NUM_HOSTS]*Flow
	var dstPorts [NUM_HOSTS]*Flow
	var currentTime float64

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
			logger <- LogEvent{Time: currentTime, Type: LOG_FLOW_ARRIVAL, Flow: flow}
			prop_delay, trans_delay := getOracleFCT(flow, bandwidth)
			flow.TimeRemaining = trans_delay
			flow.OracleFct = prop_delay + trans_delay
			activeFlows.PushBack(flow)
		case FlowSourceFree:
			flow.FinishSending = true
			flow.FinishEvent = makeCompletionEvent(currentTime+2*PROPAGATION_DELAY, flow, FlowDestFree)
			heap.Push(&eventQueue, flow.FinishEvent)
		case FlowDestFree:
			if !flow.FinishSending {
				panic("finish without finishSending")
			}
			removeFromActiveFlows(activeFlows, flow)
			flow.End = currentTime + 2*PROPAGATION_DELAY // send an ACK
			flow.Finish = true
			completedFlows = append(completedFlows, flow)
			logger <- LogEvent{Time: currentTime, Type: LOG_FLOW_FINISHED, Flow: flow}
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
		for _, f := range sortedActiveFlows {
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

	return completedFlows
}
