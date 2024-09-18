package queues

import (
	"container/heap"
	"log"

	"github.com/Marco-Guerra/Federated-Learning-Network-Workload/trace_driven_simulator/packages/writer"
)

func New(options *GlobalOptions, workload *EventHeap, rwritter *writer.Writer) *EventQueue {
	return &EventQueue{
		options:        options,
		queue:          make([]*Packet, 0),
		events:         workload,
		resultsWritter: rwritter,
	}
}

func (evq *EventQueue) Start() *Output {
	numPackets, totalBytes, totalDelay, outWorkload := evq.processEvents()

	return &Output{
		SimTime:    evq.currentTime,
		TotalBytes: totalBytes,
		Delay:      totalDelay,
		NumPackets: uint32(numPackets),
		Workload:   outWorkload,
	}
}

func (evq *EventQueue) cleanBuffer() {
	qlen := len(evq.queue)

	if qlen >= int(evq.options.MaxQueue) {
		index := -1
		low := 0
		high := qlen - 1

		for low <= high {
			mid := (low + high) / 2

			if evq.queue[mid].DepartureTime == evq.currentTime {
				index = mid
				break
			} else if evq.queue[mid].DepartureTime < evq.currentTime {
				low = mid + 1
			} else {
				high = mid - 1
			}
		}

		if index > 1 {
			evq.queue = evq.queue[index:]
		}
	}
}

func (evq *EventQueue) processEvents() (int, uint64, float64, *EventHeap) {
	numPackets := evq.events.Len()
	var totalBytes uint64 = 0
	var totalDelay float64 = 0
	var outWorkload *EventHeap = nil

	if evq.options.NetType != SERVER {
		outWorkload = &EventHeap{}
	}

	for evq.events.Len() > 0 {
		event := heap.Pop(evq.events).(*Event)
		evq.currentTime = event.Time

		switch event.Type {
		case ARRIVAL:
			qlen := len(evq.queue)

			if qlen == 0 || evq.queue[qlen-1].DepartureTime <= event.Packet.ArrivalTime {
				event.Packet.StartServiceTime = event.Packet.ArrivalTime
			} else {
				event.Packet.StartServiceTime = evq.queue[qlen-1].DepartureTime
			}

			event.Packet.DepartureTime = event.Packet.StartServiceTime + (float32(event.Packet.Size)*8)/float32(evq.options.Bandwidth)

			heap.Push(evq.events, &Event{
				Time:        event.Packet.DepartureTime,
				RoundNumber: event.RoundNumber,
				ClientID:    event.ClientID,
				Packet:      event.Packet,
				Type:        DEPARTURE,
			})

			evq.queue = append(evq.queue, event.Packet)
		case DEPARTURE:
			totalBytes += uint64(event.Packet.Size)

			if event.Packet.Type == LAST {
				individualDelay := event.Packet.DepartureTime - event.Packet.MSSArrivalTime

				evq.resultsWritter.Write(&writer.WriterRegister{
					ClientID:      event.ClientID,
					Network:       uint8(evq.options.NetType),
					RoundNumber:   event.RoundNumber,
					ArrivalTime:   event.Packet.MSSArrivalTime,
					DepartureTime: event.DepartureTime,
					Size:          event.Packet.MSSSize,
				})

				totalDelay += float64(individualDelay)
				event.Packet.MSSArrivalTime = event.Packet.DepartureTime
			}

			event.Packet.ArrivalTime = event.Packet.DepartureTime

			if evq.options.NetType != SERVER {
				heap.Push(outWorkload, &Event{
					Time:        event.Packet.DepartureTime,
					RoundNumber: event.RoundNumber,
					ClientID:    event.ClientID,
					Packet:      event.Packet,
					Type:        ARRIVAL,
				})
			}

			evq.cleanBuffer()
		default:
			log.Fatal("Unkown Event on the Event list. ", event)
		}
	}

	return numPackets, totalBytes, totalDelay, outWorkload
}
