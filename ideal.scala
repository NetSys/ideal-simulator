import scala.io.Source
import scala.collection.mutable.PriorityQueue
import scala.collection.mutable.Set

object Ideal {
    object FlowState extends Enumeration {
        type FlowState = Value
        val Pending, Active, Stopped, Finished = Value
    }

    object EventType extends Enumeration {
        type EventType = Value
        val FlowArrival, FlowCompletion = Value
    }

    import FlowState._
    import EventType._

    case class Flow(st:Float, s: Int, d: Int, sz: Int) {
        val start: Float = st
        var end: Float = -1
        val src: Int = s
        val dst: Int = d
        val size: Int = sz
        val oracleFct: Float = 0
        var state: FlowState = FlowState.Pending
        var finishEvent: Event = null
        var timeRemaining: Float = 0
        var fct: Float = 0
        var last_time: Float = -1
    }

    object FlowOrdering extends Ordering[Flow] {
        def compare(x: Flow, y: Flow):Int = {
            // SRPT
            return y.timeRemaining.compare(x.timeRemaining)
        }
    }

    case class Event(t: Float, f: Flow, et: EventType) {
        val time: Float = t
        val flow: Flow = f
        val eventType: EventType = et
        var cancelled: Boolean = false
    }

    object EventOrdering extends Ordering[Event] {
        def compare(x: Event, y: Event):Int = {
            return y.time.compare(x.time)
        }
    }
    
    def makeFlow(l: String, oracle: Flow => Float) = l.split(" ").toList match {
        case id :: size :: source :: dest :: start :: Nil => {
            var f = new Flow(start.toFloat, source.toInt, dest.toInt, size.toInt)
            f.oracleFct = oracle(f)
            f.timeRemaining = f.oracleFct
        
        case _ => null
    }

    def getOracleFct(bandwidth: Double)(flow: Flow): Float = {
        val src: Int = flow.src
        val dst: Int = flow.dst
        val in_rack = (src / 16) == (dst / 16)
        val pd = if (in_rack) 0.440 else 2.040
        val td = flow.size.toFloat / bandwidth
        return pd + td
    }

    def run(flows: Iterable[Flow], bandwidth: Int) {
        // enqueue flow arrival events
        val events = flows.map((f: Flow) => new Event(f.start, f, EventType.FlowArrival))
        var eventQueue = new PriorityQueue[Event]()(EventOrdering)
        events.foreach(eventQueue.enqueue(_))

        // is that host free?
        var srcPorts = Array().padTo(144, null)
        var dstPorts = Array().padTo(144, null)
        var activeFlows = Set[Flow]()
        var completedFlows = LinkedList[Flow]()

        while(!eventQueue.isEmpty) {
            val evt = eventQueue.dequeue
            if (evt.cancelled) {
                continue
            }

            val currentTime = evt.time
            val flow = evt.flow

            if (evt.eventType == EventType.FlowArrival) {
                //this flow is active.
                activeFlows.add(flow)
            }
            else if (evt.eventType == EventType.FlowCompletion) {
                activeFlows.remove(flow)
                flow.timeRemaining = 0
                flow.fct = currentTime - flow.start
                completedFlows += flow

                srcPorts(flow.src) = null
                dstPorts(flow.dst) = null
            }
            
            // all previously active flows should be suspended
            srcPorts.filter(_ != null).foreach((f: Flow) => {
                assert(f.last_time != -1)
                f.timeRemaining -= (currentTime - f.last_time)
                f.last_time = -1
            })
            
            var srcPorts = Array().padTo(144, null)
            var dstPorts = Array().padTo(144, null)

            val flowsToSchedule = activeFlows.toList.sorted(FlowOrdering)

            for (flow <- flowsToSchedule) {
                if (srcPorts(flow.src) == null and dstPorts(flow.dst) == null) {
                    // mark ports as busy
                    srcPorts(flow.src) = flow
                    dstPorts(flow.dst) = flow

                    flow.finishEvent.cancelled = true

                    // set an event for the flow's completion time
                    val completion_event = new Event(currentTime + flow.timeRemaining, flow, EventType.FlowCompletion)
                    flow.finishEvent = completion_event
                    eventQueue += completion_event

                    // if flow is later interrupted, need to know how much progress it made
                    flow.last_time = currentTime
                }
            }
        }
    }

    def main(args: Array[String]):Unit = {
        // args are:
        // scala ideal.scala <flows filename> <bandwidth, Gbps>
        val flowsfn = args(0)
        val bandwidth = args(1).toInt
        
        val flows = Source.fromFile(flowsfn).getLines.map(makeFlow(_))
        run(flows, getOracleFct(bandwidth))
    }
}
