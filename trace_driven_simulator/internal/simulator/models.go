package simulator

import "github.com/Marco-Guerra/Federated-Learning-Network-Workload/trace_driven_simulator/packages/writer"

type EventType uint8
type TrainingScenario uint8

const (
	ARRIVAL   EventType = 0
	DEPARTURE EventType = 1
)

const (
	CROSSSILO   TrainingScenario = 0
	CROSSDEVICE TrainingScenario = 1
)

const (
	CROSSSILOBROADCASTDELAY   float32 = 0.805
	CROSSDEVICEBROADCASTDELAY float32 = 800
)

type Packet struct {
	Size             uint16
	ArrivalTime      float32
	StartServiceTime float32
	DepartureTime    float32
	Id               uint64
}

type Event struct {
	Time float32
	*Packet
	ClientID    uint16
	RoundNumber uint16
	Type        EventType
	_           [3]byte
}

type GlobalOptions struct {
	FederatedScenario TrainingScenario
	MTU               uint16
	Bandwidth         uint32
}

type MM1Queue struct {
	options        *GlobalOptions
	queue          []*Packet
	events         *EventHeap
	resultsWritter *writer.Writer
	maxQueue       uint16
	currentTime    float32
	totalDelay     float64
	totalBytes     uint64
	totalPackets   uint64
}

// EventHeap implements heap.Interface and holds Events
type EventHeap []*Event

func (h EventHeap) Len() int { return len(h) }

func (h EventHeap) Less(i, j int) bool {
	if h[i].Time != h[j].Time {
		return h[i].Time < h[j].Time
	}
	if h[i].RoundNumber != h[j].RoundNumber {
		return h[i].RoundNumber < h[j].RoundNumber
	}
	if h[i].ClientID != h[j].ClientID {
		return h[i].ClientID < h[j].ClientID
	}
	return h[i].Packet.Id < h[j].Packet.Id
}

func (h EventHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *EventHeap) Push(x interface{}) {
	*h = append(*h, x.(*Event))
}

func (h *EventHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}
