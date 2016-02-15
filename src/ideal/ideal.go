package main

import (
	"container/heap"
	"container/list"
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

func getOracleFCT(flow *Flow, bw float64) float64 {
	var td, pd float64
	if same_rack := flow.Source/HOSTS_IN_RACK == flow.Dest/HOSTS_IN_RACK; same_rack {
		pd = IN_RACK_PROPAGATION_DELAY
	} else {
		pd = INTER_RACK_PROPAGATION_DELAY
	}

	td = (float64(flow.Size) / (bw * 1e9 / 8)) * 1e6
	return pd + td
}

// input: linked list of flows
// output: sorted slice of flows, number of flows
func getSortedFlows(actives *list.List) (SortedFlows, int) {
	sortedFlows := make(SortedFlows, actives.Len())

	i := 0
	for e := actives.Front(); e != nil; e = e.Next() {
		sortedFlows[i] = e.Value.(*Flow)
		i++
	}

	sort.Sort(sortedFlows)
	return sortedFlows, len(sortedFlows)
}

func ideal(eventQueue EventQueue, bandwidth float64) []*Flow {
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

		if ev.Type == FlowArrival {
			flow.TimeRemaining = getOracleFCT(flow, bandwidth)
			flow.OracleFct = flow.TimeRemaining
			activeFlows.PushBack(flow)
		} else {
			// FlowCompletion
			flow.End = currentTime
			removeFromActiveFlows(activeFlows, flow)
			completedFlows = append(completedFlows, flow)
		}

		for i := 0; i < len(srcPorts); i++ {
			if srcPorts[i] != nil {
				inProgressFlow := srcPorts[i]
				if inProgressFlow.LastTime == 0 {
					panic("flow in progress without LastTime set")
				}

				inProgressFlow.TimeRemaining -= (currentTime - inProgressFlow.LastTime)
				inProgressFlow.LastTime = 0

				if inProgressFlow.FinishEvent == nil {
					panic("flow in progress without FinishEvent set")
				}
				inProgressFlow.FinishEvent.Cancelled = true
				inProgressFlow.FinishEvent = nil
			}
			srcPorts[i] = nil
			dstPorts[i] = nil
		}

		sortedActiveFlows, numActiveFlows := getSortedFlows(activeFlows)

		for i := 0; i < numActiveFlows; i++ {
			f := sortedActiveFlows[i]
			src := f.Source
			dst := f.Dest
			if srcPorts[src] == nil && dstPorts[dst] == nil {
				//this flow gets scheduled.
				f.LastTime = currentTime
				srcPorts[src] = f
				dstPorts[dst] = f

				if f.FinishEvent != nil {
					panic("flow being scheduled, finish event non-nil")
				}

				f.FinishEvent = makeCompletionEvent(currentTime+f.TimeRemaining, f)
				heap.Push(&eventQueue, f.FinishEvent)
			}
		}
	}

	return completedFlows
}
