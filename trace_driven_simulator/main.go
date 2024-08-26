package main

import (
	"flag"
	"log"

	"github.com/Marco-Guerra/Federated-Learning-Network-Workload/trace_driven_simulator/internal/simulator"
)

func main() {
	minBandwidthBps := flag.Uint("min-b", 200000, "Minimum device bandwidth of the simulated network")
	maxBandwidthBps := flag.Uint("max-b", 400000, "Max device bandwidth of the simulated network")
	minAggDelay := flag.Float64("min-agg-delay", 1, "Minimum server aggregation delay of the simulated network in seconds")
	maxAggDelay := flag.Float64("max-agg-delay", 10, "Max server aggregation delay of the simulated network in seconds")
	traceFile := flag.String("t", "", "Trace file that describe the network workload during the simulation")
	mtu := flag.Uint("mtu", 1500, "MTU of the packets in the network")

	flag.Parse()

	if *traceFile == "" {
		log.Panic("Trace file path must be given")
	}

	traceDrivenSimulator := simulator.New(&simulator.GlobalOptions{
		MinBandwidth:        uint32(*minBandwidthBps),
		MaxBandwidth:        uint32(*maxBandwidthBps),
		MTU:                 uint16(*mtu),
		MinAggregationDelay: float32(*minAggDelay),
		MaxAggregationDelay: float32(*maxAggDelay),
	})

	traceDrivenSimulator.RunSimulation(*traceFile)

}
