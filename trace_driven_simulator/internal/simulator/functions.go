package simulator

import (
	"container/heap"
	"encoding/csv"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/Marco-Guerra/Federated-Learning-Network-Workload/trace_driven_simulator/packages/writer"
)

func New(options *GlobalOptions) *MM1Queue {
	return &MM1Queue{
		options:        options,
		queue:          make([]*Packet, 0),
		events:         &EventHeap{},
		resultsWritter: nil,
	}
}

func (mmq *MM1Queue) RunSimulation(trace_filename string) {
	mmq.readTrace(trace_filename)

	mmq.processEvents()

	mmq.resultsWritter.Close()

	meanDelay, throughput := mmq.calculeMetrics()

	log.Printf("Mean Delay: %f seconds", meanDelay)
	log.Printf("Throughput: %f bits per second", throughput)
}

func (mmq *MM1Queue) calculeMetrics() (float64, float32) {
	meanDelay := mmq.totalDelay / float64(mmq.totalPackets)

	if mmq.currentTime <= 0 {
		mmq.currentTime = 1
	}

	throughput := float32(mmq.totalBytes*8) / mmq.currentTime

	return meanDelay, float32(math.Floor(float64(throughput)))
}

func (mmq *MM1Queue) readTrace(traceFilename string) {
	parts := strings.Split(traceFilename, "_")
	var leafExperimentMeta string

	if len(parts) > 2 {
		leafExperimentMeta = strings.Join(parts[4:], "_")
	} else {
		log.Fatal("Unexpected patten in trace filename. ", traceFilename)
	}

	file, err := os.Open(traceFilename)
	if err != nil {
		log.Fatal("Error opening file:", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Error reading CSV file:", err)
	}

	var packetCounter uint64 = 0
	var currentTime float32 = 0.0

	// Find the maximum round number
	rounds := 0
	for i, record := range records {
		if i == 0 {
			continue // Skip header
		}
		roundNumber, _ := strconv.Atoi(record[1])
		if roundNumber > rounds {
			rounds = roundNumber
		}
	}

	for round := 1; round <= rounds; round++ {
		var clients [][]string
		for i, record := range records {
			if i == 0 {
				continue // Skip header
			}
			roundNumber, _ := strconv.Atoi(record[1])
			if roundNumber == round {
				clients = append(clients, record)
			}
		}

		var messageSize int = 0

		for _, row := range clients {
			messageSize, _ = strconv.Atoi(row[4])
			time, _ := strconv.ParseFloat(row[6], 32)
			clientID, _ := strconv.Atoi(row[0])

			packet := Packet{
				ArrivalTime: float32(time) + currentTime,
				Size:        uint32(messageSize),
				Id:          packetCounter,
			}

			event := Event{
				Time:        packet.ArrivalTime,
				RoundNumber: uint16(round),
				ClientID:    uint16(clientID),
				Packet:      &packet,
				Type:        ARRIVAL,
			}

			heap.Push(mmq.events, &event)

			packetCounter++
		}

		// Update current time
		for _, client := range clients {
			clientTime, _ := strconv.ParseFloat(client[6], 32)
			if float32(clientTime) > currentTime {
				currentTime = float32(clientTime)
			}
		}

		switch mmq.options.FederatedScenario {
		case CROSSDEVICE:
			currentTime += CROSSDEVICEBROADCASTDELAY

			for range clients {
				packet := Packet{
					ArrivalTime: currentTime,
					Size:        uint32(messageSize),
					Id:          packetCounter,
				}

				event := Event{
					Time:        packet.ArrivalTime,
					RoundNumber: uint16(round),
					ClientID:    uint16(0), // 0 == ServerID
					Packet:      &packet,
					Type:        ARRIVAL,
				}

				heap.Push(mmq.events, &event)

				packetCounter++
			}
		case CROSSSILO:
			// Assuming that all messages have the same size
			// And crosssilo have a broadcast protocol implemented
			currentTime += CROSSSILOBROADCASTDELAY

			packet := Packet{
				ArrivalTime: currentTime,
				Size:        uint32(messageSize),
				Id:          packetCounter,
			}

			event := Event{
				Time:        packet.ArrivalTime,
				RoundNumber: uint16(round),
				ClientID:    uint16(0), // 0 == ServerID
				Packet:      &packet,
				Type:        ARRIVAL,
			}

			heap.Push(mmq.events, &event)

			packetCounter++
		}
	}

	mmq.maxQueue = uint16(math.Floor((float64(mmq.events.Len()) * 0.10)))
	mmq.resultsWritter = writer.New(uint32(mmq.events.Len()), "metrics_network_"+leafExperimentMeta)
}

func (mmq *MM1Queue) processEvents() {
	go mmq.resultsWritter.Start()

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
			mmq.totalBytes += uint64(event.Packet.Size)
			mmq.totalDelay += float64(individualDelay)
			mmq.totalPackets += 1

			mmq.resultsWritter.Write(&writer.WriterRegister{
				ClientID:    event.ClientID,
				RoundNumber: event.RoundNumber,
				Time:        mmq.currentTime,
				Delay:       individualDelay,
				Size:        event.Packet.Size,
			})

			log.Printf(
				"Arrived %f : Departed %f : Delay %f : Size %d\n",
				event.Packet.ArrivalTime,
				event.Packet.DepartureTime,
				individualDelay,
				event.Packet.Size,
			)

			qlen := len(mmq.queue)

			if qlen >= int(mmq.maxQueue) {
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
}
