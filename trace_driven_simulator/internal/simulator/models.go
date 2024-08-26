package simulator

import (
	"github.com/Marco-Guerra/Federated-Learning-Network-Workload/trace_driven_simulator/internal/simulator/queues"
	"github.com/Marco-Guerra/Federated-Learning-Network-Workload/trace_driven_simulator/packages/writer"
)

const (
	SYNC_TIME               float32 = 180
	SERVER_AGG_RATE         float32 = 1
	BACKGROUND_TRAFFIC_RATE float64 = 100
)

type GlobalOptions struct {
	MTU                 uint16
	MinBandwidth        uint32
	MaxBandwidth        uint32
	MinAggregationDelay float32
	MaxAggregationDelay float32
}

type TraceDriven struct {
	options        *GlobalOptions
	queues         []*queues.MM1Queue
	resultsWritter *writer.Writer
}
