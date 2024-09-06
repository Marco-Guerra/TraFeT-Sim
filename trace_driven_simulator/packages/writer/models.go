package writer

import "encoding/csv"

type WriterRegister struct {
	Network       uint8
	ClientID      uint16
	RoundNumber   uint16
	Size          uint32
	ArrivalTime   float32
	DepartureTime float32
}

type Writer struct {
	maxBufferSize      uint32
	aggregationChannel chan *WriterRegister
	aggregationBuffer  []*WriterRegister
	filename           string
	csvWriter          *csv.Writer
}
