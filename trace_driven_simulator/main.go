package main

import (
	"flag"
	"log"

	"github.com/Marco-Guerra/Federated-Learning-Network-Workload/trace_driven_simulator/internal/simulator"
)

func main() {
	bandwidthBps := flag.Uint("b", 40000000, "Bandwidth of the simulated network")
	traceFile := flag.String("t", "", "Trace file that describe the network workload during the simulation")
	mtu := flag.Uint("mtu", 1500, "MTU of the packets in the network")
	federatedScenario := flag.Uint("fs", 0, "type of federated learning scenario: cross silo or cross-device")

	flag.Parse()

	if *traceFile == "" {
		log.Panic("Trace file path must be given")
	}

	traceDrivenSimulator := simulator.New(&simulator.GlobalOptions{
		Bandwidth:         uint32(*bandwidthBps),
		MTU:               uint16(*mtu),
		FederatedScenario: simulator.TrainingScenario(*federatedScenario),
	})

	traceDrivenSimulator.RunSimulation(*traceFile)

}
