Flow-level simulator for the "ideal" algorithm outlined by [pfabric](http://conferences.sigcomm.org/sigcomm/2013/papers/sigcomm/p435.pdf). 

Algorithm
---------
This is an algorithm that:
    1. simulates the network topology as a "big switch".
    2. greedily schedules flows in SRPT (shortest remaining processing time) order

This gives at least a 2-approximation of the optimal average FCT, which is [NP-hard to compute](http://dl.acm.org/citation.cfm?doid=378420.378792).

The following is run whenever either (a) a new flow enters the network or (b) a flow finishes.

Input: List of active flows, F: (source, destination, size).
Output: Set of flows, S, that are scheduled in this iteration.

    S = {}
    all sources are not busy
    all destinations are not busy
    for each flow f in F in order of increasing remaining size:
        if f.source and f.destination are not busy:
            add f to S
            mark f.source and f.destination as busy

Running
-------

The simulator takes the following arguments: 

1. A list of flows to simulate. This is a file with the following format: 
   "<id (ignored)> <size, bytes> <source> <destination> <start time, microseconds>"
2. Bandwidth in gigabits
