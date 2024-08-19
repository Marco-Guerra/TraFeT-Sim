package simulator

import (
	"github.com/Marco-Guerra/Federated-Learning-Network-Workload/trace_driven_simulator/internal/simulator/queues"
	"github.com/Marco-Guerra/Federated-Learning-Network-Workload/trace_driven_simulator/packages/writer"
)

type TrainingScenario uint8

const (
	CROSSSILO   TrainingScenario = 0
	CROSSDEVICE TrainingScenario = 1
)

const (
	CROSSSILOBROADCASTDELAY   float32 = 0.805
	CROSSDEVICEBROADCASTDELAY float32 = 800
)

type GlobalOptions struct {
	FederatedScenario TrainingScenario
	MTU               uint16
	Bandwidth         uint32
}

type TraceDriven struct {
	options        *GlobalOptions
	queue          *queues.MM1Queue
	resultsWritter *writer.Writer
}
