package simulator

import (
	"container/heap"
	"encoding/csv"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/Marco-Guerra/Federated-Learning-Network-Workload/trace_driven_simulator/internal/simulator/queues"
	"github.com/Marco-Guerra/Federated-Learning-Network-Workload/trace_driven_simulator/packages/writer"
)

func New(options *GlobalOptions) *TraceDriven {
	return &TraceDriven{
		options: options,
	}
}

func (td *TraceDriven) RunSimulation(trace_filename string) {
	td.readTrace(trace_filename)

	go td.resultsWritter.Start()

	qout := td.queue.Start()

	td.resultsWritter.Close()

	meanDelay, throughput := td.calculeMetrics(qout)

	log.Printf("Mean Delay: %f seconds", meanDelay)
	log.Printf("Throughput: %f bits per second", throughput)
}

func (td *TraceDriven) calculeMetrics(results *queues.Output) (float64, float32) {
	meanDelay := results.Delay / float64(results.NumPackets)

	if results.SimTime <= 0 {
		results.SimTime = 1
	}

	throughput := float32(results.TotalBytes*8) / results.SimTime

	return meanDelay, float32(math.Floor(float64(throughput)))
}

func (td *TraceDriven) readTrace(traceFilename string) {
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

	workload := queues.EventHeap{}

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

			packet := queues.Packet{
				ArrivalTime: float32(time) + currentTime,
				Size:        uint32(messageSize),
				Id:          packetCounter,
			}

			event := queues.Event{
				Time:        packet.ArrivalTime,
				RoundNumber: uint16(round),
				ClientID:    uint16(clientID),
				Packet:      &packet,
				Type:        queues.ARRIVAL,
			}

			heap.Push(&workload, &event)

			packetCounter++
		}

		// Update current time
		for _, client := range clients {
			clientTime, _ := strconv.ParseFloat(client[6], 32)
			if float32(clientTime) > currentTime {
				currentTime = float32(clientTime)
			}
		}

		switch td.options.FederatedScenario {
		case CROSSDEVICE:
			currentTime += CROSSDEVICEBROADCASTDELAY

			for range clients {
				packet := queues.Packet{
					ArrivalTime: currentTime,
					Size:        uint32(messageSize),
					Id:          packetCounter,
				}

				event := queues.Event{
					Time:        packet.ArrivalTime,
					RoundNumber: uint16(round),
					ClientID:    uint16(0), // 0 == ServerID
					Packet:      &packet,
					Type:        queues.ARRIVAL,
				}

				heap.Push(&workload, &event)

				packetCounter++
			}
		case CROSSSILO:
			// Assuming that all messages have the same size
			// And crosssilo have a broadcast protocol implemented
			currentTime += CROSSSILOBROADCASTDELAY

			packet := queues.Packet{
				ArrivalTime: currentTime,
				Size:        uint32(messageSize),
				Id:          packetCounter,
			}

			event := queues.Event{
				Time:        packet.ArrivalTime,
				RoundNumber: uint16(round),
				ClientID:    uint16(0), // 0 == ServerID
				Packet:      &packet,
				Type:        queues.ARRIVAL,
			}

			heap.Push(&workload, &event)

			packetCounter++
		}
	}

	td.resultsWritter = writer.New(uint32(workload.Len()), "metrics_network_"+leafExperimentMeta)

	queueOpt := queues.GlobalOptions{
		MaxQueue:  uint16(math.Floor((float64(workload.Len()) * 0.10))),
		MTU:       td.options.MTU,
		Bandwidth: td.options.Bandwidth,
	}

	td.queue = queues.New(&queueOpt, &workload, td.resultsWritter)
}
