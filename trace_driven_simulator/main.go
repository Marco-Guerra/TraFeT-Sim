package main

import (
	"flag"
	"log"

	"github.com/Marco-Guerra/Federated-Learning-Network-Workload/trace_driven_simulator/internal/simulator"
)

func main() {
	minBandwidthBps := flag.Uint("min-b", 200000, "Minimum device bandwidth of the simulated network")
	maxBandwidthBps := flag.Uint("max-b", 400000, "Max device bandwidth of the simulated network")
	traceFile := flag.String("t", "", "Trace file that describe the network workload during the simulation")
	mtu := flag.Uint("mtu", 1500, "MTU of the packets in the network")
	federatedScenario := flag.Uint("fs", 0, "type of federated learning scenario: cross silo or cross-device")

	flag.Parse()

	if *traceFile == "" {
		log.Panic("Trace file path must be given")
	}

	traceDrivenSimulator := simulator.New(&simulator.GlobalOptions{
		MinBandwidth:      uint32(*minBandwidthBps),
		MaxBandwidth:      uint32(*maxBandwidthBps),
		MTU:               uint16(*mtu),
		FederatedScenario: simulator.TrainingScenario(*federatedScenario),
	})

	traceDrivenSimulator.RunSimulation(*traceFile)

}
