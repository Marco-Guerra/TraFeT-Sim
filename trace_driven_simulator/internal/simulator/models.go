package simulator

import (
	"github.com/Marco-Guerra/Federated-Learning-Network-Workload/trace_driven_simulator/packages/writer"
)

const (
	SERVER_AGG_TIME         float32 = 60
	DOWNLINK_TIME           float32 = 30
	BACKGROUND_TRAFFIC_RATE float64 = 10000
	ETHERNET_HEADER         uint8   = 14
	ETHERNET_MIN_FRAME      uint8   = 64
	ETHERNET_MTU            uint16  = 1500
)

type GlobalOptions struct {
	ClientsBandwidth   uint32
	ServerBandwidth    uint32
	NBackgroundClients uint16
}

type TraceDriven struct {
	options        *GlobalOptions
	resultsWritter *writer.Writer
}
