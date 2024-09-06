package simulator

import (
	"github.com/Marco-Guerra/Federated-Learning-Network-Workload/trace_driven_simulator/packages/writer"
)

const (
	SERVER_AGG_TIME         float32 = 60
	DOWNLINK_TIME           float32 = 30
	BACKGROUND_TRAFFIC_RATE float64 = 100
	SERVER_BANDWIDTH        uint32  = 4000000
)

type GlobalOptions struct {
	MinBandwidth uint32
	MaxBandwidth uint32
	MTU          uint16
}

type TraceDriven struct {
	options        *GlobalOptions
	resultsWritter *writer.Writer
}
