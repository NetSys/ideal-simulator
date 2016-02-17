Flow-level simulator for the "ideal" algorithm outlined by [pfabric](http://conferences.sigcomm.org/sigcomm/2013/papers/sigcomm/p435.pdf). 

Algorithm
---------
This is an algorithm that:
    1. simulates the network topology as a "big switch".
    2. greedily schedules flows in SRPT (shortest remaining processing time) order

This gives at least a 2-approximation of the optimal average FCT, which is [NP-hard to compute](http://dl.acm.org/citation.cfm?doid=378420.378792).

The following, the pFabric ideal algorithm, is run whenever either (a) a new flow enters the network or (b) a flow finishes.

Input: List of active flows, F: (source, destination, size).
Output: Set of flows, S, that are scheduled in this iteration.

    S = {}
    all sources are not busy
    all destinations are not busy
    for each flow f in F in order of increasing remaining size:
        if f.source and f.destination are not busy:
            add f to S
            mark f.source and f.destination as busy

We make an improvement in our implementation. When a flow arrives in the network, only its source is marked as busy. After one propagation delay, the destination is also marked as busy. Similarly, once a flow transmits its last byte, its source is immediately made available to schedule another flow, while its destination remains marked busy until the last byte arrives.

This improvement prevents the wastage of a propagation delay amount of time at the source and destination at the end and beginning of a flow, respectively.

Running
-------

The simulator takes the following arguments: 

1. A list of flows to simulate. This is a file with the following format: 

   "[id (ignored)] [size, bytes] [source] [destination] [start time, microseconds]"

2. Bandwidth in gigabits
