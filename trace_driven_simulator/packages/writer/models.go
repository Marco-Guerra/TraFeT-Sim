package writer

import "encoding/csv"

type WriterRegister struct {
	ClientID    uint16
	RoundNumber uint16
	Size        uint16
	Time        float32
	Delay       float32
}

type Writer struct {
	maxBufferSize      uint32
	aggregationChannel chan *WriterRegister
	aggregationBuffer  []*WriterRegister
	filename           string
	csvWriter          *csv.Writer
}
