package queues

import (
	"container/heap"
	"log"

	"github.com/Marco-Guerra/Federated-Learning-Network-Workload/trace_driven_simulator/packages/writer"
)

func New(options *GlobalOptions, workload *EventHeap, rwritter *writer.Writer) *MM1Queue {
	return &MM1Queue{
		options:        options,
		queue:          make([]*Packet, 0),
		events:         workload,
		resultsWritter: rwritter,
	}
}

func (mmq *MM1Queue) Start() *Output {
	numPackets, totalBytes, totalDelay := mmq.processEvents()

	return &Output{
		SimTime:    mmq.currentTime,
		TotalBytes: totalBytes,
		Delay:      totalDelay,
		NumPackets: uint32(numPackets),
	}
}

func (mmq *MM1Queue) processEvents() (int, uint64, float64) {
	numPackets := mmq.events.Len()
	var totalBytes uint64 = 0
	var totalDelay float64 = 0

	for mmq.events.Len() > 0 {
		event := heap.Pop(mmq.events).(*Event)
		mmq.currentTime = event.Time

		switch event.Type {
		case ARRIVAL:
			qlen := len(mmq.queue)

			if qlen == 0 || mmq.queue[qlen-1].DepartureTime <= event.Packet.ArrivalTime {
				event.Packet.StartServiceTime = event.Packet.ArrivalTime
			} else {
				event.Packet.StartServiceTime = mmq.queue[qlen-1].DepartureTime
			}

			event.Packet.DepartureTime = event.Packet.StartServiceTime + (float32(event.Packet.Size)*8)/float32(mmq.options.Bandwidth)

			heap.Push(mmq.events, &Event{
				Time:        event.Packet.DepartureTime,
				RoundNumber: event.RoundNumber,
				ClientID:    event.ClientID,
				Packet:      event.Packet,
				Type:        DEPARTURE,
			})

			mmq.queue = append(mmq.queue, event.Packet)
		case DEPARTURE:
			individualDelay := event.Packet.DepartureTime - event.Packet.ArrivalTime
			totalBytes += uint64(event.Packet.Size)
			totalDelay += float64(individualDelay)

			mmq.resultsWritter.Write(&writer.WriterRegister{
				ClientID:    event.ClientID,
				RoundNumber: event.RoundNumber,
				Time:        mmq.currentTime,
				Delay:       individualDelay,
				Size:        event.Packet.Size,
			})

			qlen := len(mmq.queue)

			if qlen >= int(mmq.options.MaxQueue) {
				index := -1
				low := 0
				high := qlen - 1

				for low <= high {
					mid := (low + high) / 2

					if mmq.queue[mid].DepartureTime == mmq.currentTime {
						index = mid
						break
					} else if mmq.queue[mid].DepartureTime < mmq.currentTime {
						low = mid + 1
					} else {
						high = mid - 1
					}
				}

				if index > 1 {
					mmq.queue = mmq.queue[index:]
				}
			}
		default:
			log.Fatal("Unkown Event on the Event list. ", event)
		}
	}

	return numPackets, totalBytes, totalDelay
}
