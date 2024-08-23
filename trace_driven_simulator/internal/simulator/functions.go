package simulator

import (
	"container/heap"
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Marco-Guerra/Federated-Learning-Network-Workload/trace_driven_simulator/internal/simulator/queues"
	"github.com/Marco-Guerra/Federated-Learning-Network-Workload/trace_driven_simulator/packages/writer"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

func New(options *GlobalOptions) *TraceDriven {
	return &TraceDriven{
		options: options,
		queues:  nil,
	}
}

func (td *TraceDriven) RunSimulation(trace_filename string) {
	td.readTrace(trace_filename)

	go td.resultsWritter.Start()

	qwg := sync.WaitGroup{}

	qwg.Add(len(td.queues))

	for i := range td.queues {
		go func(qid int) {
			qout := td.queues[qid].Start()

			meanDelay, throughput := td.calculeMetrics(qout)

			resultString := fmt.Sprintf("\nClient %d Metrics\nMean Delay: %f seconds\nThroughput: %f bits per second\n",
				qid+1,
				meanDelay,
				throughput,
			)

			log.Print(resultString)

			qwg.Done()
		}(i)
	}

	qwg.Wait()

	td.resultsWritter.Close()
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

	// Find the number of clients
	nclients := 0
	lastNClients := 0
	for i, record := range records {
		if i == 0 {
			continue // Skip header
		}

		clientID, _ := strconv.Atoi(record[0])
		if clientID > nclients {
			lastNClients = nclients
			nclients = clientID
		}

		if lastNClients == nclients {
			break
		}
	}

	dqueues := make([]*queues.MM1Queue, nclients)
	workloads := make([]queues.EventHeap, nclients)

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

			heap.Push(&workloads[clientID-1], &event)

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
		case CROSSSILO:
			currentTime += CROSSSILOBROADCASTDELAY
		}
	}

	td.resultsWritter = writer.New(uint32(packetCounter), "metrics_network_"+leafExperimentMeta)
	seed := uint64(time.Now().Unix())
	rng := rand.New(rand.NewSource(seed))

	rngArrival := distuv.Exponential{
		Rate: 1, // entrada do usuário
		Src:  rand.NewSource(seed),
	}

	for i := range workloads {
		var arrivalInterval float64 = 0
		for localtime := 0.0; localtime <= float64(currentTime); localtime += arrivalInterval {
			packet := queues.Packet{
				ArrivalTime: float32(localtime),
				Size:        64 + uint32(rng.Int31n(1518-64+1)), // Distribuição uniforme
				Id:          packetCounter,
			}

			event := queues.Event{
				Time:        packet.ArrivalTime,
				RoundNumber: 1001,
				ClientID:    4096,
				Packet:      &packet,
				Type:        queues.ARRIVAL,
			}

			heap.Push(&workloads[i], &event)

			packetCounter++

			arrivalInterval = rngArrival.Rand()
		}
	}

	for i := range dqueues {
		queueOpt := queues.GlobalOptions{
			MaxQueue:  uint16(math.Floor((float64(workloads[i].Len()) * 0.10))),
			Bandwidth: td.options.MinBandwidth + uint32(rng.Int31n(int32(td.options.MaxBandwidth))-int32(td.options.MinBandwidth)+1), // Achar valores mais reais
		}

		dqueues[i] = queues.New(&queueOpt, &workloads[i], td.resultsWritter)
	}

	td.queues = dqueues
}
