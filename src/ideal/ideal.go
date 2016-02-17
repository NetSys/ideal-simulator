package main

import (
	"container/heap"
	"container/list"
	"sort"
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
	if same_rack := flow.Source/16 == flow.Dest/16; same_rack {
		pd = 0.440
	} else {
		pd = 2.040
	}

	td = float64(flow.Size) / (bw * 10e9 / 8)
	return pd + td
}

func ideal(eventQueue EventQueue, bandwidth float64) []*Flow {
	heap.Init(&eventQueue)

	activeFlows := list.New()
	completedFlows := make([]*Flow, 0)
	var srcPorts [144]*Flow
	var dstPorts [144]*Flow
	var currentTime float64

	for len(eventQueue) > 0 {
		ev := heap.Pop(&eventQueue).(*Event)
		if ev.Cancelled {
			continue
		}

		currentTime = ev.Time
		flow := ev.Flow

		if ev.Type == FlowArrival {
			flow.TimeRemaining = getOracleFCT(flow, bandwidth)
			activeFlows.PushBack(flow)
		} else {
			// FlowCompletion
			removeFromActiveFlows(activeFlows, flow)
			completedFlows = append(completedFlows, flow)
		}

		for i := 0; i < 144; i++ {
			if srcPorts[i] != nil {
				inProgressFlow := srcPorts[i]
				if inProgressFlow.LastTime == 0 {
					panic("flow in progress without LastTime set")
				}

				inProgressFlow.TimeRemaining = (currentTime - inProgressFlow.LastTime)
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

		numActiveFlows := activeFlows.Len()
		activeFlowsSlice := make(SortedFlows, numActiveFlows)

		i := 0
		for e := activeFlows.Front(); e != nil; e = e.Next() {
			activeFlowsSlice[i] = e.Value.(*Flow)
			i++
		}

		sort.Sort(activeFlowsSlice)

		for i := 0; i < numActiveFlows; i++ {
			f := activeFlowsSlice[i]
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
				eventQueue = append(eventQueue, f.FinishEvent)
			}
		}
	}

	return completedFlows
}
