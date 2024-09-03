package simulator

import (
	"github.com/Marco-Guerra/Federated-Learning-Network-Workload/trace_driven_simulator/internal/simulator/queues"
	"github.com/Marco-Guerra/Federated-Learning-Network-Workload/trace_driven_simulator/packages/writer"
)

const (
	SYNC_TIME               float32 = 180
	SERVER_AGG_TIME         float32 = 60
	DOWNLINK_TIME           float32 = 30
	BACKGROUND_TRAFFIC_RATE float64 = 100
)

type GlobalOptions struct {
	MinBandwidth uint32
	MaxBandwidth uint32
	MTU          uint16
}

type TraceDriven struct {
	options        *GlobalOptions
	queues         []*queues.EventQueue
	resultsWritter *writer.Writer
}
