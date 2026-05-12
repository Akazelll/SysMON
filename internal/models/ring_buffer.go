package models

type RingBuffer struct {
	Data []float64
	size int
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		Data: make([]float64, 0, size),
		size: size,
	}
}

func (r *RingBuffer) Add(value float64) {
	if len(r.Data) >= r.size {
		r.Data = r.Data[1:]
	}
	r.Data = append(r.Data, value)
}