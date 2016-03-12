Flow-level simulator for the "ideal" algorithm outlined by [pfabric](http://conferences.sigcomm.org/sigcomm/2013/papers/sigcomm/p435.pdf). 

Algorithm
=========

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
=======

The simulator takes one argument, the path to a configuration file. The configuration file consists of lines, each of which must be either a parameter definition or a switch. The format of the two types of lines follows.

Switch: ```[Switch Name]```

Parameter definition: ```[Field Name] [Value]```

A list of valid parameters to define and switches to use follows.

Switches
--------

Note that for each switch, no parameter other than the required parameter for that switch should be defined.

- "Read": Read flows from a trace file and run the ideal algorithm. The parameters "Bandwidth" and "TraceFile" must be defined. 
- "Generate": Generate flows using a Poisson arrival process and run the ideal algorithm. The parameters "Bandwidth", "Load", "NumFlows", and "CDF" must be defined. 
- "GenerateOnly": Generate flows using a Poisson arrival process and exit. The parameters are the same as in "Generate".

Parameters
----------

- "TraceFile": The path to a file containing the list of flows to simulate. This file should have the following format: 

   ```[id (ignored)] [size (bytes)] [source] [destination] [start time (microseconds)]```

   The source and destination fields are currently hardcoded to correspond to a 144-host topology with 9 racks of 16 nodes each. In-rack propagation delay is set to be 440 ns, and inter-rack propagation delay is set to be 2040 ns.

- "Bandwidth": Topology bandwidth in gigabits. Note that this parameter is required for all experiments.
- "NumFlows": The number of flows to generate for simulation.
- "Load": The target load at which flows should be generated.
- "CDF": The path to a CDF file defining the flow size distribution with which to generate flows. This file should have the following format: 

   ```[Flow Size (packets)] <ignored> [CDF Value]```

Example Configuration
---------------------

Configuration file:
```
Bandwidth 40
GenerateOnly
Load 0.9
NumFlows 1000000
CDF imc10.cdf
```

CDF File (imc10.cdf above):
```
1 1 0
1 1 0.500000
2 1 0.600000
3 1 0.700000
5 1 0.750000
7 1 0.800000
40 1 0.812500
72 1 0.825000
137 1 0.850000
267 1 0.900000
1187 1 0.95000
2107 1 1.0
```

